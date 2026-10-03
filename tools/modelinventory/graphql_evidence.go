package main

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	graphqlparser "github.com/vektah/gqlparser/v2"
	graphqlast "github.com/vektah/gqlparser/v2/ast"
	graphqlsyntax "github.com/vektah/gqlparser/v2/parser"
)

const (
	graphqlModelsSourcePath     = "pkg/dependencymodels/graphql_models.gen.go"
	graphqlOperationsSourcePath = "pkg/dependencies/graphql/generated.go"
	graphqlSchemaSourcePath     = "pkg/schemas/graphql/endpoints-schema.graphql"
	graphqlKindStruct           = "struct"
	graphqlKindInterface        = "interface"
	graphqlTypeNameField        = "__typename"
)

// graphqlSelectionEvidence ties each generated Go model to the checked-in
// GraphQL document and the concrete response selection or variable input path
// that causes genqlient to generate it.
type graphqlSelectionEvidence struct {
	Operation     string
	Document      string
	SelectionPath string
	Kind          string
}

type graphqlGoField struct {
	Name   string
	GoName string
	Type   goast.Expr
}

type graphqlGoModel struct {
	Kind         string
	Fields       []graphqlGoField
	Implementors []string
	Comment      string
	CustomJSON   bool
}

type graphqlEvidenceError string

func (evidenceError graphqlEvidenceError) Error() string {
	return string(evidenceError)
}

func graphqlEvidenceErrorf(format string, arguments ...any) error {
	return graphqlEvidenceError(fmt.Sprintf(format, arguments...))
}

type graphqlDocument struct {
	Path          string
	Operation     *graphqlast.OperationDefinition
	WireOperation *graphqlast.OperationDefinition
}

type graphqlInterfaceDecoderCase struct {
	Model         string
	JSONUnmarshal bool
}

type graphqlInterfaceDecoder struct {
	Cases              map[string]graphqlInterfaceDecoderCase
	ReadsTypeName      bool
	RejectsMissingType bool
	RejectsUnknownType bool
}

type graphqlInterfaceDecoderCases map[string]graphqlInterfaceDecoder

type graphqlEvidenceContext struct {
	goModels     map[string]graphqlGoModel
	registered   map[string]bool
	schema       *graphqlast.Schema
	visitedTypes map[string]bool
	traversed    map[string]bool
	traces       map[string][]graphqlSelectionEvidence
	decoderCases graphqlInterfaceDecoderCases
}

// graphqlModelEvidence resolves every model generated into dependencymodels
// from an operation response/input root. It fails closed if a model has no
// operation trace or if a traced response selection is absent from its source
// document.
func graphqlModelEvidence(models map[string]generatedModel) (map[string][]graphqlSelectionEvidence, error) {
	registered := make(map[string]bool)

	for name, model := range models {
		if model.File == graphQLGeneratedSet().Output {
			registered[name] = true
		}
	}

	goModels, err := readGraphQLGoModels()
	if err != nil {
		return nil, err
	}

	documents, err := readGraphQLDocuments()
	if err != nil {
		return nil, err
	}

	functions, err := readGraphQLOperationFunctions()
	if err != nil {
		return nil, err
	}

	schema, err := readGraphQLSchema()
	if err != nil {
		return nil, err
	}

	err = validateGraphQLDocumentsAgainstGeneratedOperations(documents, schema)
	if err != nil {
		return nil, err
	}

	decoderCases, err := readGraphQLInterfaceDecoderCases()
	if err != nil {
		return nil, err
	}

	err = validateGraphQLInterfaceDecoderCases(goModels, schema, decoderCases)
	if err != nil {
		return nil, err
	}

	evidence := graphqlEvidenceContext{
		goModels: goModels, registered: registered, schema: schema,
		visitedTypes: make(map[string]bool), traversed: make(map[string]bool),
		traces: make(map[string][]graphqlSelectionEvidence), decoderCases: decoderCases,
	}

	for operationName := range functions {
		document, found := documents[operationName]
		if !found {
			return nil, graphqlEvidenceErrorf("GraphQL generated operation %s has no checked-in source document", operationName)
		}

		err = evidence.traceOperation(operationName, functions[operationName], document)
		if err != nil {
			return nil, err
		}
	}

	// Genqlient emits helper structs solely for custom JSON marshaling. Each is
	// paired with the concrete model whose custom method constructs it.
	for name := range registered {
		if !strings.HasPrefix(name, "GraphQLpremarshal") {
			continue
		}

		base := strings.TrimPrefix(name, "GraphQLpremarshal")
		for _, trace := range evidence.traces[base] {
			trace.SelectionPath += " (custom JSON marshaling companion)"
			addGraphQLTrace(evidence.traces, name, trace)

			evidence.visitedTypes[name] = true
		}
	}

	var missing []string

	for name := range registered {
		if !evidence.visitedTypes[name] {
			missing = append(missing, name)
		}
	}

	if len(missing) != 0 {
		sort.Strings(missing)

		return evidence.traces, graphqlEvidenceErrorf("generated GraphQL models lack operation/selection evidence: %s", strings.Join(missing, ", "))
	}

	for name := range evidence.traces {
		evidence.traces[name] = uniqueGraphQLTraces(evidence.traces[name])
	}

	return evidence.traces, nil
}

func validateGraphQLDocumentsAgainstGeneratedOperations(
	documents map[string]graphqlDocument,
	schema *graphqlast.Schema,
) error {
	for operationName, document := range documents {
		validationErr := validateGeneratedGraphQLOperation(document.Operation, document.WireOperation, schema)
		if validationErr != nil {
			return fmt.Errorf("validate GraphQL operation %s: %w", operationName, validationErr)
		}
	}

	return nil
}

func (evidence *graphqlEvidenceContext) traceOperation(
	operationName string,
	function *goast.FuncDecl,
	document graphqlDocument,
) error {
	rootDefinition := operationRootDefinition(evidence.schema, document.Operation)
	if rootDefinition == nil {
		return graphqlEvidenceErrorf("GraphQL operation %s has no schema root type", operationName)
	}

	responseRoots := selectedRootFields(document.WireOperation.SelectionSet)
	if len(responseRoots) != 1 {
		return graphqlEvidenceErrorf("GraphQL operation %s selects %d root fields; expected one", operationName, len(responseRoots))
	}

	rootField := responseRoots[0]

	rootFieldDefinition := schemaField(rootDefinition, rootField.Name)
	if rootFieldDefinition == nil {
		return graphqlEvidenceErrorf("GraphQL operation %s root field %s is absent from schema type %s", operationName, rootField.Name, rootDefinition.Name)
	}

	rootGraphQLType := graphqlNamedType(rootFieldDefinition.Type)

	responseType, found := operationResponseModelName(operationName, rootField.Name, rootGraphQLType, evidence.goModels, evidence.registered)
	if !found {
		return graphqlEvidenceErrorf("GraphQL operation %s root selection %s (%s) has no generated response model", operationName, rootField.Name, rootGraphQLType)
	}

	traceErr := evidence.traceResponse(operationName, function, document, rootDefinition, responseRoots, rootField, rootGraphQLType, responseType)
	if traceErr != nil {
		return traceErr
	}

	return evidence.traceInputs(operationName, document)
}

func (evidence *graphqlEvidenceContext) traceResponse(
	operationName string,
	function *goast.FuncDecl,
	document graphqlDocument,
	rootDefinition *graphqlast.Definition,
	responseRoots []*graphqlast.Field,
	rootField *graphqlast.Field,
	rootGraphQLType, responseType string,
) error {
	responseWrapper := firstResultType(function)
	if responseWrapper != operationName+"Response" || !evidence.registered[responseWrapper] {
		return graphqlEvidenceErrorf(
			"GraphQL operation %s generated client returns %s; expected registered operation response %s",
			operationName, responseWrapper, operationName+"Response",
		)
	}

	traceErr := evidence.traceResponseWrapper(operationName, document, rootDefinition, responseRoots, rootGraphQLType, responseType, responseWrapper)
	if traceErr != nil {
		return traceErr
	}

	responseModel := evidence.goModels[responseType]
	responsePath := graphQLResponseName(rootField)
	addGraphQLTrace(evidence.traces, responseType, graphqlSelectionEvidence{
		Operation: operationName, Document: document.Path,
		SelectionPath: "$response." + responsePath, Kind: "response",
	})

	evidence.visitedTypes[responseType] = true

	return evidence.traceResponseFields(
		operationName, document, responseModel, responsePath, rootGraphQLType, responseType,
	)
}

func (evidence *graphqlEvidenceContext) traceResponseWrapper(
	operationName string,
	document graphqlDocument,
	rootDefinition *graphqlast.Definition,
	responseRoots []*graphqlast.Field,
	rootGraphQLType, responseType, responseWrapper string,
) error {
	wrapperModel := evidence.goModels[responseWrapper]
	if wrapperModel.Kind != graphqlKindStruct || len(wrapperModel.Fields) != len(responseRoots) {
		return graphqlEvidenceErrorf(
			"GraphQL operation %s response wrapper %s has %d fields for %d root selections",
			operationName, responseWrapper, len(wrapperModel.Fields), len(responseRoots),
		)
	}

	evidence.visitedTypes[responseWrapper] = true

	for _, wrapperField := range wrapperModel.Fields {
		wrapperRoot := selectedRootFieldByResponseName(responseRoots, wrapperField.Name)
		if wrapperRoot == nil {
			return graphqlEvidenceErrorf("GraphQL response wrapper %s field %s has no matching selected root field", responseWrapper, wrapperField.Name)
		}

		rootType := graphqlNamedType(schemaField(rootDefinition, wrapperRoot.Name).Type)

		wrapperRefs := graphqlGoModelReferences(wrapperField.Type, evidence.registered)
		if rootType == rootGraphQLType && !containsString(wrapperRefs, responseType) {
			return graphqlEvidenceErrorf(
				"GraphQL response wrapper %s field %s does not reference generated response model %s",
				responseWrapper, wrapperField.Name, responseType,
			)
		}

		addGraphQLTrace(evidence.traces, responseWrapper, graphqlSelectionEvidence{
			Operation: operationName, Document: document.Path,
			SelectionPath: "$response." + wrapperField.Name, Kind: "response",
		})
	}

	if graphqlTypeFromComment(evidence.goModels[responseType].Comment) != rootGraphQLType {
		return graphqlEvidenceErrorf(
			"GraphQL operation %s generated response model %s does not document GraphQL type %s",
			operationName, responseType, rootGraphQLType,
		)
	}

	return nil
}

func (evidence *graphqlEvidenceContext) traceResponseFields(
	operationName string,
	document graphqlDocument,
	responseModel graphqlGoModel,
	responsePath, rootGraphQLType, responseType string,
) error {
	for _, field := range responseModel.Fields {
		fieldDefinition := schemaField(evidence.schema.Types[rootGraphQLType], field.Name)
		if field.Name == graphqlTypeNameField && fieldDefinition == nil {
			continue
		}

		if fieldDefinition == nil {
			return graphqlEvidenceErrorf("GraphQL response model %s field %s is absent from schema type %s", responseType, field.Name, rootGraphQLType)
		}

		fieldPath := []string{responsePath, graphqlPathPart(field.Name, fieldDefinition.Type)}
		if findSelectedGraphQLField(document.WireOperation.SelectionSet, fieldPath) == nil {
			return graphqlEvidenceErrorf("GraphQL response selection %s is absent from operation %s", strings.Join(fieldPath, "."), operationName)
		}

		traceErr := traceGraphQLGoType(
			field.Type, graphqlNamedType(fieldDefinition.Type), fieldPath, "response", operationName,
			document, evidence.schema, evidence.goModels, evidence.registered,
			evidence.visitedTypes, evidence.traversed, evidence.traces, evidence.decoderCases,
		)
		if traceErr != nil {
			return traceErr
		}
	}

	return nil
}

func (evidence *graphqlEvidenceContext) traceInputs(operationName string, document graphqlDocument) error {
	inputType := "GraphQL" + operationName + "Input"
	if !evidence.registered[inputType] {
		return graphqlEvidenceErrorf("GraphQL operation %s has no generated variable model %s", operationName, inputType)
	}

	inputModel := evidence.goModels[inputType]
	if inputModel.Kind != graphqlKindStruct {
		return graphqlEvidenceErrorf("GraphQL operation %s variable model %s is not a struct", operationName, inputType)
	}

	for _, field := range inputModel.Fields {
		traceErr := evidence.traceInputField(operationName, document, inputType, field)
		if traceErr != nil {
			return traceErr
		}
	}

	return nil
}

func (evidence *graphqlEvidenceContext) traceInputField(
	operationName string,
	document graphqlDocument,
	inputType string,
	field graphqlGoField,
) error {
	variable := operationVariable(document.Operation, field.Name)
	if variable == nil {
		return graphqlEvidenceErrorf("GraphQL variable model %s field %s has no operation variable", inputType, field.Name)
	}

	variablePath := "$" + field.Name
	addGraphQLTrace(evidence.traces, inputType, graphqlSelectionEvidence{
		Operation: operationName, Document: document.Path,
		SelectionPath: variablePath, Kind: "input",
	})

	evidence.visitedTypes[inputType] = true

	for _, childType := range graphqlGoModelReferences(field.Type, evidence.registered) {
		traceErr := traceGraphQLInputType(
			childType, graphqlNamedType(variable.Type), []string{variablePath}, operationName,
			document, evidence.schema, evidence.goModels, evidence.registered,
			evidence.visitedTypes, evidence.traversed, evidence.traces,
		)
		if traceErr != nil {
			return traceErr
		}
	}

	return nil
}

func readGraphQLGoModels() (map[string]graphqlGoModel, error) {
	path, err := graphqlProjectPath(graphqlModelsSourcePath)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse generated GraphQL models: %w", err)
	}

	models, err := parseGraphQLGoModelTypes(file)
	if err != nil {
		return nil, err
	}

	markGraphQLCustomJSONMethods(file, models)
	resolveGraphQLCustomJSONFields(models)

	return models, nil
}

func parseGraphQLGoModelTypes(file *goast.File) (map[string]graphqlGoModel, error) {
	models := make(map[string]graphqlGoModel)

	for _, declaration := range file.Decls {
		group, isTypeGroup := declaration.(*goast.GenDecl)
		if !isTypeGroup || group.Tok != token.TYPE {
			continue
		}

		for _, specification := range group.Specs {
			typeSpec, isTypeSpec := specification.(*goast.TypeSpec)
			if !isTypeSpec {
				continue
			}

			model, err := parseGraphQLGoModel(typeSpec, group)
			if err != nil {
				return nil, err
			}

			models[typeSpec.Name.Name] = model
		}
	}

	return models, nil
}

func parseGraphQLGoModel(typeSpec *goast.TypeSpec, group *goast.GenDecl) (graphqlGoModel, error) {
	var model graphqlGoModel
	if typeSpec.Doc != nil {
		model.Comment = typeSpec.Doc.Text()
	} else if group.Doc != nil {
		model.Comment = group.Doc.Text()
	}

	switch expression := typeSpec.Type.(type) {
	case *goast.StructType:
		model.Kind = graphqlKindStruct

		fields, err := parseGraphQLGoFields(typeSpec.Name.Name, expression.Fields.List)
		if err != nil {
			return graphqlGoModel{}, err
		}

		model.Fields = fields
	case *goast.InterfaceType:
		model.Kind = graphqlKindInterface
		model.Implementors = parseGraphQLImplementors(model.Comment)
	default:
		model.Kind = "scalar"
	}

	return model, nil
}

func parseGraphQLGoFields(modelName string, fields []*goast.Field) ([]graphqlGoField, error) {
	var result []graphqlGoField

	for _, field := range fields {
		if len(field.Names) == 0 {
			return nil, graphqlEvidenceErrorf("anonymous field in generated GraphQL model %s needs explicit inventory handling", modelName)
		}

		name := graphqlJSONFieldName(field)
		for range field.Names {
			result = append(result, graphqlGoField{Name: name, GoName: field.Names[0].Name, Type: field.Type})
		}
	}

	return result, nil
}

func markGraphQLCustomJSONMethods(file *goast.File, models map[string]graphqlGoModel) {
	for _, declaration := range file.Decls {
		function, isFunction := declaration.(*goast.FuncDecl)
		if !isFunction || function.Recv == nil || len(function.Recv.List) == 0 {
			continue
		}

		if function.Name.Name != "MarshalJSON" && function.Name.Name != "UnmarshalJSON" {
			continue
		}

		name := graphqlGoNamedType(function.Recv.List[0].Type)

		model, found := models[name]
		if found {
			model.CustomJSON = true
			models[name] = model
		}
	}
}

func resolveGraphQLCustomJSONFields(models map[string]graphqlGoModel) {
	for name, model := range models {
		for index := range model.Fields {
			field := &model.Fields[index]
			if field.Name != "-" || !model.CustomJSON {
				continue
			}

			wireName, found := customGraphQLWireFieldName(name, field.GoName, models)
			if found {
				field.Name = wireName
			}
		}

		models[name] = model
	}
}

func customGraphQLWireFieldName(modelName, goName string, models map[string]graphqlGoModel) (string, bool) {
	companion, found := models["GraphQLpremarshal"+modelName]
	if !found || companion.Kind != graphqlKindStruct {
		return "", false
	}

	expected := lowerInitial(goName)
	for _, field := range companion.Fields {
		if field.GoName == goName && field.Name == expected {
			return expected, true
		}
	}

	return "", false
}

func graphqlJSONFieldName(field *goast.Field) string {
	if field.Tag == nil || len(field.Names) == 0 {
		if len(field.Names) == 0 {
			return "-"
		}

		return lowerInitial(field.Names[0].Name)
	}

	tag, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return lowerInitial(field.Names[0].Name)
	}

	name := reflect.StructTag(tag).Get("json")
	if name == "" {
		return lowerInitial(field.Names[0].Name)
	}

	name, _, _ = strings.Cut(name, ",")

	return name
}

func parseGraphQLImplementors(comment string) []string {
	const marker = "is implemented by the following types:"

	index := strings.Index(comment, marker)
	if index < 0 {
		return nil
	}

	lines := strings.Split(comment[index+len(marker):], "\n")

	var implementors []string

	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name != "" {
			implementors = append(implementors, name)
		}
	}

	return implementors
}

func readGraphQLDocuments() (map[string]graphqlDocument, error) {
	root, err := graphqlProjectRoot()
	if err != nil {
		return nil, err
	}

	paths, err := filepath.Glob(filepath.Join(root, "pkg", "schemas", "graphql", "*.graphql"))
	if err != nil {
		return nil, fmt.Errorf("find GraphQL operation documents: %w", err)
	}

	sort.Strings(paths)

	documents := make(map[string]graphqlDocument)

	for _, path := range paths {
		relativePath, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil, fmt.Errorf("resolve GraphQL document path %s: %w", path, relErr)
		}

		relativePath = filepath.ToSlash(relativePath)
		if relativePath == graphqlSchemaSourcePath {
			continue
		}

		//nolint:gosec // The path comes from a fixed pattern rooted in the repository.
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("read GraphQL operation document %s: %w", path, readErr)
		}

		query, parseErr := graphqlsyntax.ParseQuery(&graphqlast.Source{Name: relativePath, Input: string(data), BuiltIn: false})
		if parseErr != nil {
			return nil, fmt.Errorf("parse GraphQL operation document %s: %w", relativePath, parseErr)
		}

		for _, operation := range query.Operations {
			if _, exists := documents[operation.Name]; exists {
				return nil, graphqlEvidenceErrorf("GraphQL operation %s is defined by more than one document", operation.Name)
			}

			documents[operation.Name] = graphqlDocument{Path: relativePath, Operation: operation, WireOperation: nil}
		}
	}

	wireOperations, err := readGraphQLWireOperations()
	if err != nil {
		return nil, err
	}

	for name, document := range documents {
		wireOperation, found := wireOperations[name]
		if !found {
			return nil, graphqlEvidenceErrorf("GraphQL source operation %s has no generated wire operation", name)
		}

		document.WireOperation = wireOperation
		documents[name] = document
	}

	return documents, nil
}

func readGraphQLWireOperations() (map[string]*graphqlast.OperationDefinition, error) {
	path, err := graphqlProjectPath(graphqlOperationsSourcePath)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse generated GraphQL wire operations: %w", err)
	}

	operations := make(map[string]*graphqlast.OperationDefinition)

	for _, declaration := range file.Decls {
		group, ok := declaration.(*goast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}

		for _, specification := range group.Specs {
			valueSpec, isValueSpec := specification.(*goast.ValueSpec)
			if !isValueSpec || len(valueSpec.Names) == 0 || len(valueSpec.Values) != 1 {
				continue
			}

			name := valueSpec.Names[0].Name

			const suffix = "_Operation"

			if !strings.HasSuffix(name, suffix) {
				continue
			}

			literal, ok := valueSpec.Values[0].(*goast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return nil, graphqlEvidenceErrorf("generated GraphQL operation constant %s is not a string literal", name)
			}

			queryText, unquoteErr := strconv.Unquote(literal.Value)
			if unquoteErr != nil {
				return nil, fmt.Errorf("unquote generated GraphQL operation %s: %w", name, unquoteErr)
			}

			operationName := strings.TrimSuffix(name, suffix)

			query, parseErr := graphqlsyntax.ParseQuery(&graphqlast.Source{Name: name, Input: queryText, BuiltIn: false})
			if parseErr != nil {
				return nil, fmt.Errorf("parse generated GraphQL wire operation %s: %w", name, parseErr)
			}

			if len(query.Operations) != 1 || query.Operations[0].Name != operationName {
				return nil, graphqlEvidenceErrorf("generated GraphQL wire constant %s does not define exactly operation %s", name, operationName)
			}

			operations[operationName] = query.Operations[0]
		}
	}

	return operations, nil
}

func readGraphQLOperationFunctions() (map[string]*goast.FuncDecl, error) {
	path, err := graphqlProjectPath(graphqlOperationsSourcePath)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse generated GraphQL operation functions: %w", err)
	}

	functions := make(map[string]*goast.FuncDecl)

	for _, declaration := range file.Decls {
		function, ok := declaration.(*goast.FuncDecl)
		if ok && function.Recv == nil && function.Type.Params != nil && function.Type.Results != nil {
			functions[function.Name.Name] = function
		}
	}

	return functions, nil
}

func readGraphQLInterfaceDecoderCases() (graphqlInterfaceDecoderCases, error) {
	path, err := graphqlProjectPath(graphqlModelsSourcePath)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse generated GraphQL interface decoders: %w", err)
	}

	cases := make(graphqlInterfaceDecoderCases)

	for _, declaration := range file.Decls {
		function, isFunction := declaration.(*goast.FuncDecl)
		if !isFunction || function.Body == nil || !strings.HasPrefix(function.Name.Name, "__unmarshal") {
			continue
		}

		interfaceName := strings.TrimPrefix(function.Name.Name, "__unmarshal")
		decoder := graphqlInterfaceDecoder{
			Cases:              make(map[string]graphqlInterfaceDecoderCase),
			ReadsTypeName:      graphqlDecoderReadsTypeName(function),
			RejectsMissingType: false,
			RejectsUnknownType: false,
		}

		for _, statement := range function.Body.List {
			switchStatement, isSwitch := statement.(*goast.SwitchStmt)
			if !isSwitch || !graphqlSwitchesOnTypeName(switchStatement) {
				continue
			}

			readErr := readGraphQLDecoderSwitch(interfaceName, function.Name.Name, switchStatement, &decoder)
			if readErr != nil {
				return nil, readErr
			}
		}

		if len(decoder.Cases) != 0 || decoder.ReadsTypeName {
			cases[interfaceName] = decoder
		}
	}

	return cases, nil
}

func graphqlSwitchesOnTypeName(statement *goast.SwitchStmt) bool {
	selector, isSelector := statement.Tag.(*goast.SelectorExpr)
	if !isSelector || selector.Sel.Name != "TypeName" {
		return false
	}

	variable, isVariable := selector.X.(*goast.Ident)

	return isVariable && variable.Name == "tn"
}

func readGraphQLDecoderSwitch(
	interfaceName, functionName string,
	statement *goast.SwitchStmt,
	decoder *graphqlInterfaceDecoder,
) error {
	for _, caseStatement := range statement.Body.List {
		clause, isClause := caseStatement.(*goast.CaseClause)
		if !isClause {
			continue
		}

		for _, expression := range clause.List {
			literal, isString := expression.(*goast.BasicLit)
			if !isString || literal.Kind != token.STRING {
				continue
			}

			caseErr := registerGraphQLDecoderCase(interfaceName, functionName, clause, literal, decoder)
			if caseErr != nil {
				return caseErr
			}
		}

		if clause.List == nil {
			decoder.RejectsUnknownType = graphqlDecoderReturnsError(clause.Body)
		}
	}

	return nil
}

func registerGraphQLDecoderCase(
	interfaceName, functionName string,
	clause *goast.CaseClause,
	literal *goast.BasicLit,
	decoder *graphqlInterfaceDecoder,
) error {
	concreteType, err := strconv.Unquote(literal.Value)
	if err != nil {
		return fmt.Errorf("unquote generated GraphQL decoder case in %s: %w", functionName, err)
	}

	if concreteType == "" {
		decoder.RejectsMissingType = graphqlDecoderReturnsError(clause.Body)

		return nil
	}

	modelName := graphqlDecoderInstantiatedType(clause.Body)
	if modelName == "" {
		return graphqlEvidenceErrorf("GraphQL decoder %s case %s does not instantiate a generated model", interfaceName, concreteType)
	}

	if !graphqlDecoderUnmarshalsIntoValue(clause.Body) {
		return graphqlEvidenceErrorf("GraphQL decoder %s case %s does not decode JSON into the selected generated model", interfaceName, concreteType)
	}

	if previous, exists := decoder.Cases[concreteType]; exists {
		return graphqlEvidenceErrorf("GraphQL decoder %s maps %s to both %s and %s", interfaceName, concreteType, previous.Model, modelName)
	}

	decoder.Cases[concreteType] = graphqlInterfaceDecoderCase{Model: modelName, JSONUnmarshal: true}

	return nil
}

func graphqlDecoderReadsTypeName(function *goast.FuncDecl) bool {
	var hasTaggedTypeName bool

	goast.Inspect(function.Body, func(node goast.Node) bool {
		valueSpec, isValueSpec := node.(*goast.ValueSpec)
		if !isValueSpec || len(valueSpec.Names) != 1 || valueSpec.Names[0].Name != "tn" {
			return true
		}

		structure, ok := valueSpec.Type.(*goast.StructType)
		if !ok {
			return true
		}

		for _, field := range structure.Fields.List {
			if len(field.Names) != 1 || field.Names[0].Name != "TypeName" || field.Tag == nil {
				continue
			}

			tag, err := strconv.Unquote(field.Tag.Value)
			if err == nil && reflect.StructTag(tag).Get("json") == graphqlTypeNameField {
				hasTaggedTypeName = true
			}
		}

		return true
	})

	if !hasTaggedTypeName {
		return false
	}

	var readsTypeName bool

	goast.Inspect(function.Body, func(node goast.Node) bool {
		call, ok := node.(*goast.CallExpr)
		if ok && isJSONUnmarshalCall(call, "b", "tn", true) {
			readsTypeName = true
		}

		return true
	})

	return readsTypeName
}

func graphqlDecoderUnmarshalsIntoValue(statements []goast.Stmt) bool {
	for _, statement := range statements {
		returnStatement, isReturn := statement.(*goast.ReturnStmt)
		if !isReturn || len(returnStatement.Results) == 0 {
			continue
		}

		call, ok := returnStatement.Results[len(returnStatement.Results)-1].(*goast.CallExpr)
		if ok && isJSONUnmarshalIntoInterface(call) {
			return true
		}
	}

	return false
}

func isJSONUnmarshalIntoInterface(call *goast.CallExpr) bool {
	if !isJSONUnmarshalSelector(call) || len(call.Args) != 2 {
		return false
	}

	input, inputOK := call.Args[0].(*goast.Ident)

	value, valueOK := call.Args[1].(*goast.StarExpr)
	if !valueOK {
		return false
	}

	interfaceValue, interfaceOK := value.X.(*goast.Ident)

	return inputOK && input.Name == "b" && interfaceOK && interfaceValue.Name == "v"
}

func isJSONUnmarshalCall(call *goast.CallExpr, inputName, outputName string, addressOutput bool) bool {
	if !isJSONUnmarshalSelector(call) || len(call.Args) != 2 {
		return false
	}

	input, inputOK := call.Args[0].(*goast.Ident)
	if !inputOK || input.Name != inputName {
		return false
	}

	if addressOutput {
		address, isAddress := call.Args[1].(*goast.UnaryExpr)
		if !isAddress || address.Op != token.AND {
			return false
		}

		output, ok := address.X.(*goast.Ident)

		return ok && output.Name == outputName
	}

	output, ok := call.Args[1].(*goast.Ident)

	return ok && output.Name == outputName
}

func isJSONUnmarshalSelector(call *goast.CallExpr) bool {
	selector, isSelector := call.Fun.(*goast.SelectorExpr)
	if !isSelector || selector.Sel.Name != "Unmarshal" {
		return false
	}

	packageName, ok := selector.X.(*goast.Ident)

	return ok && packageName.Name == "json"
}

func graphqlDecoderReturnsError(statements []goast.Stmt) bool {
	var found bool

	for _, statement := range statements {
		goast.Inspect(statement, func(node goast.Node) bool {
			returnStatement, isReturn := node.(*goast.ReturnStmt)
			if !isReturn || len(returnStatement.Results) == 0 {
				return true
			}

			call, isCall := returnStatement.Results[len(returnStatement.Results)-1].(*goast.CallExpr)
			if !isCall {
				return true
			}

			selector, isSelector := call.Fun.(*goast.SelectorExpr)
			if !isSelector || selector.Sel.Name != "Errorf" {
				return true
			}

			packageName, ok := selector.X.(*goast.Ident)
			if ok && packageName.Name == "fmt" {
				found = true
			}

			return true
		})
	}

	return found
}

func validateGraphQLInterfaceDecoderCases(
	models map[string]graphqlGoModel,
	schema *graphqlast.Schema,
	decoders graphqlInterfaceDecoderCases,
) error {
	for name, model := range models {
		if model.Kind == graphqlKindInterface {
			validationErr := validateGraphQLInterfaceDecoder(name, model, models, schema, decoders)
			if validationErr != nil {
				return validationErr
			}
		}
	}

	for name := range decoders {
		model, found := models[name]
		if found && model.Kind == graphqlKindInterface {
			continue
		}

		if !found || model.Kind != graphqlKindInterface {
			return graphqlEvidenceErrorf("generated GraphQL decoder %s has no matching generated interface", name)
		}
	}

	return nil
}

func validateGraphQLInterfaceDecoder(
	name string,
	model graphqlGoModel,
	models map[string]graphqlGoModel,
	schema *graphqlast.Schema,
	decoders graphqlInterfaceDecoderCases,
) error {
	decoder, found := decoders[name]
	if !found {
		return graphqlEvidenceErrorf("generated GraphQL interface %s has no custom JSON decoder", name)
	}

	if !decoder.ReadsTypeName || !decoder.RejectsMissingType || !decoder.RejectsUnknownType {
		return graphqlEvidenceErrorf("generated GraphQL decoder %s must read __typename and reject missing or unknown values", name)
	}

	interfaceType := graphqlTypeFromComment(model.Comment)
	if interfaceType == "" || !graphqlSchemaAbstractType(schema, interfaceType) {
		return graphqlEvidenceErrorf("generated GraphQL interface %s has no matching SDL interface or union type", name)
	}

	expected, err := expectedGraphQLDecoderCases(name, interfaceType, model.Implementors, models, schema)
	if err != nil {
		return err
	}

	return validateGraphQLDecoderMappings(name, expected, decoder.Cases)
}

func expectedGraphQLDecoderCases(
	interfaceName, interfaceType string,
	implementors []string,
	models map[string]graphqlGoModel,
	schema *graphqlast.Schema,
) (map[string]string, error) {
	expected := make(map[string]string, len(implementors))

	for index := range implementors {
		if graphqlTypeFromComment(models[implementors[index]].Comment) == "" {
			return nil, graphqlEvidenceErrorf("generated GraphQL interface %s lists %s without a GraphQL type comment", interfaceName, implementors[index])
		}

		concreteType := graphqlTypeFromComment(models[implementors[index]].Comment)
		if _, duplicate := expected[concreteType]; duplicate {
			return nil, graphqlEvidenceErrorf("generated GraphQL interface %s repeats concrete GraphQL type %s", interfaceName, concreteType)
		}

		if !graphqlSchemaPossibleType(schema, interfaceType, concreteType) {
			return nil, graphqlEvidenceErrorf(
				"generated GraphQL interface %s lists %s, but SDL type %s is not a possible type of %s",
				interfaceName, implementors[index], concreteType, interfaceType,
			)
		}

		expected[concreteType] = implementors[index]
	}

	return expected, nil
}

func validateGraphQLDecoderMappings(
	name string,
	expected map[string]string,
	actual map[string]graphqlInterfaceDecoderCase,
) error {
	if len(actual) != len(expected) {
		return graphqlEvidenceErrorf("generated GraphQL decoder %s has %d concrete cases for %d generated implementors", name, len(actual), len(expected))
	}

	for concreteType, implementor := range expected {
		decoderCase, exists := actual[concreteType]
		if !exists || decoderCase.Model != implementor || !decoderCase.JSONUnmarshal {
			return graphqlEvidenceErrorf("generated GraphQL decoder %s does not map SDL type %s to generated model %s and decode it", name, concreteType, implementor)
		}
	}

	for concreteType, decoderCase := range actual {
		if expected[concreteType] != decoderCase.Model {
			return graphqlEvidenceErrorf("generated GraphQL decoder %s has unexpected case %s => %s", name, concreteType, decoderCase.Model)
		}
	}

	return nil
}

func graphqlDecoderInstantiatedType(statements []goast.Stmt) string {
	for _, statement := range statements {
		assignment, isAssignment := statement.(*goast.AssignStmt)
		if !isAssignment || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
			continue
		}

		left, isPointer := assignment.Lhs[0].(*goast.StarExpr)
		if !isPointer {
			continue
		}

		value, isIdentifier := left.X.(*goast.Ident)
		if !isIdentifier || value.Name != "v" {
			continue
		}

		call, isCall := assignment.Rhs[0].(*goast.CallExpr)
		if !isCall || len(call.Args) != 1 {
			continue
		}

		function, isFunction := call.Fun.(*goast.Ident)
		if !isFunction || function.Name != "new" {
			continue
		}

		argument, isIdentifier := call.Args[0].(*goast.Ident)
		if isIdentifier {
			return argument.Name
		}
	}

	return ""
}

func readGraphQLSchema() (*graphqlast.Schema, error) {
	path, err := graphqlProjectPath(graphqlSchemaSourcePath)
	if err != nil {
		return nil, err
	}

	//nolint:gosec // The path is a fixed schema path rooted in the repository.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read GraphQL schema %s: %w", graphqlSchemaSourcePath, err)
	}

	schema, err := graphqlparser.LoadSchema(&graphqlast.Source{Name: graphqlSchemaSourcePath, Input: string(data), BuiltIn: false})
	if err != nil {
		return nil, fmt.Errorf("parse GraphQL schema %s: %w", graphqlSchemaSourcePath, err)
	}

	return schema, nil
}

func validateGeneratedGraphQLOperation(
	source, generated *graphqlast.OperationDefinition,
	schema *graphqlast.Schema,
) error {
	if source == nil || generated == nil {
		return graphqlEvidenceErrorf("source and generated operation are required")
	}

	if source.Name != generated.Name || source.Operation != generated.Operation {
		return graphqlEvidenceErrorf(
			"generated operation identity %s %s differs from source %s %s",
			generated.Operation, generated.Name, source.Operation, source.Name,
		)
	}

	if graphqlVariableDefinitionsKey(source.VariableDefinitions) != graphqlVariableDefinitionsKey(generated.VariableDefinitions) {
		return graphqlEvidenceErrorf("generated variable definitions differ from the checked-in operation")
	}

	if graphqlDirectivesKey(source.Directives) != graphqlDirectivesKey(generated.Directives) {
		return graphqlEvidenceErrorf("generated operation directives differ from the checked-in operation")
	}

	root := operationRootDefinition(schema, source)
	if root == nil {
		return graphqlEvidenceErrorf("operation has no matching SDL root type")
	}

	return validateGraphQLSelectionSets(source.SelectionSet, generated.SelectionSet, root.Name, schema)
}

func validateGraphQLSelectionSets(
	source, generated graphqlast.SelectionSet,
	parentType string,
	schema *graphqlast.Schema,
) error {
	matched := make([]bool, len(generated))

	for _, sourceSelection := range source {
		generatedIndex, err := validateGraphQLSourceSelection(sourceSelection, generated, matched, parentType, schema)
		if err != nil {
			return err
		}

		matched[generatedIndex] = true
	}

	return validateGraphQLGeneratedSelections(generated, matched, parentType, schema)
}

func validateGraphQLSourceSelection(
	sourceSelection graphqlast.Selection,
	generated graphqlast.SelectionSet,
	matched []bool,
	parentType string,
	schema *graphqlast.Schema,
) (int, error) {
	switch sourceItem := sourceSelection.(type) {
	case *graphqlast.Field:
		return validateGraphQLSourceField(sourceItem, generated, matched, parentType, schema)
	case *graphqlast.InlineFragment:
		return validateGraphQLSourceInlineFragment(sourceItem, generated, matched, parentType, schema)
	case *graphqlast.FragmentSpread:
		return -1, graphqlEvidenceErrorf("named fragment spread %s needs explicit generated-fragment inventory support", sourceItem.Name)
	default:
		return -1, graphqlEvidenceErrorf("unsupported GraphQL source selection %T", sourceSelection)
	}
}

func validateGraphQLSourceField(
	sourceField *graphqlast.Field,
	generated graphqlast.SelectionSet,
	matched []bool,
	parentType string,
	schema *graphqlast.Schema,
) (int, error) {
	generatedIndex := matchingGraphQLField(generated, matched, sourceField)
	if generatedIndex < 0 {
		return -1, graphqlEvidenceErrorf("generated operation is missing source field %s on %s", graphQLResponseName(sourceField), parentType)
	}

	generatedField, isField := generated[generatedIndex].(*graphqlast.Field)
	if !isField {
		return -1, graphqlEvidenceErrorf("generated selection matching field %s changed type", graphQLResponseName(sourceField))
	}

	if sourceField.Name == graphqlTypeNameField {
		if len(sourceField.SelectionSet) != 0 || len(generatedField.SelectionSet) != 0 {
			return -1, graphqlEvidenceErrorf("__typename cannot have a nested selection")
		}

		return generatedIndex, nil
	}

	fieldDefinition := schemaField(schema.Types[parentType], sourceField.Name)
	if fieldDefinition == nil {
		return -1, graphqlEvidenceErrorf("operation field %s is absent from SDL type %s", sourceField.Name, parentType)
	}

	err := validateGraphQLSelectionSets(sourceField.SelectionSet, generatedField.SelectionSet, graphqlNamedType(fieldDefinition.Type), schema)
	if err != nil {
		return -1, err
	}

	return generatedIndex, nil
}

func validateGraphQLSourceInlineFragment(
	sourceFragment *graphqlast.InlineFragment,
	generated graphqlast.SelectionSet,
	matched []bool,
	parentType string,
	schema *graphqlast.Schema,
) (int, error) {
	generatedIndex := matchingGraphQLInlineFragment(generated, matched, sourceFragment)
	if generatedIndex < 0 {
		return -1, graphqlEvidenceErrorf("generated operation is missing source inline fragment on %s", sourceFragment.TypeCondition)
	}

	if !graphqlTypeConditionPossible(schema, parentType, sourceFragment.TypeCondition) {
		return -1, graphqlEvidenceErrorf("inline fragment type %s is not possible for SDL type %s", sourceFragment.TypeCondition, parentType)
	}

	generatedFragment, isFragment := generated[generatedIndex].(*graphqlast.InlineFragment)
	if !isFragment {
		return -1, graphqlEvidenceErrorf("generated selection matching inline fragment %s changed type", sourceFragment.TypeCondition)
	}

	err := validateGraphQLSelectionSets(sourceFragment.SelectionSet, generatedFragment.SelectionSet, sourceFragment.TypeCondition, schema)
	if err != nil {
		return -1, err
	}

	return generatedIndex, nil
}

func validateGraphQLGeneratedSelections(
	generated graphqlast.SelectionSet,
	matched []bool,
	parentType string,
	schema *graphqlast.Schema,
) error {
	for index, selection := range generated {
		if matched[index] {
			continue
		}

		selectionErr := validateExtraGraphQLSelection(selection, parentType)
		if selectionErr != nil {
			return selectionErr
		}

		if !graphqlSchemaAbstractType(schema, parentType) {
			return graphqlEvidenceErrorf("generated operation adds __typename on non-abstract SDL type %s", parentType)
		}
	}

	if graphqlSchemaAbstractType(schema, parentType) && len(generated) != 0 && !hasDirectGraphQLField(generated, graphqlTypeNameField) {
		return graphqlEvidenceErrorf("generated operation omits __typename discriminator for abstract SDL type %s", parentType)
	}

	return nil
}

func validateExtraGraphQLSelection(selection graphqlast.Selection, parentType string) error {
	field, isField := selection.(*graphqlast.Field)
	if !isField {
		return graphqlEvidenceErrorf("generated operation adds unmodeled selection %T on SDL type %s", selection, parentType)
	}

	if field.Name != graphqlTypeNameField || graphqlFieldAliasKey(field) != "" || len(field.Arguments) != 0 ||
		len(field.Directives) != 0 || len(field.SelectionSet) != 0 {
		return graphqlEvidenceErrorf(
			"generated operation adds unmodeled field %s alias=%q arguments=%d directives=%d children=%d on SDL type %s",
			field.Name, field.Alias, len(field.Arguments), len(field.Directives), len(field.SelectionSet), parentType,
		)
	}

	return nil
}

func matchingGraphQLField(selection graphqlast.SelectionSet, matched []bool, expected *graphqlast.Field) int {
	for index, candidate := range selection {
		field, ok := candidate.(*graphqlast.Field)
		if ok && !matched[index] && field.Name == expected.Name &&
			graphqlFieldAliasKey(field) == graphqlFieldAliasKey(expected) &&
			graphqlArgumentsKey(field.Arguments) == graphqlArgumentsKey(expected.Arguments) &&
			graphqlDirectivesKey(field.Directives) == graphqlDirectivesKey(expected.Directives) {
			return index
		}
	}

	return -1
}

func matchingGraphQLInlineFragment(selection graphqlast.SelectionSet, matched []bool, expected *graphqlast.InlineFragment) int {
	for index, candidate := range selection {
		fragment, ok := candidate.(*graphqlast.InlineFragment)
		if ok && !matched[index] && fragment.TypeCondition == expected.TypeCondition &&
			graphqlDirectivesKey(fragment.Directives) == graphqlDirectivesKey(expected.Directives) {
			return index
		}
	}

	return -1
}

func graphqlTypeConditionPossible(schema *graphqlast.Schema, parentType, condition string) bool {
	if schema.Types[condition] == nil {
		return false
	}

	if parentType == condition {
		return true
	}

	return graphqlSchemaPossibleType(schema, parentType, condition)
}

func graphqlSchemaAbstractType(schema *graphqlast.Schema, name string) bool {
	definition := schema.Types[name]

	return definition != nil && (definition.Kind == graphqlast.Interface || definition.Kind == graphqlast.Union)
}

func hasDirectGraphQLField(selection graphqlast.SelectionSet, name string) bool {
	for _, item := range selection {
		field, ok := item.(*graphqlast.Field)
		if ok && field.Name == name && graphqlFieldAliasKey(field) == "" {
			return true
		}
	}

	return false
}

func graphqlFieldAliasKey(field *graphqlast.Field) string {
	if field.Alias == field.Name {
		return ""
	}

	return field.Alias
}

func graphqlArgumentsKey(arguments graphqlast.ArgumentList) string {
	values := make([]string, 0, len(arguments))

	for _, argument := range arguments {
		value := ""
		if argument.Value != nil {
			value = argument.Value.String()
		}

		values = append(values, argument.Name+":"+value)
	}

	sort.Strings(values)

	return strings.Join(values, ",")
}

func graphqlDirectivesKey(directives graphqlast.DirectiveList) string {
	values := make([]string, 0, len(directives))
	for _, directive := range directives {
		values = append(values, directive.Name+"("+graphqlArgumentsKey(directive.Arguments)+")")
	}

	sort.Strings(values)

	return strings.Join(values, ",")
}

func graphqlVariableDefinitionsKey(definitions graphqlast.VariableDefinitionList) string {
	values := make([]string, 0, len(definitions))

	for _, definition := range definitions {
		defaultValue := ""
		if definition.DefaultValue != nil {
			defaultValue = definition.DefaultValue.String()
		}

		values = append(values, definition.Variable+":"+definition.Type.String()+"="+defaultValue+";"+graphqlDirectivesKey(definition.Directives))
	}

	sort.Strings(values)

	return strings.Join(values, ",")
}

func graphqlProjectPath(relative string) (string, error) {
	root, err := graphqlProjectRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, filepath.FromSlash(relative)), nil
}

func graphqlProjectRoot() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("locate repository root: %w", err)
	}

	for {
		_, statErr := os.Stat(filepath.Join(directory, "go.mod"))
		if statErr == nil {
			return directory, nil
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return "", graphqlEvidenceErrorf("locate repository root from %s: go.mod not found", directory)
		}

		directory = parent
	}
}

func graphqlGoNamedType(expression goast.Expr) string {
	switch value := expression.(type) {
	case *goast.Ident:
		return value.Name
	case *goast.StarExpr:
		return graphqlGoNamedType(value.X)
	case *goast.ArrayType:
		return graphqlGoNamedType(value.Elt)
	case *goast.MapType:
		return graphqlGoNamedType(value.Value)
	default:
		return ""
	}
}

func firstResultType(function *goast.FuncDecl) string {
	if function.Type.Results == nil || len(function.Type.Results.List) == 0 {
		return ""
	}

	return graphqlGoNamedType(function.Type.Results.List[0].Type)
}

func operationResponseModelName(operationName, rootField, schemaType string, models map[string]graphqlGoModel, registered map[string]bool) (string, bool) {
	candidates := []string{
		operationName + schemaType,
		operationName + upperInitial(rootField) + schemaType,
	}
	for _, candidate := range candidates {
		if registered[candidate] && graphqlTypeFromComment(models[candidate].Comment) == schemaType {
			return candidate, true
		}
	}

	return "", false
}

func selectedRootFieldByResponseName(fields []*graphqlast.Field, responseName string) *graphqlast.Field {
	for _, field := range fields {
		if graphQLResponseName(field) == responseName {
			return field
		}
	}

	return nil
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}

	return false
}

func operationRootDefinition(schema *graphqlast.Schema, operation *graphqlast.OperationDefinition) *graphqlast.Definition {
	switch operation.Operation {
	case graphqlast.Query:
		return schema.Query
	case graphqlast.Mutation:
		return schema.Mutation
	case graphqlast.Subscription:
		return schema.Subscription
	default:
		return nil
	}
}

func selectedRootFields(selection graphqlast.SelectionSet) []*graphqlast.Field {
	var fields []*graphqlast.Field

	for _, item := range selection {
		if field, ok := item.(*graphqlast.Field); ok {
			fields = append(fields, field)
		}
	}

	return fields
}

func graphQLResponseName(field *graphqlast.Field) string {
	if field.Alias != "" {
		return field.Alias
	}

	return field.Name
}

func operationVariable(operation *graphqlast.OperationDefinition, name string) *graphqlast.VariableDefinition {
	for _, variable := range operation.VariableDefinitions {
		if strings.TrimPrefix(variable.Variable, "$") == name {
			return variable
		}
	}

	return nil
}

func schemaField(definition *graphqlast.Definition, name string) *graphqlast.FieldDefinition {
	if definition == nil {
		return nil
	}

	for _, field := range definition.Fields {
		if field.Name == name {
			return field
		}
	}

	return nil
}

func graphqlPathPart(fieldName string, fieldType *graphqlast.Type) string {
	if fieldType != nil && fieldType.Elem != nil {
		return fieldName + "[]"
	}

	return fieldName
}

func graphqlNamedType(fieldType *graphqlast.Type) string {
	if fieldType == nil {
		return ""
	}

	if fieldType.Elem != nil {
		return graphqlNamedType(fieldType.Elem)
	}

	return fieldType.NamedType
}

func traceGraphQLGoType(
	expression goast.Expr,
	gqlType string,
	path []string,
	kind, operation string,
	document graphqlDocument,
	schema *graphqlast.Schema,
	goModels map[string]graphqlGoModel,
	registered, visitedTypes, traversed map[string]bool,
	traces map[string][]graphqlSelectionEvidence,
	decoderCases graphqlInterfaceDecoderCases,
) error {
	return traceGraphQLGoTypeWithInterfaceProjection(
		expression, gqlType, path, kind, operation, document, schema, goModels,
		registered, visitedTypes, traversed, traces, decoderCases, false, "", "",
	)
}

func traceGraphQLGoTypeWithInterfaceProjection(
	expression goast.Expr,
	gqlType string,
	path []string,
	kind, operation string,
	document graphqlDocument,
	schema *graphqlast.Schema,
	goModels map[string]graphqlGoModel,
	registered, visitedTypes, traversed map[string]bool,
	traces map[string][]graphqlSelectionEvidence,
	decoderCases graphqlInterfaceDecoderCases,
	interfaceProjection bool,
	abstractType string,
	abstractGoType string,
) error {
	evidence := graphqlEvidenceContext{
		goModels: goModels, registered: registered, schema: schema,
		visitedTypes: visitedTypes, traversed: traversed, traces: traces,
		decoderCases: decoderCases,
	}
	trace := graphqlGoTraceContext{evidence: &evidence, kind: kind, operation: operation, document: document}

	return trace.traceTypes(expression, gqlType, path, interfaceProjection, abstractType, abstractGoType)
}

type graphqlGoTraceContext struct {
	evidence  *graphqlEvidenceContext
	kind      string
	operation string
	document  graphqlDocument
}

func (trace *graphqlGoTraceContext) traceTypes(
	expression goast.Expr,
	gqlType string,
	path []string,
	interfaceProjection bool,
	abstractType, abstractGoType string,
) error {
	for _, name := range graphqlGoModelReferences(expression, trace.evidence.registered) {
		traceErr := trace.traceModel(name, gqlType, path, interfaceProjection, abstractType, abstractGoType)
		if traceErr != nil {
			return traceErr
		}
	}

	return nil
}

func (trace *graphqlGoTraceContext) traceModel(
	name, gqlType string,
	path []string,
	interfaceProjection bool,
	abstractType, abstractGoType string,
) error {
	model := trace.evidence.goModels[name]
	modelPath := strings.Join(path, ".")

	if interfaceProjection {
		var (
			shouldTrace bool
			err         error
		)

		modelPath, shouldTrace, err = trace.interfaceProjectionPath(name, model, gqlType, path, abstractType, abstractGoType)
		if err != nil {
			return err
		}

		if !shouldTrace {
			return nil
		}
	}

	trace.evidence.visitedTypes[name] = true

	visitKey := graphqlVisitKey(trace.operation, name, trace.kind)
	if trace.evidence.traversed[visitKey] {
		return nil
	}

	trace.evidence.traversed[visitKey] = true
	addGraphQLTrace(trace.evidence.traces, name, graphqlSelectionEvidence{
		Operation: trace.operation, Document: trace.document.Path,
		SelectionPath: modelPath, Kind: trace.kind,
	})

	if model.Kind == graphqlKindInterface {
		implementorErr := trace.traceImplementors(name, model, gqlType, path)
		if implementorErr != nil {
			return implementorErr
		}
	}

	if model.Kind != graphqlKindStruct {
		return nil
	}

	return trace.traceFields(name, model, gqlType, path, interfaceProjection)
}

func (trace *graphqlGoTraceContext) interfaceProjectionPath(
	name string,
	model graphqlGoModel,
	gqlType string,
	path []string,
	abstractType, abstractGoType string,
) (string, bool, error) {
	selection := trace.document.WireOperation.SelectionSet
	for _, field := range model.Fields {
		fieldPath := graphQLInterfaceFieldPath(selection, trace.evidence.schema, gqlType, path, field)
		if len(fieldPath) != 0 {
			return strings.Join(fieldPath, "."), true, nil
		}
	}

	if abstractType == "" || !graphqlSchemaPossibleType(trace.evidence.schema, abstractType, gqlType) {
		return "", false, nil
	}

	typenamePath := appendPath(path, graphqlTypeNameField)
	if findSelectedGraphQLField(selection, typenamePath) == nil {
		return "", false, graphqlEvidenceErrorf("GraphQL operation %s omits __typename for interface %s generated model %s", trace.operation, abstractType, name)
	}

	decoderCase, found := trace.evidence.decoderCases[abstractGoType].Cases[gqlType]
	if !found || decoderCase.Model != name || !decoderCase.JSONUnmarshal {
		return "", false, graphqlEvidenceErrorf("GraphQL interface decoder %s has no __typename case mapping %s to generated model %s", abstractGoType, gqlType, name)
	}

	modelPath := strings.Join(path, ".") + ".__typename (decoder case " + gqlType + " => " + name + ")"

	return modelPath, true, nil
}

func (trace *graphqlGoTraceContext) traceImplementors(name string, model graphqlGoModel, gqlType string, path []string) error {
	for _, implementor := range model.Implementors {
		concreteType, err := trace.validateGraphQLImplementor(name, implementor, gqlType)
		if err != nil {
			return err
		}

		traceErr := trace.traceTypes(goast.NewIdent(implementor), concreteType, path, true, gqlType, name)
		if traceErr != nil {
			return traceErr
		}
	}

	return nil
}

func (trace *graphqlGoTraceContext) validateGraphQLImplementor(
	interfaceName, implementor, gqlType string,
) (string, error) {
	concreteType := graphqlTypeFromComment(trace.evidence.goModels[implementor].Comment)
	if concreteType == "" {
		return "", graphqlEvidenceErrorf("GraphQL interface %s lists %s without a GraphQL type comment", interfaceName, implementor)
	}

	if !graphqlSchemaPossibleType(trace.evidence.schema, gqlType, concreteType) {
		return "", graphqlEvidenceErrorf("GraphQL interface %s lists %s, but schema type %s is not a possible type", interfaceName, implementor, concreteType)
	}

	decoderCase, found := trace.evidence.decoderCases[interfaceName].Cases[concreteType]
	if !found || decoderCase.Model != implementor || !decoderCase.JSONUnmarshal {
		return "", graphqlEvidenceErrorf("GraphQL interface decoder %s has no case mapping %s to generated model %s", interfaceName, concreteType, implementor)
	}

	return concreteType, nil
}

func (trace *graphqlGoTraceContext) traceFields(
	modelName string,
	model graphqlGoModel,
	gqlType string,
	path []string,
	interfaceProjection bool,
) error {
	for _, field := range model.Fields {
		var err error
		if interfaceProjection {
			err = trace.traceProjectedField(modelName, gqlType, path, field)
		} else {
			err = trace.traceSelectedField(modelName, gqlType, path, field)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (trace *graphqlGoTraceContext) traceProjectedField(
	modelName, gqlType string,
	path []string,
	field graphqlGoField,
) error {
	fieldDefinition := schemaField(trace.evidence.schema.Types[gqlType], field.Name)
	if field.Name != graphqlTypeNameField && fieldDefinition == nil {
		return graphqlEvidenceErrorf("GraphQL interface implementation %s field %s is absent from schema type %s", modelName, field.Name, gqlType)
	}

	fieldPath := graphQLInterfaceFieldPath(trace.document.WireOperation.SelectionSet, trace.evidence.schema, gqlType, path, field)
	if len(fieldPath) == 0 || fieldDefinition == nil {
		return nil
	}

	return trace.traceTypes(field.Type, graphqlNamedType(fieldDefinition.Type), fieldPath, false, "", "")
}

func (trace *graphqlGoTraceContext) traceSelectedField(
	modelName, gqlType string,
	path []string,
	field graphqlGoField,
) error {
	fieldDefinition := schemaField(trace.evidence.schema.Types[gqlType], field.Name)
	if field.Name == graphqlTypeNameField && fieldDefinition == nil {
		return nil
	}

	if fieldDefinition == nil {
		return graphqlEvidenceErrorf("GraphQL model %s field %s is absent from schema type %s", modelName, field.Name, gqlType)
	}

	fieldPath := appendPath(path, graphqlPathPart(field.Name, fieldDefinition.Type))
	if findSelectedGraphQLField(trace.document.WireOperation.SelectionSet, fieldPath) == nil {
		return graphqlEvidenceErrorf("GraphQL selection %s is absent from operation %s in %s", strings.Join(fieldPath, "."), trace.operation, trace.document.Path)
	}

	return trace.traceTypes(field.Type, graphqlNamedType(fieldDefinition.Type), fieldPath, false, "", "")
}

func graphqlSchemaPossibleType(schema *graphqlast.Schema, abstractType, concreteType string) bool {
	for _, possibleType := range schema.PossibleTypes[abstractType] {
		if possibleType.Name == concreteType {
			return true
		}
	}

	return false
}

func graphQLInterfaceFieldPath(selection graphqlast.SelectionSet, schema *graphqlast.Schema, gqlType string, path []string, field graphqlGoField) []string {
	if field.Name == graphqlTypeNameField {
		return nil
	}

	fieldDefinition := schemaField(schema.Types[gqlType], field.Name)
	if fieldDefinition == nil {
		return nil
	}

	fragmentPath := appendPath(path, "{"+gqlType+"}")

	fragmentFieldPath := appendPath(fragmentPath, graphqlPathPart(field.Name, fieldDefinition.Type))
	if findSelectedGraphQLField(selection, fragmentFieldPath) != nil {
		return fragmentFieldPath
	}

	fieldPath := appendPath(path, graphqlPathPart(field.Name, fieldDefinition.Type))
	if findSelectedGraphQLField(selection, fieldPath) != nil {
		return fieldPath
	}

	return nil
}

func traceGraphQLInputType(
	goType, gqlType string,
	path []string,
	operation string,
	document graphqlDocument,
	schema *graphqlast.Schema,
	goModels map[string]graphqlGoModel,
	registered, visitedTypes, traversed map[string]bool,
	traces map[string][]graphqlSelectionEvidence,
) error {
	visitKey := graphqlVisitKey(operation, goType, "input")
	if traversed[visitKey] {
		return nil
	}

	traversed[visitKey] = true
	visitedTypes[goType] = true

	addGraphQLTrace(traces, goType, graphqlSelectionEvidence{
		Operation: operation, Document: document.Path,
		SelectionPath: strings.Join(path, "."), Kind: "input",
	})

	model := goModels[goType]
	if model.Kind == graphqlKindInterface {
		return graphqlEvidenceErrorf("GraphQL input model %s unexpectedly uses an interface", goType)
	}

	if model.Kind != graphqlKindStruct {
		return nil
	}

	definition := schema.Types[gqlType]
	if definition == nil || !definition.IsInputType() {
		return graphqlEvidenceErrorf("GraphQL input model %s traces to non-input schema type %s", goType, gqlType)
	}

	commentType := graphqlTypeFromComment(model.Comment)
	if commentType != "" && commentType != gqlType {
		return graphqlEvidenceErrorf("GraphQL input model %s documents GraphQL type %s, not %s", goType, commentType, gqlType)
	}

	for _, field := range model.Fields {
		fieldDef := schemaField(definition, field.Name)
		if fieldDef == nil {
			return graphqlEvidenceErrorf("GraphQL input model %s field %s is absent from schema type %s", goType, field.Name, gqlType)
		}

		fieldPath := appendPath(path, graphqlPathPart(field.Name, fieldDef.Type))
		for _, childType := range graphqlGoModelReferences(field.Type, registered) {
			traceErr := traceGraphQLInputType(
				childType, graphqlNamedType(fieldDef.Type), fieldPath, operation, document,
				schema, goModels, registered, visitedTypes, traversed, traces,
			)
			if traceErr != nil {
				return traceErr
			}
		}
	}

	return nil
}

func graphqlGoModelReferences(expression goast.Expr, registered map[string]bool) []string {
	name := graphqlGoNamedType(expression)
	if registered[name] {
		return []string{name}
	}

	switch value := expression.(type) {
	case *goast.ArrayType:
		return graphqlGoModelReferences(value.Elt, registered)
	case *goast.StarExpr:
		return graphqlGoModelReferences(value.X, registered)
	case *goast.MapType:
		return graphqlGoModelReferences(value.Value, registered)
	default:
		return nil
	}
}

func graphqlTypeFromComment(comment string) string {
	for _, marker := range []string{"GraphQL type ", "GraphQL interface ", "GraphQL union "} {
		index := strings.Index(comment, marker)
		if index < 0 {
			continue
		}

		value := comment[index+len(marker):]
		name, _, _ := strings.Cut(value, ".")

		return strings.TrimSpace(name)
	}

	return ""
}

func findSelectedGraphQLField(selection graphqlast.SelectionSet, path []string) *graphqlast.Field {
	if len(path) == 0 {
		return nil
	}

	segment := path[0]
	if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
		return findSelectedGraphQLFragment(selection, path, segment)
	}

	return findSelectedGraphQLPathSegment(selection, path, strings.TrimSuffix(segment, "[]"))
}

func findSelectedGraphQLFragment(
	selection graphqlast.SelectionSet,
	path []string,
	segment string,
) *graphqlast.Field {
	typeCondition := strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")

	for _, item := range selection {
		fragment, isFragment := item.(*graphqlast.InlineFragment)
		if !isFragment || fragment.TypeCondition != typeCondition {
			continue
		}

		if len(path) == 1 {
			return nil
		}

		return findSelectedGraphQLField(fragment.SelectionSet, path[1:])
	}

	return nil
}

func findSelectedGraphQLPathSegment(
	selection graphqlast.SelectionSet,
	path []string,
	segment string,
) *graphqlast.Field {
	for _, item := range selection {
		switch selectionItem := item.(type) {
		case *graphqlast.Field:
			if graphQLResponseName(selectionItem) == segment {
				return selectedGraphQLChild(selectionItem, path)
			}
		case *graphqlast.InlineFragment:
			if found := findSelectedGraphQLField(selectionItem.SelectionSet, path); found != nil {
				return found
			}
		}
	}

	return nil
}

func selectedGraphQLChild(field *graphqlast.Field, path []string) *graphqlast.Field {
	if len(path) == 1 {
		return field
	}

	return findSelectedGraphQLField(field.SelectionSet, path[1:])
}

func appendPath(path []string, segment string) []string {
	result := append([]string(nil), path...)

	return append(result, segment)
}

func addGraphQLTrace(traces map[string][]graphqlSelectionEvidence, name string, trace graphqlSelectionEvidence) {
	traces[name] = append(traces[name], trace)
}

func graphqlVisitKey(operation, name, kind string) string {
	return operation + "|" + kind + "|" + name
}

func uniqueGraphQLTraces(traces []graphqlSelectionEvidence) []graphqlSelectionEvidence {
	sort.Slice(traces, func(left, right int) bool {
		if traces[left].Operation != traces[right].Operation {
			return traces[left].Operation < traces[right].Operation
		}

		if traces[left].SelectionPath != traces[right].SelectionPath {
			return traces[left].SelectionPath < traces[right].SelectionPath
		}

		return traces[left].Kind < traces[right].Kind
	})

	unique := traces[:0]
	for _, trace := range traces {
		if len(unique) == 0 || unique[len(unique)-1] != trace {
			unique = append(unique, trace)
		}
	}

	return unique
}

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	behaviorSchemaPath       = "api/behaviors.yaml"
	featureControlSchemaPath = "api/feature-controls.yaml"
	compatControlSchemaPath  = "api/compat/controls.yaml"
	behaviorRESTSourcePath   = "pkg/dependencies/rest/endpoints.go"
	featureGraphQLSourcePath = "pkg/dependencies/graphql/features.go"
	controlActionSourcePath  = "pkg/alexa/client_control.go"
	callerOpenParamsField    = "Params"
	behaviorSchemaObjectType = "object"
)

func checkBehaviorPayloadInventory() error {
	behaviorSchema, err := readPayloadSchema(behaviorSchemaPath)
	if err != nil {
		return err
	}

	err = checkBehaviorPayloadSchema(behaviorSchema)
	if err != nil {
		return err
	}

	featureSchema, err := readPayloadSchema(featureControlSchemaPath)
	if err != nil {
		return err
	}

	compatControlsSchema, err := readPayloadSchema(compatControlSchemaPath)
	if err != nil {
		return err
	}

	err = checkFeaturePayloadOpenBoundary(featureSchema, compatControlsSchema)
	if err != nil {
		return err
	}

	repositoryRoot, err := behaviorRepoRoot()
	if err != nil {
		return err
	}

	err = checkCallerOpenPayloadBoundaries(repositoryRoot, behaviorSchema, featureSchema)
	if err != nil {
		return err
	}

	return checkDynamicPayloadMapConstructions(repositoryRoot)
}

func readPayloadSchema(path string) (schemaDocument, error) {
	repositoryRoot, err := behaviorRepoRoot()
	if err != nil {
		return schemaDocument{}, err
	}

	data, err := fs.ReadFile(os.DirFS(repositoryRoot), filepath.ToSlash(path))
	if err != nil {
		return schemaDocument{}, fmt.Errorf("read payload schema %s: %w", path, err)
	}

	var document schemaDocument

	err = yaml.Unmarshal(data, &document)
	if err != nil {
		return schemaDocument{}, fmt.Errorf("parse payload schema %s: %w", path, err)
	}

	return document, nil
}

func checkBehaviorPayloadSchema(document schemaDocument) error {
	schemas := document.Components.Schemas

	payloadKeys, err := schemaEnumValues(schemas, "BehaviorPayloadKey")
	if err != nil {
		return err
	}

	operationTypes, err := schemaEnumValues(schemas, "BehaviorOperationType")
	if err != nil {
		return err
	}

	operationNodes := 0

	for name, nodeSchema := range schemas {
		if !strings.HasSuffix(name, "OperationNode") || name == "OpaquePayloadOperationNode" {
			continue
		}

		operationNodes++

		err = checkBehaviorOperationNodeSchema(name, nodeSchema, schemas, payloadKeys, operationTypes)
		if err != nil {
			return err
		}
	}

	if operationNodes == 0 {
		return diagnosticf("behavior schema has no typed operation nodes")
	}

	return nil
}

func checkBehaviorOperationNodeSchema(
	name string,
	nodeSchema schemaNode,
	schemas map[string]schemaNode,
	payloadKeys, operationTypes map[string]bool,
) error {
	payloadSchema, payloadSchemaFound := nodeSchema.Properties["operationPayload"]
	if !payloadSchemaFound || payloadSchema.Ref == "" {
		return diagnosticf("behavior node %s has no generated operationPayload reference", name)
	}

	payloadName := schemaComponentName(payloadSchema.Ref)
	if !strings.HasSuffix(payloadName, "Payload") {
		return diagnosticf("behavior node %s operationPayload references non-payload schema %s", name, payloadName)
	}

	if _, payloadDefinitionFound := schemas[payloadName]; !payloadDefinitionFound {
		return diagnosticf("behavior node %s operationPayload references missing schema %s", name, payloadName)
	}

	operationType, operationTypeFound := nodeSchema.Properties["type"]
	if !operationTypeFound {
		return diagnosticf("behavior node %s has no operation type", name)
	}

	if operationType.Ref == "" {
		if operationType.Type != schemaStringType || name != "FireTVOperationNode" {
			return diagnosticf("behavior node %s must reference a generated operation enum or preserve the Fire TV open string", name)
		}
	} else {
		values, err := schemaEnumValues(schemas, schemaComponentName(operationType.Ref))
		if err != nil {
			return err
		}

		for value := range values {
			if !operationTypes[value] {
				return diagnosticf("behavior operation type %q from %s is missing from BehaviorOperationType", value, name)
			}
		}
	}

	err := checkPayloadComponentKeys(schemas, payloadName, payloadKeys, map[string]bool{})
	if err != nil {
		return err
	}

	return nil
}

func schemaEnumValues(schemas map[string]schemaNode, name string) (map[string]bool, error) {
	schema, schemaFound := schemas[name]
	if !schemaFound || schema.Type != schemaStringType || len(schema.Enum) == 0 {
		return nil, diagnosticf("behavior schema is missing string enum %s", name)
	}

	values := make(map[string]bool, len(schema.Enum))

	for _, rawValue := range schema.Enum {
		value, valueIsString := rawValue.(string)
		if !valueIsString || value == "" {
			return nil, diagnosticf("behavior schema enum %s has a non-string or empty member", name)
		}

		values[value] = true
	}

	return values, nil
}

func checkPayloadComponentKeys(schemas map[string]schemaNode, name string, payloadKeys map[string]bool, visiting map[string]bool) error {
	if visiting[name] {
		return nil
	}

	schema, schemaFound := schemas[name]
	if !schemaFound {
		return diagnosticf("behavior payload references missing schema %s", name)
	}

	if schema.Type != behaviorSchemaObjectType {
		return nil
	}

	visiting[name] = true

	for propertyName, property := range schema.Properties {
		if !payloadKeys[propertyName] {
			return diagnosticf("behavior payload schema %s property %q is missing from BehaviorPayloadKey", name, propertyName)
		}

		if property.Ref != "" {
			err := checkPayloadComponentKeys(schemas, schemaComponentName(property.Ref), payloadKeys, visiting)
			if err != nil {
				return err
			}
		}

		if property.Items != nil && property.Items.Ref != "" {
			err := checkPayloadComponentKeys(schemas, schemaComponentName(property.Items.Ref), payloadKeys, visiting)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func schemaComponentName(reference string) string {
	return strings.TrimPrefix(reference, "#/components/schemas/")
}

func checkFeaturePayloadOpenBoundary(featureSchema, compatSchema schemaDocument) error {
	actionPayload, actionPayloadFound := featureSchema.Components.Schemas["ActionPayload"]
	if !actionPayloadFound || !schemaAllowsAdditionalProperties(actionPayload.AdditionalProperties) {
		return diagnosticf("feature-controls ActionPayload must explicitly preserve caller-open additional properties")
	}

	if len(actionPayload.Properties) != 1 || actionPayload.Properties["action"].Type != schemaStringType {
		return diagnosticf("feature-controls ActionPayload must keep only the library-defined action field closed")
	}

	for name, schema := range featureSchema.Components.Schemas {
		if name != "ActionPayload" && schemaAllowsAdditionalProperties(schema.AdditionalProperties) {
			return diagnosticf("feature-controls schema %s unexpectedly permits additional properties", name)
		}
	}

	actionRequest, actionRequestFound := compatSchema.Components.Schemas["ActionControlRequest"]
	if !actionRequestFound {
		return diagnosticf("compatibility controls schema is missing ActionControlRequest")
	}

	params, paramsFound := actionRequest.Properties["params"]
	if !paramsFound || params.Type != behaviorSchemaObjectType || !schemaAllowsAdditionalProperties(params.AdditionalProperties) {
		return diagnosticf("ActionControlRequest.params must remain a caller-open object in the compatibility schema")
	}

	return nil
}

func checkCallerOpenPayloadBoundaries(repositoryRoot string, behaviorSchema, featureSchema schemaDocument) error {
	opaqueNode, opaqueNodeFound := behaviorSchema.Components.Schemas["OpaquePayloadOperationNode"]
	if !opaqueNodeFound {
		return diagnosticf("behavior schema is missing the legacy OpaquePayloadOperationNode")
	}

	operationPayload, operationPayloadFound := opaqueNode.Properties["operationPayload"]
	if !operationPayloadFound || operationPayload.Type != behaviorSchemaObjectType ||
		!schemaAllowsAdditionalProperties(operationPayload.AdditionalProperties) {
		return diagnosticf("OpaquePayloadOperationNode.operationPayload must remain a caller-open object")
	}

	behaviorSource, err := readRepositorySource(repositoryRoot, behaviorRESTSourcePath)
	if err != nil {
		return err
	}

	err = checkSendSequenceOpenBoundary(behaviorRESTSourcePath, behaviorSource)
	if err != nil {
		return err
	}

	featureSource, err := readRepositorySource(repositoryRoot, featureGraphQLSourcePath)
	if err != nil {
		return err
	}

	err = checkActionPayloadOpenBoundary(featureGraphQLSourcePath, featureSource, featureSchema)
	if err != nil {
		return err
	}

	controlSource, err := readRepositorySource(repositoryRoot, controlActionSourcePath)
	if err != nil {
		return err
	}

	err = checkActionControlRequestOpenBoundary(controlActionSourcePath, controlSource)
	if err != nil {
		return err
	}

	return nil
}

func checkSendSequenceOpenBoundary(path string, source []byte) error {
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, path, source, 0)
	if err != nil {
		return fmt.Errorf("parse behavior source %s: %w", path, err)
	}

	for _, declaration := range file.Decls {
		function, isFunction := declaration.(*ast.FuncDecl)
		if !isFunction || function.Recv == nil || function.Name.Name != "SendSequence" {
			continue
		}

		if hasCallerOpenMapParameter(function.Type.Params, "operationPayload") {
			return nil
		}
	}

	return diagnosticf("%s must keep Client.SendSequence's operationPayload map input open", path)
}

func hasCallerOpenMapParameter(parameters *ast.FieldList, name string) bool {
	if parameters == nil {
		return false
	}

	for _, field := range parameters.List {
		if !isDynamicStringMap(field.Type) {
			continue
		}

		for _, identifier := range field.Names {
			if identifier.Name == name {
				return true
			}
		}
	}

	return false
}

func checkActionPayloadOpenBoundary(path string, source []byte, featureSchema schemaDocument) error {
	if !schemaAllowsAdditionalProperties(featureSchema.Components.Schemas["ActionPayload"].AdditionalProperties) {
		return diagnosticf("feature ActionPayload schema no longer accepts caller params")
	}

	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, path, source, 0)
	if err != nil {
		return fmt.Errorf("parse feature payload source %s: %w", path, err)
	}

	for _, declaration := range file.Decls {
		function, isFunction := declaration.(*ast.FuncDecl)
		if !isFunction || function.Name.Name != "ControlActionFeature" {
			continue
		}

		var found bool

		ast.Inspect(function.Body, func(node ast.Node) bool {
			literal, isLiteral := node.(*ast.CompositeLit)
			if !isLiteral || !isNamedCompositeType(literal.Type, "ActionPayload") {
				return true
			}

			for _, element := range literal.Elts {
				field, isField := element.(*ast.KeyValueExpr)
				if !isField || !isIdent(field.Key, "AdditionalProperties") {
					continue
				}

				selector, isSelector := field.Value.(*ast.SelectorExpr)
				if isSelector && isIdent(selector.X, "req") && selector.Sel.Name == callerOpenParamsField {
					found = true
				}
			}

			return true
		})

		if found {
			return nil
		}
	}

	return diagnosticf("%s Client.ControlActionFeature must pass caller req.Params to generated ActionPayload.AdditionalProperties", path)
}

func checkActionControlRequestOpenBoundary(path string, source []byte) error {
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, path, source, 0)
	if err != nil {
		return fmt.Errorf("parse action control source %s: %w", path, err)
	}

	for _, declaration := range file.Decls {
		function, isFunction := declaration.(*ast.FuncDecl)
		if !isFunction || function.Name.Name != "controlAction" {
			continue
		}

		var found bool

		ast.Inspect(function.Body, func(node ast.Node) bool {
			literal, isLiteral := node.(*ast.CompositeLit)
			if !isLiteral || !isNamedCompositeType(literal.Type, "ActionControlRequest") {
				return true
			}

			for _, element := range literal.Elts {
				field, isField := element.(*ast.KeyValueExpr)
				if !isField || !isIdent(field.Key, "Params") {
					continue
				}

				selector, isSelector := field.Value.(*ast.SelectorExpr)
				if isSelector && isIdent(selector.X, "payload") && selector.Sel.Name == callerOpenParamsField {
					found = true
				}
			}

			return true
		})

		if found {
			return nil
		}
	}

	return diagnosticf("%s Client.controlAction must pass caller payload.Params into ActionControlRequest.Params", path)
}

func checkDynamicPayloadMapConstructions(repositoryRoot string) error {
	filesystem := os.DirFS(repositoryRoot)

	err := fs.WalkDir(filesystem, "pkg", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk production Go sources at %s: %w", path, walkErr)
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") ||
			strings.HasSuffix(path, ".gen.go") {
			return nil
		}

		source, err := fs.ReadFile(filesystem, path)
		if err != nil {
			return fmt.Errorf("read production Go source %s: %w", path, err)
		}

		return checkDynamicPayloadMapSource(path, source)
	})
	if err != nil {
		return fmt.Errorf("scan production payload maps: %w", err)
	}

	return nil
}

func checkDynamicPayloadMapSource(path string, source []byte) error {
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, path, source, 0)
	if err != nil {
		return fmt.Errorf("parse production Go source %s: %w", path, err)
	}

	dynamicAliases := discoverDynamicMapAliases(file)

	err = findDynamicPayloadMapDefinition(fileSet, file, path, dynamicAliases)
	if err != nil {
		return err
	}

	err = findDynamicPayloadMapConstruction(fileSet, file, path, dynamicAliases)
	if err != nil {
		return err
	}

	return checkDynamicPayloadMapMutations(fileSet, file, path, dynamicAliases)
}

func discoverDynamicMapAliases(file *ast.File) map[string]bool {
	aliases := make(map[string]bool)
	changed := true

	for changed {
		changed = false

		ast.Inspect(file, func(node ast.Node) bool {
			typeSpec, isType := node.(*ast.TypeSpec)
			if !isType || aliases[typeSpec.Name.Name] || !isDynamicStringMapWithAliases(typeSpec.Type, aliases) {
				return true
			}

			aliases[typeSpec.Name.Name] = true
			changed = true

			return true
		})
	}

	return aliases
}

func findDynamicPayloadMapDefinition(fileSet *token.FileSet, file *ast.File, path string, aliases map[string]bool) error {
	var violation error

	ast.Inspect(file, func(node ast.Node) bool {
		typeSpec, isType := node.(*ast.TypeSpec)
		if !isType || !aliases[typeSpec.Name.Name] {
			return true
		}

		position := fileSet.Position(typeSpec.Pos())
		violation = diagnosticf("%s:%d declares a handwritten dynamic payload map type", path, position.Line)

		return false
	})

	return violation
}

func findDynamicPayloadMapConstruction(fileSet *token.FileSet, file *ast.File, path string, aliases map[string]bool) error {
	var violation error

	ast.Inspect(file, func(node ast.Node) bool {
		if !isDynamicPayloadMapConstruction(node, aliases) {
			return true
		}

		position := fileSet.Position(node.Pos())
		violation = diagnosticf(
			"%s:%d constructs a handwritten dynamic payload map; use a schema-generated DTO "+
				"or the inventoried caller-open boundary",
			path,
			position.Line,
		)

		return false
	})

	return violation
}

func checkDynamicPayloadMapMutations(fileSet *token.FileSet, file *ast.File, path string, aliases map[string]bool) error {
	var violation error

	ast.Inspect(file, func(node ast.Node) bool {
		if violation != nil {
			return false
		}

		function, isFunction := node.(*ast.FuncDecl)
		if isFunction && function.Body != nil {
			violation = checkDynamicMapMutationBody(fileSet, function.Body, function.Type.Params, path, aliases)

			return false
		}

		literal, isLiteral := node.(*ast.FuncLit)
		if isLiteral {
			violation = checkDynamicMapMutationBody(fileSet, literal.Body, literal.Type.Params, path, aliases)

			return false
		}

		return true
	})

	return violation
}

func checkDynamicMapMutationBody(
	fileSet *token.FileSet,
	body *ast.BlockStmt,
	parameters *ast.FieldList,
	path string,
	aliases map[string]bool,
) error {
	return checkDynamicMapMutationBodyWithInheritedNames(fileSet, body, parameters, path, aliases, nil)
}

func checkDynamicMapMutationBodyWithInheritedNames(
	fileSet *token.FileSet,
	body *ast.BlockStmt,
	parameters *ast.FieldList,
	path string,
	aliases map[string]bool,
	inheritedNames map[string]bool,
) error {
	dynamicNames := cloneDynamicMapNames(inheritedNames)
	dynamicNames = discoverFunctionDynamicMapNames(body, parameters, aliases, dynamicNames)

	var violation error

	ast.Inspect(body, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.AssignStmt:
			if !writesDynamicMapIndex(typed, aliases, dynamicNames) {
				return true
			}
		case *ast.IncDecStmt:
			if !isDynamicMapIndex(typed.X, aliases, dynamicNames) {
				return true
			}
		case *ast.CallExpr:
			if !mutatesDynamicMap(typed, aliases, dynamicNames) {
				return true
			}
		case *ast.FuncLit:
			violation = checkDynamicMapMutationBodyWithInheritedNames(
				fileSet,
				typed.Body,
				typed.Type.Params,
				path,
				aliases,
				dynamicNames,
			)

			return false
		default:
			return true
		}

		position := fileSet.Position(node.Pos())
		violation = diagnosticf("%s:%d mutates or adds a key to a caller-open dynamic payload map", path, position.Line)

		return false
	})

	return violation
}

func discoverFunctionDynamicMapNames(
	body *ast.BlockStmt,
	parameters *ast.FieldList,
	aliases map[string]bool,
	dynamicNames map[string]bool,
) map[string]bool {
	if parameters != nil {
		addDynamicFunctionParameters(dynamicNames, parameters, aliases)
	}

	changed := true

	for changed {
		changed = collectFunctionDynamicMapNames(body, dynamicNames, aliases)
	}

	return dynamicNames
}

func addDynamicFunctionParameters(names map[string]bool, parameters *ast.FieldList, aliases map[string]bool) {
	for _, field := range parameters.List {
		if isDynamicStringMapWithAliases(field.Type, aliases) {
			addDynamicNames(names, field.Names)
		}
	}
}

func collectFunctionDynamicMapNames(body *ast.BlockStmt, dynamicNames map[string]bool, aliases map[string]bool) bool {
	changed := false

	ast.Inspect(body, func(node ast.Node) bool {
		if _, isNestedFunction := node.(*ast.FuncLit); isNestedFunction {
			return false
		}

		switch typed := node.(type) {
		case *ast.ValueSpec:
			changed = addValueSpecDynamicMapNames(dynamicNames, typed, aliases) || changed
		case *ast.AssignStmt:
			changed = addAssignmentDynamicMapNames(dynamicNames, typed, aliases) || changed
		}

		return true
	})

	return changed
}

func addValueSpecDynamicMapNames(names map[string]bool, valueSpec *ast.ValueSpec, aliases map[string]bool) bool {
	changed := false
	if isDynamicStringMapWithAliases(valueSpec.Type, aliases) {
		changed = addDynamicNames(names, valueSpec.Names) || changed
	}

	for index, name := range valueSpec.Names {
		rightIndex := index
		if len(valueSpec.Values) == 1 {
			rightIndex = 0
		}

		if rightIndex < len(valueSpec.Values) && isDynamicMapExpression(valueSpec.Values[rightIndex], aliases, names) {
			changed = addDynamicName(names, name.Name) || changed
		}
	}

	return changed
}

func addAssignmentDynamicMapNames(names map[string]bool, assignment *ast.AssignStmt, aliases map[string]bool) bool {
	changed := false

	for index, left := range assignment.Lhs {
		identifier, isIdentifier := left.(*ast.Ident)
		if !isIdentifier {
			continue
		}

		rightIndex := index
		if len(assignment.Rhs) == 1 {
			rightIndex = 0
		}

		if rightIndex < len(assignment.Rhs) && isDynamicMapExpression(assignment.Rhs[rightIndex], aliases, names) {
			changed = addDynamicName(names, identifier.Name) || changed
		}
	}

	return changed
}

func cloneDynamicMapNames(names map[string]bool) map[string]bool {
	cloned := make(map[string]bool, len(names))
	for name := range names {
		cloned[name] = true
	}

	return cloned
}

func addDynamicNames(names map[string]bool, identifiers []*ast.Ident) bool {
	changed := false
	for _, identifier := range identifiers {
		changed = addDynamicName(names, identifier.Name) || changed
	}

	return changed
}

func addDynamicName(names map[string]bool, name string) bool {
	if names[name] {
		return false
	}

	names[name] = true

	return true
}

func isDynamicPayloadMapConstruction(node ast.Node, aliases map[string]bool) bool {
	switch typed := node.(type) {
	case *ast.CompositeLit:
		return isDynamicStringMapWithAliases(typed.Type, aliases)
	case *ast.CallExpr:
		if isDynamicStringMapWithAliases(typed.Fun, aliases) {
			return true
		}

		function, isBuiltin := unparenthesize(typed.Fun).(*ast.Ident)
		if !isBuiltin || (function.Name != "make" && function.Name != builtinNewName) || len(typed.Args) == 0 {
			return false
		}

		return isDynamicStringMapWithAliases(typed.Args[0], aliases)
	default:
		return false
	}
}

func writesDynamicMapIndex(
	assignment *ast.AssignStmt,
	aliases map[string]bool,
	dynamicNames map[string]bool,
) bool {
	for _, left := range assignment.Lhs {
		index, isIndex := left.(*ast.IndexExpr)
		if !isIndex || !isDynamicMapExpression(index.X, aliases, dynamicNames) {
			continue
		}

		return true
	}

	return false
}

func isDynamicMapIndex(expression ast.Expr, aliases map[string]bool, dynamicNames map[string]bool) bool {
	index, isIndex := expression.(*ast.IndexExpr)

	return isIndex && isDynamicMapExpression(index.X, aliases, dynamicNames)
}

func mutatesDynamicMap(call *ast.CallExpr, aliases map[string]bool, dynamicNames map[string]bool) bool {
	function, isBuiltin := unparenthesize(call.Fun).(*ast.Ident)
	if !isBuiltin || len(call.Args) == 0 || (function.Name != builtinDeleteName && function.Name != builtinClearName) {
		return false
	}

	return isDynamicMapExpression(call.Args[0], aliases, dynamicNames)
}

func isDynamicMapExpression(expression ast.Expr, aliases map[string]bool, dynamicNames map[string]bool) bool {
	switch typed := expression.(type) {
	case *ast.Ident:
		return dynamicNames[typed.Name]
	case *ast.SelectorExpr:
		return typed.Sel.Name == callerOpenParamsField
	case *ast.ParenExpr:
		return isDynamicMapExpression(typed.X, aliases, dynamicNames)
	case *ast.StarExpr:
		return isDynamicMapExpression(typed.X, aliases, dynamicNames)
	case *ast.TypeAssertExpr:
		return isDynamicStringMapWithAliases(typed.Type, aliases)
	default:
		return false
	}
}

func isDynamicStringMapWithAliases(expression ast.Expr, aliases map[string]bool) bool {
	expression = unparenthesize(expression)

	if isDynamicStringMap(expression) {
		return true
	}

	identifier, isIdentifier := expression.(*ast.Ident)

	return isIdentifier && aliases[identifier.Name]
}

func isDynamicStringMap(expression ast.Expr) bool {
	expression = unparenthesize(expression)

	mapType, isMap := expression.(*ast.MapType)
	if !isMap {
		return false
	}

	key, keyIsString := mapType.Key.(*ast.Ident)
	if !keyIsString || key.Name != schemaStringType {
		return false
	}

	if value, isIdentifier := mapType.Value.(*ast.Ident); isIdentifier {
		return value.Name == "any"
	}

	value, isInterface := mapType.Value.(*ast.InterfaceType)

	return isInterface && len(value.Methods.List) == 0
}

func unparenthesize(expression ast.Expr) ast.Expr {
	for {
		parenthesized, isParenthesized := expression.(*ast.ParenExpr)
		if !isParenthesized {
			return expression
		}

		expression = parenthesized.X
	}
}

func isNamedCompositeType(expression ast.Expr, name string) bool {
	selector, isSelector := expression.(*ast.SelectorExpr)
	if isSelector {
		return selector.Sel.Name == name
	}

	identifier, isIdentifier := expression.(*ast.Ident)

	return isIdentifier && identifier.Name == name
}

func isIdent(expression ast.Expr, name string) bool {
	identifier, isIdentifier := expression.(*ast.Ident)

	return isIdentifier && identifier.Name == name
}

func readRepositorySource(repositoryRoot, path string) ([]byte, error) {
	source, err := fs.ReadFile(os.DirFS(repositoryRoot), filepath.ToSlash(path))
	if err != nil {
		return nil, fmt.Errorf("read source %s: %w", path, err)
	}

	return source, nil
}

func behaviorRepoRoot() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory for payload inventory: %w", err)
	}

	for _, candidate := range []string{workingDirectory, filepath.Join(workingDirectory, "..", "..")} {
		repositoryRoot, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("resolve payload inventory repository root %s: %w", candidate, err)
		}

		_, err = os.Stat(filepath.Join(repositoryRoot, filepath.FromSlash(behaviorSchemaPath)))
		if err == nil {
			return repositoryRoot, nil
		}
	}

	return "", diagnosticf("could not locate behavior schema %s from the working directory", behaviorSchemaPath)
}

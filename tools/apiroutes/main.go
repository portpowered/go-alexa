// Command apiroutes generates the HTTP route and stream framing definitions
// consumed by go-alexa from its checked-in OpenAPI and AsyncAPI documents.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	apiRoutesPackage                = "apiroutes"
	newRequestWithContextName       = "NewRequestWithContext"
	fullURLRequestHelperName        = "doRequestWithFullURL"
	fullURLJSONHelperName           = "doJSONRequestWithFullURL"
	requestHeaderFieldName          = "Header"
	querySetMethodName              = "Set"
	queryAddMethodName              = "Add"
	queryEncodeMethodName           = "Encode"
	routeOutputDirectoryPermissions = 0o750
	routeOutputFilePermissions      = 0o600
	regionalServerCapacity          = 3
	requestWithContextArgumentCount = 3
	requestArgumentCount            = 2
)

type openAPIDocument struct {
	Servers        []regionalServer `yaml:"servers"`
	RequestHeaders []string         `yaml:"x-go-alexa-request-headers"`
	Paths          map[string]map[string]struct {
		OperationID string           `yaml:"operationId"`
		Servers     []regionalServer `yaml:"servers"`
		Parameters  []apiParameter   `yaml:"parameters"`
	} `yaml:"paths"`
	Components struct {
		Parameters map[string]apiParameter `yaml:"parameters"`
	} `yaml:"components"`
}

type apiParameter struct {
	Ref  string `yaml:"$ref"`
	Name string `yaml:"name"`
	In   string `yaml:"in"`
}

type namedParameter struct {
	name, value string
}

type asyncAPIDocument struct {
	Servers  map[string]regionalServer `yaml:"servers"`
	Channels map[string]struct {
		Address string        `yaml:"address"`
		Framing streamFraming `yaml:"x-go-alexa-stream-framing"`
	} `yaml:"channels"`
	Operations map[string]struct {
		Action  string `yaml:"action"`
		Channel struct {
			Ref string `yaml:"$ref"`
		} `yaml:"channel"`
	} `yaml:"operations"`
}

type route struct {
	name, method, path string
}

type regionalServer struct {
	URL    string `yaml:"url"`
	Host   string `yaml:"host"`
	Region string `yaml:"x-go-alexa-region"`
}

type regionalEndpoints struct {
	AlexaAPI            map[string]string
	AmazonAPI           map[string]string
	AlexaWeb            map[string]string
	Event               map[string]string
	ExchangeURLTemplate string
}

type streamFraming struct {
	LineDelimiter               string `yaml:"line-delimiter"`
	TrimWhitespace              bool   `yaml:"trim-whitespace"`
	BoundaryPrefix              string `yaml:"boundary-prefix"`
	ContentTypePrefix           string `yaml:"content-type-prefix"`
	SkipEmptyLines              bool   `yaml:"skip-empty-lines"`
	AuthenticationFailureMarker string `yaml:"authentication-failure-marker"`
}

var pathLiteralPattern = regexp.MustCompile(`^(?:/|https?://[^/]+/)`)

func main() {
	check := flag.Bool("check", false, "check generated output and production route literals")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	err := run(*root, *check)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, check bool) error {
	var openAPI openAPIDocument

	err := readYAML(filepath.Join(root, "api", "openapi.yaml"), &openAPI)
	if err != nil {
		return err
	}

	var asyncAPI asyncAPIDocument

	err = readYAML(filepath.Join(root, "api", "asyncapi.yaml"), &asyncAPI)
	if err != nil {
		return err
	}

	routes, err := collectRoutes(openAPI)
	if err != nil {
		return err
	}

	parameters, err := collectParameters(openAPI)
	if err != nil {
		return err
	}

	endpoints, err := collectRegionalEndpoints(openAPI, asyncAPI)
	if err != nil {
		return err
	}

	channel, framing, err := collectDirectiveChannel(asyncAPI, routes)
	if err != nil {
		return err
	}

	want, err := format.Source(render(routes, parameters, channel, framing, endpoints))
	if err != nil {
		return fmt.Errorf("format generated route definitions: %w", err)
	}

	outputPath := filepath.Join(root, "pkg", "internal", "apiroutes", "routes.gen.go")
	if check {
		//nolint:gosec // The path is the generated route file beneath the repository root.
		got, err := os.ReadFile(outputPath)
		if err != nil {
			return fmt.Errorf("read generated routes: %w", err)
		}

		if !bytes.Equal(got, want) {
			return fmt.Errorf("%s is stale; run make generate-api", filepath.ToSlash(outputPath))
		}

		return checkProductionRouteLiterals(root, routes, endpoints)
	}

	err = os.MkdirAll(filepath.Dir(outputPath), routeOutputDirectoryPermissions)
	if err != nil {
		return fmt.Errorf("create route output directory: %w", err)
	}

	err = os.WriteFile(outputPath, want, routeOutputFilePermissions)
	if err != nil {
		return fmt.Errorf("write generated route definitions: %w", err)
	}

	return nil
}

func readYAML(path string, target any) error {
	//nolint:gosec // The CLI reads the explicit local OpenAPI or AsyncAPI path.
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read YAML %s: %w", path, err)
	}

	err = yaml.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("parse %s: %w", filepath.ToSlash(path), err)
	}

	return nil
}

func collectRoutes(doc openAPIDocument) ([]route, error) {
	var routes []route

	seenNames := make(map[string]struct{})

	for path, operations := range doc.Paths {
		for method, operation := range operations {
			if operation.OperationID == "" {
				return nil, fmt.Errorf("OpenAPI operation %s %s has no operationId", strings.ToUpper(method), path)
			}

			name := exportedIdentifier(operation.OperationID)
			if _, exists := seenNames[name]; exists {
				return nil, fmt.Errorf("duplicate OpenAPI operationId %q", operation.OperationID)
			}

			seenNames[name] = struct{}{}

			if strings.Contains(path, "{") != strings.Contains(path, "}") {
				return nil, fmt.Errorf("unbalanced path parameter in OpenAPI path %q", path)
			}

			goPath := regexp.MustCompile(`\{[^{}]+\}`).ReplaceAllString(path, "%s")
			routes = append(routes, route{name: name, method: strings.ToUpper(method), path: goPath})
		}
	}

	if len(routes) == 0 {
		return nil, errors.New("OpenAPI document contains no HTTP operations")
	}

	sort.Slice(routes, func(i, j int) bool { return routes[i].name < routes[j].name })

	return routes, nil
}

func collectParameters(doc openAPIDocument) ([]namedParameter, error) {
	seen := make(map[string]string)

	for _, operations := range doc.Paths {
		for _, operation := range operations {
			if operation.OperationID == "" {
				continue
			}

			for _, parameter := range operation.Parameters {
				resolved, err := resolveParameterReference(doc, parameter)
				if err != nil {
					return nil, err
				}

				err = addNamedParameter(seen, resolved)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	for _, header := range doc.RequestHeaders {
		err := addNamedParameter(seen, apiParameter{Name: header, In: "header", Ref: ""})
		if err != nil {
			return nil, err
		}
	}

	parameters := make([]namedParameter, 0, len(seen))
	for name, value := range seen {
		parameters = append(parameters, namedParameter{name: name, value: value})
	}

	sort.Slice(parameters, func(i, j int) bool { return parameters[i].name < parameters[j].name })

	return parameters, nil
}

func resolveParameterReference(doc openAPIDocument, parameter apiParameter) (apiParameter, error) {
	if parameter.Ref == "" {
		return parameter, nil
	}

	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(parameter.Ref, prefix) {
		return apiParameter{}, fmt.Errorf("unsupported parameter reference %q", parameter.Ref)
	}

	ref := parameter.Ref

	resolved, ok := doc.Components.Parameters[strings.TrimPrefix(ref, prefix)]
	if !ok {
		return apiParameter{}, fmt.Errorf("undefined parameter reference %q", ref)
	}

	return resolved, nil
}

func addNamedParameter(seen map[string]string, parameter apiParameter) error {
	if parameter.Name == "" || (parameter.In != "query" && parameter.In != "header" && parameter.In != "path") {
		return fmt.Errorf("OpenAPI parameter has missing name or unsupported location: %+v", parameter)
	}

	name := exportedIdentifier(parameter.In) + exportedIdentifier(parameter.Name)
	if parameter.In == "query" {
		name = "QueryParam" + exportedIdentifier(parameter.Name)
	}

	if old, ok := seen[name]; ok && old != parameter.Name {
		return fmt.Errorf("parameter identifier collision %q", name)
	}

	seen[name] = parameter.Name

	return nil
}

func collectRegionalEndpoints(openAPI openAPIDocument, asyncAPI asyncAPIDocument) (regionalEndpoints, error) {
	endpoints := regionalEndpoints{
		AlexaAPI:            nil,
		AmazonAPI:           nil,
		AlexaWeb:            nil,
		Event:               nil,
		ExchangeURLTemplate: "",
	}

	var err error

	endpoints.AlexaAPI, err = collectRegionalURLs(openAPI.Servers, "url", "OpenAPI top-level servers")
	if err != nil {
		return endpoints, err
	}

	authorizationServers, err := operationServers(openAPI, "openAuthorizationPage")
	if err != nil {
		return endpoints, err
	}

	endpoints.AmazonAPI, err = collectRegionalURLs(authorizationServers, "url", "openAuthorizationPage servers")
	if err != nil {
		return endpoints, err
	}

	webServers, err := operationServers(openAPI, "getUserInfo")
	if err != nil {
		return endpoints, err
	}

	endpoints.AlexaWeb, err = collectRegionalURLs(webServers, "url", "getUserInfo servers")
	if err != nil {
		return endpoints, err
	}

	eventServers := make([]regionalServer, 0, len(asyncAPI.Servers))
	for _, server := range asyncAPI.Servers {
		eventServers = append(eventServers, server)
	}

	endpoints.Event, err = collectRegionalURLs(eventServers, "host", "AsyncAPI servers")
	if err != nil {
		return endpoints, err
	}

	exchangeServers, err := operationServers(openAPI, "exchangeRefreshTokenForCookies")
	if err != nil {
		return endpoints, err
	}

	if len(exchangeServers) != 1 || strings.Count(exchangeServers[0].URL, "{") != 1 || strings.Count(exchangeServers[0].URL, "}") != 1 {
		return endpoints, errors.New("exchangeRefreshTokenForCookies must declare one variable in its server URL")
	}

	endpoints.ExchangeURLTemplate = regexp.MustCompile(`\{[^{}]+\}`).ReplaceAllString(exchangeServers[0].URL, "%s")

	return endpoints, nil
}

func operationServers(doc openAPIDocument, operationID string) ([]regionalServer, error) {
	for _, operations := range doc.Paths {
		for _, operation := range operations {
			if operation.OperationID == operationID {
				if len(operation.Servers) == 0 {
					return nil, fmt.Errorf("OpenAPI operation %q has no regional server list", operationID)
				}

				return operation.Servers, nil
			}
		}
	}

	return nil, fmt.Errorf("OpenAPI operation %q not found", operationID)
}

func collectRegionalURLs(servers []regionalServer, field, source string) (map[string]string, error) {
	urls := make(map[string]string, regionalServerCapacity)

	for _, server := range servers {
		var value string
		if field == "url" {
			value = server.URL
		} else {
			value = server.Host
		}

		if value == "" || server.Region == "" {
			return nil, fmt.Errorf("%s must label every server with x-go-alexa-region and a %s", source, field)
		}

		if server.Region != "na" && server.Region != "eu" && server.Region != "jp" {
			return nil, fmt.Errorf("%s has unsupported region %q", source, server.Region)
		}

		if _, exists := urls[server.Region]; exists {
			return nil, fmt.Errorf("%s has duplicate %q region server", source, server.Region)
		}

		urls[server.Region] = value
	}

	for _, region := range []string{"na", "eu", "jp"} {
		if urls[region] == "" {
			return nil, fmt.Errorf("%s is missing the %q region server", source, region)
		}
	}

	return urls, nil
}

func collectDirectiveChannel(doc asyncAPIDocument, routes []route) (string, streamFraming, error) {
	var framing streamFraming

	channel, ok := doc.Channels["directives"]
	if !ok {
		return "", framing, errors.New("AsyncAPI document is missing the directives channel")
	}

	if channel.Address == "" || !strings.HasPrefix(channel.Address, "/") {
		return "", framing, errors.New("AsyncAPI directives channel must have an absolute path address")
	}

	foundOperation := false

	for _, operation := range doc.Operations {
		if operation.Channel.Ref == "#/channels/directives" {
			foundOperation = true

			break
		}
	}

	if !foundOperation {
		return "", framing, errors.New("AsyncAPI directives channel has no operation")
	}

	matched := false

	for _, route := range routes {
		if route.method == "GET" && route.path == channel.Address {
			matched = true

			break
		}
	}

	if !matched {
		return "", framing, fmt.Errorf("AsyncAPI directives address %q has no matching GET operation in OpenAPI", channel.Address)
	}

	framing = channel.Framing
	if len(framing.LineDelimiter) != 1 || framing.BoundaryPrefix == "" || framing.ContentTypePrefix == "" || framing.AuthenticationFailureMarker == "" {
		return "", framing, errors.New("AsyncAPI directives channel has incomplete x-go-alexa-stream-framing configuration")
	}

	return channel.Address, framing, nil
}

func render(routes []route, parameters []namedParameter, channel string, framing streamFraming, endpoints regionalEndpoints) []byte {
	var out strings.Builder

	out.WriteString("// Code generated by tools/apiroutes; DO NOT EDIT.\n")
	out.WriteString("package apiroutes\n\n")
	out.WriteString("// HTTP operation method and path templates are generated from api/openapi.yaml.\n")
	out.WriteString("const (\n")

	for _, route := range routes {
		fmt.Fprintf(&out, "\tMethod%s = %s\n", route.name, strconv.Quote(route.method))
		fmt.Fprintf(&out, "\tPath%s = %s\n", route.name, strconv.Quote(route.path))
	}

	for _, parameter := range parameters {
		fmt.Fprintf(&out, "\t%s = %s\n", parameter.name, strconv.Quote(parameter.value))
	}

	out.WriteString("\tServerAlexaAPIBaseNa = " + strconv.Quote(endpoints.AlexaAPI["na"]) + "\n")
	out.WriteString("\tServerAlexaAPIBaseEu = " + strconv.Quote(endpoints.AlexaAPI["eu"]) + "\n")
	out.WriteString("\tServerAlexaAPIBaseJp = " + strconv.Quote(endpoints.AlexaAPI["jp"]) + "\n")
	out.WriteString("\tServerAmazonAPIBaseNa = " + strconv.Quote(endpoints.AmazonAPI["na"]) + "\n")
	out.WriteString("\tServerAmazonAPIBaseEu = " + strconv.Quote(endpoints.AmazonAPI["eu"]) + "\n")
	out.WriteString("\tServerAmazonAPIBaseJp = " + strconv.Quote(endpoints.AmazonAPI["jp"]) + "\n")
	out.WriteString("\tServerAlexaWebBaseNa = " + strconv.Quote(endpoints.AlexaWeb["na"]) + "\n")
	out.WriteString("\tServerAlexaWebBaseEu = " + strconv.Quote(endpoints.AlexaWeb["eu"]) + "\n")
	out.WriteString("\tServerAlexaWebBaseJp = " + strconv.Quote(endpoints.AlexaWeb["jp"]) + "\n")
	out.WriteString("\tServerEventAuthorityNa = " + strconv.Quote(endpoints.Event["na"]) + "\n")
	out.WriteString("\tServerEventAuthorityEu = " + strconv.Quote(endpoints.Event["eu"]) + "\n")
	out.WriteString("\tServerEventAuthorityJp = " + strconv.Quote(endpoints.Event["jp"]) + "\n")
	out.WriteString("\tServerExchangeRefreshTokenForCookies = " + strconv.Quote(endpoints.ExchangeURLTemplate) + "\n")
	out.WriteString(")\n\n")
	out.WriteString("// IsKnownRequestHeader reports whether a request header is declared in api/openapi.yaml.\n")
	out.WriteString("func IsKnownRequestHeader(name string) bool {\n\tswitch name {\n\tcase ")

	var headerNames []string

	for _, parameter := range parameters {
		if strings.HasPrefix(parameter.name, "Header") {
			headerNames = append(headerNames, parameter.name)
		}
	}

	out.WriteString(strings.Join(headerNames, ", "))
	out.WriteString(":\n\t\treturn true\n\tdefault:\n\t\treturn false\n\t}\n}\n\n")
	out.WriteString("// ChannelDirectivesAddress and DirectiveStreamFraming are generated from api/asyncapi.yaml.\n")
	fmt.Fprintf(&out, "const ChannelDirectivesAddress = %s\n\n", strconv.Quote(channel))
	out.WriteString("type DirectiveStreamFraming struct {\n")
	out.WriteString("\tLineDelimiter byte\n\tTrimWhitespace bool\n")
	out.WriteString("\tBoundaryPrefix string\n\tContentTypePrefix string\n")
	out.WriteString("\tSkipEmptyLines bool\n\tAuthenticationFailureMarker string\n")
	out.WriteString("}\n\n")
	out.WriteString("var DirectiveFraming = DirectiveStreamFraming{\n")
	fmt.Fprintf(&out, "\tLineDelimiter: %s[0],\n", strconv.Quote(framing.LineDelimiter))
	fmt.Fprintf(&out, "\tTrimWhitespace: %t,\n", framing.TrimWhitespace)
	fmt.Fprintf(&out, "\tBoundaryPrefix: %s,\n", strconv.Quote(framing.BoundaryPrefix))
	fmt.Fprintf(&out, "\tContentTypePrefix: %s,\n", strconv.Quote(framing.ContentTypePrefix))
	fmt.Fprintf(&out, "\tSkipEmptyLines: %t,\n", framing.SkipEmptyLines)
	fmt.Fprintf(&out, "\tAuthenticationFailureMarker: %s,\n", strconv.Quote(framing.AuthenticationFailureMarker))
	out.WriteString("}\n")

	return []byte(out.String())
}

func exportedIdentifier(value string) string {
	if value == "" {
		return value
	}

	parts := strings.FieldsFunc(value, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9')
	})
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, "")
}

func checkProductionRouteLiterals(root string, routes []route, endpoints regionalEndpoints) error {
	base := filepath.Join(root, "pkg")

	var violations []string

	aliases, err := routeAliases(filepath.Join(base, "dependencymodels", "constants.go"))
	if err != nil {
		return err
	}

	knownMethods := make(map[string]string, len(routes))
	for _, route := range routes {
		knownMethods[route.name] = route.path
	}

	err = filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}
		// pkg/testing replays inbound response headers; it does not send provider requests.
		if strings.HasPrefix(path, filepath.Join(base, "testing")+string(filepath.Separator)) {
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".gen.go") {
			return nil
		}

		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse Go file %s: %w", path, err)
		}

		exemptLiterals := nonRouteURLLiterals(file)
		checkWireCallsites(file, fset, aliases, knownMethods, strings.Count(endpoints.ExchangeURLTemplate, "%s"), &violations)
		checkNetworkInventory(file, fset, filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator))), &violations)
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}

			value, err := strconv.Unquote(literal.Value)
			if err == nil && value != "/" && !exemptLiterals[literal.Pos()] && pathLiteralPattern.MatchString(value) {
				position := fset.Position(literal.Pos())
				violations = append(violations, fmt.Sprintf("%s:%d contains a raw route or URL literal %q", filepath.ToSlash(position.Filename), position.Line, value))
			}

			return true
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("scan Go files for route literals: %w", err)
	}

	if len(violations) > 0 {
		sort.Strings(violations)

		return fmt.Errorf("production route strings must come from generated schema definitions:\n%s", strings.Join(violations, "\n"))
	}

	return nil
}

func routeAliases(path string) (map[string]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse route source %s: %w", path, err)
	}

	aliases := make(map[string]string)

	ast.Inspect(file, func(node ast.Node) bool {
		value, ok := node.(*ast.ValueSpec)
		if !ok {
			return true
		}

		for i, name := range value.Names {
			if i >= len(value.Values) {
				continue
			}

			if selector, ok := value.Values[i].(*ast.SelectorExpr); ok && strings.HasPrefix(selector.Sel.Name, "Path") {
				aliases[name.Name] = strings.TrimPrefix(selector.Sel.Name, "Path")
			}
		}

		return true
	})

	return aliases, nil
}

func checkWireCallsites(
	file *ast.File,
	fset *token.FileSet,
	aliases map[string]string,
	methods map[string]string,
	serverTemplateArity int,
	violations *[]string,
) {
	checker := wireCallsiteChecker{
		fset: fset, aliases: aliases, methods: methods, serverTemplateArity: serverTemplateArity, violations: violations,
	}
	checker.checkTransportMethodValues(file)

	ast.Inspect(file, func(node ast.Node) bool {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			return true
		}

		analysis := analyzeWireFunction(decl.Body)
		checker.checkFunction(decl, analysis)

		return false
	})
}

type wireFunctionAnalysis struct {
	assignments    map[string][]routeAssignment
	queryValues    map[any]bool
	headerAliases  map[string]bool
	addressEscapes map[any][]token.Pos
}

type wireCallsiteChecker struct {
	fset                *token.FileSet
	aliases             map[string]string
	methods             map[string]string
	serverTemplateArity int
	violations          *[]string
}

func analyzeWireFunction(body *ast.BlockStmt) wireFunctionAnalysis {
	conditional := conditionalAssignmentPositions(body)
	analysis := wireFunctionAnalysis{
		assignments:    make(map[string][]routeAssignment),
		queryValues:    make(map[any]bool),
		headerAliases:  make(map[string]bool),
		addressEscapes: make(map[any][]token.Pos),
	}

	ast.Inspect(body, func(node ast.Node) bool {
		if address, ok := node.(*ast.UnaryExpr); ok && address.Op == token.AND {
			if variable, ok := unparenthesizedIdent(address.X); ok {
				analysis.addressEscapes[variable.Obj] = append(analysis.addressEscapes[variable.Obj], address.Pos())
			}
		}

		if declaration, ok := node.(*ast.ValueSpec); ok {
			analyzeHeaderDeclarations(declaration, analysis.headerAliases)
			analyzeQueryDeclarations(declaration, analysis.queryValues)
			analyzeWireDeclarations(declaration, analysis.assignments)
		}

		assignment, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}

		analyzeWireAssignments(assignment, &analysis, conditional[assignment.Pos()])

		return true
	})
	markCustomHeaderAliases(body, &analysis)

	return analysis
}

func unparenthesizedIdent(value ast.Expr) (*ast.Ident, bool) {
	variable, ok := unparenthesizedExpr(value).(*ast.Ident)

	return variable, ok
}

func unparenthesizedExpr(value ast.Expr) ast.Expr {
	for {
		if parenthesized, ok := value.(*ast.ParenExpr); ok {
			value = parenthesized.X

			continue
		}

		return value
	}
}

func conditionalAssignmentPositions(body ast.Node) map[token.Pos]token.Pos {
	positions := make(map[token.Pos]token.Pos)

	var parents []ast.Node

	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil {
			parents = parents[:len(parents)-1]

			return true
		}

		if assignment, ok := node.(*ast.AssignStmt); ok {
			positions[assignment.Pos()] = conditionalScopeEnd(parents, assignment.Pos())
		}

		parents = append(parents, node)

		return true
	})

	return positions
}

func conditionalScopeEnd(parents []ast.Node, assignment token.Pos) token.Pos {
	for index := len(parents) - 1; index >= 0; index-- {
		if !isControlFlowNode(parents[index]) {
			continue
		}

		if index+1 < len(parents) {
			return parents[index+1].End()
		}

		return assignment
	}

	return token.NoPos
}

func isControlFlowNode(node ast.Node) bool {
	switch node.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
		return true
	default:
		return false
	}
}

func markCustomHeaderAliases(body *ast.BlockStmt, analysis *wireFunctionAnalysis) {
	const customHeaderArgumentIndex = 4

	analysis.headerAliases["customHeaders"] = true

	ast.Inspect(body, func(node ast.Node) bool {
		call, isCall := node.(*ast.CallExpr)
		if !isCall || len(call.Args) <= customHeaderArgumentIndex {
			return true
		}

		selector, isSelector := call.Fun.(*ast.SelectorExpr)
		if !isSelector || (selector.Sel.Name != fullURLJSONHelperName && selector.Sel.Name != fullURLRequestHelperName) {
			return true
		}

		if variable, isIdent := call.Args[customHeaderArgumentIndex].(*ast.Ident); isIdent && variable.Name != "nil" {
			analysis.headerAliases[variable.Name] = true
		}

		return true
	})

	for changed := true; changed; {
		changed = false

		for name, assignments := range analysis.assignments {
			for _, assignment := range assignments {
				alias, isIdent := assignment.value.(*ast.Ident)
				if !isIdent {
					continue
				}

				if analysis.headerAliases[name] && !analysis.headerAliases[alias.Name] {
					analysis.headerAliases[alias.Name] = true
					changed = true
				}

				if analysis.headerAliases[alias.Name] && !analysis.headerAliases[name] {
					analysis.headerAliases[name] = true
					changed = true
				}
			}
		}
	}
}

func analyzeWireDeclarations(declaration *ast.ValueSpec, assignments map[string][]routeAssignment) {
	for index, value := range declaration.Values {
		if index < len(declaration.Names) {
			name := declaration.Names[index].Name
			assignments[name] = append(assignments[name], routeAssignment{
				value: value, position: declaration.Pos(), op: token.ASSIGN, object: declaration.Names[index].Obj,
				scopeEnd: token.NoPos,
			})
		}
	}
}

func analyzeHeaderDeclarations(declaration *ast.ValueSpec, aliases map[string]bool) {
	for i, value := range declaration.Values {
		if i < len(declaration.Names) && isHeaderExpression(value, aliases) {
			aliases[declaration.Names[i].Name] = true
		}
	}
}

func analyzeQueryDeclarations(declaration *ast.ValueSpec, aliases map[any]bool) {
	for index, value := range declaration.Values {
		if index < len(declaration.Names) && isQueryValuesExpression(value, aliases) {
			aliases[declaration.Names[index].Obj] = true
		}
	}
}

func analyzeWireAssignments(assignment *ast.AssignStmt, analysis *wireFunctionAnalysis, scopeEnd token.Pos) {
	for i, right := range assignment.Rhs {
		if i >= len(assignment.Lhs) {
			continue
		}

		name, ok := assignment.Lhs[i].(*ast.Ident)
		if !ok {
			continue
		}

		analysis.assignments[name.Name] = append(analysis.assignments[name.Name], routeAssignment{
			value:    right,
			position: assignment.Pos(),
			op:       assignment.Tok,
			object:   name.Obj,
			scopeEnd: scopeEnd,
		})
		if isHeaderExpression(right, analysis.headerAliases) {
			analysis.headerAliases[name.Name] = true
		}

		if isQueryValuesExpression(right, analysis.queryValues) {
			analysis.queryValues[name.Obj] = true
		}
	}
}

func (checker *wireCallsiteChecker) checkFunction(decl *ast.FuncDecl, analysis wireFunctionAnalysis) {
	if decl.Name.Name == "setFullURLAuthentication" {
		analysis.headerAliases["headers"] = true
	}

	checker.checkSchemaMapMethodValues(decl.Body, analysis)

	ast.Inspect(decl.Body, func(node ast.Node) bool {
		checker.checkIndex(node, analysis)
		checker.checkQueryComposite(node)

		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		checker.checkRequestCall(decl, call, analysis)
		checker.checkParameterCall(decl, call, analysis)
		checker.checkQueryMapEscape(call, analysis)

		if decl.Name.Name != fullURLJSONHelperName {
			checker.checkCustomHeaderArgument(call, analysis)
		}

		checker.checkCustomHeaderEscape(call, analysis)

		return true
	})
}

func (checker *wireCallsiteChecker) checkSchemaMapMethodValues(body ast.Node, analysis wireFunctionAnalysis) {
	var parents []ast.Node

	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil {
			parents = parents[:len(parents)-1]

			return true
		}

		if selector, ok := node.(*ast.SelectorExpr); ok && isSchemaMapMethod(selector, analysis) {
			var direct bool

			if len(parents) != 0 {
				call, ok := parents[len(parents)-1].(*ast.CallExpr)
				direct = ok && call.Fun == selector
			}

			if !direct {
				*checker.violations = append(*checker.violations, fmt.Sprintf(
					"%s: schema-keyed map method value must not be aliased",
					checker.fset.Position(selector.Pos()),
				))
			}
		}

		parents = append(parents, node)

		return true
	})
}

func isSchemaMapMethod(selector *ast.SelectorExpr, analysis wireFunctionAnalysis) bool {
	switch selector.Sel.Name {
	case querySetMethodName, queryAddMethodName, "Del", "Get", "Has", queryEncodeMethodName, "Clone", "Values":
	default:
		return false
	}

	if receiver, ok := unparenthesizedIdent(selector.X); ok {
		return analysis.queryValues[receiver.Obj] || analysis.headerAliases[receiver.Name]
	}

	return isHeaderExpression(selector.X, analysis.headerAliases)
}

func (checker *wireCallsiteChecker) checkQueryMapEscape(call *ast.CallExpr, analysis wireFunctionAnalysis) {
	if builtin, ok := call.Fun.(*ast.Ident); ok && builtin.Name == "len" && builtin.Obj == nil && len(call.Args) == 1 {
		return
	}

	for _, argument := range call.Args {
		if containsQueryMap(argument, analysis.queryValues) {
			checker.addParameterViolation(call, "query parameter map must not escape to an unverified helper")

			return
		}
	}

	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector || !containsQueryMap(selector.X, analysis.queryValues) {
		return
	}

	switch selector.Sel.Name {
	case querySetMethodName, queryAddMethodName, queryEncodeMethodName:
		return
	default:
		checker.addParameterViolation(call, "query parameter map must not escape to an unverified helper")
	}
}

func containsQueryMap(value ast.Expr, aliases map[any]bool) bool {
	found := false

	ast.Inspect(value, func(node ast.Node) bool {
		if variable, ok := node.(*ast.Ident); ok && variable.Obj != nil && aliases[variable.Obj] {
			found = true
		}

		return !found
	})

	return found
}

func (checker *wireCallsiteChecker) checkTransportMethodValues(body ast.Node) {
	directCalls := make(map[token.Pos]bool)

	ast.Inspect(body, func(node ast.Node) bool {
		call, isCall := node.(*ast.CallExpr)
		if isCall {
			if selector, isSelector := call.Fun.(*ast.SelectorExpr); isSelector {
				directCalls[selector.Pos()] = true
			}
		}

		return true
	})

	ast.Inspect(body, func(node ast.Node) bool {
		selector, isSelector := node.(*ast.SelectorExpr)
		if isSelector && isTransportHelper(selector.Sel.Name) && !directCalls[selector.Pos()] {
			*checker.violations = append(*checker.violations, fmt.Sprintf(
				"%s: transport helper method value must not be aliased",
				checker.fset.Position(selector.Pos()),
			))
		}

		return true
	})
}

func (checker *wireCallsiteChecker) checkCustomHeaderArgument(call *ast.CallExpr, analysis wireFunctionAnalysis) {
	const customHeaderArgumentIndex = 4

	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector || (selector.Sel.Name != fullURLJSONHelperName && selector.Sel.Name != fullURLRequestHelperName) ||
		len(call.Args) <= customHeaderArgumentIndex {
		return
	}

	checker.checkHeaderMapExpression(call.Args[customHeaderArgumentIndex], analysis.assignments,
		analysis.addressEscapes, call.Pos())
}

func (checker *wireCallsiteChecker) checkCustomHeaderEscape(call *ast.CallExpr, analysis wireFunctionAnalysis) {
	const customHeaderArgumentIndex = 4

	selector, isSelector := call.Fun.(*ast.SelectorExpr)

	for index, argument := range call.Args {
		variable, isIdent := argument.(*ast.Ident)
		if !isIdent || !analysis.headerAliases[variable.Name] {
			continue
		}

		if isSelector && index == customHeaderArgumentIndex &&
			(selector.Sel.Name == fullURLJSONHelperName || selector.Sel.Name == fullURLRequestHelperName) {
			continue
		}

		*checker.violations = append(*checker.violations, fmt.Sprintf(
			"%s: custom header map must not escape to an unverified helper",
			checker.fset.Position(argument.Pos()),
		))
	}
}

func (checker *wireCallsiteChecker) checkHeaderMapExpression(
	expr ast.Expr,
	assignments map[string][]routeAssignment,
	addressEscapes map[any][]token.Pos,
	before token.Pos,
) {
	switch value := expr.(type) {
	case *ast.Ident:
		if value.Name == "nil" {
			return
		}

		if addressEscapedBefore(value.Obj, before, addressEscapes) {
			checker.addUnverifiedHeaderMap(value.Pos())

			return
		}

		latest, found := latestScopedRouteAssignment(assignments[value.Name], before, value.Obj)
		if found {
			if latest.scopeEnd != token.NoPos && before > latest.scopeEnd {
				checker.addUnverifiedHeaderMap(value.Pos())

				return
			}

			checker.checkHeaderMapExpression(latest.value, assignments, addressEscapes, latest.position)

			return
		}

		checker.addUnverifiedHeaderMap(value.Pos())
	case *ast.ParenExpr:
		checker.checkHeaderMapExpression(value.X, assignments, addressEscapes, before)
	case *ast.CompositeLit:
		for _, element := range value.Elts {
			pair, hasKey := element.(*ast.KeyValueExpr)
			if hasKey && !generatedParameter(pair.Key, requestHeaderFieldName) {
				*checker.violations = append(*checker.violations, fmt.Sprintf(
					"%s: custom request header map key must use a schema-generated Header constant",
					checker.fset.Position(pair.Key.Pos()),
				))
			}
		}
	case *ast.CallExpr:
		if isIdentifier(value.Fun, "make") {
			return
		}

		checker.addUnverifiedHeaderMap(value.Pos())
	default:
		checker.addUnverifiedHeaderMap(value.Pos())
	}
}

func (checker *wireCallsiteChecker) addUnverifiedHeaderMap(position token.Pos) {
	*checker.violations = append(*checker.violations, fmt.Sprintf(
		"%s: custom header map must be constructed from generated Header keys at this call site",
		checker.fset.Position(position),
	))
}

func (checker *wireCallsiteChecker) checkQueryComposite(node ast.Node) {
	composite, isComposite := node.(*ast.CompositeLit)
	if !isComposite || !isURLValuesType(composite.Type) {
		return
	}

	for _, element := range composite.Elts {
		pair, hasKey := element.(*ast.KeyValueExpr)
		if hasKey && !generatedParameter(pair.Key, "QueryParam") {
			*checker.violations = append(*checker.violations, fmt.Sprintf(
				"%s: query parameter map key must use a schema-generated QueryParam constant",
				checker.fset.Position(pair.Key.Pos()),
			))
		}
	}
}

func isURLValuesType(expr ast.Expr) bool {
	selector, isSelector := expr.(*ast.SelectorExpr)

	return isSelector && isPackageQualifier(selector.X, "url") && selector.Sel.Name == "Values"
}

func (checker *wireCallsiteChecker) checkIndex(node ast.Node, analysis wireFunctionAnalysis) {
	index, ok := node.(*ast.IndexExpr)
	if !ok {
		return
	}

	receiver := unparenthesizedExpr(index.X)

	if isIdentifier(receiver, "customHeaders") && !generatedParameter(index.Index, requestHeaderFieldName) {
		*checker.violations = append(*checker.violations, fmt.Sprintf(
			"%s: custom request header name must use a schema-generated Header constant",
			checker.fset.Position(index.Pos()),
		))

		return
	}

	if isHeaderExpression(receiver, analysis.headerAliases) && !generatedParameter(index.Index, requestHeaderFieldName) {
		*checker.violations = append(*checker.violations, fmt.Sprintf(
			"%s: request header map key must use a schema-generated Header constant",
			checker.fset.Position(index.Pos()),
		))
	}

	if variable, isIdent := receiver.(*ast.Ident); isIdent && analysis.queryValues[variable.Obj] &&
		!generatedParameter(index.Index, "QueryParam") {
		*checker.violations = append(*checker.violations, fmt.Sprintf(
			"%s: query parameter map key must use a schema-generated QueryParam constant",
			checker.fset.Position(index.Pos()),
		))
	}
}

func (checker *wireCallsiteChecker) checkRequestCall(
	decl *ast.FuncDecl,
	call *ast.CallExpr,
	analysis wireFunctionAnalysis,
) {
	method, target, isRequest := requestTarget(call)
	if !isRequest || isTransportHelper(decl.Name.Name) {
		return
	}

	operation := generatedMethod(method)

	_, knownMethod := checker.methods[operation]
	if operation == "" || !knownMethod {
		*checker.violations = append(*checker.violations, fmt.Sprintf(
			"%s: outbound request method must be a generated OpenAPI operation",
			checker.fset.Position(call.Pos()),
		))

		return
	}

	found := routeNames(target, analysis.assignments, checker.aliases, analysis.queryValues, analysis.addressEscapes,
		checker.methods, checker.serverTemplateArity, receiverObject(decl), call.Pos())

	expected := operation
	if operation == "OpenDirectiveStream" {
		expected = "channel:directives"
	}

	if len(found) != 1 || !found[expected] {
		*checker.violations = append(*checker.violations, fmt.Sprintf(
			"%s: method %s is not paired with its generated route or channel at this request call",
			checker.fset.Position(call.Pos()), operation,
		))
	}
}

func receiverObject(decl *ast.FuncDecl) any {
	if decl.Recv == nil || len(decl.Recv.List) == 0 || len(decl.Recv.List[0].Names) == 0 {
		return nil
	}

	return decl.Recv.List[0].Names[0].Obj
}

func (checker *wireCallsiteChecker) checkParameterCall(
	decl *ast.FuncDecl,
	call *ast.CallExpr,
	analysis wireFunctionAnalysis,
) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || len(call.Args) == 0 || (selector.Sel.Name != querySetMethodName && selector.Sel.Name != queryAddMethodName) {
		return
	}

	if receiver, ok := unparenthesizedIdent(selector.X); ok && analysis.queryValues[receiver.Obj] {
		if !generatedParameter(call.Args[0], "QueryParam") {
			checker.addParameterViolation(call, "query parameter key must use a schema-generated QueryParam constant")
		}
	}

	if isHeaderExpression(selector.X, analysis.headerAliases) {
		allowedDynamicHeader := decl.Name.Name == fullURLRequestHelperName && isIdentifier(call.Args[0], "key")
		if !generatedParameter(call.Args[0], requestHeaderFieldName) && !allowedDynamicHeader {
			checker.addParameterViolation(call, "request header name must use a schema-generated Header constant")
		}
	}
}

func (checker *wireCallsiteChecker) addParameterViolation(call *ast.CallExpr, message string) {
	*checker.violations = append(*checker.violations, fmt.Sprintf("%s: %s", checker.fset.Position(call.Pos()), message))
}

func isHeaderExpression(expr ast.Expr, aliases map[string]bool) bool {
	expr = unparenthesizedExpr(expr)
	if call, ok := expr.(*ast.CallExpr); ok && len(call.Args) == 1 {
		return isHeaderExpression(call.Args[0], aliases)
	}

	if selector, ok := expr.(*ast.SelectorExpr); ok {
		return selector.Sel.Name == requestHeaderFieldName
	}

	if ident, ok := expr.(*ast.Ident); ok {
		return aliases[ident.Name]
	}

	return false
}

// The only outbound send sites are these request constructors followed by an
// injected Client.Do in the same function. Adding another network edge requires
// an explicit inventory update and a schema-bound method/route check above.
var allowedNetworkFunctions = map[string]bool{
	"pkg/alexa/http2.go#Connect":                                   true,
	"pkg/alexa/http2.go#ping":                                      true,
	"pkg/dependencies/graphql/client.go#Execute":                   true,
	"pkg/dependencies/graphql/queries.go#MakeRequest":              true,
	"pkg/dependencies/rest/auth.go#ExchangeRefreshTokenForCookies": true,
	"pkg/dependencies/rest/client.go#fetchCSRFTokenFromAPI":        true,
	"pkg/dependencies/rest/client.go#doRequest":                    true,
	"pkg/dependencies/rest/client.go#doRequestWithFullURL":         true,
	"pkg/dependencies/rest/client.go#doUnauthenticatedRequest":     true,
}

var injectedClientReceivers = map[string]string{
	"pkg/alexa/http2.go#Connect":                                   "c.client",
	"pkg/alexa/http2.go#ping":                                      "c.client",
	"pkg/dependencies/graphql/client.go#Execute":                   "c.httpClient",
	"pkg/dependencies/graphql/queries.go#MakeRequest":              "a.httpClient",
	"pkg/dependencies/rest/auth.go#ExchangeRefreshTokenForCookies": "c.httpClient",
	"pkg/dependencies/rest/client.go#fetchCSRFTokenFromAPI":        "c.httpClient",
	"pkg/dependencies/rest/client.go#doRequest":                    "c.httpClient",
	"pkg/dependencies/rest/client.go#doRequestWithFullURL":         "c.httpClient",
	"pkg/dependencies/rest/client.go#doUnauthenticatedRequest":     "c.httpClient",
}

var allowedNetworkImports = map[string]map[string]bool{
	"pkg/alexa/client.go":                 {"net/http": true},
	"pkg/alexa/interface.go":              {"net/http": true},
	"pkg/alexa/http2.go":                  {"net/http": true, "golang.org/x/net/http2": true},
	"pkg/alexaapimodels/errors.go":        {"net/http": true},
	"pkg/dependencies/graphql/client.go":  {"net/http": true},
	"pkg/dependencies/graphql/queries.go": {"net/http": true},
	"pkg/dependencies/rest/auth.go":       {"net/http": true},
	"pkg/dependencies/rest/client.go":     {"net/http": true, "net/http/cookiejar": true},
}

func isNetworkImport(path string) bool {
	lower := strings.ToLower(path)

	return (strings.HasPrefix(path, "net/") && path != "net/url") || path == "net" ||
		strings.HasPrefix(path, "golang.org/x/net/") || strings.Contains(lower, "websocket") ||
		strings.Contains(lower, "mqtt") || strings.Contains(lower, "grpc") || strings.Contains(lower, "quic")
}

func networkPrimitive(call *ast.CallExpr) string {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	return networkPrimitiveSelector(selector)
}

func networkPrimitiveSelector(selector *ast.SelectorExpr) string {
	switch selector.Sel.Name {
	case "Get", "Post", "PostForm", "Head", "Do", "RoundTrip", "Dial", "DialContext", "DialTLS", "Upgrade", "NewRequest", newRequestWithContextName:
		return selector.Sel.Name
	default:
		return ""
	}
}

type networkSend struct {
	variable string
	position token.Pos
}

type networkFunctionScan struct {
	fset               *token.FileSet
	functionKey        string
	receiver           any
	allowed            bool
	constructors       map[string]token.Pos
	requestAssignments map[string][]token.Pos
	requestAliases     map[string][]string
	requestEscapes     map[string][]token.Pos
	sends              []networkSend
	seen               map[string]int
	violations         *[]string
}

func checkNetworkInventory(file *ast.File, fset *token.FileSet, path string, violations *[]string) {
	checkNetworkImports(file, fset, path, violations)
	checkNetworkMethodValues(file, fset, violations)

	for _, declaration := range file.Decls {
		decl, ok := declaration.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			continue
		}

		checkNetworkFunction(decl, fset, path, violations)
	}
}

func checkNetworkImports(file *ast.File, fset *token.FileSet, path string, violations *[]string) {
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err == nil && isNetworkImport(importPath) && !allowedNetworkImports[path][importPath] {
			*violations = append(*violations, fmt.Sprintf("%s: unregistered network import %q", fset.Position(spec.Pos()), importPath))
		}
	}
}

func checkNetworkFunction(decl *ast.FuncDecl, fset *token.FileSet, path string, violations *[]string) {
	scan := networkFunctionScan{
		fset:               fset,
		functionKey:        path + "#" + decl.Name.Name,
		receiver:           receiverObject(decl),
		allowed:            allowedNetworkFunctions[path+"#"+decl.Name.Name],
		constructors:       make(map[string]token.Pos),
		requestAssignments: make(map[string][]token.Pos),
		requestAliases:     make(map[string][]string),
		requestEscapes:     make(map[string][]token.Pos),
		sends:              nil,
		seen:               make(map[string]int),
		violations:         violations,
	}

	ast.Inspect(decl.Body, func(node ast.Node) bool {
		scan.inspectNode(node)

		return true
	})

	for _, send := range scan.sends {
		scan.checkRequestSend(send)
	}

	scan.checkRequiredCalls(decl.Pos())
}

func checkNetworkMethodValues(body ast.Node, fset *token.FileSet, violations *[]string) {
	var parents []ast.Node

	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil {
			parents = parents[:len(parents)-1]

			return true
		}

		if selector, ok := node.(*ast.SelectorExpr); ok && networkPrimitiveSelector(selector) != "" {
			var direct bool

			if len(parents) != 0 {
				call, ok := parents[len(parents)-1].(*ast.CallExpr)
				direct = ok && call.Fun == selector
			}

			if !direct {
				*violations = append(*violations, fmt.Sprintf("%s: outbound network method value must be called directly", fset.Position(selector.Pos())))
			}
		}

		parents = append(parents, node)

		return true
	})
}

func (scan *networkFunctionScan) inspectNode(node ast.Node) {
	scan.recordRequestAssignments(node)

	call, ok := node.(*ast.CallExpr)
	if !ok {
		return
	}

	primitive := networkPrimitive(call)
	if primitive != "Do" {
		if selector, ok := call.Fun.(*ast.SelectorExpr); ok && !isDirectRequestHeaderMethod(selector) {
			for _, variable := range routeMutableArgumentSources(selector.X) {
				scan.requestEscapes[variable] = append(scan.requestEscapes[variable], call.Pos())
			}
		}

		for _, argument := range call.Args {
			if scan.verifiedHeaderHelperArgument(call, argument) {
				continue
			}

			for _, variable := range routeMutableArgumentSources(argument) {
				scan.requestEscapes[variable] = append(scan.requestEscapes[variable], call.Pos())
			}
		}
	}

	if primitive == "" {
		return
	}

	scan.seen[primitive]++
	if !scan.allowed || (primitive != newRequestWithContextName && primitive != "Do") || scan.seen[primitive] > 1 {
		scan.addViolation(call.Pos(), "unregistered outbound network primitive "+primitive)
	}

	if primitive == newRequestWithContextName {
		scan.checkRequestConstructor(call)
	}

	if primitive == "Do" {
		scan.recordNetworkSend(call)
	}
}

func (scan *networkFunctionScan) verifiedHeaderHelperArgument(call *ast.CallExpr, argument ast.Expr) bool {
	if scan.functionKey != "pkg/dependencies/rest/client.go#doRequestWithFullURL" {
		return false
	}

	helper, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector || helper.Sel.Name != "setFullURLAuthentication" {
		return false
	}

	header, isHeader := unparenthesizedExpr(argument).(*ast.SelectorExpr)

	return isHeader && header.Sel.Name == requestHeaderFieldName && rootIdentifier(header.X) == "req"
}

func isDirectRequestHeaderMethod(selector *ast.SelectorExpr) bool {
	header, ok := unparenthesizedExpr(selector.X).(*ast.SelectorExpr)

	return ok && header.Sel.Name == requestHeaderFieldName
}

func routeMutableArgumentSources(expr ast.Expr) []string {
	if variable := rootIdentifier(expr); variable != "" {
		return []string{variable}
	}

	return requestAliasSources(expr)
}

func (scan *networkFunctionScan) recordRequestAssignments(node ast.Node) {
	if declaration, ok := node.(*ast.ValueSpec); ok {
		for index, variable := range declaration.Names {
			scan.requestAssignments[variable.Name] = append(scan.requestAssignments[variable.Name], declaration.Pos())

			if index < len(declaration.Values) {
				scan.requestAliases[variable.Name] = append(
					scan.requestAliases[variable.Name], requestAliasSources(declaration.Values[index])...,
				)
			}
		}

		return
	}

	assignment, ok := node.(*ast.AssignStmt)
	if !ok {
		return
	}

	for _, target := range assignment.Lhs {
		if variable := rootIdentifier(target); variable != "" {
			scan.requestAssignments[variable] = append(scan.requestAssignments[variable], assignment.Pos())
		}
	}

	for index, value := range assignment.Rhs {
		if index < len(assignment.Lhs) {
			if variable := rootIdentifier(assignment.Lhs[index]); variable != "" && scan.constructors[variable] == token.NoPos {
				scan.requestAliases[variable] = append(scan.requestAliases[variable], requestAliasSources(value)...)
			}
		}

		call, isCall := value.(*ast.CallExpr)
		if !isCall || networkPrimitive(call) != newRequestWithContextName || index >= len(assignment.Lhs) {
			continue
		}

		if variable, isIdent := assignment.Lhs[index].(*ast.Ident); isIdent {
			scan.constructors[variable.Name] = call.Pos()
		}
	}
}

// A request or its URL can be changed through a local alias after construction.
// Reject creating such aliases instead of trying to prove every later use safe.
func requestAliasSources(expr ast.Expr) []string {
	switch value := expr.(type) {
	case *ast.Ident:
		return []string{value.Name}
	case *ast.SelectorExpr:
		if value.Sel.Name == "URL" || value.Sel.Name == requestHeaderFieldName {
			return []string{rootIdentifier(value.X)}
		}
	case *ast.CallExpr:
		selector, isSelector := value.Fun.(*ast.SelectorExpr)
		if isSelector && isIdentifier(selector.X, "http") && selector.Sel.Name == requestHeaderFieldName && len(value.Args) == 1 {
			return requestAliasSources(value.Args[0])
		}
	case *ast.UnaryExpr:
		return requestAliasSources(value.X)
	case *ast.ParenExpr:
		return requestAliasSources(value.X)
	case *ast.KeyValueExpr:
		return requestAliasSources(value.Value)
	case *ast.CompositeLit:
		var sources []string

		for _, element := range value.Elts {
			sources = append(sources, requestAliasSources(element)...)
		}

		return sources
	}

	return nil
}

func rootIdentifier(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return rootIdentifier(value.X)
	case *ast.StarExpr:
		return rootIdentifier(value.X)
	case *ast.ParenExpr:
		return rootIdentifier(value.X)
	case *ast.IndexExpr:
		return rootIdentifier(value.X)
	}

	return ""
}

func (scan *networkFunctionScan) checkRequestConstructor(call *ast.CallExpr) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || isPackageQualifier(selector.X, "http") {
		return
	}

	scan.addViolation(call.Pos(), "request constructor must be http.NewRequestWithContext")
}

func (scan *networkFunctionScan) recordNetworkSend(call *ast.CallExpr) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	clientName, owner := injectedClientReceiver(selector.X)
	if clientName != injectedClientReceivers[scan.functionKey] || owner != scan.receiver {
		scan.addViolation(call.Pos(), "Client.Do must use the inventoried injected client")
	}

	variable := ""

	if len(call.Args) == 1 {
		if ident, ok := call.Args[0].(*ast.Ident); ok {
			variable = ident.Name
		}
	}

	scan.sends = append(scan.sends, networkSend{variable: variable, position: call.Pos()})
}

func injectedClientReceiver(expr ast.Expr) (string, any) {
	field, isSelector := expr.(*ast.SelectorExpr)
	if !isSelector {
		return "", nil
	}

	owner, isIdentifier := field.X.(*ast.Ident)
	if !isIdentifier {
		return "", nil
	}

	return owner.Name + "." + field.Sel.Name, owner.Obj
}

func (scan *networkFunctionScan) checkRequestSend(send networkSend) {
	constructor := scan.constructors[send.variable]
	if constructor == token.NoPos || constructor >= send.position {
		scan.addViolation(send.position, "Client.Do must send a request constructed in this function")

		return
	}

	for variable, positions := range scan.requestAssignments {
		if !scan.isRequestAlias(variable, send.variable, make(map[string]bool)) {
			continue
		}

		for _, position := range positions {
			if position > constructor && position < send.position {
				scan.addViolation(send.position, "Client.Do request was reassigned or mutated after its schema-bound constructor")

				return
			}
		}
	}

	for variable, positions := range scan.requestEscapes {
		if !scan.isRequestAlias(variable, send.variable, make(map[string]bool)) {
			continue
		}

		for _, position := range positions {
			if position > constructor && position < send.position {
				scan.addViolation(send.position, "Client.Do request or URL escaped to a helper after its schema-bound constructor")

				return
			}
		}
	}
}

func (scan *networkFunctionScan) isRequestAlias(variable, request string, seen map[string]bool) bool {
	if variable == request {
		return true
	}

	if seen[variable] {
		return false
	}

	seen[variable] = true

	for _, owner := range scan.requestAliases[variable] {
		if scan.isRequestAlias(owner, request, seen) {
			return true
		}
	}

	return false
}

func (scan *networkFunctionScan) checkRequiredCalls(position token.Pos) {
	if scan.allowed && (scan.seen[newRequestWithContextName] != 1 || scan.seen["Do"] != 1) {
		scan.addViolation(position, fmt.Sprintf(
			"inventoried network function %s must have one request constructor and one send",
			scan.functionKey,
		))
	}
}

func (scan *networkFunctionScan) addViolation(position token.Pos, message string) {
	*scan.violations = append(*scan.violations, fmt.Sprintf("%s: %s", scan.fset.Position(position), message))
}

func isTransportHelper(name string) bool {
	switch name {
	case "doRequest", "doJSONRequest", fullURLRequestHelperName, fullURLJSONHelperName,
		"doUnauthenticatedRequest", "doUnauthenticatedJSONRequest":
		return true
	default:
		return false
	}
}

func isIdentifier(expr ast.Expr, name string) bool {
	identifier, ok := expr.(*ast.Ident)

	return ok && identifier.Name == name
}

func isPackageQualifier(expr ast.Expr, name string) bool {
	identifier, ok := expr.(*ast.Ident)

	return ok && identifier.Name == name && identifier.Obj == nil
}

func generatedParameter(expr ast.Expr, prefix string) bool {
	selector, isSelector := expr.(*ast.SelectorExpr)
	if !isSelector || !strings.HasPrefix(selector.Sel.Name, prefix) {
		return false
	}

	qualifier, isQualifier := selector.X.(*ast.Ident)

	return isQualifier && qualifier.Obj == nil &&
		(qualifier.Name == apiRoutesPackage || qualifier.Name == "alexamodels")
}

func generatedMethod(expr ast.Expr) string {
	selector, isSelector := expr.(*ast.SelectorExpr)
	if !isSelector || !strings.HasPrefix(selector.Sel.Name, "Method") {
		return ""
	}

	qualifier, isQualifier := selector.X.(*ast.Ident)
	if !isQualifier || qualifier.Name != apiRoutesPackage || qualifier.Obj != nil {
		return ""
	}

	return strings.TrimPrefix(selector.Sel.Name, "Method")
}

func requestTarget(call *ast.CallExpr) (ast.Expr, ast.Expr, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, nil, false
	}

	switch selector.Sel.Name {
	case newRequestWithContextName, "doRequest", "doJSONRequest", fullURLRequestHelperName,
		fullURLJSONHelperName, "doUnauthenticatedRequest", "doUnauthenticatedJSONRequest":
		if len(call.Args) >= requestWithContextArgumentCount {
			return call.Args[1], call.Args[2], true
		}
	case "NewRequest":
		if len(call.Args) >= requestArgumentCount {
			return call.Args[0], call.Args[1], true
		}
	}

	return nil, nil, false
}

type routeAssignment struct {
	value    ast.Expr
	position token.Pos
	op       token.Token
	object   any
	scopeEnd token.Pos
}

func routeNames(
	expr ast.Expr,
	assignments map[string][]routeAssignment,
	aliases map[string]string,
	queryValues map[any]bool,
	addressEscapes map[any][]token.Pos,
	routeTemplates map[string]string,
	serverTemplateArity int,
	receiver any,
	before token.Pos,
) map[string]bool {
	scan := routeNameScan{
		assignments:         assignments,
		aliases:             aliases,
		queryValues:         queryValues,
		addressEscapes:      addressEscapes,
		routeTemplates:      routeTemplates,
		serverTemplateArity: serverTemplateArity,
		receiver:            receiver,
	}

	name := scan.expression(expr, before)
	if name == "" {
		return nil
	}

	return map[string]bool{name: true}
}

type routeNameScan struct {
	assignments         map[string][]routeAssignment
	aliases             map[string]string
	queryValues         map[any]bool
	addressEscapes      map[any][]token.Pos
	routeTemplates      map[string]string
	serverTemplateArity int
	receiver            any
}

func (scan *routeNameScan) expression(value ast.Expr, cutoff token.Pos) string {
	switch part := value.(type) {
	case *ast.Ident:
		if scan.routeAddressEscaped(part, cutoff) {
			return ""
		}

		latest, ok := latestScopedRouteAssignment(scan.assignments[part.Name], cutoff, part.Obj)
		if !ok {
			return ""
		}

		if latest.op == token.ADD_ASSIGN {
			if scan.querySuffix(latest.value) {
				return scan.expression(part, latest.position)
			}

			return ""
		}

		if latest.scopeEnd != token.NoPos && cutoff > latest.scopeEnd ||
			(latest.op != token.ASSIGN && latest.op != token.DEFINE) {
			return ""
		}

		return scan.expression(latest.value, latest.position)
	case *ast.SelectorExpr:
		return scan.selectorName(part)
	case *ast.ParenExpr:
		return scan.expression(part.X, cutoff)
	case *ast.BinaryExpr:
		if part.Op != token.ADD {
			return ""
		}

		left := scan.expression(part.X, cutoff)

		right := scan.expression(part.Y, cutoff)
		if left == "" {
			if right != "" && scan.routePrefix(part.X) {
				return right
			}

			return ""
		}

		if right == "" && scan.querySuffix(part.Y) {
			return left
		}
	case *ast.CallExpr:
		return scan.callName(part, cutoff)
	}

	return ""
}

func (scan *routeNameScan) routeAddressEscaped(variable *ast.Ident, before token.Pos) bool {
	return addressEscapedBefore(variable.Obj, before, scan.addressEscapes)
}

func addressEscapedBefore(object any, before token.Pos, addressEscapes map[any][]token.Pos) bool {
	for _, address := range addressEscapes[object] {
		if address < before {
			return true
		}
	}

	return false
}

func latestScopedRouteAssignment(assignments []routeAssignment, cutoff token.Pos, object any) (routeAssignment, bool) {
	var latest routeAssignment

	for _, candidate := range assignments {
		if object != nil && candidate.object != object {
			continue
		}

		if candidate.position < cutoff && (latest.position == token.NoPos || candidate.position > latest.position) {
			latest = candidate
		}
	}

	return latest, latest.position != token.NoPos
}

func (scan *routeNameScan) selectorName(selector *ast.SelectorExpr) string {
	qualifier, ok := selector.X.(*ast.Ident)
	if !ok || qualifier.Obj != nil {
		return ""
	}

	switch qualifier.Name {
	case apiRoutesPackage:
		if strings.HasPrefix(selector.Sel.Name, "Path") {
			return strings.TrimPrefix(selector.Sel.Name, "Path")
		} else if selector.Sel.Name == "ChannelDirectivesAddress" {
			return "channel:directives"
		}
	case "alexamodels":
		if routeName, ok := scan.aliases[selector.Sel.Name]; ok {
			return routeName
		}
	}

	return ""
}

func (scan *routeNameScan) callName(call *ast.CallExpr, cutoff token.Pos) string {
	if len(call.Args) == 0 {
		return ""
	}

	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector || !isPackageQualifier(selector.X, "fmt") || selector.Sel.Name != "Sprintf" {
		return ""
	}

	if name := scan.expression(call.Args[0], cutoff); name != "" {
		template, known := scan.routeTemplates[name]
		if known && strings.Count(template, "%s") == len(call.Args)-1 {
			return name
		}

		return ""
	}

	format, ok := call.Args[0].(*ast.BasicLit)
	if !ok || format.Kind != token.STRING || len(call.Args) < 2 {
		return ""
	}

	pattern, err := strconv.Unquote(format.Value)
	if err != nil || pattern != "https://%s%s" || len(call.Args) != 3 || !scan.isEventAuthority(call.Args[1]) {
		return ""
	}

	return scan.expression(call.Args[len(call.Args)-1], cutoff)
}

func (scan *routeNameScan) isEventAuthority(value ast.Expr) bool {
	selector, ok := value.(*ast.SelectorExpr)

	return ok && scan.isTrustedOwner(selector.X, "c") && selector.Sel.Name == "authority"
}

func (scan *routeNameScan) isTrustedOwner(value ast.Expr, name string) bool {
	owner, ok := value.(*ast.Ident)

	return ok && owner.Name == name && owner.Obj == scan.receiver
}

func (scan *routeNameScan) routePrefix(expr ast.Expr) bool {
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		if !scan.isTrustedOwner(selector.X, "c") && !scan.isTrustedOwner(selector.X, "a") {
			return false
		}

		switch selector.Sel.Name {
		case "alexaAmazonBaseURI", "amazonapiBaseURI", "baseURL":
			return true
		}
	}

	call, isCall := expr.(*ast.CallExpr)
	if !isCall || len(call.Args) == 0 {
		return false
	}

	formatter, isFormatter := call.Fun.(*ast.SelectorExpr)
	if !isFormatter || !isPackageQualifier(formatter.X, "fmt") || formatter.Sel.Name != "Sprintf" {
		return false
	}

	format, isSelector := call.Args[0].(*ast.SelectorExpr)

	return isSelector && isPackageQualifier(format.X, apiRoutesPackage) &&
		format.Sel.Name == "ServerExchangeRefreshTokenForCookies" && len(call.Args)-1 == scan.serverTemplateArity
}

func (scan *routeNameScan) querySuffix(expr ast.Expr) bool {
	addition, isAddition := expr.(*ast.BinaryExpr)
	if !isAddition || addition.Op != token.ADD {
		return false
	}

	separator, isLiteral := addition.X.(*ast.BasicLit)
	if !isLiteral || separator.Kind != token.STRING || separator.Value != `"?"` {
		return false
	}

	call, isCall := addition.Y.(*ast.CallExpr)
	if !isCall || len(call.Args) != 0 {
		return false
	}

	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector || selector.Sel.Name != queryEncodeMethodName {
		return false
	}

	receiver, isIdent := selector.X.(*ast.Ident)

	return isIdent && trustedQueryMapAt(receiver, call.Pos(), scan.assignments, scan.addressEscapes)
}

func trustedQueryMapAt(variable *ast.Ident, before token.Pos, assignments map[string][]routeAssignment,
	addressEscapes map[any][]token.Pos,
) bool {
	if addressEscapedBefore(variable.Obj, before, addressEscapes) {
		return false
	}

	latest, found := latestScopedRouteAssignment(assignments[variable.Name], before, variable.Obj)
	if !found || (latest.scopeEnd != token.NoPos && before > latest.scopeEnd) ||
		(latest.op != token.ASSIGN && latest.op != token.DEFINE) {
		return false
	}

	switch value := latest.value.(type) {
	case *ast.Ident:
		return trustedQueryMapAt(value, latest.position, assignments, addressEscapes)
	case *ast.ParenExpr:
		return trustedQueryMapExpression(value.X, latest.position, assignments, addressEscapes)
	default:
		return trustedQueryMapExpression(value, latest.position, assignments, addressEscapes)
	}
}

func trustedQueryMapExpression(value ast.Expr, before token.Pos, assignments map[string][]routeAssignment,
	addressEscapes map[any][]token.Pos,
) bool {
	if variable, ok := value.(*ast.Ident); ok {
		return trustedQueryMapAt(variable, before, assignments, addressEscapes)
	}

	if composite, ok := value.(*ast.CompositeLit); ok {
		if !isURLValuesType(composite.Type) {
			return false
		}

		for _, element := range composite.Elts {
			pair, isPair := element.(*ast.KeyValueExpr)
			if !isPair || !generatedParameter(pair.Key, "QueryParam") {
				return false
			}
		}

		return true
	}

	if call, ok := value.(*ast.CallExpr); ok && isIdentifier(call.Fun, "make") && len(call.Args) == 1 {
		return isURLValuesType(call.Args[0])
	}

	return false
}

func isQueryValuesExpression(value ast.Expr, aliases map[any]bool) bool {
	if variable, ok := value.(*ast.Ident); ok {
		return variable.Obj != nil && aliases[variable.Obj]
	}

	if parenthesized, ok := value.(*ast.ParenExpr); ok {
		return isQueryValuesExpression(parenthesized.X, aliases)
	}

	if composite, ok := value.(*ast.CompositeLit); ok {
		return isURLValuesType(composite.Type)
	}

	if call, ok := value.(*ast.CallExpr); ok && isIdentifier(call.Fun, "make") && len(call.Args) > 0 {
		return isURLValuesType(call.Args[0])
	}

	return false
}

func nonRouteURLLiterals(file *ast.File) map[token.Pos]bool {
	exempt := make(map[token.Pos]bool)

	ast.Inspect(file, func(node ast.Node) bool {
		call, isCall := node.(*ast.CallExpr)
		if !isCall || len(call.Args) < 2 {
			return true
		}

		selector, isSelector := call.Fun.(*ast.SelectorExpr)
		if !isSelector || selector.Sel.Name != querySetMethodName {
			return true
		}

		if !generatedParameter(call.Args[0], "HeaderReferer") {
			return true
		}

		value, ok := call.Args[1].(*ast.BasicLit)
		if ok && value.Kind == token.STRING {
			exempt[value.Pos()] = true
		}

		return true
	})

	return exempt
}

// Command apiroutes generates the HTTP route and stream framing definitions
// consumed by go-alexa from its checked-in OpenAPI and AsyncAPI documents.
package main

import (
	"bytes"
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

	if err := run(*root, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, check bool) error {
	openAPI, err := readYAML[openAPIDocument](filepath.Join(root, "api", "openapi.yaml"))
	if err != nil {
		return err
	}
	asyncAPI, err := readYAML[asyncAPIDocument](filepath.Join(root, "api", "asyncapi.yaml"))
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
		got, err := os.ReadFile(outputPath)
		if err != nil {
			return fmt.Errorf("read generated routes: %w", err)
		}
		if !bytes.Equal(got, want) {
			return fmt.Errorf("%s is stale; run make generate-api", filepath.ToSlash(outputPath))
		}
		return checkProductionRouteLiterals(root, routes)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, want, 0o644)
}

func readYAML[T any](path string) (T, error) {
	var doc T
	data, err := os.ReadFile(path)
	if err != nil {
		return doc, err
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return doc, fmt.Errorf("parse %s: %w", filepath.ToSlash(path), err)
	}
	return doc, nil
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
		return nil, fmt.Errorf("OpenAPI document contains no HTTP operations")
	}
	sort.Slice(routes, func(i, j int) bool { return routes[i].name < routes[j].name })
	return routes, nil
}

func collectParameters(doc openAPIDocument) ([]namedParameter, error) {
	seen := make(map[string]string)
	add := func(parameter apiParameter) error {
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
	for _, operations := range doc.Paths {
		for _, operation := range operations {
			if operation.OperationID == "" {
				continue
			}
			for _, parameter := range operation.Parameters {
				if parameter.Ref != "" {
					const prefix = "#/components/parameters/"
					if !strings.HasPrefix(parameter.Ref, prefix) {
						return nil, fmt.Errorf("unsupported parameter reference %q", parameter.Ref)
					}
					var ok bool
					parameter, ok = doc.Components.Parameters[strings.TrimPrefix(parameter.Ref, prefix)]
					if !ok {
						return nil, fmt.Errorf("undefined parameter reference %q", parameter.Ref)
					}
				}
				if err := add(parameter); err != nil {
					return nil, err
				}
			}
		}
	}
	for _, header := range doc.RequestHeaders {
		if err := add(apiParameter{Name: header, In: "header"}); err != nil {
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

func collectRegionalEndpoints(openAPI openAPIDocument, asyncAPI asyncAPIDocument) (regionalEndpoints, error) {
	endpoints := regionalEndpoints{}
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
		return endpoints, fmt.Errorf("exchangeRefreshTokenForCookies must declare one variable in its server URL")
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
	urls := make(map[string]string, 3)
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
		return "", framing, fmt.Errorf("AsyncAPI document is missing the directives channel")
	}
	if channel.Address == "" || !strings.HasPrefix(channel.Address, "/") {
		return "", framing, fmt.Errorf("AsyncAPI directives channel must have an absolute path address")
	}
	foundOperation := false
	for _, operation := range doc.Operations {
		if operation.Channel.Ref == "#/channels/directives" {
			foundOperation = true
			break
		}
	}
	if !foundOperation {
		return "", framing, fmt.Errorf("AsyncAPI directives channel has no operation")
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
		return "", framing, fmt.Errorf("AsyncAPI directives channel has incomplete x-go-alexa-stream-framing configuration")
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
	out.WriteString("\tLineDelimiter byte\n\tTrimWhitespace bool\n\tBoundaryPrefix string\n\tContentTypePrefix string\n\tSkipEmptyLines bool\n\tAuthenticationFailureMarker string\n")
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
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	})
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func checkProductionRouteLiterals(root string, routes []route) error {
	base := filepath.Join(root, "pkg")
	var violations []string
	aliases, err := routeAliases(filepath.Join(base, "dependencymodels", "constants.go"))
	if err != nil {
		return err
	}
	knownMethods := make(map[string]bool, len(routes))
	for _, route := range routes {
		knownMethods[route.name] = true
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
			return err
		}
		exemptLiterals := nonRouteURLLiterals(file)
		checkWireCallsites(file, fset, aliases, knownMethods, &violations)
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
		return err
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
		return nil, err
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

func checkWireCallsites(file *ast.File, fset *token.FileSet, aliases map[string]string, methods map[string]bool, violations *[]string) {
	ast.Inspect(file, func(node ast.Node) bool {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			return true
		}
		assignments := make(map[string][]routeAssignment)
		queryValues := make(map[string]bool)
		headerAliases := make(map[string]bool)
		ast.Inspect(decl.Body, func(node ast.Node) bool {
			if declaration, ok := node.(*ast.ValueSpec); ok {
				for i, value := range declaration.Values {
					if i < len(declaration.Names) && isHeaderExpression(value, headerAliases) {
						headerAliases[declaration.Names[i].Name] = true
					}
				}
			}
			assignment, ok := node.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for i, right := range assignment.Rhs {
				if i >= len(assignment.Lhs) {
					continue
				}
				if name, ok := assignment.Lhs[i].(*ast.Ident); ok {
					assignments[name.Name] = append(assignments[name.Name], routeAssignment{value: right, position: assignment.Pos(), op: assignment.Tok})
					if isHeaderExpression(right, headerAliases) {
						headerAliases[name.Name] = true
					}
					if isQueryValuesExpression(right) {
						queryValues[name.Name] = true
					}
				}
			}
			return true
		})
		ast.Inspect(decl.Body, func(node ast.Node) bool {
			if index, ok := node.(*ast.IndexExpr); ok {
				if isIdentifier(index.X, "customHeaders") && !generatedParameter(index.Index, "Header") {
					*violations = append(*violations, fmt.Sprintf("%s: custom request header name must use a schema-generated Header constant", fset.Position(index.Pos())))
				}
				if isHeaderExpression(index.X, headerAliases) && !generatedParameter(index.Index, "Header") {
					*violations = append(*violations, fmt.Sprintf("%s: request header map key must use a schema-generated Header constant", fset.Position(index.Pos())))
				}
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if method, target, isRequest := requestTarget(call); isRequest && !isTransportHelper(decl.Name.Name) {
				operation := generatedMethod(method)
				if operation == "" || !methods[operation] {
					*violations = append(*violations, fmt.Sprintf("%s: outbound request method must be a generated OpenAPI operation", fset.Position(call.Pos())))
				} else {
					found := routeNames(target, assignments, aliases, call.Pos())
					expected := operation
					if operation == "OpenDirectiveStream" {
						expected = "channel:directives"
					}
					if len(found) != 1 || !found[expected] {
						*violations = append(*violations, fmt.Sprintf("%s: method %s is not paired with its generated route or channel at this request call", fset.Position(call.Pos()), operation))
					}
				}
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || len(call.Args) == 0 || (selector.Sel.Name != "Set" && selector.Sel.Name != "Add") {
				return true
			}
			if receiver, ok := selector.X.(*ast.Ident); ok && queryValues[receiver.Name] {
				if !generatedParameter(call.Args[0], "QueryParam") {
					*violations = append(*violations, fmt.Sprintf("%s: query parameter key must use a schema-generated QueryParam constant", fset.Position(call.Pos())))
				}
			}
			if header, ok := selector.X.(*ast.SelectorExpr); ok && header.Sel.Name == "Header" {
				if !generatedParameter(call.Args[0], "Header") && !(decl.Name.Name == "doRequestWithFullURL" && isIdentifier(call.Args[0], "key")) {
					*violations = append(*violations, fmt.Sprintf("%s: request header name must use a schema-generated Header constant", fset.Position(call.Pos())))
				}
			}
			if receiver, ok := selector.X.(*ast.Ident); ok && headerAliases[receiver.Name] && !generatedParameter(call.Args[0], "Header") {
				*violations = append(*violations, fmt.Sprintf("%s: request header name must use a schema-generated Header constant", fset.Position(call.Pos())))
			}
			return true
		})
		return false
	})
}

func isHeaderExpression(expr ast.Expr, aliases map[string]bool) bool {
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		return selector.Sel.Name == "Header"
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
	switch selector.Sel.Name {
	case "Get", "Post", "PostForm", "Head", "Do", "RoundTrip", "Dial", "DialContext", "DialTLS", "Upgrade", "NewRequest", "NewRequestWithContext":
		return selector.Sel.Name
	default:
		return ""
	}
}

func checkNetworkInventory(file *ast.File, fset *token.FileSet, path string, violations *[]string) {
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err == nil && isNetworkImport(importPath) && !allowedNetworkImports[path][importPath] {
			*violations = append(*violations, fmt.Sprintf("%s: unregistered network import %q", fset.Position(spec.Pos()), importPath))
		}
	}
	for _, declaration := range file.Decls {
		decl, ok := declaration.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			continue
		}
		key := path + "#" + decl.Name.Name
		constructors := make(map[string]token.Pos)
		requestAssignments := make(map[string][]token.Pos)
		var sends []struct {
			variable string
			position token.Pos
		}
		seen := make(map[string]int)
		ast.Inspect(decl.Body, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if ok {
				for _, target := range assignment.Lhs {
					if variable, isIdent := target.(*ast.Ident); isIdent {
						requestAssignments[variable.Name] = append(requestAssignments[variable.Name], assignment.Pos())
					}
				}
				for i, value := range assignment.Rhs {
					call, isCall := value.(*ast.CallExpr)
					if isCall && networkPrimitive(call) == "NewRequestWithContext" && i < len(assignment.Lhs) {
						if variable, isIdent := assignment.Lhs[i].(*ast.Ident); isIdent {
							constructors[variable.Name] = call.Pos()
						}
					}
				}
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			primitive := networkPrimitive(call)
			if primitive == "" {
				return true
			}
			seen[primitive]++
			if !allowedNetworkFunctions[key] || (primitive != "NewRequestWithContext" && primitive != "Do") || seen[primitive] > 1 {
				*violations = append(*violations, fmt.Sprintf("%s: unregistered outbound network primitive %s", fset.Position(call.Pos()), primitive))
			}
			if primitive == "NewRequestWithContext" {
				selector := call.Fun.(*ast.SelectorExpr)
				if !isIdentifier(selector.X, "http") {
					*violations = append(*violations, fmt.Sprintf("%s: request constructor must be http.NewRequestWithContext", fset.Position(call.Pos())))
				}
			}
			if primitive == "Do" {
				selector := call.Fun.(*ast.SelectorExpr)
				receiver := ""
				if field, ok := selector.X.(*ast.SelectorExpr); ok {
					if owner, ok := field.X.(*ast.Ident); ok {
						receiver = owner.Name + "." + field.Sel.Name
					}
				}
				if receiver != injectedClientReceivers[key] {
					*violations = append(*violations, fmt.Sprintf("%s: Client.Do must use the inventoried injected client", fset.Position(call.Pos())))
				}
				variable := ""
				if len(call.Args) == 1 {
					if ident, ok := call.Args[0].(*ast.Ident); ok {
						variable = ident.Name
					}
				}
				sends = append(sends, struct {
					variable string
					position token.Pos
				}{variable, call.Pos()})
			}
			return true
		})
		for _, send := range sends {
			constructor := constructors[send.variable]
			if constructor == token.NoPos || constructor >= send.position {
				*violations = append(*violations, fmt.Sprintf("%s: Client.Do must send a request constructed in this function", fset.Position(send.position)))
			} else {
				for _, position := range requestAssignments[send.variable] {
					if position > constructor && position < send.position {
						*violations = append(*violations, fmt.Sprintf("%s: Client.Do request was reassigned after its schema-bound constructor", fset.Position(send.position)))
						break
					}
				}
			}
		}
		if allowedNetworkFunctions[key] && (seen["NewRequestWithContext"] != 1 || seen["Do"] != 1) {
			*violations = append(*violations, fmt.Sprintf("%s: inventoried network function %s must have one request constructor and one send", fset.Position(decl.Pos()), key))
		}
	}
}

func isTransportHelper(name string) bool {
	switch name {
	case "doRequest", "doRequestWithFullURL", "doUnauthenticatedRequest", "doUnauthenticatedJSONRequest":
		return true
	default:
		return false
	}
}

func isIdentifier(expr ast.Expr, name string) bool {
	identifier, ok := expr.(*ast.Ident)
	return ok && identifier.Name == name
}

func generatedParameter(expr ast.Expr, prefix string) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || !strings.HasPrefix(selector.Sel.Name, prefix) {
		return false
	}
	qualifier, ok := selector.X.(*ast.Ident)
	return ok && (qualifier.Name == "apiroutes" || qualifier.Name == "alexamodels")
}

func generatedMethod(expr ast.Expr) string {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || !strings.HasPrefix(selector.Sel.Name, "Method") {
		return ""
	}
	qualifier, ok := selector.X.(*ast.Ident)
	if !ok || qualifier.Name != "apiroutes" {
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
	case "NewRequestWithContext", "doJSONRequest", "doJSONRequestWithFullURL", "doUnauthenticatedJSONRequest":
		if len(call.Args) >= 3 {
			return call.Args[1], call.Args[2], true
		}
	case "NewRequest":
		if len(call.Args) >= 2 {
			return call.Args[0], call.Args[1], true
		}
	}
	return nil, nil, false
}

type routeAssignment struct {
	value    ast.Expr
	position token.Pos
	op       token.Token
}

func routeNames(expr ast.Expr, assignments map[string][]routeAssignment, aliases map[string]string, before token.Pos) map[string]bool {
	found := make(map[string]bool)
	var scan func(ast.Expr, token.Pos)
	scan = func(value ast.Expr, cutoff token.Pos) {
		ast.Inspect(value, func(node ast.Node) bool {
			switch part := node.(type) {
			case *ast.Ident:
				var latest *routeAssignment
				for i := range assignments[part.Name] {
					candidate := &assignments[part.Name][i]
					if candidate.position < cutoff && (latest == nil || candidate.position > latest.position) {
						latest = candidate
					}
				}
				if latest != nil {
					if latest.op == token.ADD_ASSIGN {
						scan(part, latest.position)
					}
					scan(latest.value, latest.position)
				}
			case *ast.SelectorExpr:
				qualifier, ok := part.X.(*ast.Ident)
				if !ok {
					break
				}
				if qualifier.Name == "apiroutes" {
					if strings.HasPrefix(part.Sel.Name, "Path") {
						found[strings.TrimPrefix(part.Sel.Name, "Path")] = true
					} else if part.Sel.Name == "ChannelDirectivesAddress" {
						found["channel:directives"] = true
					}
				} else if qualifier.Name == "alexamodels" {
					if routeName, ok := aliases[part.Sel.Name]; ok {
						found[routeName] = true
					}
				}
			}
			return true
		})
	}
	scan(expr, before)
	return found
}

func isQueryValuesExpression(value ast.Expr) bool {
	if composite, ok := value.(*ast.CompositeLit); ok {
		if selector, ok := composite.Type.(*ast.SelectorExpr); ok {
			if qualifier, ok := selector.X.(*ast.Ident); ok {
				return qualifier.Name == "url" && selector.Sel.Name == "Values"
			}
		}
	}
	if call, ok := value.(*ast.CallExpr); ok {
		if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
			return selector.Sel.Name == "Query" || selector.Sel.Name == "ParseQuery"
		}
	}
	return false
}

func nonRouteURLLiterals(file *ast.File) map[token.Pos]bool {
	exempt := make(map[token.Pos]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Set" {
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

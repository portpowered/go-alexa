package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	graphqlast "github.com/vektah/gqlparser/v2/ast"
)

const sdkImportPath = "github.com/portpowered/go-alexa/pkg/alexaapimodels"
const wireImportPath = "github.com/portpowered/go-alexa/pkg/dependencymodels"

type primitivePackage struct {
	ImportPath string
	Name       string
	Directory  string
}

func checkSDKGraphQLBindings(schema schemaNode, graphql *graphqlast.Schema) error {
	err := checkSDKGraphQLBindingCompleteness(schema, graphql)
	if err != nil {
		return err
	}

	for name, binding := range schema.SDKGraphQLValues {
		parts := strings.Split(binding, ".")
		if len(parts) != evidenceParts {
			return diagnosticf("SDK primitive %s has an invalid GraphQL binding", name)
		}

		value, exists := schema.CompatibilityConstants["SDK"+name]

		definition := graphql.Types[parts[0]]

		if !exists || value != parts[1] || definition == nil || definition.Kind != graphqlast.Enum ||
			definition.EnumValues.ForName(parts[1]) == nil {
			return diagnosticf("SDK primitive %s differs from its GraphQL enum binding %s", name, binding)
		}
	}

	return nil
}

func checkSDKGraphQLBindingCompleteness(schema schemaNode, graphql *graphqlast.Schema) error {
	enumName := schema.SDKType
	// The public command semantic type omits the provider event UNKNOWN state.
	if enumName == "PowerState" {
		enumName = "PowerStateValue"
	}

	definition := graphql.Types[enumName]
	if definition == nil || definition.Kind != graphqlast.Enum {
		return nil
	}

	for name, value := range schema.CompatibilityConstants {
		if !strings.HasPrefix(name, "SDK") || definition.EnumValues.ForName(value) == nil {
			continue
		}

		publicName := strings.TrimPrefix(name, "SDK")
		if schema.SDKGraphQLValues[publicName] != enumName+"."+value {
			return diagnosticf("SDK primitive %s lacks its shared GraphQL enum binding", publicName)
		}
	}

	return nil
}

// SDK and GraphQL projections share some names. Resolve SDK package imports so
// the inventory does not report uses of the GraphQL projection as SDK uses.
func findPrimitiveCallSites(names map[string]sdkPrimitive, owner primitivePackage) (map[string][]string, error) {
	usages := make(map[string][]string)

	err := filepath.WalkDir("pkg", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk SDK primitive uses: %w", walkErr)
		}

		path = filepath.ToSlash(path)
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || (path == sdkPrimitiveOutput || path == wirePrimitiveOutput) {
			return nil
		}

		set := token.NewFileSet()

		file, parseErr := parser.ParseFile(set, path, mustRead(path), 0)
		if parseErr != nil {
			return fmt.Errorf("parse SDK primitive uses %s: %w", path, parseErr)
		}

		collectPrimitiveUses(file, path, set, names, usages, owner)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("inventory SDK primitive uses: %w", err)
	}

	for name, sites := range usages {
		sort.Strings(sites)
		usages[name] = compact(sites)
	}

	return usages, nil
}

func collectPrimitiveUses(
	file *ast.File, path string, set *token.FileSet,
	names map[string]sdkPrimitive, usages map[string][]string, owner primitivePackage,
) {
	aliases := primitiveImportAliases(file, owner)
	self := filepath.ToSlash(filepath.Dir(path)) == owner.Directory

	ast.Inspect(file, func(node ast.Node) bool {
		name := ""

		switch expression := node.(type) {
		case *ast.SelectorExpr:
			identifier, isIdentifier := expression.X.(*ast.Ident)
			if isIdentifier && identifier.Obj == nil && aliases[identifier.Name] {
				name = expression.Sel.Name
			}
		case *ast.Ident:
			if self && expression.Obj == nil {
				name = expression.Name
			}
		}

		if _, exists := names[name]; exists {
			usages[name] = append(usages[name], fmt.Sprintf("%s:%d", path, set.Position(node.Pos()).Line))
		}

		return true
	})
}

func primitiveImportAliases(file *ast.File, owner primitivePackage) map[string]bool {
	aliases := make(map[string]bool)

	for _, dependency := range file.Imports {
		path, err := strconv.Unquote(dependency.Path.Value)
		if err != nil || path != owner.ImportPath {
			continue
		}

		name := owner.Name
		if dependency.Name != nil {
			name = dependency.Name.Name
		}

		aliases[name] = true
	}

	return aliases
}

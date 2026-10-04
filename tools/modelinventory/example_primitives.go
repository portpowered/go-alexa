package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

func rejectRawExampleInputPrimitives(root string, values map[string]bool) error {
	owner := primitivePackage{ImportPath: sdkImportPath, Name: "alexaapimodels", Directory: "pkg/alexaapimodels"}

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk example wire inputs: %w", walkErr)
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		set := token.NewFileSet()

		file, err := parser.ParseFile(set, path, mustRead(path), 0)
		if err != nil {
			return fmt.Errorf("parse example wire inputs: %w", err)
		}

		return checkExampleInputPrimitives(file, set, path, owner, values)
	})
	if err != nil {
		return fmt.Errorf("check example wire inputs: %w", err)
	}

	return nil
}

func checkExampleInputPrimitives(
	file *ast.File, set *token.FileSet, path string, owner primitivePackage, values map[string]bool,
) error {
	aliases := primitiveImportAliases(file, owner)
	assignments := indexWireSourceAssignments(file)

	var problems []string

	ast.Inspect(file, func(node ast.Node) bool {
		literal, isLiteral := node.(*ast.CompositeLit)
		if !isLiteral {
			return true
		}

		selector, isSelector := literal.Type.(*ast.SelectorExpr)
		if !isSelector {
			return true
		}

		identifier, isIdentifier := selector.X.(*ast.Ident)
		if !isIdentifier || identifier.Obj != nil || !aliases[identifier.Name] {
			return true
		}

		for _, element := range literal.Elts {
			if field, isField := element.(*ast.KeyValueExpr); isField {
				element = field.Value
			}

			inspectWireValueLiterals(element, make(map[wireSourceVariable]bool), func(raw *ast.BasicLit) {
				value, err := strconv.Unquote(raw.Value)
				if err == nil && values[value] {
					problems = append(problems, fmt.Sprintf("%s: example input value %q must use its generated declaration", set.Position(raw.Pos()), value))
				}
			}, assignments)
		}

		return true
	})

	if len(problems) > 0 {
		return diagnosticf("%s: %s", path, strings.Join(problems, "\n"))
	}

	return nil
}

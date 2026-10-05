package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

type wireSourceFile struct {
	path string
	file *ast.File
}

func checkGeneratedWireConstructionRoot(root string, generated map[string]bool, models map[string]generatedModel) error {
	set := token.NewFileSet()
	packages := make(map[string][]wireSourceFile)

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk wire construction sources: %w", walkErr)
		}

		path = filepath.ToSlash(path)
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || generated[path] {
			return nil
		}

		file, parseErr := parser.ParseFile(set, path, mustRead(path), 0)
		if parseErr != nil {
			return fmt.Errorf("parse wire construction source %s: %w", path, parseErr)
		}

		key := filepath.Dir(path) + "/" + file.Name.Name
		packages[key] = append(packages[key], wireSourceFile{path: path, file: file})

		return nil
	})
	if err != nil {
		return fmt.Errorf("collect wire construction sources: %w", err)
	}

	for _, files := range packages {
		err := checkWireSourcePackage(files, set, models)
		if err != nil {
			return err
		}
	}

	return nil
}

func resolveWirePackageNames(files []wireSourceFile) {
	owners := make(map[string]*ast.File)

	for _, source := range files {
		for name := range source.file.Scope.Objects {
			owners[name] = source.file
		}
	}

	for _, source := range files {
		for _, identifier := range source.file.Unresolved {
			owner := owners[identifier.Name]
			if identifier.Obj == nil && owner != nil {
				identifier.Obj = owner.Scope.Objects[identifier.Name]
			}
		}
	}
}

func checkWireSourcePackage(files []wireSourceFile, set *token.FileSet, models map[string]generatedModel) error {
	resolveWirePackageNames(files)

	assignments := wireSourceAssignments{
		values:              make(map[wireSourceVariable][]ast.Expr),
		callArguments:       make(map[wireSourceVariable][]ast.Expr),
		boundParameters:     make(map[wireSourceVariable]bool),
		generatedParameters: make(map[wireSourceVariable]bool),
		globalVariables:     make(map[wireSourceVariable]bool),
		sdkDependencies:     make(map[string]bool),
		sdkOwner:            nil,
		callableFields:      make(map[string][]ast.Expr),
		functionBodies:      make(map[*ast.FuncType]*ast.BlockStmt),
		functionDecls:       make(map[*ast.FuncType]*ast.FuncDecl),
		methods:             make(map[string][]*ast.FuncDecl),
		sourceContexts:      make(map[ast.Node]wireSourceContext),
		resultIndexes:       make(map[wireSourceVariable]map[ast.Expr]int),
		resultSelections:    make(map[*ast.CallExpr]int),
	}

	for _, source := range files {
		local := indexWireSourceAssignments(source.file, source.path)
		for node, context := range local.sourceContexts {
			assignments.sourceContexts[node] = context
		}

		for variable, indexes := range local.resultIndexes {
			if assignments.resultIndexes[variable] == nil {
				assignments.resultIndexes[variable] = make(map[ast.Expr]int)
			}

			for expression, index := range indexes {
				assignments.resultIndexes[variable][expression] = index
			}
		}

		for variable, global := range local.globalVariables {
			assignments.globalVariables[variable] = global
		}

		if local.sdkOwner != nil {
			assignments.sdkOwner = local.sdkOwner
		}

		for function, body := range local.functionBodies {
			assignments.functionBodies[function] = body
		}

		for function, declaration := range local.functionDecls {
			assignments.functionDecls[function] = declaration
		}

		for name, values := range local.callableFields {
			assignments.callableFields[name] = append(assignments.callableFields[name], values...)
		}

		for name, verified := range local.sdkDependencies {
			assignments.sdkDependencies[name] = verified
		}

		for name, methods := range local.methods {
			assignments.methods[name] = append(assignments.methods[name], methods...)
		}

		for variable, values := range local.values {
			assignments.values[variable] = append(assignments.values[variable], values...)
		}
	}

	indexWirePackageHelperParameters(files, models, assignments)
	indexWirePackageValueArguments(files, assignments)

	for _, source := range files {
		err := rejectRawGeneratedWireConstructionsWithAssignments(source.file, set, source.path, models, assignments)
		if err != nil {
			return fmt.Errorf("check generated wire construction: %w", err)
		}
	}

	return nil
}

func indexWirePackageHelperParameters(files []wireSourceFile, models map[string]generatedModel, assignments wireSourceAssignments) {
	owner := primitivePackage{ImportPath: wireImportPath, Name: "alexamodels", Directory: "pkg/dependencymodels"}

	for {
		before := len(assignments.generatedParameters) + len(assignments.values)

		for _, source := range files {
			aliases := primitiveImportAliases(source.file, owner)
			indexWireHelperParameters(source.file, aliases, source.path, models, assignments)
		}

		if before == len(assignments.generatedParameters)+len(assignments.values) {
			return
		}
	}
}

//nolint:gocognit // The ordered parameter walk binds helper calls to positional and variadic arguments.
func indexWirePackageValueArguments(files []wireSourceFile, assignments wireSourceAssignments) {
	for _, source := range files {
		ast.Inspect(source.file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			for _, function := range wireLocalHelpers(call.Fun, assignments, make(map[wireSourceVariable]bool)) {
				if function == nil || function.Params == nil {
					continue
				}

				argument := 0

				for _, field := range function.Params.List {
					variadic := false
					if _, ok := field.Type.(*ast.Ellipsis); ok {
						variadic = true
					}

					if len(field.Names) == 0 {
						argument++

						continue
					}

					for _, name := range field.Names {
						if argument >= len(call.Args) {
							break
						}

						key := wireSourceVariable{declaration: field, name: name.Name}

						if variadic {
							for _, value := range call.Args[argument:] {
								assignWireCallArgument(assignments, key, value)
							}

							argument = len(call.Args)

							break
						}

						assignWireCallArgument(assignments, key, call.Args[argument])

						argument++
					}
				}
			}

			return true
		})
	}
}

func assignWireCallArgument(assignments wireSourceAssignments, key wireSourceVariable, expression ast.Expr) {
	for _, previous := range assignments.callArguments[key] {
		if previous == expression {
			return
		}
	}

	assignments.callArguments[key] = append(assignments.callArguments[key], expression)
}

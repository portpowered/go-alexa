// Command graphqlschemacheck validates shipped operations against the captured live schema.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	const directory = "pkg/schemas/graphql"

	data, err := os.ReadFile(filepath.Join(directory, "snapshots", "live-schema.graphql"))
	if err != nil {
		return fmt.Errorf("read live schema: %w", err)
	}

	schema, err := gqlparser.LoadSchema(&ast.Source{Name: "live-schema.graphql", Input: string(data), BuiltIn: false})
	if err != nil {
		return fmt.Errorf("parse live schema: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(directory, "*.graphql"))
	if err != nil {
		return fmt.Errorf("find operation documents: %w", err)
	}

	count := 0

	for _, file := range files {
		if filepath.Base(file) == "endpoints-schema.graphql" {
			continue
		}

		//nolint:gosec // Paths come from the fixed repository schema directory glob.
		query, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}

		_, errors := gqlparser.LoadQuery(schema, string(query))
		if len(errors) > 0 {
			return fmt.Errorf("%s: %w", file, errors)
		}

		count++
	}

	_, err = fmt.Fprintf(os.Stdout, "Validated %d GraphQL documents against the captured live schema\n", count)
	if err != nil {
		return fmt.Errorf("write verification result: %w", err)
	}

	return nil
}

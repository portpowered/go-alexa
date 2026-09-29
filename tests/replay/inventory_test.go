package replay_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	replaymodels "github.com/portpowered/go-alexa/pkg/testing"
	"gopkg.in/yaml.v2"
)

func readPairs(t *testing.T, name string) []replaymodels.SyntheticExchange {
	t.Helper()

	//nolint:gosec // The filename is supplied by this test's fixed replay inventory.
	data, err := os.ReadFile(filepath.Join("fixtures", "synthetic", name))
	if err != nil {
		t.Fatal(err)
	}

	var pairs []replaymodels.SyntheticExchange
	{
		err := json.Unmarshal(data, &pairs)
		if err != nil {
			t.Fatal(err)
		}
	}

	return pairs
}

func TestSyntheticPairsCoverSchemaInventory(t *testing.T) {
	t.Parallel()

	all := append(readPairs(t, "alexa-rest.json"), readPairs(t, "alexa-graphql.json")...)
	all = append(all, readPairs(t, "alexa-events.json")...)
	covered := make(map[string]bool)

	for _, pair := range all {
		if pair.Source != "synthetic" {
			t.Fatalf("%s has non-synthetic provenance", pair.Operation)
		}

		covered[pair.Operation] = true
	}

	assertOpenAPIOperationsCovered(t, covered)
	assertGraphQLOperationsCovered(t, covered)
	assertEventReplayInventory(t)
}

func assertOpenAPIOperationsCovered(t *testing.T, covered map[string]bool) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	var api struct {
		Paths map[string]map[string]struct {
			OperationID string `yaml:"operationId"`
		} `yaml:"paths"`
	}

	{
		err := yaml.Unmarshal(data, &api)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, methods := range api.Paths {
		for _, operation := range methods {
			if operation.OperationID == "" || operation.OperationID == "openAuthorizationPage" || operation.OperationID == "executeNexusGraphQL" {
				continue
			}

			found := covered[operation.OperationID]
			for name := range covered {
				found = found || strings.HasPrefix(name, operation.OperationID+"_")
			}

			if !found {
				t.Errorf("OpenAPI operation %s has no paired replay", operation.OperationID)
			}
		}
	}
}

func assertGraphQLOperationsCovered(t *testing.T, covered map[string]bool) {
	t.Helper()

	operations, err := filepath.Glob(filepath.Join("..", "..", "pkg", "schemas", "graphql", "*.graphql"))
	if err != nil {
		t.Fatal(err)
	}

	pattern := regexp.MustCompile(`(?m)^\s*(?:query|mutation)\s+(\w+)`)
	count := 0

	for _, path := range operations {
		if strings.HasSuffix(path, "endpoints-schema.graphql") {
			continue
		}

		//nolint:gosec // The glob is scoped to checked-in GraphQL schema files.
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		matches := pattern.FindSubmatch(contents)
		if len(matches) < 2 {
			t.Errorf("no named operation in %s", path)

			continue
		}

		count++

		if !covered[string(matches[1])] {
			t.Errorf("GraphQL operation %s has no paired replay", matches[1])
		}
	}

	if count != 7 {
		t.Errorf("expected seven generated GraphQL operations, found %d", count)
	}
}

func assertEventReplayInventory(t *testing.T) {
	t.Helper()

	events := readPairs(t, "alexa-events.json")
	if len(events) != 2 || events[0].Operation != "openDirectiveStream" || events[1].Operation != "pingDirectiveStream" {
		t.Fatal("event exchange order is incomplete")
	}
}

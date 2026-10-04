// Command openapibundle combines the API responsibility specs into api/openapi.yaml.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	sourceManifestPath             = "api/openapi/sources.yaml"
	bundlePermissions              = 0o600
	errInvalidBundle   bundleError = "invalid API bundle"
)

type sourceManifest struct {
	Bundle     string           `yaml:"bundle"`
	Sources    []sourceSpec     `yaml:"sources"`
	Standalone []standaloneSpec `yaml:"standalone"`
}

type sourceSpec struct {
	File           string `yaml:"file"`
	Responsibility string `yaml:"responsibility"`
	Tag            string `yaml:"tag,omitempty"`
}

type standaloneSpec struct {
	File           string `yaml:"file"`
	Responsibility string `yaml:"responsibility"`
}

type sourceDocument struct {
	File           string
	Responsibility string
	Tag            string
	Spec           map[string]interface{}
}

func main() {
	write := flag.Bool("write", false, "write the deterministic master OpenAPI document")
	check := flag.Bool("check", false, "verify the master OpenAPI document matches its sources")
	flag.Parse()

	if *write && *check {
		fatalf("choose either -write or -check")
	}

	err := run(*write)
	if err != nil {
		fatalf("%v", err)
	}
}

func run(write bool) error {
	content, outputPath, err := buildBundle(".")
	if err != nil {
		return err
	}

	if write {
		err := os.WriteFile(filepath.FromSlash(outputPath), content, bundlePermissions)
		if err != nil {
			return fmt.Errorf("write %s: %w", outputPath, err)
		}

		return nil
	}

	actual, err := os.ReadFile(filepath.FromSlash(outputPath))
	if err != nil {
		return fmt.Errorf("read %s: %w (run go run ./tools/openapibundle -write)", outputPath, err)
	}

	if !bytes.Equal(actual, content) {
		return bundleDiagnostic("%s is stale; run go run ./tools/openapibundle -write", outputPath)
	}

	return nil
}

func buildBundle(root string) ([]byte, string, error) {
	//nolint:gosec // Reads the repository manifest under the caller-selected project root.
	manifestData, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sourceManifestPath)))
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", sourceManifestPath, err)
	}

	var manifest sourceManifest

	err = yaml.Unmarshal(manifestData, &manifest)
	if err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", sourceManifestPath, err)
	}

	if manifest.Bundle == "" || len(manifest.Sources) == 0 {
		return nil, "", bundleDiagnostic("OpenAPI source manifest needs a bundle path and at least one source")
	}

	sources := append([]sourceSpec(nil), manifest.Sources...)
	sort.Slice(sources, func(i, j int) bool { return sources[i].File < sources[j].File })

	documents := make([]sourceDocument, 0, len(sources))
	seenFiles := make(map[string]bool, len(sources))

	seenResponsibilities := make(map[string]bool, len(sources))

	for _, source := range sources {
		if source.File == "" || source.Responsibility == "" {
			return nil, "", bundleDiagnostic("each OpenAPI source needs a file and responsibility")
		}

		if source.File == manifest.Bundle || seenFiles[source.File] {
			return nil, "", bundleDiagnostic("OpenAPI source path %q is duplicated or names the bundle", source.File)
		}

		seenFiles[source.File] = true

		if seenResponsibilities[source.Responsibility] {
			return nil, "", bundleDiagnostic("OpenAPI responsibility %q is duplicated", source.Responsibility)
		}

		seenResponsibilities[source.Responsibility] = true

		spec, err := readOpenAPISpec(filepath.Join(root, filepath.FromSlash(source.File)))
		if err != nil {
			return nil, "", fmt.Errorf("source %s: %w", source.File, err)
		}

		err = validateSourceSpec(source, spec)
		if err != nil {
			return nil, "", fmt.Errorf("source %s: %w", source.File, err)
		}

		err = validateComponentReferences(spec)
		if err != nil {
			return nil, "", fmt.Errorf("source %s: %w", source.File, err)
		}

		documents = append(documents, sourceDocument{
			File: source.File, Responsibility: source.Responsibility, Tag: source.Tag, Spec: spec,
		})
	}

	err = validateStandaloneSources(root, manifest, seenFiles)
	if err != nil {
		return nil, "", err
	}

	bundle, err := mergeDocuments(documents)
	if err != nil {
		return nil, "", err
	}

	content, err := yaml.Marshal(bundle)
	if err != nil {
		return nil, "", fmt.Errorf("marshal bundled OpenAPI document: %w", err)
	}

	return content, manifest.Bundle, nil
}

func readOpenAPISpec(path string) (map[string]interface{}, error) {
	//nolint:gosec // Reads explicit source paths from the repository schema manifest.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read schema %s: %w", path, err)
	}

	var decoded interface{}

	err = yaml.Unmarshal(data, &decoded)
	if err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	spec, ok := normalize(decoded).(map[string]interface{})
	if !ok {
		return nil, bundleDiagnostic("OpenAPI document root must be a mapping")
	}

	if version, ok := spec["openapi"].(string); !ok || !strings.HasPrefix(version, "3.") {
		return nil, bundleDiagnostic("document must declare an OpenAPI 3.x version")
	}

	if _, ok := spec["info"].(map[string]interface{}); !ok {
		return nil, bundleDiagnostic("document must declare an info object")
	}

	if _, ok := spec["paths"].(map[string]interface{}); !ok {
		return nil, bundleDiagnostic("document must declare a paths object")
	}

	return spec, nil
}

func normalize(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(typed))

		for key, entry := range typed {
			name, ok := key.(string)
			if !ok {
				return value
			}

			result[name] = normalize(entry)
		}

		return result
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, entry := range typed {
			result[key] = normalize(entry)
		}

		return result
	case []interface{}:
		result := make([]interface{}, len(typed))
		for index, entry := range typed {
			result[index] = normalize(entry)
		}

		return result
	default:
		return value
	}
}

func validateComponentReferences(document map[string]interface{}) error {
	return visitComponentReferences(document, document)
}

func visitComponentReferences(value interface{}, document map[string]interface{}) error {
	switch typed := value.(type) {
	case map[string]interface{}:
		if rawReference, exists := typed["$ref"]; exists {
			reference, ok := rawReference.(string)
			if !ok {
				return bundleDiagnostic("$ref must be a string")
			}

			if strings.HasPrefix(reference, "#/components/") {
				_, err := resolveJSONPointer(document, reference)
				if err != nil {
					return err
				}
			}
		}

		for _, child := range typed {
			err := visitComponentReferences(child, document)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for _, child := range typed {
			err := visitComponentReferences(child, document)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func resolveJSONPointer(document map[string]interface{}, reference string) (interface{}, error) {
	segments := strings.Split(strings.TrimPrefix(reference, "#/"), "/")

	var current interface{} = document

	for _, segment := range segments {
		segment = strings.ReplaceAll(strings.ReplaceAll(segment, "~1", "/"), "~0", "~")

		switch value := current.(type) {
		case map[string]interface{}:
			var exists bool

			current, exists = value[segment]
			if !exists {
				return nil, bundleDiagnostic("unresolved component reference %q", reference)
			}
		case []interface{}:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(value) {
				return nil, bundleDiagnostic("unresolved component reference %q", reference)
			}

			current = value[index]
		default:
			return nil, bundleDiagnostic("unresolved component reference %q", reference)
		}
	}

	return current, nil
}

func validateSourceSpec(source sourceSpec, spec map[string]interface{}) error {
	paths, err := stringMap(spec["paths"], "paths")
	if err != nil {
		return err
	}

	if source.Tag == "" {
		if len(paths) != 0 {
			return bundleDiagnostic("shared source must not contain API paths")
		}

		return nil
	}

	if len(paths) == 0 {
		return bundleDiagnostic("responsibility %q has no API paths", source.Responsibility)
	}

	tags, err := tagNames(spec["tags"])
	if err != nil {
		return err
	}

	if !tags[source.Tag] {
		return bundleDiagnostic("source does not define its responsibility tag %q", source.Tag)
	}

	for path, pathItemValue := range paths {
		pathItem, err := stringMap(pathItemValue, "path item "+path)
		if err != nil {
			return err
		}

		for method, operationValue := range pathItem {
			if !isOperation(method) {
				continue
			}

			operation, err := stringMap(operationValue, "operation "+method+" "+path)
			if err != nil {
				return err
			}

			operationTags, err := stringSlice(operation["tags"], "operation tags")
			if err != nil {
				return fmt.Errorf("operation %s %s: %w", method, path, err)
			}

			if len(operationTags) != 1 || operationTags[0] != source.Tag {
				return bundleDiagnostic("operation %s %s must have only the responsibility tag %q", method, path, source.Tag)
			}
		}
	}

	return nil
}

func mergeDocuments(documents []sourceDocument) (map[string]interface{}, error) {
	if len(documents) == 0 {
		return nil, bundleDiagnostic("cannot bundle an empty OpenAPI source list")
	}

	merged := make(map[string]interface{})
	paths := make(map[string]interface{})
	pathOwners := make(map[string]string)
	components := make(map[string]interface{})
	tags := make(map[string]interface{})

	var common map[string]interface{}

	for _, document := range documents {
		currentCommon := make(map[string]interface{})

		for key, value := range document.Spec {
			if key == "paths" || key == "components" || key == "tags" {
				continue
			}

			currentCommon[key] = value
		}

		if common == nil {
			common = currentCommon
		} else if !reflect.DeepEqual(common, currentCommon) {
			return nil, bundleDiagnostic("source %s has root metadata that differs from other responsibility specs", document.File)
		}

		sourcePaths, err := stringMap(document.Spec["paths"], "paths")
		if err != nil {
			return nil, fmt.Errorf("source %s: %w", document.File, err)
		}

		for path, pathValue := range sourcePaths {
			if previous, exists := pathOwners[path]; exists {
				return nil, bundleDiagnostic("source %s duplicates path %s from source %s", document.File, path, previous)
			}

			paths[path] = pathValue
			pathOwners[path] = document.File
		}

		err = mergeComponents(components, document.Spec["components"], document.File)
		if err != nil {
			return nil, err
		}

		err = mergeTags(tags, document.Spec["tags"], document.File)
		if err != nil {
			return nil, err
		}
	}

	for key, value := range common {
		merged[key] = value
	}

	merged["paths"] = paths
	merged["components"] = components
	merged["tags"] = sortedTagList(tags)

	return merged, nil
}

func mergeComponents(target map[string]interface{}, raw interface{}, source string) error {
	if raw == nil {
		return nil
	}

	components, err := stringMap(raw, "components")
	if err != nil {
		return fmt.Errorf("source %s: %w", source, err)
	}

	for kind, values := range components {
		entries, err := stringMap(values, "components."+kind)
		if err != nil {
			return fmt.Errorf("source %s: %w", source, err)
		}

		merged, _ := target[kind].(map[string]interface{})
		if merged == nil {
			merged = make(map[string]interface{})
			target[kind] = merged
		}

		for name, value := range entries {
			if previous, exists := merged[name]; exists {
				if !reflect.DeepEqual(previous, value) {
					return bundleDiagnostic("source %s conflicts with components.%s.%s", source, kind, name)
				}

				continue
			}

			merged[name] = value
		}
	}

	return nil
}

func mergeTags(target map[string]interface{}, raw interface{}, source string) error {
	if raw == nil {
		return nil
	}

	tagList, ok := raw.([]interface{})
	if !ok {
		return bundleDiagnostic("source %s tags must be an array", source)
	}

	for _, value := range tagList {
		tag, err := stringMap(value, "tag")
		if err != nil {
			return fmt.Errorf("source %s: %w", source, err)
		}

		name, ok := tag["name"].(string)
		if !ok || name == "" {
			return bundleDiagnostic("source %s has a tag without a name", source)
		}

		if previous, exists := target[name]; exists {
			if !reflect.DeepEqual(previous, tag) {
				return bundleDiagnostic("source %s conflicts with tag %q", source, name)
			}

			continue
		}

		target[name] = tag
	}

	return nil
}

func sortedTagList(tags map[string]interface{}) []interface{} {
	names := make([]string, 0, len(tags))
	for name := range tags {
		names = append(names, name)
	}

	sort.Strings(names)

	result := make([]interface{}, 0, len(names))
	for _, name := range names {
		result = append(result, tags[name])
	}

	return result
}

func stringMap(value interface{}, label string) (map[string]interface{}, error) {
	result, ok := value.(map[string]interface{})
	if !ok {
		return nil, bundleDiagnostic("%s must be a mapping", label)
	}

	return result, nil
}

func stringSlice(value interface{}, label string) ([]string, error) {
	items, ok := value.([]interface{})
	if !ok {
		return nil, bundleDiagnostic("%s must be an array", label)
	}

	result := make([]string, 0, len(items))

	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil, bundleDiagnostic("%s must contain only strings", label)
		}

		result = append(result, text)
	}

	return result, nil
}

func tagNames(value interface{}) (map[string]bool, error) {
	items, ok := value.([]interface{})
	if !ok {
		return nil, bundleDiagnostic("tags must be an array")
	}

	result := make(map[string]bool, len(items))

	for _, item := range items {
		tag, err := stringMap(item, "tag")
		if err != nil {
			return nil, err
		}

		name, ok := tag["name"].(string)
		if !ok || name == "" {
			return nil, bundleDiagnostic("tag must have a name")
		}

		result[name] = true
	}

	return result, nil
}

func isOperation(method string) bool {
	switch method {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		return true
	default:
		return false
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// bundleError identifies a rejected schema composition.
type bundleError string

func (err bundleError) Error() string { return string(err) }
func bundleDiagnostic(format string, values ...any) error {
	return fmt.Errorf("%w: %s", errInvalidBundle, fmt.Sprintf(format, values...))
}

func validateStandaloneSources(root string, manifest sourceManifest, seenFiles map[string]bool) error {
	for _, standalone := range manifest.Standalone {
		if standalone.File == "" || standalone.Responsibility == "" {
			return bundleDiagnostic("each standalone OpenAPI source needs a file and responsibility")
		}

		if standalone.File == manifest.Bundle || seenFiles[standalone.File] {
			return bundleDiagnostic("standalone OpenAPI path %q is duplicated or names the bundle", standalone.File)
		}

		seenFiles[standalone.File] = true

		spec, err := readOpenAPISpec(filepath.Join(root, filepath.FromSlash(standalone.File)))
		if err != nil {
			return fmt.Errorf("standalone source %s: %w", standalone.File, err)
		}

		err = validateComponentReferences(spec)
		if err != nil {
			return fmt.Errorf("standalone source %s: %w", standalone.File, err)
		}
	}

	return nil
}

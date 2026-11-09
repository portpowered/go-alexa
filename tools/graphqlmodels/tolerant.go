package main

import (
	"fmt"
	"go/format"
	"regexp"
)

// genqlient decodes GraphQL unions strictly: one entry whose __typename is
// unknown or missing fails the whole response. Alexa adds and omits feature
// kinds without notice, so the moved models decode the feature property list
// and the feature configuration/semantics unions per entry instead. An entry
// that cannot be decoded is dropped (properties) or left nil (configuration,
// semantics); the rest of the response is kept. The enclosing envelope is
// still decoded strictly, so malformed or truncated JSON remains an error.
var (
	tolerantPropertiesPattern = regexp.MustCompile(
		`\*dst = make\(\s*\[\](\w+),\s*len\(src\)\)\s*` +
			`for i, src := range src \{\s*` +
			`dst := &\(\*dst\)\[i\]\s*` +
			`if len\(src\) != 0 && string\(src\) != "null" \{\s*` +
			`err = (__unmarshal\w+FeaturesFeaturePropertiesFeatureProperty)\(\s*src, dst\)\s*` +
			`if err != nil \{\s*return fmt\.Errorf\(\s*"[^"]*", err\)\s*\}\s*\}\s*\}`)

	tolerantUnionPattern = regexp.MustCompile(
		`err = (__unmarshal\w+FeaturesFeature(?:Configuration|Semantics))\(\s*src, dst\)\s*` +
			`if err != nil \{\s*return fmt\.Errorf\(\s*"[^"]*", err\)\s*\}`)
)

const tolerantPropertiesReplacement = `*dst = make([]${1}, 0, len(src))
for _, src := range src {
	if len(src) == 0 || string(src) == "null" {
		continue
	}

	var property ${1}
	if unmarshalErr := ${2}(src, &property); unmarshalErr != nil || property == nil {
		// Unknown, missing or malformed property kind: skip only this property.
		continue
	}

	*dst = append(*dst, property)
}`

const tolerantUnionReplacement = `if unmarshalErr := ${1}(src, dst); unmarshalErr != nil {
	// Unknown or malformed kind: leave the union unset rather than failing the response.
	*dst = nil
}`

func makeFeatureUnionsTolerant(source []byte) ([]byte, error) {
	if !tolerantPropertiesPattern.Match(source) {
		return nil, fmt.Errorf("%w: feature properties decoder", errTolerantPatternMissing)
	}

	if !tolerantUnionPattern.Match(source) {
		return nil, fmt.Errorf("%w: feature configuration/semantics decoder", errTolerantPatternMissing)
	}

	out := tolerantPropertiesPattern.ReplaceAll(source, []byte(tolerantPropertiesReplacement))
	out = tolerantUnionPattern.ReplaceAll(out, []byte(tolerantUnionReplacement))

	formatted, err := format.Source(out)
	if err != nil {
		return nil, fmt.Errorf("format tolerant GraphQL models: %w", err)
	}

	return formatted, nil
}

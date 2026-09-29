// Command coverage reports statement coverage for non-generated library code.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	defaultMinimumCoveragePercent = 80
	coverageProfileFieldCount     = 3
	filteredProfilePermissions    = 0o600
	coverageErrorPreviewBytes     = 2048
)

type totals struct {
	covered int64
	all     int64
}

type coverageBlock struct {
	source     string
	statements int64
	covered    bool
}

func main() {
	profile := flag.String("profile", "coverage.out", "Go coverage profile")
	minimum := flag.Float64("min", defaultMinimumCoveragePercent, "minimum combined statement coverage percent")
	percentOnly := flag.Bool("percent-only", false, "print only the combined percentage")
	filteredProfile := flag.String("filtered-profile", "", "write a profile without generated files")
	flag.Parse()

	err := report(*profile, *minimum, *percentOnly, *filteredProfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func report(profile string, minimum float64, percentOnly bool, filteredProfile string) (reportErr error) {
	module, err := modulePath()
	if err != nil {
		return err
	}

	//nolint:gosec // The CLI accepts a caller-selected local coverage profile.
	file, err := os.Open(profile)
	if err != nil {
		return fmt.Errorf("open coverage profile %s: %w", profile, err)
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil && reportErr == nil {
			reportErr = fmt.Errorf("close coverage profile: %w", closeErr)
		}
	}()

	mode, blocks, err := readCoverageBlocks(file, profile, module)
	if err != nil {
		return err
	}

	if mode == "" {
		return errors.New("coverage profile has no mode header")
	}

	summary := summarizeCoverage(mode, blocks)

	if filteredProfile != "" {
		err := os.WriteFile(filteredProfile, []byte(summary.filteredProfile), filteredProfilePermissions)
		if err != nil {
			return fmt.Errorf("write filtered coverage profile %s: %w", filteredProfile, err)
		}
	}

	return writeCoverageSummary(summary, minimum, percentOnly)
}

func readCoverageBlocks(file *os.File, profile, module string) (string, map[string]coverageBlock, error) {
	generated := make(map[string]bool)
	blocks := make(map[string]coverageBlock)
	mode := ""
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			updatedMode, err := updateCoverageMode(mode, line)
			if err != nil {
				return "", nil, err
			}

			mode = updatedMode

			continue
		}

		location, block, include, err := parseCoverageRecord(line, module, generated)
		if err != nil {
			return "", nil, err
		}

		if include {
			err = mergeCoverageRecord(blocks, location, block)
			if err != nil {
				return "", nil, err
			}
		}
	}

	scanErr := scanner.Err()
	if scanErr != nil {
		return "", nil, fmt.Errorf("read coverage profile %s: %w", profile, scanErr)
	}

	return mode, blocks, nil
}

func updateCoverageMode(mode, line string) (string, error) {
	if mode != "" && mode != line {
		return mode, fmt.Errorf("coverage profile has conflicting modes: %q and %q", mode, line)
	}

	return line, nil
}

func parseCoverageRecord(
	line string,
	module string,
	generated map[string]bool,
) (string, coverageBlock, bool, error) {
	fields := strings.Fields(line)
	if len(fields) != coverageProfileFieldCount {
		return "", emptyCoverageBlock(), false, fmt.Errorf("invalid coverage record: %q", line)
	}

	location := fields[0]

	colon := strings.LastIndexByte(location, ':')
	if colon < 0 {
		return "", emptyCoverageBlock(), false, fmt.Errorf("invalid coverage location: %q", location)
	}

	source := strings.TrimPrefix(strings.ReplaceAll(location[:colon], "\\", "/"), module+"/")
	if !strings.HasPrefix(source, "pkg/") || strings.HasPrefix(source, "pkg/testing/") {
		return "", emptyCoverageBlock(), false, nil
	}

	isGenerated, ok := generated[source]
	if !ok {
		generatedSource, err := generatedFile(source)
		if err != nil {
			return "", emptyCoverageBlock(), false, err
		}

		isGenerated = generatedSource
		generated[source] = isGenerated
	}

	if isGenerated {
		return "", emptyCoverageBlock(), false, nil
	}

	statements, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || statements < 0 {
		return "", emptyCoverageBlock(), false, fmt.Errorf("invalid statement count in %q", line)
	}

	count, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || count < 0 {
		return "", emptyCoverageBlock(), false, fmt.Errorf("invalid execution count in %q", line)
	}

	block := coverageBlock{source: source, statements: statements, covered: count > 0}

	return location, block, true, nil
}

func emptyCoverageBlock() coverageBlock {
	return coverageBlock{source: "", statements: 0, covered: false}
}

func mergeCoverageRecord(blocks map[string]coverageBlock, location string, block coverageBlock) error {
	previous, exists := blocks[location]
	if !exists {
		blocks[location] = block

		return nil
	}

	if previous.statements != block.statements {
		return fmt.Errorf("coverage block %q has conflicting statement counts", location)
	}

	previous.covered = previous.covered || block.covered
	blocks[location] = previous

	return nil
}

type coverageSummary struct {
	byPackage       map[string]totals
	packages        []string
	combined        totals
	filteredProfile string
}

func summarizeCoverage(mode string, blocks map[string]coverageBlock) coverageSummary {
	summary := coverageSummary{
		byPackage:       make(map[string]totals),
		packages:        nil,
		combined:        totals{covered: 0, all: 0},
		filteredProfile: "",
	}

	keys := make([]string, 0, len(blocks))
	for location := range blocks {
		keys = append(keys, location)
	}

	sort.Strings(keys)

	var filtered strings.Builder

	filtered.WriteString(mode + "\n")

	for _, location := range keys {
		block := blocks[location]

		count := 0
		if block.covered {
			count = 1
		}

		fmt.Fprintf(&filtered, "%s %d %d\n", location, block.statements, count)

		packageName := path.Dir(block.source)
		entry := summary.byPackage[packageName]

		entry.all += block.statements
		if block.covered {
			entry.covered += block.statements
		}

		summary.byPackage[packageName] = entry
	}

	summary.filteredProfile = filtered.String()

	for name := range summary.byPackage {
		summary.packages = append(summary.packages, name)
	}

	sort.Strings(summary.packages)

	for _, name := range summary.packages {
		entry := summary.byPackage[name]
		summary.combined.all += entry.all
		summary.combined.covered += entry.covered
	}

	return summary
}

func writeCoverageSummary(summary coverageSummary, minimum float64, percentOnly bool) error {
	for _, name := range summary.packages {
		entry := summary.byPackage[name]
		if entry.all > 0 && !percentOnly {
			outputf("%s: %.1f%% (%d/%d statements)\n", name, percent(entry), entry.covered, entry.all)
		}
	}

	if summary.combined.all == 0 {
		return errors.New("coverage profile has no non-generated pkg statements")
	}

	measured := percent(summary.combined)
	if percentOnly {
		outputf("%.1f\n", measured)
	} else {
		outputf("combined non-generated pkg coverage: %.1f%% (%d/%d statements)\n", measured, summary.combined.covered, summary.combined.all)
	}

	if measured+1e-9 < minimum {
		return fmt.Errorf("coverage %.1f%% is below %.1f%% minimum", measured, minimum)
	}

	return nil
}

func modulePath() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("read module file go.mod: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}

	return "", errors.New("module directive not found in go.mod")
}

func generatedFile(filename string) (bool, error) {
	if strings.HasSuffix(filename, ".gen.go") {
		return true, nil
	}

	//nolint:gosec // The coverage profile supplies local source paths to inspect.
	data, err := os.ReadFile(filename)
	if err != nil {
		return false, fmt.Errorf("read source file %s: %w", filename, err)
	}

	if len(data) > coverageErrorPreviewBytes {
		data = data[:coverageErrorPreviewBytes]
	}

	return strings.Contains(string(data), "Code generated") && strings.Contains(string(data), "DO NOT EDIT"), nil
}

func percent(value totals) float64 {
	return 100 * float64(value.covered) / float64(value.all)
}

func outputf(format string, values ...any) {
	_, err := fmt.Fprintf(os.Stdout, format, values...)
	if err != nil {
		panic(err)
	}
}

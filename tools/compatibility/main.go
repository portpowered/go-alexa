// Command compatibility compares the public Go API with a Git baseline.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const parsedVersionFieldCount = 4

const (
	policyRelease            = "release"
	policyReport             = "report"
	defaultModulePath        = "github.com/portpowered/go-alexa"
	defaultPublicPackages    = "pkg/alexa,pkg/alexaapimodels"
	apiDiffTool              = "golang.org/x/exp/cmd/apidiff@v0.0.0-20260908205506-85c1c2202aba"
	previousRelease          = "previous-release"
	toolDirectoryPermissions = 0o755
	summaryFilePermissions   = 0o600
)

var (
	stableTagPattern = regexp.MustCompile("^v(0|[1-9][0-9]*)[.](0|[1-9][0-9]*)[.](0|[1-9][0-9]*)$")
)

type gateError struct {
	message string
	cause   error
}

func (err gateError) Error() string {
	if err.cause == nil {
		return err.message
	}

	return err.message + ": " + err.cause.Error()
}

func (err gateError) Unwrap() error {
	return err.cause
}

type version struct {
	major int
	minor int
	patch int
}

type incompatibleChange struct {
	packageName string
	details     string
}

type comparisonOptions struct {
	baseRef        string
	releaseVersion string
	policy         string
	modulePath     string
	packages       []string
}

type comparisonWorkspace struct {
	root        string
	tempDir     string
	baseDir     string
	tool        string
	hasWorktree bool
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	options, err := parseComparisonOptions()
	if err != nil {
		return err
	}

	return compareCompatibility(options)
}

func parseComparisonOptions() (comparisonOptions, error) {
	baseRef := flag.String("base", "", "Git ref to compare against, or previous-release")
	releaseVersion := flag.String("version", "", "release tag, used with -policy=release")
	policy := flag.String("policy", policyReport, "report or release")
	modulePath := flag.String("module", defaultModulePath, "module import path")
	packageList := flag.String("packages", defaultPublicPackages, "comma-separated public package paths relative to the module root")
	flag.Parse()

	if *baseRef == "" {
		return comparisonOptions{}, gateError{message: "-base is required", cause: nil}
	}

	if *policy != policyReport && *policy != policyRelease {
		return comparisonOptions{}, gateError{message: "-policy must be report or release", cause: nil}
	}

	if *policy == policyRelease && *releaseVersion == "" {
		return comparisonOptions{}, gateError{message: "-version is required with -policy=release", cause: nil}
	}

	if *policy == policyRelease {
		err := validateReleaseTag(*releaseVersion)
		if err != nil {
			return comparisonOptions{}, gateError{message: "validate release version", cause: err}
		}
	}

	if strings.TrimSpace(*modulePath) == "" || strings.HasSuffix(*modulePath, "/") {
		return comparisonOptions{}, gateError{message: "-module must be a non-empty module import path", cause: nil}
	}

	packages, err := parsePackages(*packageList)
	if err != nil {
		return comparisonOptions{}, gateError{message: "parse -packages", cause: err}
	}

	return comparisonOptions{
		baseRef:        *baseRef,
		releaseVersion: *releaseVersion,
		policy:         *policy,
		modulePath:     *modulePath,
		packages:       packages,
	}, nil
}

func compareCompatibility(options comparisonOptions) error {
	root, err := repositoryRoot()
	if err != nil {
		return gateError{message: "find repository root", cause: err}
	}

	requestedBase := options.baseRef
	if isZeroRef(requestedBase) {
		requestedBase = previousRelease
	}

	base, found, err := resolveBase(root, requestedBase, options.releaseVersion)
	if err != nil {
		return gateError{message: "resolve baseline", cause: err}
	}

	if !found {
		outputLine("No previous stable release tag found; skipping API compatibility comparison.")

		return nil
	}

	err = validateComparisonReleaseOrder(options, base)
	if err != nil {
		return err
	}

	workspace, err := createComparisonWorkspace(root, base)
	if err != nil {
		return err
	}
	defer workspace.cleanup()

	changes, err := comparePackageAPIs(options, workspace)
	if err != nil {
		return err
	}

	return reportCompatibility(options, base, changes)
}

func validateComparisonReleaseOrder(options comparisonOptions, base string) error {
	if options.policy != policyRelease {
		return nil
	}

	err := validateReleaseOrder(base, options.releaseVersion)
	if err != nil {
		return gateError{message: "validate release version", cause: err}
	}

	return nil
}

func createComparisonWorkspace(root, base string) (comparisonWorkspace, error) {
	tempDir, err := os.MkdirTemp("", "go-api-compatibility-")
	if err != nil {
		return comparisonWorkspace{}, gateError{message: "create temporary directory", cause: err}
	}

	workspace := comparisonWorkspace{
		root:        root,
		tempDir:     tempDir,
		baseDir:     filepath.Join(tempDir, "base"),
		tool:        "",
		hasWorktree: false,
	}

	err = git(root, "worktree", "add", "--detach", workspace.baseDir, base)
	if err != nil {
		_ = os.RemoveAll(tempDir)

		return comparisonWorkspace{}, gateError{message: "create baseline worktree", cause: err}
	}

	workspace.hasWorktree = true

	toolDir := filepath.Join(tempDir, "bin")

	err = os.Mkdir(toolDir, toolDirectoryPermissions)
	if err != nil {
		workspace.cleanup()

		return comparisonWorkspace{}, gateError{message: "create tool directory", cause: err}
	}

	err = installAPIDiff(root, toolDir)
	if err != nil {
		workspace.cleanup()

		return comparisonWorkspace{}, gateError{message: "install API comparison tool", cause: err}
	}

	workspace.tool = filepath.Join(toolDir, "apidiff")
	if runtime.GOOS == "windows" {
		workspace.tool += ".exe"
	}

	return workspace, nil
}

func (workspace comparisonWorkspace) cleanup() {
	if workspace.hasWorktree {
		err := git(workspace.root, "worktree", "remove", "--force", workspace.baseDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "remove baseline worktree: %v\n", err)
		}
	}

	_ = os.RemoveAll(workspace.tempDir)
}

func comparePackageAPIs(options comparisonOptions, workspace comparisonWorkspace) ([]incompatibleChange, error) {
	changes := make([]incompatibleChange, 0)

	for _, packageName := range options.packages {
		packagePath := fullPackagePath(options.modulePath, packageName)
		baseFileName := strings.ReplaceAll(packageName, "/", "-") + "-base.export"
		currentFileName := strings.ReplaceAll(packageName, "/", "-") + "-current.export"
		oldData := filepath.Join(workspace.tempDir, baseFileName)
		newData := filepath.Join(workspace.tempDir, currentFileName)

		err := writeExportData(workspace.tool, workspace.baseDir, packagePath, oldData)
		if err != nil {
			return nil, gateError{message: "read baseline API for " + packagePath, cause: err}
		}

		err = writeExportData(workspace.tool, workspace.root, packagePath, newData)
		if err != nil {
			return nil, gateError{message: "read current API for " + packagePath, cause: err}
		}

		packageChanges, err := compareAPIs(workspace.tool, oldData, newData)
		if err != nil {
			return nil, gateError{message: "compare API for " + packagePath, cause: err}
		}

		if packageChanges != "" {
			changes = append(changes, incompatibleChange{packageName: packageName, details: packageChanges})
		}
	}

	return changes, nil
}

func reportCompatibility(options comparisonOptions, base string, changes []incompatibleChange) error {
	releaseMessage, err := releaseMessageForChanges(options, base, changes)
	if err != nil {
		return err
	}

	err = writeReportSummary(options.policy, base, changes, releaseMessage, false)
	if err != nil {
		return gateError{message: "write CI summary", cause: err}
	}

	return printCompatibilityResult(options, base, changes, releaseMessage)
}

func releaseMessageForChanges(options comparisonOptions, base string, changes []incompatibleChange) (string, error) {
	if len(changes) == 0 || options.policy != policyRelease {
		return "", nil
	}

	allowed, reason, err := releaseAllowsBreak(base, options.releaseVersion)
	if err != nil {
		return "", gateError{message: "validate release version", cause: err}
	}

	if allowed {
		return reason, nil
	}

	err = writeReportSummary(options.policy, base, changes, "", true)
	if err != nil {
		return "", gateError{message: "write CI summary", cause: err}
	}

	return "", gateError{
		message: fmt.Sprintf(
			"release %s has incompatible API changes from %s; the version increase does not permit these changes",
			options.releaseVersion,
			base,
		),
		cause: nil,
	}
}

func printCompatibilityResult(
	options comparisonOptions,
	base string,
	changes []incompatibleChange,
	releaseMessage string,
) error {
	if len(changes) == 0 {
		outputf("Public Go API is compatible with %s.\n", base)

		return nil
	}

	for _, change := range changes {
		outputf("Incompatible changes in %s:\n%s\n", change.packageName, change.details)
	}

	if options.policy == policyReport {
		outputLine("Report only: reviewers must explicitly acknowledge this report before approving an intentional API break.")

		return nil
	}

	outputf("Release %s permits these incompatible changes: %s.\n", options.releaseVersion, releaseMessage)

	return nil
}

func parsePackages(value string) ([]string, error) {
	packages := make([]string, 0)
	seen := make(map[string]struct{})

	for _, item := range strings.Split(value, ",") {
		name := strings.TrimSpace(item)
		if name == "" {
			return nil, gateError{message: "package paths must not be empty", cause: nil}
		}

		if name != "." && (strings.HasPrefix(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..")) {
			return nil, gateError{message: "package path must be relative to the module root: " + name, cause: nil}
		}

		if _, exists := seen[name]; exists {
			return nil, gateError{message: "duplicate public package: " + name, cause: nil}
		}

		seen[name] = struct{}{}

		packages = append(packages, name)
	}

	return packages, nil
}

func fullPackagePath(modulePath, packageName string) string {
	if packageName == "." {
		return modulePath
	}

	return modulePath + "/" + packageName
}

func repositoryRoot() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current working directory: %w", err)
	}

	output, err := runGit(workingDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func resolveBase(root, requested, releaseVersion string) (string, bool, error) {
	if requested != previousRelease {
		return requested, true, nil
	}

	tags, err := runGit(root, "tag", "--list", "v*", "--sort=-version:refname")
	if err != nil {
		return "", false, err
	}

	var target version
	if releaseVersion != "" {
		target, err = parseVersion(releaseVersion, stableTagPattern)
		if err != nil {
			return "", false, gateError{
				message: "release version must be a semantic vMAJOR.MINOR.PATCH tag",
				cause:   err,
			}
		}
	}

	hasOtherStableTag := false

	for _, tag := range strings.Fields(string(tags)) {
		candidate, parseErr := parseVersion(tag, stableTagPattern)
		if parseErr != nil {
			continue
		}

		if releaseVersion == "" {
			return tag, true, nil
		}

		if tag != releaseVersion {
			hasOtherStableTag = true
		}

		if compareVersions(candidate, target) < 0 {
			return tag, true, nil
		}
	}

	if !hasOtherStableTag {
		return "", false, nil
	}

	return "", false, gateError{message: "no stable release tag is older than " + releaseVersion, cause: nil}
}

func releaseAllowsBreak(baseTag, releaseTag string) (bool, string, error) {
	base, err := parseVersion(baseTag, stableTagPattern)
	if err != nil {
		return false, "", err
	}

	release, err := parseVersion(releaseTag, stableTagPattern)
	if err != nil {
		return false, "", err
	}

	if release.major > base.major {
		return true, "the major version increased", nil
	}

	if base.major == 0 && release.major == 0 && release.minor > base.minor {
		return true, "the v0 minor version increased", nil
	}

	return false, "", nil
}

func validateReleaseOrder(baseTag, releaseTag string) error {
	base, err := parseVersion(baseTag, stableTagPattern)
	if err != nil {
		return err
	}

	release, err := parseVersion(releaseTag, stableTagPattern)
	if err != nil {
		return err
	}

	if compareVersions(release, base) <= 0 {
		return gateError{message: fmt.Sprintf("release tag %s must be newer than baseline %s", releaseTag, baseTag), cause: nil}
	}

	return nil
}

func validateReleaseTag(releaseTag string) error {
	{
		_, err := parseVersion(releaseTag, stableTagPattern)
		if err != nil {
			return gateError{message: "release version must be a semantic vMAJOR.MINOR.PATCH tag", cause: err}
		}
	}

	return nil
}

func parseVersion(tag string, pattern *regexp.Regexp) (version, error) {
	match := pattern.FindStringSubmatch(tag)
	if len(match) != parsedVersionFieldCount {
		return version{}, gateError{message: "invalid version tag " + tag, cause: nil}
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		return version{}, gateError{message: "parse major version in " + tag, cause: err}
	}

	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return version{}, gateError{message: "parse minor version in " + tag, cause: err}
	}

	patch, err := strconv.Atoi(match[3])
	if err != nil {
		return version{}, gateError{message: "parse patch version in " + tag, cause: err}
	}

	return version{major: major, minor: minor, patch: patch}, nil
}

func compareVersions(left, right version) int {
	if left.major != right.major {
		if left.major < right.major {
			return -1
		}

		return 1
	}

	if left.minor != right.minor {
		if left.minor < right.minor {
			return -1
		}

		return 1
	}

	if left.patch < right.patch {
		return -1
	}

	if left.patch > right.patch {
		return 1
	}

	return 0
}

func installAPIDiff(root, toolDir string) error {
	cmd := exec.CommandContext(context.Background(), "go", "install", apiDiffTool)
	cmd.Dir = root
	cmd.Env = withEnvironment(os.Environ(), "GOBIN", toolDir)
	cmd.Env = withEnvironment(cmd.Env, "GOTOOLCHAIN", "auto")
	cmd.Env = withEnvironment(cmd.Env, "GOWORK", "off")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("install API diff tool: %w", err)
	}

	return nil
}

func writeExportData(tool, directory, packagePath, outputPath string) error {
	cmd := exec.CommandContext(context.Background(), tool, "-w", outputPath, packagePath)
	cmd.Dir = directory
	cmd.Env = withEnvironment(os.Environ(), "GOTOOLCHAIN", "auto")
	cmd.Env = withEnvironment(cmd.Env, "GOWORK", "off")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("write export data for %s: %w", packagePath, err)
	}

	return nil
}

func compareAPIs(tool, oldData, newData string) (string, error) {
	cmd := exec.CommandContext(context.Background(), tool, "-incompatible", oldData, newData)
	cmd.Env = withEnvironment(os.Environ(), "GOTOOLCHAIN", "auto")
	cmd.Env = withEnvironment(cmd.Env, "GOWORK", "off")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", gateError{message: strings.TrimSpace(string(output)), cause: err}
	}

	return strings.TrimSpace(string(output)), nil
}

func writeReportSummary(policy, base string, changes []incompatibleChange, releaseMessage string, releaseFailed bool) error {
	var summary strings.Builder

	summary.WriteString("## Public Go API compatibility\n\n")
	summary.WriteString("Baseline: " + base + "\n\n")

	if len(changes) == 0 {
		summary.WriteString("No incompatible API changes were found.\n")
	} else {
		summary.WriteString("Incompatible API changes were found.\n\n")

		if policy == policyReport {
			summary.WriteString("This check reports changes without blocking merges. Reviewers must explicitly acknowledge an intentional API break.\n\n")
		} else {
			summary.WriteString("The release check evaluates these changes against the release version policy.\n\n")
		}

		for _, change := range changes {
			summary.WriteString("### " + change.packageName + "\n\n")
			summary.WriteString(change.details + "\n\n")
		}
	}

	switch {
	case policy == policyRelease && releaseFailed:
		summary.WriteString("Release compatibility check failed because the version does not permit the incompatible changes.\n")
	case policy == policyRelease && releaseMessage == "" && len(changes) == 0:
		summary.WriteString("Release compatibility check passed.\n")
	case policy == policyRelease && releaseMessage != "":
		summary.WriteString("Release compatibility check passed: " + releaseMessage + ".\n")
	}

	text := summary.String()
	outputText(text)

	path := os.Getenv("GITHUB_STEP_SUMMARY")
	if path == "" {
		return nil
	}

	//nolint:gosec // The CI runner supplies its explicit local summary path.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, summaryFilePermissions)
	if err != nil {
		return fmt.Errorf("open GitHub step summary: %w", err)
	}

	_, writeErr := file.WriteString(text + "\n")
	closeErr := file.Close()

	if writeErr != nil {
		return fmt.Errorf("write GitHub step summary: %w", writeErr)
	}

	if closeErr != nil {
		return fmt.Errorf("close GitHub step summary: %w", closeErr)
	}

	return nil
}

func isZeroRef(ref string) bool {
	return ref != "" && strings.Trim(ref, "0") == ""
}

func git(root string, args ...string) error {
	_, err := runGit(root, args...)

	return err
}

func runGit(root string, args ...string) ([]byte, error) {
	gitArgs := append([]string{"-c", "safe.directory=" + filepath.ToSlash(root), "-C", root}, args...)
	//nolint:gosec // The executable is fixed to git and argv is never passed through a shell.
	cmd := exec.CommandContext(context.Background(), "git", gitArgs...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run git %s: %w", strings.Join(args, " "), err)
	}

	return output, nil
}

func withEnvironment(environment []string, name, value string) []string {
	prefix := name + "="

	filtered := make([]string, 0, len(environment)+1)

	for _, entry := range environment {
		if !strings.HasPrefix(entry, prefix) {
			filtered = append(filtered, entry)
		}
	}

	return append(filtered, prefix+value)
}

func outputLine(values ...any) {
	_, err := fmt.Fprintln(os.Stdout, values...)
	if err != nil {
		panic(err)
	}
}

func outputf(format string, values ...any) {
	_, err := fmt.Fprintf(os.Stdout, format, values...)
	if err != nil {
		panic(err)
	}
}

func outputText(values ...any) {
	_, err := fmt.Fprint(os.Stdout, values...)
	if err != nil {
		panic(err)
	}
}

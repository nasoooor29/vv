package depscan

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	classifier "github.com/google/licenseclassifier/v2"
	"golang.org/x/mod/modfile"
)

// Dependency represents a resolved dependency with license information
type Dependency struct {
	Path        string // Module path (e.g., github.com/user/repo)
	Version     string // Semantic version
	Indirect    bool   // Whether this is an indirect dependency
	License     string // SPDX identifier or "UNKNOWN"
	LicenseFile string // Relative path to license file from module root
	Dir         string // Actual directory of the module
}

// goModuleFromGoList represents the JSON structure from go list -m -json all
type goModuleFromGoList struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
	Sum     string `json:"Sum"`
	Dir     string `json:"Dir"`
	Main    bool   `json:"Main"`
	Replace *struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
		Dir     string `json:"Dir"`
	} `json:"Replace"`
	GoMod     string `json:"GoMod"`
	GoVersion string `json:"GoVersion"`
}

// Config holds configuration for dependency scanning
type Config struct {
	GoModPath    string   // Path to go.mod file
	LicenseNames []string // License file names to search for (in order)
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig(goModPath string) Config {
	return Config{
		GoModPath: goModPath,
		LicenseNames: []string{
			"LICENSE",
			"LICENSE.md",
			"LICENSE.txt",
			"COPYING",
			"COPYRIGHT",
		},
	}
}

// parseGoMod parses the go.mod file and returns a map of module path -> isIndirect
func parseGoMod(goModPath string) (map[string]bool, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %w", err)
	}

	file, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod: %w", err)
	}

	indirectMap := make(map[string]bool)
	for _, req := range file.Require {
		indirectMap[req.Mod.Path] = req.Indirect
	}

	return indirectMap, nil
}

// resolveModules executes go list -m -json all and returns resolved modules
func resolveModules() ([]goModuleFromGoList, error) {
	cmd := exec.Command("go", "list", "-m", "-json", "all")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute go list: %w", err)
	}

	var modules []goModuleFromGoList
	decoder := json.NewDecoder(strings.NewReader(string(output)))

	for decoder.More() {
		var mod goModuleFromGoList
		if err := decoder.Decode(&mod); err != nil {
			return nil, fmt.Errorf("failed to decode go list output: %w", err)
		}
		modules = append(modules, mod)
	}

	return modules, nil
}

// findLicenseFile searches for a license file in the given directory
func findLicenseFile(dir string, licenseNames []string) (string, error) {
	if dir == "" {
		return "", errors.New("empty directory path")
	}

	for _, name := range licenseNames {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			// Return relative path from module root for clarity
			return name, nil
		}
	}

	return "", errors.New("no license file found")
}

// classifyLicense uses licenseclassifier to detect SPDX license
func classifyLicense(licenseFilePath string) (string, error) {
	data, err := os.ReadFile(licenseFilePath)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to read license file: %w", err)
	}

	// Create a classifier with default threshold (0.8)
	cl := classifier.NewClassifier(0.8)
	if cl == nil {
		return "UNKNOWN", errors.New("failed to create classifier")
	}

	result := cl.Match(data)
	if len(result.Matches) > 0 {
		// Return the top match's SPDX identifier
		topMatch := result.Matches[0]
		if topMatch.MatchType == "License" {
			return topMatch.Name, nil
		}
	}

	// Fallback: try simple keyword matching for common licenses
	return detectLicenseByKeywords(string(data)), nil
}

// detectLicenseByKeywords implements a simple fallback for common license detection
func detectLicenseByKeywords(content string) string {
	contentUpper := strings.ToUpper(content)

	// Check for common licenses in order
	licensePatterns := map[string][]string{
		"MIT":          {"MIT LICENSE", "PERMISSION IS HEREBY GRANTED, FREE OF CHARGE"},
		"Apache-2.0":   {"APACHE LICENSE", "APACHE SOFTWARE FOUNDATION"},
		"GPL-2.0":      {"GNU GENERAL PUBLIC LICENSE", "VERSION 2"},
		"GPL-3.0":      {"GNU GENERAL PUBLIC LICENSE", "VERSION 3"},
		"AGPL-3.0":     {"AFFERO GENERAL PUBLIC LICENSE", "VERSION 3"},
		"BSD-2-Clause": {"BSD 2-CLAUSE LICENSE", "SIMPLIFIED BSD LICENSE"},
		"BSD-3-Clause": {"BSD 3-CLAUSE LICENSE"},
		"ISC":          {"ISC LICENSE"},
		"MPL-2.0":      {"MOZILLA PUBLIC LICENSE", "VERSION 2.0"},
		"LGPL-2.1":     {"LESSER GENERAL PUBLIC LICENSE", "VERSION 2.1"},
		"LGPL-3.0":     {"LESSER GENERAL PUBLIC LICENSE", "VERSION 3"},
	}

	for license, patterns := range licensePatterns {
		found := true
		for _, pattern := range patterns {
			if !strings.Contains(contentUpper, pattern) {
				found = false
				break
			}
		}
		if found {
			return license
		}
	}

	return "UNKNOWN"
}

// CollectDependencies is the main entry point that collects and classifies all dependencies
func CollectDependencies(goModPath string) ([]Dependency, error) {
	return CollectDependenciesWithConfig(DefaultConfig(goModPath))
}

// CollectDependenciesWithConfig collects dependencies using custom config
func CollectDependenciesWithConfig(cfg Config) ([]Dependency, error) {
	// Step 1: Parse go.mod to identify direct vs indirect
	indirectMap, err := parseGoMod(cfg.GoModPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod: %w", err)
	}

	// Step 2: Resolve all modules using go list
	modules, err := resolveModules()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve modules: %w", err)
	}

	// Step 3: Build dependencies, skipping main module
	var dependencies []Dependency

	for _, mod := range modules {
		// Skip the main module
		if mod.Main {
			continue
		}

		// Determine the actual directory (respect replace directives)
		moduleDir := mod.Dir
		if mod.Replace != nil && mod.Replace.Dir != "" {
			moduleDir = mod.Replace.Dir
		}

		// Find license file
		licenseFile, licenseErr := findLicenseFile(moduleDir, cfg.LicenseNames)
		if licenseErr != nil {
			licenseFile = ""
		}

		// Classify license
		license := "UNKNOWN"
		if licenseFile != "" {
			licenseFilePath := filepath.Join(moduleDir, licenseFile)
			classifiedLicense, classifyErr := classifyLicense(licenseFilePath)
			if classifyErr == nil {
				license = classifiedLicense
			}
		}

		// Get indirect flag from go.mod (not from go list)
		isIndirect := indirectMap[mod.Path]

		dependencies = append(dependencies, Dependency{
			Path:        mod.Path,
			Version:     mod.Version,
			Indirect:    isIndirect,
			License:     license,
			LicenseFile: licenseFile,
			Dir:         moduleDir,
		})
	}

	return dependencies, nil
}

// FilterByLicense filters dependencies by license type
func FilterByLicense(deps []Dependency, licenses ...string) []Dependency {
	licenseSet := make(map[string]bool)
	for _, l := range licenses {
		licenseSet[l] = true
	}

	var filtered []Dependency
	for _, dep := range deps {
		if licenseSet[dep.License] {
			filtered = append(filtered, dep)
		}
	}
	return filtered
}

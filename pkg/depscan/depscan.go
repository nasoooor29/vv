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
		// License file names to search for - we'll do case-insensitive matching
		// and also check for common misspellings (licence vs license)
		LicenseNames: []string{
			// Standard spellings
			"LICENSE",
			"LICENSE.md",
			"LICENSE.txt",
			"LICENSE.MIT",
			"LICENSE-MIT",
			"LICENSE.APACHE",
			"LICENSE-APACHE",
			// British spelling (licence)
			"LICENCE",
			"LICENCE.md",
			"LICENCE.txt",
			// Other common names
			"COPYING",
			"COPYING.md",
			"COPYING.txt",
			"COPYRIGHT",
			"COPYRIGHT.md",
			"COPYRIGHT.txt",
			// Lowercase variants
			"license",
			"license.md",
			"license.txt",
			"licence",
			"licence.md",
			"licence.txt",
			"copying",
			"copyright",
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
// It performs case-insensitive matching and checks for common misspellings
func findLicenseFile(dir string, licenseNames []string) (string, error) {
	if dir == "" {
		return "", errors.New("empty directory path")
	}

	// First, try exact matches from the provided list
	for _, name := range licenseNames {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return name, nil
		}
	}

	// If no exact match, scan directory for case-insensitive matches
	// This catches any case variations we might have missed
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	// Patterns to match (case-insensitive) - includes common misspellings
	licensePatterns := []string{
		"license", // American spelling
		"licence", // British spelling
		"copying",
		"copyright",
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		nameLower := strings.ToLower(entry.Name())
		for _, pattern := range licensePatterns {
			if strings.HasPrefix(nameLower, pattern) {
				return entry.Name(), nil
			}
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

	// Check licenses in priority order (more specific first)
	// Each check uses characteristic phrases from the actual license text

	// ISC License - "Permission to use, copy, modify, and distribute" + "WITH OR WITHOUT FEE"
	if strings.Contains(contentUpper, "PERMISSION TO USE, COPY, MODIFY, AND DISTRIBUTE") &&
		strings.Contains(contentUpper, "WITH OR WITHOUT FEE") {
		return "ISC"
	}

	// MIT License - "Permission is hereby granted, free of charge"
	if strings.Contains(contentUpper, "PERMISSION IS HEREBY GRANTED, FREE OF CHARGE") {
		return "MIT"
	}

	// Apache 2.0 - "Apache License" + "Version 2.0"
	if strings.Contains(contentUpper, "APACHE LICENSE") &&
		strings.Contains(contentUpper, "VERSION 2.0") {
		return "Apache-2.0"
	}

	// BSD-3-Clause - Three conditions including "neither the name" clause
	if strings.Contains(contentUpper, "REDISTRIBUTION AND USE IN SOURCE AND BINARY FORMS") &&
		strings.Contains(contentUpper, "NEITHER THE NAME") &&
		strings.Contains(contentUpper, "THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS") {
		return "BSD-3-Clause"
	}

	// BSD-2-Clause - Two conditions, no "neither the name" clause
	if strings.Contains(contentUpper, "REDISTRIBUTION AND USE IN SOURCE AND BINARY FORMS") &&
		strings.Contains(contentUpper, "THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS") &&
		!strings.Contains(contentUpper, "NEITHER THE NAME") {
		return "BSD-2-Clause"
	}

	// MPL-2.0
	if strings.Contains(contentUpper, "MOZILLA PUBLIC LICENSE") &&
		strings.Contains(contentUpper, "VERSION 2.0") {
		return "MPL-2.0"
	}

	// LGPL-3.0
	if strings.Contains(contentUpper, "GNU LESSER GENERAL PUBLIC LICENSE") &&
		strings.Contains(contentUpper, "VERSION 3") {
		return "LGPL-3.0"
	}

	// LGPL-2.1
	if strings.Contains(contentUpper, "GNU LESSER GENERAL PUBLIC LICENSE") &&
		strings.Contains(contentUpper, "VERSION 2.1") {
		return "LGPL-2.1"
	}

	// AGPL-3.0
	if strings.Contains(contentUpper, "GNU AFFERO GENERAL PUBLIC LICENSE") &&
		strings.Contains(contentUpper, "VERSION 3") {
		return "AGPL-3.0"
	}

	// GPL-3.0
	if strings.Contains(contentUpper, "GNU GENERAL PUBLIC LICENSE") &&
		strings.Contains(contentUpper, "VERSION 3") {
		return "GPL-3.0"
	}

	// GPL-2.0
	if strings.Contains(contentUpper, "GNU GENERAL PUBLIC LICENSE") &&
		strings.Contains(contentUpper, "VERSION 2") {
		return "GPL-2.0"
	}

	// Unlicense
	if strings.Contains(contentUpper, "THIS IS FREE AND UNENCUMBERED SOFTWARE") &&
		strings.Contains(contentUpper, "PUBLIC DOMAIN") {
		return "Unlicense"
	}

	// CC0-1.0
	if strings.Contains(contentUpper, "CC0 1.0 UNIVERSAL") ||
		strings.Contains(contentUpper, "CREATIVE COMMONS ZERO") {
		return "CC0-1.0"
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
		// If module is not in go.mod at all, it's a transitive dependency (indirect)
		isIndirect, inGoMod := indirectMap[mod.Path]
		if !inGoMod {
			// Module not in go.mod means it's a transitive dependency
			isIndirect = true
		}

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

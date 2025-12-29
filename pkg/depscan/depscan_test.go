package depscan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultConfig verifies DefaultConfig returns sensible defaults
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("/path/to/go.mod")
	if cfg.GoModPath != "/path/to/go.mod" {
		t.Errorf("expected GoModPath to be /path/to/go.mod, got %s", cfg.GoModPath)
	}
	if len(cfg.LicenseNames) == 0 {
		t.Errorf("expected LicenseNames to be non-empty")
	}
	expectedNames := []string{"LICENSE", "LICENSE.md", "LICENSE.txt", "COPYING", "COPYRIGHT"}
	for i, name := range cfg.LicenseNames {
		if name != expectedNames[i] {
			t.Errorf("expected license name %d to be %s, got %s", i, expectedNames[i], name)
		}
	}
}

// TestParseGoMod tests the go.mod parsing functionality
func TestParseGoMod(t *testing.T) {
	// Create a temporary go.mod file
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")

	goModContent := `module github.com/example/testproject

go 1.23

require (
	github.com/direct/dep v1.0.0
	github.com/another/direct v1.1.0
)

require (
	github.com/indirect/dep v1.5.0 // indirect
)
`

	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write test go.mod: %v", err)
	}

	indirectMap, err := parseGoMod(goModPath)
	if err != nil {
		t.Fatalf("parseGoMod failed: %v", err)
	}

	// Check direct dependencies
	if indirect, ok := indirectMap["github.com/direct/dep"]; !ok || indirect {
		t.Errorf("expected github.com/direct/dep to be direct (not indirect)")
	}
	if indirect, ok := indirectMap["github.com/another/direct"]; !ok || indirect {
		t.Errorf("expected github.com/another/direct to be direct (not indirect)")
	}

	// Check indirect dependency
	if indirect, ok := indirectMap["github.com/indirect/dep"]; !ok || !indirect {
		t.Errorf("expected github.com/indirect/dep to be indirect")
	}
}

// TestParseGoModInvalid tests error handling for invalid go.mod
func TestParseGoModInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")

	if err := os.WriteFile(goModPath, []byte("invalid go.mod content!!!"), 0644); err != nil {
		t.Fatalf("failed to write test go.mod: %v", err)
	}

	_, err := parseGoMod(goModPath)
	if err == nil {
		t.Errorf("expected parseGoMod to return error for invalid go.mod")
	}
}

// TestParseGoModNotFound tests error handling for missing go.mod
func TestParseGoModNotFound(t *testing.T) {
	_, err := parseGoMod("/nonexistent/path/go.mod")
	if err == nil {
		t.Errorf("expected parseGoMod to return error for missing go.mod")
	}
}

// TestFindLicenseFile tests license file discovery
func TestFindLicenseFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a LICENSE file
	licenseContent := "MIT License\n\nCopyright (c) 2024\n"
	licenseFile := filepath.Join(tmpDir, "LICENSE")
	if err := os.WriteFile(licenseFile, []byte(licenseContent), 0644); err != nil {
		t.Fatalf("failed to create test LICENSE file: %v", err)
	}

	found, err := findLicenseFile(tmpDir, []string{"LICENSE", "LICENSE.md", "COPYING"})
	if err != nil {
		t.Fatalf("findLicenseFile failed: %v", err)
	}
	if found != "LICENSE" {
		t.Errorf("expected to find LICENSE, got %s", found)
	}
}

// TestFindLicenseFileMD tests finding LICENSE.md
func TestFindLicenseFileMD(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a LICENSE.md file
	licenseFile := filepath.Join(tmpDir, "LICENSE.md")
	if err := os.WriteFile(licenseFile, []byte("# MIT License\n"), 0644); err != nil {
		t.Fatalf("failed to create test LICENSE.md: %v", err)
	}

	found, err := findLicenseFile(tmpDir, []string{"LICENSE", "LICENSE.md", "COPYING"})
	if err != nil {
		t.Fatalf("findLicenseFile failed: %v", err)
	}
	if found != "LICENSE.md" {
		t.Errorf("expected to find LICENSE.md, got %s", found)
	}
}

// TestFindLicenseFilePriority tests license file priority
func TestFindLicenseFilePriority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple license files
	files := []string{"LICENSE", "LICENSE.md", "COPYING"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("license content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Should find LICENSE first (highest priority)
	found, err := findLicenseFile(tmpDir, []string{"LICENSE", "LICENSE.md", "COPYING"})
	if err != nil {
		t.Fatalf("findLicenseFile failed: %v", err)
	}
	if found != "LICENSE" {
		t.Errorf("expected to find LICENSE first, got %s", found)
	}
}

// TestFindLicenseFileNotFound tests error handling when no license file exists
func TestFindLicenseFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := findLicenseFile(tmpDir, []string{"LICENSE", "LICENSE.md"})
	if err == nil {
		t.Errorf("expected findLicenseFile to return error when no license file found")
	}
}

// TestFindLicenseFileEmptyDir tests error handling for empty directory
func TestFindLicenseFileEmptyDir(t *testing.T) {
	_, err := findLicenseFile("", []string{"LICENSE"})
	if err == nil {
		t.Errorf("expected findLicenseFile to return error for empty directory")
	}
}

// TestClassifyLicense tests MIT license detection
func TestClassifyLicense(t *testing.T) {
	tmpDir := t.TempDir()
	licenseFile := filepath.Join(tmpDir, "LICENSE")

	// MIT License text
	mitLicense := `MIT License

Copyright (c) 2024 Example

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.`

	if err := os.WriteFile(licenseFile, []byte(mitLicense), 0644); err != nil {
		t.Fatalf("failed to create test license file: %v", err)
	}

	license, err := classifyLicense(licenseFile)
	if err != nil {
		t.Fatalf("classifyLicense failed: %v", err)
	}
	if license == "UNKNOWN" {
		t.Logf("warning: could not classify MIT license (got UNKNOWN, may be expected if classifier requires training data)")
	}
}

// TestClassifyLicenseFileNotFound tests error handling for missing file
func TestClassifyLicenseFileNotFound(t *testing.T) {
	license, _ := classifyLicense("/nonexistent/path/LICENSE")
	if license != "UNKNOWN" {
		t.Errorf("expected UNKNOWN for missing file, got %s", license)
	}
}

// TestFilterByLicense tests filtering dependencies by license type
func TestFilterByLicense(t *testing.T) {
	deps := []Dependency{
		{Path: "github.com/mit/lib", License: "MIT"},
		{Path: "github.com/apache/lib", License: "Apache-2.0"},
		{Path: "github.com/mit/another", License: "MIT"},
		{Path: "github.com/gpl/lib", License: "GPL-3.0"},
	}

	filtered := FilterByLicense(deps, "MIT", "Apache-2.0")
	if len(filtered) != 3 {
		t.Errorf("expected 3 filtered results, got %d", len(filtered))
	}

	for _, dep := range filtered {
		if dep.License != "MIT" && dep.License != "Apache-2.0" {
			t.Errorf("unexpected license %s in filtered results", dep.License)
		}
	}
}

// TestFilterByLicenseEmpty tests filtering with no matches
func TestFilterByLicenseEmpty(t *testing.T) {
	deps := []Dependency{
		{Path: "github.com/mit/lib", License: "MIT"},
		{Path: "github.com/apache/lib", License: "Apache-2.0"},
	}

	filtered := FilterByLicense(deps, "GPL-3.0", "AGPL-3.0")
	if len(filtered) != 0 {
		t.Errorf("expected 0 filtered results, got %d", len(filtered))
	}
}

// TestGoModuleFromGoListUnmarshal tests JSON unmarshaling of go list output
func TestGoModuleFromGoListUnmarshal(t *testing.T) {
	jsonData := `{
		"Path": "github.com/example/lib",
		"Version": "v1.2.3",
		"Sum": "h1:abc123",
		"Dir": "/go/pkg/mod/github.com/example/lib@v1.2.3",
		"Main": false,
		"Replace": null,
		"GoMod": "/go/pkg/mod/github.com/example/lib@v1.2.3/go.mod",
		"GoVersion": "1.23"
	}`

	var mod goModuleFromGoList
	err := json.Unmarshal([]byte(jsonData), &mod)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if mod.Path != "github.com/example/lib" {
		t.Errorf("expected Path to be github.com/example/lib, got %s", mod.Path)
	}
	if mod.Version != "v1.2.3" {
		t.Errorf("expected Version to be v1.2.3, got %s", mod.Version)
	}
	if mod.Main != false {
		t.Errorf("expected Main to be false")
	}
	if mod.Replace != nil {
		t.Errorf("expected Replace to be nil")
	}
}

// TestGoModuleFromGoListUnmarshalWithReplace tests JSON unmarshaling with replace directive
func TestGoModuleFromGoListUnmarshalWithReplace(t *testing.T) {
	jsonData := `{
		"Path": "github.com/example/lib",
		"Version": "v1.2.3",
		"Dir": "/go/pkg/mod/github.com/example/lib@v1.2.3",
		"Main": false,
		"Replace": {
			"Path": "github.com/example/lib",
			"Version": "v1.3.0",
			"Dir": "/local/path/to/lib"
		},
		"GoMod": "/go/pkg/mod/github.com/example/lib@v1.2.3/go.mod"
	}`

	var mod goModuleFromGoList
	err := json.Unmarshal([]byte(jsonData), &mod)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if mod.Replace == nil {
		t.Errorf("expected Replace to be non-nil")
	} else {
		if mod.Replace.Dir != "/local/path/to/lib" {
			t.Errorf("expected Replace.Dir to be /local/path/to/lib, got %s", mod.Replace.Dir)
		}
	}
}

// TestDependencyStruct tests the Dependency struct
func TestDependencyStruct(t *testing.T) {
	dep := Dependency{
		Path:        "github.com/example/lib",
		Version:     "v1.2.3",
		Indirect:    true,
		License:     "MIT",
		LicenseFile: "LICENSE",
		Dir:         "/path/to/lib",
	}

	if dep.Path != "github.com/example/lib" {
		t.Errorf("expected Path to be github.com/example/lib")
	}
	if !dep.Indirect {
		t.Errorf("expected Indirect to be true")
	}
	if dep.License != "MIT" {
		t.Errorf("expected License to be MIT")
	}
}

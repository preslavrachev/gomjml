package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type ComponentDefaults struct {
	Name       string            `json:"-"`
	Attributes map[string]string `json:"-"`
}

type ExtractorConfig struct {
	InputPath  string
	OutputPath string
	Verbose    bool
}

var (
	// Regex to find default_attribute function
	defaultAttrFuncRegex = regexp.MustCompile(`fn\s+default_attribute\s*\([^)]*\)\s*->[^{]*\{([^}]+)\}`)

	// Regex to extract match arms
	matchArmRegex = regexp.MustCompile(`"([^"]+)"\s*=>\s*Some\("([^"]+)"\)`)

	// Regex to extract match arms with constants
	matchArmConstRegex = regexp.MustCompile(`"([^"]+)"\s*=>\s*Some\(([A-Z_][A-Z0-9_]*)\)`)

	// Regex to find NAME constant
	nameConstRegex = regexp.MustCompile(`pub\s+const\s+NAME:\s*&str\s*=\s*"([^"]+)"`)

	// Regex to find constant definitions
	constDefRegex = regexp.MustCompile(`const\s+([A-Z_][A-Z0-9_]*)\s*:\s*&str\s*=\s*"([^"]+)"`)
)

func main() {
	config := parseFlags()

	if config.Verbose {
		log.Printf("Extracting default attributes from: %s", config.InputPath)
	}

	components, err := extractDefaultAttributes(config)
	if err != nil {
		log.Fatalf("Error extracting attributes: %v", err)
	}

	if config.Verbose {
		log.Printf("Found %d components with default attributes", len(components))
	}

	err = writeJSONOutput(components, config.OutputPath, config.Verbose)
	if err != nil {
		log.Fatalf("Error writing output: %v", err)
	}

	if config.Verbose {
		log.Printf("Successfully wrote default attributes to: %s", config.OutputPath)
	}
}

func parseFlags() ExtractorConfig {
	var config ExtractorConfig

	flag.StringVar(&config.InputPath, "input", "mrml/packages/mrml-core/src", "Path to MRML source directory")
	flag.StringVar(&config.OutputPath, "output", "default-css-attrs.json", "Output JSON file path")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	return config
}

func extractDefaultAttributes(config ExtractorConfig) (map[string]map[string]string, error) {
	components := make(map[string]map[string]string)

	// Find all mj_* directories
	componentDirs, err := findComponentDirectories(config.InputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find component directories: %w", err)
	}

	if config.Verbose {
		log.Printf("Found %d component directories", len(componentDirs))
	}

	for _, dir := range componentDirs {
		componentName, defaults, err := extractComponentDefaults(dir, config.Verbose)
		if err != nil {
			if config.Verbose {
				log.Printf("Skipping %s: %v", dir, err)
			}
			continue
		}

		if len(defaults) > 0 {
			components[componentName] = defaults
			if config.Verbose {
				log.Printf("Extracted %d defaults for %s", len(defaults), componentName)
			}
		}
	}

	return components, nil
}

func findComponentDirectories(basePath string) ([]string, error) {
	var dirs []string

	err := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && strings.HasPrefix(d.Name(), "mj_") {
			dirs = append(dirs, path)
		}

		return nil
	})

	return dirs, err
}

func extractComponentDefaults(componentDir string, verbose bool) (string, map[string]string, error) {
	// First, get the component name from mod.rs
	componentName, err := getComponentName(componentDir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get component name: %w", err)
	}

	// Read render.rs file
	renderFile := filepath.Join(componentDir, "render.rs")
	content, err := os.ReadFile(renderFile)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read render.rs: %w", err)
	}

	// Extract constants first
	constants := extractConstants(string(content))

	// Extract default_attribute function
	defaults, err := parseDefaultAttributeFunction(string(content), constants, verbose)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse default_attribute function: %w", err)
	}

	return componentName, defaults, nil
}

func getComponentName(componentDir string) (string, error) {
	modFile := filepath.Join(componentDir, "mod.rs")
	content, err := os.ReadFile(modFile)
	if err != nil {
		return "", err
	}

	matches := nameConstRegex.FindStringSubmatch(string(content))
	if len(matches) != 2 {
		return "", fmt.Errorf("NAME constant not found in mod.rs")
	}

	return matches[1], nil
}

func extractConstants(content string) map[string]string {
	constants := make(map[string]string)

	matches := constDefRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) == 3 {
			constants[match[1]] = match[2]
		}
	}

	return constants
}

func parseDefaultAttributeFunction(content string, constants map[string]string, verbose bool) (map[string]string, error) {
	// Find the default_attribute function
	funcMatches := defaultAttrFuncRegex.FindStringSubmatch(content)
	if len(funcMatches) != 2 {
		return nil, fmt.Errorf("default_attribute function not found")
	}

	funcBody := funcMatches[1]
	defaults := make(map[string]string)

	// Extract direct string matches
	stringMatches := matchArmRegex.FindAllStringSubmatch(funcBody, -1)
	for _, match := range stringMatches {
		if len(match) == 3 {
			defaults[match[1]] = match[2]
		}
	}

	// Extract constant matches
	constMatches := matchArmConstRegex.FindAllStringSubmatch(funcBody, -1)
	for _, match := range constMatches {
		if len(match) == 3 {
			key := match[1]
			constName := match[2]
			if value, exists := constants[constName]; exists {
				defaults[key] = value
			} else if verbose {
				log.Printf("Warning: constant %s not found", constName)
			}
		}
	}

	return defaults, nil
}

func writeJSONOutput(components map[string]map[string]string, outputPath string, verbose bool) error {
	// Create ordered output for consistent JSON
	type orderedComponent struct {
		Name       string            `json:"-"`
		Attributes map[string]string `json:"-"`
	}

	// Sort component names
	var names []string
	for name := range components {
		names = append(names, name)
	}
	sort.Strings(names)

	// Create ordered output map
	output := make(map[string]map[string]string)
	totalProps := 0

	for _, name := range names {
		// Sort attributes within each component
		attrs := components[name]
		var keys []string
		for key := range attrs {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		sortedAttrs := make(map[string]string)
		for _, key := range keys {
			sortedAttrs[key] = attrs[key]
		}

		output[name] = sortedAttrs
		totalProps += len(sortedAttrs)
	}

	// Write JSON with proper formatting
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if verbose {
		fmt.Printf("Statistics:\n")
		fmt.Printf("  Components: %d\n", len(components))
		fmt.Printf("  Total properties: %d\n", totalProps)
		fmt.Printf("  Average properties per component: %.1f\n", float64(totalProps)/float64(len(components)))
	}

	return nil
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Color constants for diff output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

type Config struct {
	TestCase     string
	KeepFiles    bool
	Verbose      bool
	DiffLines    int
	ScriptDir    string
	ProjectRoot  string
	OutputDir    string
	GomjmlBinary string
	TestDataDir  string
}

func main() {
	var config Config

	flag.StringVar(&config.TestCase, "test", "", "Test case name (required)")
	flag.StringVar(&config.TestDataDir, "testdata-dir", "", "Path to testdata directory (defaults to mjml/testdata)")
	flag.BoolVar(&config.KeepFiles, "keep-files", false, "Keep temporary files for inspection")
	flag.BoolVar(&config.KeepFiles, "k", false, "Keep temporary files for inspection (short)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Show more diff context (50 lines instead of 20)")
	flag.BoolVar(&config.Verbose, "v", false, "Show more diff context (short)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] test-case-name\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Compares reference HTML vs gomjml output for test cases\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # From mjml/testdata directory:\n")
		fmt.Fprintf(os.Stderr, "  %s basic                           # Compare basic.mjml vs basic.html\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s basic --keep-files              # Keep files for inspection\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  # From project root:\n")
		fmt.Fprintf(os.Stderr, "  %s basic --testdata-dir mjml/testdata    # Specify testdata directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s basic -k -v                           # Keep files and verbose output\n", os.Args[0])
	}

	flag.Parse()

	// Handle positional argument
	if flag.NArg() > 0 && config.TestCase == "" {
		config.TestCase = flag.Arg(0)
	}

	if config.TestCase == "" {
		fmt.Fprintf(os.Stderr, "Error: Test case name is required\n")
		flag.Usage()
		os.Exit(1)
	}

	config.DiffLines = 20
	if config.Verbose {
		config.DiffLines = 50
	}

	if err := setupPaths(&config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := runComparison(&config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func setupPaths(config *Config) error {
	// Get script directory (where this binary is located)
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %v", err)
	}
	config.ScriptDir = filepath.Dir(executable)

	// Find project root by looking for go.mod
	config.ProjectRoot = config.ScriptDir
	for config.ProjectRoot != "/" {
		if _, err := os.Stat(filepath.Join(config.ProjectRoot, "go.mod")); err == nil {
			break
		}
		config.ProjectRoot = filepath.Dir(config.ProjectRoot)
	}

	if config.ProjectRoot == "/" {
		return fmt.Errorf("could not find project root (go.mod not found)")
	}

	fmt.Printf("Detected project root: %s\n", config.ProjectRoot)

	// Set up testdata directory
	if config.TestDataDir == "" {
		// Try to detect if we're in testdata directory
		if cwd, err := os.Getwd(); err == nil {
			if filepath.Base(cwd) == "testdata" && filepath.Base(filepath.Dir(cwd)) == "mjml" {
				config.TestDataDir = cwd
			} else {
				config.TestDataDir = filepath.Join(config.ProjectRoot, "mjml", "testdata")
			}
		} else {
			config.TestDataDir = filepath.Join(config.ProjectRoot, "mjml", "testdata")
		}
	}

	config.OutputDir = "./comparison"
	config.GomjmlBinary = filepath.Join(config.ProjectRoot, "bin", "gomjml-debug")

	return nil
}

func runComparison(config *Config) error {
	// Set up file paths using testdata directory
	inputFile := filepath.Join(config.TestDataDir, config.TestCase+".mjml")
	referenceFile := filepath.Join(config.TestDataDir, config.TestCase+".html")

	// Check if both files exist
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("MJML file '%s' not found", inputFile)
	}
	if _, err := os.Stat(referenceFile); os.IsNotExist(err) {
		return fmt.Errorf("reference HTML file '%s' not found", referenceFile)
	}

	fmt.Printf("Comparing reference HTML vs gomjml for test case: %s\n", config.TestCase)
	fmt.Printf("  MJML file: %s\n", inputFile)
	fmt.Printf("  Reference HTML: %s\n", referenceFile)
	fmt.Printf("  Output directory: %s\n", config.OutputDir)

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("could not create output directory: %v", err)
	}

	// Build gomjml
	if err := buildGomjml(config); err != nil {
		return fmt.Errorf("gomjml build failed: %v", err)
	}

	// Generate gomjml output
	gomjmlOutput := filepath.Join(config.OutputDir, config.TestCase+"_gomjml.html")
	if err := compileWithGomjml(config, inputFile, gomjmlOutput); err != nil {
		return fmt.Errorf("gomjml compilation failed: %v", err)
	}

	// Copy reference HTML
	referenceOutput := filepath.Join(config.OutputDir, config.TestCase+"_reference.html")
	if err := copyFile(referenceFile, referenceOutput); err != nil {
		return fmt.Errorf("could not copy reference file: %v", err)
	}

	// Beautify both HTML files
	referencePretty := filepath.Join(config.OutputDir, config.TestCase+"_reference_pretty.html")
	gomjmlPretty := filepath.Join(config.OutputDir, config.TestCase+"_gomjml_pretty.html")

	fmt.Println("Beautifying HTML outputs...")
	if err := beautifyHTML(referenceOutput, referencePretty); err != nil {
		return fmt.Errorf("could not beautify reference HTML: %v", err)
	}
	if err := beautifyHTML(gomjmlOutput, gomjmlPretty); err != nil {
		return fmt.Errorf("could not beautify gomjml HTML: %v", err)
	}

	// Generate diff
	diffOutput := filepath.Join(config.OutputDir, config.TestCase+"_diff.txt")
	if err := generateDiff(referencePretty, gomjmlPretty, diffOutput, config); err != nil {
		return fmt.Errorf("could not generate diff: %v", err)
	}

	fmt.Println("Comparison complete!")
	return nil
}

func buildGomjml(config *Config) error {
	fmt.Println("Building gomjml with debug tags...")

	cmd := exec.Command("go", "build", "-tags", "debug", "-o", config.GomjmlBinary, "./cmd/gomjml")
	cmd.Dir = config.ProjectRoot

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Build output: %s\n", string(output))
		return err
	}

	return nil
}

func compileWithGomjml(config *Config, inputFile, outputFile string) error {
	fmt.Println("Compiling with gomjml...")

	cmd := exec.Command(config.GomjmlBinary, "compile", inputFile, "-o", outputFile)

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Compilation output: %s\n", string(output))
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	fmt.Printf("Using reference HTML file: %s\n", src)

	cmd := exec.Command("cp", src, dst)
	return cmd.Run()
}

func beautifyHTML(inputFile, outputFile string) error {
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	beautified := normalizeHTML(string(content))

	return os.WriteFile(outputFile, []byte(beautified), 0o644)
}

func normalizeHTML(content string) string {
	// Normalize whitespace first
	whitespaceRe := regexp.MustCompile(`\s+`)
	content = whitespaceRe.ReplaceAllString(content, " ")
	content = strings.TrimSpace(content)

	// Key fix: Separate text content from tags consistently
	// This handles both '>Hello World!</div>' and '>\n  Hello World!\n  </div>'
	textContentRe := regexp.MustCompile(`(>)\s*([^<\s][^<]*?)\s*(<)`)
	content = textContentRe.ReplaceAllString(content, "$1\n$2\n$3")

	// Add newlines around all tags
	tagRe := regexp.MustCompile(`(<[^>]*>)`)
	content = tagRe.ReplaceAllString(content, "\n$1\n")

	// Clean up multiple newlines
	multiNewlineRe := regexp.MustCompile(`\n+`)
	content = multiNewlineRe.ReplaceAllString(content, "\n")

	lines := strings.Split(content, "\n")
	var result []string
	indent := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Decrease indent for closing tags
		if strings.HasPrefix(line, "</") {
			if indent > 0 {
				indent--
			}
		}

		// Add indented line
		result = append(result, strings.Repeat("  ", indent)+line)

		// Increase indent for opening tags (but not self-closing or closing)
		if strings.HasPrefix(line, "<") && !strings.HasPrefix(line, "</") && !strings.HasSuffix(line, "/>") {
			// Don't increase indent for inline tags
			lowerLine := strings.ToLower(line)
			if !strings.Contains(lowerLine, "<br") && !strings.Contains(lowerLine, "<img") &&
				!strings.Contains(lowerLine, "<input") && !strings.Contains(lowerLine, "<meta") &&
				!strings.Contains(lowerLine, "<link") {
				indent++
			}
		}
	}

	return strings.Join(result, "\n")
}

func generateDiff(referenceFile, gomjmlFile, outputFile string, config *Config) error {
	fmt.Println("Generating diff...")

	// Use semantic diff - ignore whitespace differences
	cmd := exec.Command("diff", "-u", "-w", "-B", referenceFile, gomjmlFile)

	output, _ := cmd.CombinedOutput()
	diffContent := string(output)

	// Write diff to file
	if err := os.WriteFile(outputFile, output, 0o644); err != nil {
		return err
	}

	if len(diffContent) > 0 && strings.TrimSpace(diffContent) != "" {
		fmt.Printf("Differences found! Check: %s\n", outputFile)
		fmt.Println("Files generated:")
		fmt.Printf("  Reference output (pretty): %s\n", referenceFile)
		fmt.Printf("  gomjml output (pretty): %s\n", gomjmlFile)
		fmt.Printf("  Diff: %s\n", outputFile)

		// Show preview of diff
		fmt.Printf("\nDiff preview (first %d lines):\n", config.DiffLines)
		showDiffPreview(diffContent, config.DiffLines)

		if !config.KeepFiles {
			cleanup(config)
		} else {
			fmt.Println("\nFiles preserved for inspection:")
			fmt.Printf("  Reference output: %s\n", filepath.Join(config.OutputDir, config.TestCase+"_reference.html"))
			fmt.Printf("  gomjml output: %s\n", filepath.Join(config.OutputDir, config.TestCase+"_gomjml.html"))
			fmt.Printf("  Reference output (pretty): %s\n", referenceFile)
			fmt.Printf("  gomjml output (pretty): %s\n", gomjmlFile)
			fmt.Printf("  Diff: %s\n", outputFile)
		}
	} else {
		fmt.Println("No differences found! HTML outputs are identical.")
		if !config.KeepFiles {
			cleanup(config)
		} else {
			fmt.Println("Files preserved for inspection:")
			fmt.Printf("  Reference output: %s\n", filepath.Join(config.OutputDir, config.TestCase+"_reference.html"))
			fmt.Printf("  gomjml output: %s\n", filepath.Join(config.OutputDir, config.TestCase+"_gomjml.html"))
			fmt.Printf("  Reference output (pretty): %s\n", referenceFile)
			fmt.Printf("  gomjml output (pretty): %s\n", gomjmlFile)
		}
	}

	return nil
}

func showDiffPreview(content string, maxLines int) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineCount := 0

	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()
		coloredLine := colorDiffLine(line)
		fmt.Println(coloredLine)
		lineCount++
	}
}

func colorDiffLine(line string) string {
	if len(line) == 0 {
		return line
	}

	switch line[0] {
	case '+':
		return ColorGreen + line + ColorReset
	case '-':
		return ColorRed + line + ColorReset
	case '@':
		if strings.HasPrefix(line, "@@") {
			return ColorYellow + line + ColorReset
		}
		return line
	default:
		return line
	}
}

func cleanup(config *Config) {
	fmt.Println("\nCleaning up temporary files...")

	files := []string{
		filepath.Join(config.OutputDir, config.TestCase+"_reference.html"),
		filepath.Join(config.OutputDir, config.TestCase+"_gomjml.html"),
		filepath.Join(config.OutputDir, config.TestCase+"_reference_pretty.html"),
		filepath.Join(config.OutputDir, config.TestCase+"_gomjml_pretty.html"),
		filepath.Join(config.OutputDir, config.TestCase+"_diff.txt"),
	}

	for _, file := range files {
		os.Remove(file) // Ignore errors
	}

	// Remove directory if empty
	os.Remove(config.OutputDir)

	fmt.Println("Temporary files cleaned up.")
}

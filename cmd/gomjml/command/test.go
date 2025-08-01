package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewTestCommand creates the test command
func NewTestCommand() *cobra.Command {
	var (
		verbose bool
		pattern string
	)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run test suite against MRML",
		Long: `Run the Go MJML implementation test suite, comparing output against MRML (Rust implementation).

This command runs the integration tests that compare the Go implementation output
with the reference MRML implementation to ensure compatibility.

Examples:
  gomjml test                    # Run all tests
  gomjml test -v                 # Run with verbose output
  gomjml test -pattern "basic"   # Run tests matching pattern`,
		Run: func(cmd *cobra.Command, args []string) {
			// Change to the mjml package directory to run tests
			mjmlDir := filepath.Join("mjml")

			// Build go test command
			testCmd := exec.Command("go", "test")
			if verbose {
				testCmd.Args = append(testCmd.Args, "-v")
			}
			if pattern != "" {
				testCmd.Args = append(testCmd.Args, "-run", pattern)
			}
			testCmd.Args = append(testCmd.Args, "./"+mjmlDir)

			// Set up command execution
			testCmd.Stdout = os.Stdout
			testCmd.Stderr = os.Stderr
			testCmd.Dir = "."

			// Run tests
			fmt.Printf("Running Go MJML tests...\n")
			err := testCmd.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Tests failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("All tests passed!")
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose test output")
	cmd.Flags().StringVarP(&pattern, "pattern", "p", "", "run only tests matching pattern")

	return cmd
}

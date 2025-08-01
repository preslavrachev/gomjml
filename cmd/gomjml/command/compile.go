package command

import (
	"fmt"
	"os"

	"github.com/preslavrachev/gomjml/mjml"
	"github.com/spf13/cobra"
)

// NewCompileCommand creates the compile command
func NewCompileCommand() *cobra.Command {
	var (
		outputFile string
		stdout     bool
		debug      bool
	)

	cmd := &cobra.Command{
		Use:   "compile [input]",
		Short: "Compile MJML to HTML",
		Long: `Compile MJML markup to responsive HTML.

Examples:
  gomjml compile input.mjml -o output.html
  gomjml compile input.mjml -s
  gomjml compile input.mjml --debug`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			inputFile := args[0]

			// Read MJML file
			mjmlContent, err := os.ReadFile(inputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				os.Exit(1)
			}

			// Render MJML to HTML using library
			var html string
			if debug {
				html, err = mjml.Render(string(mjmlContent), mjml.WithDebugTags(true))
			} else {
				html, err = mjml.Render(string(mjmlContent))
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error rendering MJML: %v\n", err)
				os.Exit(1)
			}

			// Output HTML
			if outputFile != "" {
				err := os.WriteFile(outputFile, []byte(html), 0o644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Print(html)
			}
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path")
	cmd.Flags().BoolVarP(&stdout, "stdout", "s", false, "output to stdout")
	cmd.Flags().BoolVar(&debug, "debug", false, "include debug attributes in output")

	return cmd
}

package command

import (
	"fmt"
	"os"
	"time"

	"github.com/preslavrachev/gomjml/mjml"
	"github.com/spf13/cobra"
)

// NewCompileCommand creates the compile command
func NewCompileCommand() *cobra.Command {
	var (
		outputFile    string
		stdout        bool
		debug         bool
		cache         bool
		cacheTTL      time.Duration
		cacheInterval time.Duration
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

			if cacheTTL > 0 {
				mjml.SetASTCacheTTLOnce(cacheTTL)
			}
			if cacheInterval > 0 {
				mjml.SetASTCacheCleanupIntervalOnce(cacheInterval)
			}

			// Render MJML to HTML using library
			opts := []mjml.RenderOption{}
			if debug {
				opts = append(opts, mjml.WithDebugTags(true))
			}
			if cache {
				opts = append(opts, mjml.WithCache())
			}
			html, err := mjml.Render(string(mjmlContent), opts...)
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
	cmd.Flags().BoolVar(&cache, "cache", false, "enable experimental AST caching")
	cmd.Flags().DurationVar(&cacheTTL, "cache-ttl", 0, "AST cache TTL (e.g. 10m)")
	cmd.Flags().DurationVar(&cacheInterval, "cache-cleanup-interval", 0, "AST cache cleanup interval")

	return cmd
}

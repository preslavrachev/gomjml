package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute runs the root command
func Execute() {
	rootCmd := &cobra.Command{
		Use:   "gomjml",
		Short: "MJML compiler written in Go - converts MJML to responsive HTML",
		Long: `gomjml is a native Go implementation of the MJML email framework.
It compiles MJML markup into responsive HTML suitable for email clients.

Available Commands:
  compile    Compile MJML to HTML (default)
  test       Run test suite against MRML
  version    Show version information`,
	}

	// Add subcommands
	rootCmd.AddCommand(NewCompileCommand())
	rootCmd.AddCommand(NewTestCommand())

	// If no command is specified, default to compile
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		// Default to compile command behavior
		compileCmd := NewCompileCommand()
		compileCmd.Run(cmd, args)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

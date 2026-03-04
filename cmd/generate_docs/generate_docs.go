package generate_docs

import (
	"fmt"
	"os"

	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var Cmd = &cobra.Command{
	Use:     "generate-docs",
	Short:   "Generate Markdown docs",
	Long:    "This command generates Markdown documentation for the CLI commands using Cobra's doc generator.",
	Aliases: []string{"gen-docs", "docs"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		path := "./cobra-docs/"
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating docs directory: %v\n", err)
			os.Exit(1)
		}
		if err := doc.GenMarkdownTree(root.RootCmd, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating docs: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}

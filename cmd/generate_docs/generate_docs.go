package generate_docs

import (
	"os"

	"github.com/VojtechPastyrik/vp-utils/cmd/root"
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
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			panic(err)
		}
		err = doc.GenMarkdownTree(root.RootCmd, path)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}

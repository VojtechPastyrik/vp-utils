package remove_context

import (
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/k8s"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

var FlagContextName string

var Cmd = &cobra.Command{
	Use:     "remove-kubeconfig",
	Short:   "Remove context from kubeconfig",
	Aliases: []string{"rc"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		removeContext(FlagContextName)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagContextName, "context-name",
		"c",
		"",
		"Context name")
	Cmd.MarkFlagRequired("context-name")
}

func removeContext(contextName string) {
	cmd := exec.Command("kubectl", "config", "unset", "contexts."+contextName)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Nepodařilo se odebrat kontext %s: %v", contextName, err)
	} else {
		log.Printf("Kontext %s byl úspěšně odebrán.", contextName)
	}
}

package remove_context

import (
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/k8s"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "remove-context <name>",
	Short:   "Remove context from kubeconfig",
	Aliases: []string{"rc"},
	Args:    cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		contextName := args[0]
		removeContext(contextName)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

func removeContext(contextName string) {
	cmd := exec.Command("kubectl", "config", "unset", "contexts."+contextName)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to remove context %s: %v", contextName, err)
	} else {
		log.Printf("Context %s was successfully removed.", contextName)
	}
}

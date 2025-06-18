package remove_context

import (
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/k8s"
	"log"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
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
	// Load rules for kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configPath := loadingRules.GetDefaultFilename()

	// Load configuration
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// Check if context exists
	if _, exists := config.Contexts[contextName]; !exists {
		log.Fatalf("Context %s doesn't exist", contextName)
	}

	// Remove context
	delete(config.Contexts, contextName)

	// Clear current context if it was deleted
	if config.CurrentContext == contextName {
		config.CurrentContext = ""
	}

	// Save modified configuration
	if err := clientcmd.WriteToFile(*config, configPath); err != nil {
		log.Fatalf("Failed to write to kubeconfig: %v", err)
	} else {
		log.Printf("Context %s was successfully removed.", contextName)
	}
}

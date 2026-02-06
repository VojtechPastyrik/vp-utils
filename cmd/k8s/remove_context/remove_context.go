package remove_context

import (
	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/k8s"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
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
		logger.Fatalf("failed to load kubeconfig: %v", err)
	}

	// Check if context exists
	if _, exists := config.Contexts[contextName]; !exists {
		logger.Fatalf("context %s doesn't exist", contextName)
	}

	// Remove context
	delete(config.Contexts, contextName)

	// Clear current context if it was deleted
	if config.CurrentContext == contextName {
		config.CurrentContext = ""
	}

	// Save modified configuration
	if err := clientcmd.WriteToFile(*config, configPath); err != nil {
		logger.Fatalf("failed to write to kubeconfig: %v", err)
	} else {
		logger.Infof("context %s was successfully removed.", contextName)
	}
}

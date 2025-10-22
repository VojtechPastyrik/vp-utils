package add_context

import (
	"os"
	"path/filepath"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/k8s"
	"github.com/VojtechPastyrik/vp-utils/pkg/logger"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var FlagFilePath string

var Cmd = &cobra.Command{
	Use:     "add-context",
	Short:   "Add a context to the kubeconfig file. This command merges the specified kubeconfig file into the default kubeconfig file located at ~/.kube/config.",
	Aliases: []string{"ac"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		mergeKubeConfigs(FlagFilePath)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagFilePath, "file-path",
		"f",
		"",
		"Path to the kubeconfig file to merge into ~/.kube/config")
	Cmd.MarkFlagRequired("file-path")
}

func mergeKubeConfigs(filePath string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Fatalf("failed to get home directory: %v", err)
	}

	kubeDir := filepath.Join(homeDir, ".kube")
	existingConfigPath := filepath.Join(kubeDir, "config")

	// Check if input file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Fatalf("input kubeconfig file does not exist: %s", filePath)
	}

	logger.Infof("merging kubeconfig from: %s", filePath)

	// Create .kube directory if it doesn't exist
	if _, err := os.Stat(kubeDir); os.IsNotExist(err) {
		if err := os.MkdirAll(kubeDir, 0755); err != nil {
			logger.Fatalf("failed to create .kube directory: %v", err)
		}
	}

	// Load the existing config
	existingConfig := clientcmdapi.NewConfig()
	if _, err := os.Stat(existingConfigPath); !os.IsNotExist(err) {
		existingConfig, err = clientcmd.LoadFromFile(existingConfigPath)
		if err != nil {
			logger.Fatalf("failed to load existing kubeconfig: %v", err)
		}
	}

	// Load the new config
	newConfig, err := clientcmd.LoadFromFile(filePath)
	if err != nil {
		logger.Fatalf("failed to load new kubeconfig: %v", err)
	}

	// Merge configs
	// Add clusters from new config
	for name, cluster := range newConfig.Clusters {
		existingConfig.Clusters[name] = cluster
	}

	// Add auth info from new config
	for name, authInfo := range newConfig.AuthInfos {
		existingConfig.AuthInfos[name] = authInfo
	}

	// Add contexts from new config
	for name, context := range newConfig.Contexts {
		existingConfig.Contexts[name] = context
	}

	// Preserve existing current context if set
	if existingConfig.CurrentContext == "" && newConfig.CurrentContext != "" {
		existingConfig.CurrentContext = newConfig.CurrentContext
	}

	// Save the merged config
	if err := clientcmd.WriteToFile(*existingConfig, existingConfigPath); err != nil {
		logger.Fatalf("failed to save merged kubeconfig: %v", err)
	}

	logger.Infof("kubeconfig successfully updated and merged with context from: %s", filePath)
}

package add_context

import (
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/k8s"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
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
	// Merge the kubeconfig file specified by filePath into the default kubeconfig location
	mergedConfigPath := os.Getenv("HOME") + "/.kube/merged-config"
	cmd := exec.Command("kubectl", "config", "view", "--kubeconfig", filePath, "--merge", "--flatten", "-o", mergedConfigPath)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to merge kubeconfig files: %v", err)
	}

	// Move the merged config to the final location
	finalConfigPath := os.Getenv("HOME") + "/.kube/config"
	if err := os.Rename(mergedConfigPath, finalConfigPath); err != nil {
		log.Fatalf("Failed to move merged config: %v", err)
	}

	log.Println("Kubeconfig successfully updated and merged.")
}

package add_context

import (
	"bytes"
	"fmt"
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/k8s"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	kubeDir := filepath.Join(homeDir, ".kube")
	existingConfigPath := filepath.Join(kubeDir, "config")
	mergedConfigPath := filepath.Join(kubeDir, "merged-config")

	// Check if input file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Input kubeconfig file does not exist: %s", filePath)
	}

	log.Printf("Merging kubeconfig from: %s", filePath)

	// Create .kube directory if it doesn't exist
	if _, err := os.Stat(kubeDir); os.IsNotExist(err) {
		if err := os.MkdirAll(kubeDir, 0755); err != nil {
			log.Fatalf("Failed to create .kube directory: %v", err)
		}
	}

	// Use kubectl directly to merge ONLY the two configs we want
	// Explicitly use only the base config and new file, ignoring KUBECONFIG env var
	cmd := exec.Command("kubectl", "config", "view", "--flatten")

	// Important: Only include these two files, ignore KUBECONFIG from environment
	mergeEnv := fmt.Sprintf("KUBECONFIG=%s:%s", existingConfigPath, filePath)

	// Create a clean environment without inheriting KUBECONFIG
	newEnv := []string{}
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "KUBECONFIG=") {
			newEnv = append(newEnv, env)
		}
	}
	cmd.Env = append(newEnv, mergeEnv)

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to merge kubeconfig files: %v", err)
	}

	// Validate the output contains clusters and contexts sections
	if !bytes.Contains(output, []byte("clusters:")) ||
		!bytes.Contains(output, []byte("contexts:")) {
		log.Fatalf("Generated config appears invalid (missing key sections)")
	}

	// Write output to temporary file
	if err := os.WriteFile(mergedConfigPath, output, 0600); err != nil {
		log.Fatalf("Failed to write merged config: %v", err)
	}

	// Replace existing config with merged config
	if err := os.Rename(mergedConfigPath, existingConfigPath); err != nil {
		log.Fatalf("Failed to move merged config: %v", err)
	}

	log.Printf("Kubeconfig successfully updated and merged with context from: %s", filePath)
}

package init_unseal

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/vault"
	"github.com/spf13/cobra"
)

var FlagPath string
var FlagNamespace string

var Cmd = &cobra.Command{
	Use:   "vault-init-unseal",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		vaultInitUnseal(FlagPath, FlagNamespace)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.MarkFlagRequired("path")
	Cmd.Flags().StringVarP(
		&FlagPath,
		"path",
		"p",
		"",
		"Path to save vault keys",
	)
	Cmd.MarkFlagRequired("namespace")
	Cmd.Flags().StringVarP(
		&FlagNamespace,
		"namespace",
		"n",
		"vault",
		"Namespace where Vault is running",
	)
}

func vaultInitUnseal(path string, namespace string) {
	podNames := getPods(namespace)
	if len(podNames) == 0 {
		log.Fatalf("No Vault pods found in namespace %s", namespace)
	}
	vaultKeys := vaultInit(podNames[0], path, namespace)

	extractedValueKeys, treshold := extractVaultKeys(vaultKeys)

	for _, podName := range podNames {
		log.Printf("Unsealing pod %s", podName)
		unsealPod(podName, namespace, extractedValueKeys, treshold)
	}

	log.Printf("Vault initialization and unsealing completed successfully. Keys saved to %s/vault_keys.json\n", path)
}

func getPods(namespace string) []string {
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-o", "jsonpath={.items[*].metadata.name}")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing kubectl command: %v output %s", err, string(output))
	}

	lines := strings.Split(string(output), " ")
	var podNames []string
	for _, line := range lines {
		if strings.Contains(line, "vault") {
			podNames = append(podNames, line)
		}
	}

	return podNames
}

func vaultInit(pod, path, namespace string) string {
	log.Printf("Executing vault init on pod %s in namespace %s\n", pod, namespace)
	cmd := exec.Command("kubectl", "exec", pod, "-n", namespace, "--", "vault", "operator", "init", "-format=json")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing vault init command: %s\n", string(output))
	}

	err = os.WriteFile(filepath.Join(path, "vault_keys.json"), output, 0644)
	if err != nil {
		log.Printf("[WARNING] Error writing output to file: %v", err)
		log.Printf("Output: %s\n", string(output))
	}

	return string(output)
}

func extractVaultKeys(jsonData string) ([]string, int) {

	var response struct {
		UnsealKeysB64   []string `json:"unseal_keys_b64"`
		UnsealThreshold int      `json:"unseal_threshold"`
	}

	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	return response.UnsealKeysB64, response.UnsealThreshold
}

func unsealPod(podName, namespace string, vaultKeys []string, treshold int) {
	for i, key := range vaultKeys {
		if i >= treshold {
			break
		}
		cmd := exec.Command("kubectl", "exec", podName, "-n", namespace, "--", "vault", "operator", "unseal", key)
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Error unsealing pod %s with key %s: %v\nOutput: %s", podName, key, err, output)
		}
		waitForPodReady(podName, namespace)
	}
}

func waitForPodReady(pod, namespace string) {
	for {
		cmd := exec.Command("kubectl", "get", "pod", pod, "-n", namespace, "-o", "jsonpath={.status.phase}")
		cmd.Env = os.Environ()
		output, err := cmd.Output()
		if err != nil {
			log.Fatalf("Error checking pod status: %v", err)
		}

		status := strings.TrimSpace(string(output))
		if status == "Running" {
			break
		}

		log.Printf("Waiting for pod %s to be ready (current status: %s)\n", pod, status)
		time.Sleep(1 * time.Second)
	}
}

package init_unseal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/vault"
	"github.com/VojtechPastyrik/vp-utils/pkg/logger"
	"github.com/spf13/cobra"
)

var FlagPath string
var FlagNamespace string
var FlagKeyShares int
var FlagKeyThreshold int

var Cmd = &cobra.Command{
	Use:   "init-unseal",
	Short: "Initialize and unseal Vault",
	Long: `This command initializes Vault and unseals it using the generated keys.
It requires a running Vault instance in the specified namespace and saves the keys to the specified path.`,
	Example: `vp vault init-unseal --path /path/to/save/keys --namespace vault`,
	Args:    cobra.NoArgs,
	Aliases: []string{"iu"},
	Run: func(cmd *cobra.Command, args []string) {
		vaultInitUnseal(FlagPath, FlagNamespace, FlagKeyShares, FlagKeyThreshold)
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
	Cmd.Flags().IntVarP(
		&FlagKeyShares,
		"key-shares",
		"s",
		5,
		"Number of key shares to generate during Vault initialization",
	)
	Cmd.Flags().IntVarP(
		&FlagKeyThreshold,
		"key-threshold",
		"t",
		3,
		"Number of key shares required to unseal Vault",
	)
}

func vaultInitUnseal(path, namespace string, keyShares, keyThreshold int) {
	if keyShares < 1 || keyThreshold < 1 {
		logger.Fatal("key shares and threshold must be greater than 0")
	}
	if keyThreshold > keyShares {
		logger.Fatal("key threshold cannot be greater than key shares")
	}

	podNames := getPods(namespace)
	if len(podNames) == 0 {
		logger.Fatalf("no Vault pods found in namespace %s", namespace)
	}
	vaultKeys := vaultInit(podNames[0], path, namespace, keyShares, keyThreshold)

	extractedValueKeys, treshold := extractVaultKeys(vaultKeys)

	for _, podName := range podNames {
		logger.Infof("unsealing pod %s", podName)
		unsealPod(podName, namespace, extractedValueKeys, treshold)
	}

	logger.Infof("vault initialization and unsealing completed successfully. Keys saved to %s/vault_keys.json\n", path)
}

func getPods(namespace string) []string {
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-o", "jsonpath={.items[*].metadata.name}")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error executing kubectl command: %v output %s", err, string(output))
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

func vaultInit(pod, path, namespace string, keyShares, keyThreshold int) string {
	logger.Infof("executing vault init on pod %s in namespace %s\n", pod, namespace)
	cmd := exec.Command("kubectl", "exec", pod, "-n", namespace, "--", "vault", "operator", "init", "-format=json", "-key-shares", fmt.Sprintf("%d", keyShares), "-key-threshold", fmt.Sprintf("%d", keyThreshold))
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error executing vault init command: %s\n", string(output))
	}

	err = os.WriteFile(filepath.Join(path, "vault_keys.json"), output, 0644)
	if err != nil {
		logger.Infof("[WARNING] error writing output to file: %v", err)
		logger.Infof("output: %s\n", string(output))
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
		logger.Fatalf("error unmarshalling JSON: %v", err)
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
			logger.Fatalf("error unsealing pod %s with key %s: %v\nOutput: %s", podName, key, err, output)
		}
	}
	waitForPodReady(podName, namespace)
}

func waitForPodReady(pod, namespace string) {
	cmd := exec.Command("kubectl", "wait", "pod", pod, "-n", namespace, "--for=condition=Ready", "--timeout=60s")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error waiting for pod to be ready: %v\nOutput: %s", err, output)
	}
}

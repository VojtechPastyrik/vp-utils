package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/VojtechPastyrik/vpd/pkg/logger"
)

func GetPods(namespace string) []string {
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace,
		"-l", "app.kubernetes.io/name=vault,component=server",
		"-o", "jsonpath={.items[*].metadata.name}")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error executing kubectl command: %v output %s", err, string(output))
	}

	pods := strings.Split(strings.TrimSpace(string(output)), " ")
	var podNames []string
	for _, pod := range pods {
		if pod != "" {
			podNames = append(podNames, pod)
		}
	}

	return podNames
}

func ExtractVaultKeys(data string) ([]string, int) {
	var response struct {
		UnsealKeysB64   []string `json:"unseal_keys_b64"`
		UnsealThreshold int      `json:"unseal_threshold"`
	}

	err := json.Unmarshal([]byte(data), &response)
	if err == nil {
		return response.UnsealKeysB64, response.UnsealThreshold
	}

	keys, parseErr := ParseVaultKeysText(data)
	if parseErr != nil {
		logger.Fatalf("error parsing vault keys (tried JSON and text format): %v", parseErr)
	}

	return keys, len(keys)
}

func ParseVaultKeysText(text string) ([]string, error) {
	re := regexp.MustCompile(`(?m)^Unseal Key \d+:\s*(.+)$`)
	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no unseal keys found in text")
	}

	var keys []string
	for _, match := range matches {
		keys = append(keys, strings.TrimSpace(match[1]))
	}

	return keys, nil
}

func UnsealPod(podName, namespace string, vaultKeys []string, threshold int) {
	for i, key := range vaultKeys {
		if i >= threshold {
			break
		}
		cmd := exec.Command("kubectl", "exec", podName, "-n", namespace, "--", "vault", "operator", "unseal", key)
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Fatalf("error unsealing pod %s with key %s: %v\nOutput: %s", podName, key, err, output)
		}
	}
	WaitForPodReady(podName, namespace)
}

func WaitForPodReady(pod, namespace string) {
	cmd := exec.Command("kubectl", "wait", "pod", pod, "-n", namespace, "--for=condition=Ready", "--timeout=60s")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error waiting for pod to be ready: %v\nOutput: %s", err, output)
	}
}

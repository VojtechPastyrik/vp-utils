package list

import (
	"fmt"
	"os/exec"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/gpg"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List all GPG secret keys",
	Run: func(cmd *cobra.Command, args []string) {
		out, err := exec.Command("gpg", "--list-secret-keys", "--keyid-format", "long").CombinedOutput()
		if err != nil {
			logger.Fatalf("gpg command failed: %v\n%s", err, string(out))
		}
		output := strings.TrimSpace(string(out))
		if output == "" {
			fmt.Println("No GPG secret keys found")
			return
		}
		fmt.Println(output)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

package find

import (
	"fmt"
	"os/exec"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/gpg"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "find <email-or-name>",
	Short: "Search for a GPG key by email or name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		out, err := exec.Command("gpg", "--list-secret-keys", "--keyid-format", "long", query).CombinedOutput()
		if err != nil {
			logger.Fatalf("no key found for %q: %v", query, err)
		}
		output := strings.TrimSpace(string(out))
		if output == "" {
			fmt.Printf("No GPG key found matching %q\n", query)
			return
		}
		fmt.Println(output)
		fmt.Println()
		fmt.Println("To configure git signing, run:")
		fmt.Println("  git config --global user.signingkey <KEY_ID>")
		fmt.Println("  git config --global commit.gpgsign true")
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

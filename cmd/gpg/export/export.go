package export

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/gpg"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var outputFile string

var Cmd = &cobra.Command{
	Use:   "export <key-id-or-email>",
	Short: "Export GPG public key in ASCII armor format",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyID := args[0]
		out, err := exec.Command("gpg", "--armor", "--export", keyID).CombinedOutput()
		if err != nil {
			logger.Fatalf("gpg export failed: %v\n%s", err, string(out))
		}
		output := strings.TrimSpace(string(out))
		if output == "" {
			logger.Fatalf("no public key found for %q", keyID)
		}

		if outputFile != "" {
			if err := os.WriteFile(outputFile, out, 0644); err != nil {
				logger.Fatalf("error writing to file: %v", err)
			}
			fmt.Printf("Public key written to %s\n", outputFile)
			return
		}
		fmt.Println(output)
	},
}

func init() {
	Cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write key to file instead of stdout")
	parent_cmd.Cmd.AddCommand(Cmd)
}

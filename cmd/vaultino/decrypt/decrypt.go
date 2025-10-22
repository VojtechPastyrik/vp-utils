package decrypt

import (
	"log"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/vaultino"
	vaultinoUtils "github.com/VojtechPastyrik/vp-utils/utils/vaultino"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "decrypt",
	Aliases: []string{"crt"},
	Short:   "Decrypt Vaultino encrypted file",
	Long:    "Decrypt Vaultino encrypted file. It will prompt for a password and create a decrypted file.",
	Example: "vp-utils vaultino decrypt <path_tp_file>",
	Run: func(cmd *cobra.Command, args []string) {
		if args == nil || len(args) < 1 {
			log.Fatalf("Path to the encrypted file is required as the first argument")
		}
		vaultinoUtils.DecryptVaultToFile(args[0])
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

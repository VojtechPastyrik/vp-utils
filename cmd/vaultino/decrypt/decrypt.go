package decrypt

import (
	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/vaultino"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	vaultinoUtils "github.com/VojtechPastyrik/vpd/utils/vaultino"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "decrypt",
	Aliases: []string{"crt"},
	Short:   "Decrypt Vaultino encrypted file",
	Long:    "Decrypt Vaultino encrypted file. It will prompt for a password and create a decrypted file.",
	Example: "vpd vaultino decrypt <path_tp_file>",
	Run: func(cmd *cobra.Command, args []string) {
		if args == nil || len(args) < 1 {
			logger.Fatalf("path to the encrypted file is required as the first argument")
		}
		err := vaultinoUtils.DecryptVaultToFile(args[0])
		if err != nil {
			logger.Fatalf("failed to decrypt vault: %v", err)
		}
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

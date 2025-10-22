package edit

import (
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/vaultino"
	"github.com/VojtechPastyrik/vp-utils/pkg/logger"
	vaultinoUtils "github.com/VojtechPastyrik/vp-utils/utils/vaultino"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Short:   "Edit Vaultino encrypted file",
	Long:    "Edit Vaultino encrypted file. It will prompt for a password, decrypt the file, open it in the default editor, and re-encrypt it upon saving.",
	Example: "vp-utils vaultino edit <path_tp_file>",
	Run: func(cmd *cobra.Command, args []string) {
		if args == nil || len(args) < 1 {
			logger.Fatalf("path to the encrypted file is required as the first argument")
		}
		err := vaultinoUtils.EditVault(args[0])
		if err != nil {
			logger.Fatalf("failed to edit vault: %v", err)
		}
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

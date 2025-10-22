package edit

import (
	"log"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/vaultino"
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
			log.Fatalf("Path to the encrypted file is required as the first argument")
		}
		vaultinoUtils.EditVault(args[0])
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

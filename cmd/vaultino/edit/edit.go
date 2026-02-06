package edit

import (
	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/vaultino"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	vaultinoUtils "github.com/VojtechPastyrik/vpd/utils/vaultino"

	"github.com/spf13/cobra"
)

var FlagChangePassword bool

var Cmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Short:   "Edit Vaultino encrypted file",
	Long:    "Edit Vaultino encrypted file. It will prompt for a password, decrypt the file, open it in the default editor, and re-encrypt it upon saving.",
	Example: "vpd vaultino edit <path_tp_file>\nvpd vaultino edit --change-password <path_tp_file>",
	Run: func(cmd *cobra.Command, args []string) {
		if args == nil || len(args) < 1 {
			logger.Fatalf("path to the encrypted file is required as the first argument")
		}

		var err error
		if FlagChangePassword {
			err = vaultinoUtils.EditVaultWithPasswordChange(args[0])
		} else {
			err = vaultinoUtils.EditVault(args[0])
		}

		if err != nil {
			logger.Fatalf("failed to edit vault: %v", err)
		}
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().BoolVarP(&FlagChangePassword, "change-password", "p", false, "Change the vault password after editing")
}

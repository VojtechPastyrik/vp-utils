package get

import (
	"fmt"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/vaultino"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	vaultinoUtils "github.com/VojtechPastyrik/vpd/utils/vaultino"
	"github.com/spf13/cobra"
)

var (
	Flagkey      string
	FlagPassword string
)

var Cmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get value from Vaultino encrypted file",
	Long:    "Get value from Vaultino encrypted file for the provided key. It will prompt for a password and retrieve the value. Supported file formats are YAML and JSON.",
	Example: "vpd vaultino get <path_to_file> -k <key>",
	Run: func(cmd *cobra.Command, args []string) {
		if args == nil || len(args) < 1 {
			logger.Fatalf("path to the encrypted file is required as the first argument")
		}
		value, err := vaultinoUtils.GetSecretFromVault(args[0], Flagkey, FlagPassword)
		if err != nil {
			logger.Fatalf("failed to get secret: %v", err)
		}
		fmt.Println(value)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&Flagkey, "key", "k", "", "Key to get value for")
	Cmd.MarkFlagRequired("key")
	Cmd.Flags().StringVarP(&FlagPassword, "password", "p", "", "Password to decrypt vault (if not provided, will prompt interactively)")
}

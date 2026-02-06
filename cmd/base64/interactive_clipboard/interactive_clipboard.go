package interactive_clipboard

import (
	"encoding/base64"
	"fmt"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/base64"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var FlagDecode bool

var Cmd = &cobra.Command{
	Use:     "interactive-clipboard",
	Short:   "Base64 Interactive Encode/Decode from Clipboard",
	Aliases: []string{"ic"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		clipboardData, err := clipboard.ReadAll()
		if err != nil {
			logger.Fatalf("%v", err)
		}
		if FlagDecode {
			dataPlainText, err := base64.StdEncoding.DecodeString(clipboardData)
			if err == nil {
				fmt.Println(string(dataPlainText))
			}
		} else {
			encodedData := base64.StdEncoding.EncodeToString([]byte(clipboardData))
			fmt.Println(string(encodedData))
		}
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().BoolVarP(&FlagDecode, "decode", "d", false, "Decode the base64 encoded clipboard content")
}

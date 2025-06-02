package jwt

import (
	"encoding/base64"
	"fmt"
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	"github.com/spf13/cobra"
	"log"
)

var FlagDecode bool

var Cmd = &cobra.Command{
	Use:   "base64",
	Short: "Base64 Utils",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		encodeDecode(FlagDecode, args[0])
	},
}

func init() {
	root.RootCmd.AddCommand(Cmd)
	Cmd.Flags().BoolVarP(&FlagDecode, "decode", "d", false, "Decode the base64 encoded file or string")

}

func encodeDecode(decode bool, inputString string) {
	if inputString == "" {
		log.Print("Input string is empty")
		return
	}

	if decode {
		decodedBytes, err := base64.StdEncoding.DecodeString(inputString)
		if err != nil {
			log.Fatalf("Error decoding base64: %v", err)
		}
		fmt.Println(string(decodedBytes))
	} else {
		encodedString := base64.StdEncoding.EncodeToString([]byte(inputString))
		fmt.Println(encodedString)
	}
}

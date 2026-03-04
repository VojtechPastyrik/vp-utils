package pretty_json

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/convert"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pretty-json [file]",
	Short: "Pretty-print JSON",
	Long:  "Pretty-print JSON with indentation. Accepts a file path as argument or reads from stdin.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := readInput(args)
		if err != nil {
			logger.Fatalf("error reading input: %v", err)
		}

		var obj interface{}
		if err := json.Unmarshal(data, &obj); err != nil {
			logger.Fatalf("invalid JSON: %v", err)
		}

		out, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			logger.Fatalf("error formatting JSON: %v", err)
		}
		fmt.Println(string(out))
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

func readInput(args []string) ([]byte, error) {
	if len(args) == 1 {
		return os.ReadFile(args[0])
	}
	return io.ReadAll(os.Stdin)
}

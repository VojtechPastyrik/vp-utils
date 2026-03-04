package validate_json

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
	Use:   "validate-json [file]",
	Short: "Validate JSON syntax",
	Long:  "Validate JSON syntax. Accepts a file path as argument or reads from stdin.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := readInput(args)
		if err != nil {
			logger.Fatalf("error reading input: %v", err)
		}

		if json.Valid(data) {
			fmt.Println("Valid JSON")
		} else {
			logger.Fatal("Invalid JSON")
		}
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

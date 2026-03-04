package json2yaml

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/convert"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var Cmd = &cobra.Command{
	Use:   "json2yaml [file]",
	Short: "Convert JSON to YAML",
	Long:  "Convert JSON to YAML. Accepts a file path as argument or reads from stdin.",
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

		out, err := yaml.Marshal(obj)
		if err != nil {
			logger.Fatalf("error converting to YAML: %v", err)
		}
		fmt.Print(string(out))
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

package yaml2json

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
	Use:   "yaml2json [file]",
	Short: "Convert YAML to JSON",
	Long:  "Convert YAML to JSON. Accepts a file path as argument or reads from stdin.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := readInput(args)
		if err != nil {
			logger.Fatalf("error reading input: %v", err)
		}

		var obj interface{}
		if err := yaml.Unmarshal(data, &obj); err != nil {
			logger.Fatalf("invalid YAML: %v", err)
		}

		out, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			logger.Fatalf("error converting to JSON: %v", err)
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

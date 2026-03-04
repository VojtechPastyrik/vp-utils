package stats

import (
	"fmt"
	"os/exec"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/docker"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "stats",
	Short: "Show running container resource usage (one-shot)",
	Run: func(cmd *cobra.Command, args []string) {
		out, err := exec.Command("docker", "stats", "--no-stream", "--format",
			"table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.PIDs}}").CombinedOutput()
		if err != nil {
			logger.Fatalf("docker stats failed: %v\n%s", err, string(out))
		}
		output := strings.TrimSpace(string(out))
		if output == "" {
			fmt.Println("No running containers")
			return
		}
		fmt.Println(output)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

package cleanup

import (
	"fmt"
	"os/exec"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/docker"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var dryRun bool

var Cmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Prune stopped containers, dangling images, unused volumes and networks",
	Run: func(cmd *cobra.Command, args []string) {
		commands := []struct {
			name string
			args []string
		}{
			{"containers", []string{"container", "prune", "-f"}},
			{"images", []string{"image", "prune", "-f"}},
			{"volumes", []string{"volume", "prune", "-f"}},
			{"networks", []string{"network", "prune", "-f"}},
		}

		for _, c := range commands {
			if dryRun {
				fmt.Printf("[dry-run] docker %s\n", strings.Join(c.args, " "))
				continue
			}
			fmt.Printf("Pruning %s...\n", c.name)
			out, err := exec.Command("docker", c.args...).CombinedOutput()
			if err != nil {
				logger.Errorf("failed to prune %s: %v\n%s", c.name, err, string(out))
				continue
			}
			output := strings.TrimSpace(string(out))
			if output != "" {
				fmt.Println(output)
			}
		}

		if !dryRun {
			logger.Success("Docker cleanup complete")
		}
	},
}

func init() {
	Cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be removed without doing it")
	parent_cmd.Cmd.AddCommand(Cmd)
}

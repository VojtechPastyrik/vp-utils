package info

import (
	"fmt"
	"runtime"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/cpu"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "info",
	Short: "Show CPU information",
	Long: `Displays basic information about the CPU:
- Number of logical cores
- Architecture
- Model and frequency (if available)
`,
	Run: func(cmd *cobra.Command, args []string) {
		printCPUInfo()
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

func printCPUInfo() {
	fmt.Printf("Logical cores: %d\n", runtime.NumCPU())
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)

	info, err := cpu.Info()
	if err == nil && len(info) > 0 {
		fmt.Printf("Model: %s\n", info[0].ModelName)
		fmt.Printf("Frequency: %.2f MHz\n", info[0].Mhz)
	}
}

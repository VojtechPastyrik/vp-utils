package info

import (
	"fmt"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/memory"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
)

var SystemCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system memory info",
	Long:  "Displays memory statistics for the whole system.",
	Run: func(cmd *cobra.Command, args []string) {
		v, err := mem.VirtualMemory()
		if err != nil {
			fmt.Println("Error getting system memory info:", err)
			return
		}
		fmt.Printf("Total: %.2f GB\n", float64(v.Total)/1024/1024/1024)
		fmt.Printf("Used: %.2f GB\n", float64(v.Used)/1024/1024/1024)
		fmt.Printf("Available: %.2f GB\n", float64(v.Available)/1024/1024/1024)
		fmt.Printf("UsedPercent: %.2f%%\n", v.UsedPercent)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(SystemCmd)
}

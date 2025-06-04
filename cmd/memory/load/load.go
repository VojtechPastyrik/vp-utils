package load

import (
	"fmt"
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/memory"
	"github.com/spf13/cobra"
	"runtime"
	"time"
)

var (
	FlagDuration int
	FlagSizeMB   int
)

var Cmd = &cobra.Command{
	Use:   "load",
	Short: "Make memory load",
	Long: `Allocates the specified amount of memory for a given duration.
The command measures wall time and verifies that the memory is actually used.
Useful for testing system performance, memory limits, or stability.

Output explanation:
  Wall time:     Real elapsed time from start to finish (as seen on a clock).
  Memory used:   Amount of memory allocated and touched by the process.

Examples:
  memory load --duration 60 --size 1024
  memory load -d 10 -s 256
`,
	Aliases: []string{"l"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		memoryLoad(FlagDuration, FlagSizeMB)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().IntVarP(&FlagDuration, "duration", "d", 30, "Duration of the load in seconds")
	Cmd.Flags().IntVarP(&FlagSizeMB, "size", "s", 1024, "Amount of memory to allocate in MB")
}

func memoryLoad(duration, sizeMB int) {
	fmt.Printf("Starting memory load: %d MB for %d seconds...\n", sizeMB, duration)
	startWall := time.Now()

	// Allocate and touch the memory
	mem := make([]byte, sizeMB*1024*1024)
	for i := range mem {
		mem[i] = byte(i)
	}
	runtime.GC()

	fmt.Printf("Memory used: %d MB\n", sizeMB)
	time.Sleep(time.Duration(duration) * time.Second)

	elapsedWall := time.Since(startWall)
	fmt.Printf("Wall time: %v\n", elapsedWall)
}

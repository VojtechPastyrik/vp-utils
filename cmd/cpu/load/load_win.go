//go:build windows

package load

import (
	"fmt"
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/cpu"
	"github.com/spf13/cobra"
	"runtime"
	"time"
)

var (
	FlagDuration int
	FlagWorkers  int
)

var Cmd = &cobra.Command{
	Use:   "load",
	Short: "Make CPU load",
	Long: `Starts a CPU load using the specified number of parallel workers (threads) for a given duration.
Each worker fully loads the CPU by performing intensive prime number calculations. Only wall time is measured on Windows.
Examples:
  cpu load --duration 60 --workers 8
  cpu load -d 10 -w 2
`,
	Aliases: []string{"l"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		cpuLoad(FlagDuration, FlagWorkers)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().IntVarP(&FlagDuration, "duration", "d", 30, "Duration of the load in seconds")
	Cmd.Flags().IntVarP(&FlagWorkers, "workers", "w", runtime.NumCPU(), "Number of parallel workers (CPU cores)")
}

func cpuLoad(duration, workers int) {
	fmt.Printf("Starting CPU load: %d workers for %d seconds...\n", workers, duration)
	startWall := time.Now()

	stop := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-stop:
					return
				default:
					_ = isPrime(982451653)
				}
			}
		}()
	}

	time.Sleep(time.Duration(duration) * time.Second)
	close(stop)

	elapsedWall := time.Since(startWall)

	fmt.Printf("Wall time: %v\n", elapsedWall)
	fmt.Println("CPU time: not supported on Windows")
	fmt.Println("CPU utilization: not supported on Windows")
}

func isPrime(n int) bool {
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

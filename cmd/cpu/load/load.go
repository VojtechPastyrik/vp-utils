//go:build linux || darwin

package load

import (
	"fmt"
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/cpu"
	"github.com/spf13/cobra"
	"runtime"
	"syscall"
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
Each worker fully loads the CPU by performing intensive prime number calculations. The command measures wall time, total CPU time, and average CPU utilization in percent.
Useful for testing system performance, cooling, or stability.

Output explanation:
  Wall time:     Real elapsed time from start to finish (as seen on a clock).
  CPU time:      Total time spent by all threads on the CPU (can be higher than wall time if running in parallel).
  CPU utilization: Average CPU usage during the test, in percent. Values above 100% mean multiple CPU cores were used.

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
	startCPU := getCPUTime()

	stop := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-stop:
					return
				default:
					_ = isPrime(982451653) // heavy prime check
				}
			}
		}()
	}

	time.Sleep(time.Duration(duration) * time.Second)
	close(stop)

	elapsedWall := time.Since(startWall)
	elapsedCPU := getCPUTime() - startCPU
	cpuPercent := float64(elapsedCPU) / float64(elapsedWall) * 100

	fmt.Printf("Wall time: %v\n", elapsedWall)
	fmt.Printf("CPU time: %v\n", elapsedCPU)
	fmt.Printf("CPU utilization: %.1f%%\n", cpuPercent)
}

func isPrime(n int) bool {
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func getCPUTime() time.Duration {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err == nil {
		user := time.Duration(ru.Utime.Sec)*time.Second + time.Duration(ru.Utime.Usec)*time.Microsecond
		sys := time.Duration(ru.Stime.Sec)*time.Second + time.Duration(ru.Stime.Usec)*time.Microsecond
		return user + sys
	}
	return 0
}

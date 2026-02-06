package benchmark

import (
	"fmt"
	"runtime"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/cpu"
	"github.com/spf13/cobra"
)

var (
	FlagSeconds int
)

var Cmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run a simple CPU benchmark",
	Long: `Performs a simple CPU benchmark by running arithmetic operations in a loop.
Reports the number of operations per second as a score.`,
	Run: func(cmd *cobra.Command, args []string) {
		runBenchmark(FlagSeconds)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().IntVarP(&FlagSeconds, "seconds", "s", 5, "Benchmark duration in seconds")
}

func runBenchmark(seconds int) {
	fmt.Printf("Running CPU benchmark for %d seconds on %d logical cores...\n", seconds, runtime.NumCPU())
	var ops uint64 = 0
	start := time.Now()
	end := start.Add(time.Duration(seconds) * time.Second)
	var a float64 = 1.0001

	for time.Now().Before(end) {
		for i := 0; i < 1000000; i++ {
			a *= 1.0000001
			a /= 1.0000001
		}
		ops += 1000000
	}
	elapsed := time.Since(start).Seconds()
	fmt.Printf("Benchmark finished.\n")
	fmt.Printf("Total operations: %d\n", ops)
	fmt.Printf("Elapsed time: %.2f s\n", elapsed)
	fmt.Printf("Score: %.0f ops/sec\n", float64(ops)/elapsed)
}

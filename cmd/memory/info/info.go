package leaktest

import (
	"fmt"
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/memory"
	"github.com/spf13/cobra"
	"time"
)

var (
	blockSize   int
	iterations  int
	sleepMillis int
)

var Cmd = &cobra.Command{
	Use:   "leak-test",
	Short: "Simulate a memory leak",
	Long:  "Repeatedly allocates memory without releasing it, simulating a memory leak.",
	Run: func(cmd *cobra.Command, args []string) {
		simulateLeak()
	},
}

func init() {
	Cmd.Flags().IntVarP(&blockSize, "block", "b", 1024*1024, "Block size in bytes")
	Cmd.Flags().IntVarP(&iterations, "count", "c", 100, "Number of allocations")
	Cmd.Flags().IntVarP(&sleepMillis, "sleep", "s", 100, "Delay between allocations (ms)")
	parent_cmd.Cmd.AddCommand(Cmd)
}

func simulateLeak() {
	fmt.Printf("Simulating memory leak: %d blocks of %d bytes\n", iterations, blockSize)
	leaks := make([][]byte, 0, iterations)
	for i := 0; i < iterations; i++ {
		leaks = append(leaks, make([]byte, blockSize))
		fmt.Printf("Allocated: %d MB\n", (i+1)*blockSize/1024/1024)
		time.Sleep(time.Duration(sleepMillis) * time.Millisecond)
	}
	fmt.Println("Test finished. Memory was not released.")
}

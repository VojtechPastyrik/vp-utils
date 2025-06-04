package allocpatterns

import (
	"fmt"
	"math/rand"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/memory"
	"github.com/spf13/cobra"
)

var (
	pattern    string
	blockSize  int
	iterations int
)

var Cmd = &cobra.Command{
	Use:   "alloc-patterns",
	Short: "Simulate various memory allocation/freeing patterns",
	Long:  "Simulates different allocation and freeing patterns to test memory management and fragmentation.",
	Run: func(cmd *cobra.Command, args []string) {
		switch pattern {
		case "alternating":
			alternatingPattern()
		case "random-free":
			randomFreePattern()
		default:
			fmt.Println("Unknown pattern. Use 'alternating' or 'random-free'.")
		}
	},
}

func init() {
	Cmd.Flags().StringVarP(&pattern, "pattern", "p", "alternating", "Allocation pattern: alternating, random-free")
	Cmd.Flags().IntVarP(&blockSize, "block", "b", 1024*1024, "Block size in bytes")
	Cmd.Flags().IntVarP(&iterations, "count", "c", 100, "Number of allocations")
	parent_cmd.Cmd.AddCommand(Cmd)
}

// Allocates memory in an alternating pattern, freeing every second block
func alternatingPattern() {
	fmt.Println("Alternating allocation/free pattern")
	for i := 0; i < iterations; i++ {
		b := make([]byte, blockSize)
		_ = b
		if i%2 == 0 {
			b = nil
		}
	}
	fmt.Println("Done.")
}

// Allocates memory and then randomly frees some blocks
func randomFreePattern() {
	fmt.Println("Random free allocation pattern")
	blocks := make([][]byte, iterations)
	for i := 0; i < iterations; i++ {
		blocks[i] = make([]byte, blockSize)
	}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < iterations/2; i++ {
		idx := rand.Intn(iterations)
		blocks[idx] = nil
	}
	fmt.Println("Done.")
}

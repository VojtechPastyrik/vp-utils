package stress

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/cpu"
	"github.com/spf13/cobra"
)

var (
	FlagType     string
	FlagThreads  int
	FlagDuration int
)

var Cmd = &cobra.Command{
	Use:   "stress",
	Short: "Stress the CPU with different load types",
	Long: `Generates CPU load using different operation types:
- math: floating point and integer operations
- branch: heavy branching (if/switch)
- io: simulated I/O wait (sleep)
`,
	Run: func(cmd *cobra.Command, args []string) {
		runStress(FlagType, FlagThreads, FlagDuration)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagType, "type", "t", "all", "Type of stress: math, branch, io, all")
	Cmd.Flags().IntVarP(&FlagThreads, "threads", "n", runtime.NumCPU(), "Number of threads (goroutines)")
	Cmd.Flags().IntVarP(&FlagDuration, "duration", "d", 10, "Duration in seconds")
}

func runStress(stressType string, threads, duration int) {
	fmt.Printf("Starting CPU stress: type=%s, threads=%d, duration=%ds\n", stressType, threads, duration)
	var wg sync.WaitGroup
	stop := make(chan struct{})

	types := []string{}
	switch stressType {
	case "all":
		types = []string{"math", "branch", "io"}
	default:
		types = []string{stressType}
	}

	threadsPerType := threads / len(types)
	extra := threads % len(types)
	threadCounts := make([]int, len(types))
	for i := range types {
		threadCounts[i] = threadsPerType
		if i < extra {
			threadCounts[i]++
		}
	}

	for i, t := range types {
		for j := 0; j < threadCounts[i]; j++ {
			wg.Add(1)
			go func(st string) {
				defer wg.Done()
				switch st {
				case "math":
					stressMath(stop)
				case "branch":
					stressBranch(stop)
				case "io":
					stressIO(stop)
				default:
					fmt.Fprintf(os.Stderr, "Unknown stress type: %s\n", st)
				}
			}(t)
		}
	}

	time.Sleep(time.Duration(duration) * time.Second)
	close(stop)
	wg.Wait()
	fmt.Println("CPU stress finished.")
}

func stressMath(stop <-chan struct{}) {
	x := 1.0001
	for {
		select {
		case <-stop:
			return
		default:
			for i := 0; i < 100000; i++ {
				x *= 1.000001
				x /= 1.000001
			}
		}
	}
}

func stressBranch(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
			for i := 0; i < 100000; i++ {
				if rand.Intn(2) == 0 {
					_ = 1
				} else {
					_ = 2
				}
			}
		}
	}
}

func stressIO(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}

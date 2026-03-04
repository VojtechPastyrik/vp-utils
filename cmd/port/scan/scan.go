package scan

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/port"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	portRange  string
	timeout    time.Duration
	concurrent int
)

var Cmd = &cobra.Command{
	Use:   "scan <host>",
	Short: "Scan a range of ports on a host",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		start, end, err := parseRange(portRange)
		if err != nil {
			logger.Fatalf("invalid port range: %v", err)
		}

		openPorts := scanPorts(host, start, end)
		if len(openPorts) == 0 {
			fmt.Printf("No open ports found on %s (%d-%d)\n", host, start, end)
			return
		}

		sort.Ints(openPorts)
		fmt.Printf("Open ports on %s:\n", host)
		for _, p := range openPorts {
			fmt.Printf("  %d/tcp\topen\n", p)
		}
	},
}

func init() {
	Cmd.Flags().StringVar(&portRange, "range", "1-1024", "Port range to scan (e.g. 80-443)")
	Cmd.Flags().DurationVar(&timeout, "timeout", 1*time.Second, "Connection timeout per port")
	Cmd.Flags().IntVar(&concurrent, "concurrent", 100, "Number of concurrent workers")
	parent_cmd.Cmd.AddCommand(Cmd)
}

func parseRange(r string) (int, int, error) {
	parts := strings.SplitN(r, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected format: start-end")
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	if start < 1 || end > 65535 || start > end {
		return 0, 0, fmt.Errorf("invalid range %d-%d", start, end)
	}
	return start, end, nil
}

func scanPorts(host string, start, end int) []int {
	var mu sync.Mutex
	var openPorts []int

	sem := make(chan struct{}, concurrent)
	var wg sync.WaitGroup

	for p := start; p <= end; p++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(port int) {
			defer wg.Done()
			defer func() { <-sem }()

			addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
			conn, err := net.DialTimeout("tcp", addr, timeout)
			if err == nil {
				conn.Close()
				mu.Lock()
				openPorts = append(openPorts, port)
				mu.Unlock()
			}
		}(p)
	}
	wg.Wait()
	return openPorts
}

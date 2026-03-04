package check

import (
	"fmt"
	"net"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/port"
	"github.com/spf13/cobra"
)

var timeout time.Duration

var Cmd = &cobra.Command{
	Use:   "check <host:port>",
	Short: "Check if a specific port is open",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		address := args[0]
		start := time.Now()
		conn, err := net.DialTimeout("tcp", address, timeout)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("❌ %s is closed (%v)\n", address, elapsed.Round(time.Millisecond))
			return
		}
		conn.Close()
		fmt.Printf("✅ %s is open (%v)\n", address, elapsed.Round(time.Millisecond))
	},
}

func init() {
	Cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "Connection timeout")
	parent_cmd.Cmd.AddCommand(Cmd)
}

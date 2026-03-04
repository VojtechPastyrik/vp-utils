package reverse

import (
	"fmt"
	"net"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/dns"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "reverse <ip>",
	Short: "Reverse DNS lookup (PTR) for an IP address",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ip := args[0]
		names, err := net.LookupAddr(ip)
		if err != nil {
			logger.Fatalf("reverse lookup failed: %v", err)
		}
		if len(names) == 0 {
			fmt.Println("No PTR records found")
			return
		}
		for _, name := range names {
			fmt.Printf("PTR\t%s\n", name)
		}
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

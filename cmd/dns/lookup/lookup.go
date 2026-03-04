package lookup

import (
	"fmt"
	"net"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/dns"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var recordType string

var Cmd = &cobra.Command{
	Use:   "lookup <domain>",
	Short: "Resolve DNS records for a domain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		rtype := strings.ToUpper(recordType)

		if rtype == "ANY" {
			for _, t := range []string{"A", "AAAA", "MX", "CNAME", "TXT", "NS"} {
				lookupType(domain, t)
			}
			return
		}
		lookupType(domain, rtype)
	},
}

func init() {
	Cmd.Flags().StringVarP(&recordType, "type", "t", "A", "Record type: A, AAAA, MX, CNAME, TXT, NS, ANY")
	parent_cmd.Cmd.AddCommand(Cmd)
}

func lookupType(domain, rtype string) {
	switch rtype {
	case "A":
		ips, err := net.LookupHost(domain)
		if err != nil {
			logger.Errorf("A lookup failed: %v", err)
			return
		}
		for _, ip := range ips {
			if net.ParseIP(ip).To4() != nil {
				fmt.Printf("A\t%s\n", ip)
			}
		}
	case "AAAA":
		ips, err := net.LookupHost(domain)
		if err != nil {
			logger.Errorf("AAAA lookup failed: %v", err)
			return
		}
		for _, ip := range ips {
			if net.ParseIP(ip).To4() == nil {
				fmt.Printf("AAAA\t%s\n", ip)
			}
		}
	case "MX":
		mxs, err := net.LookupMX(domain)
		if err != nil {
			logger.Errorf("MX lookup failed: %v", err)
			return
		}
		for _, mx := range mxs {
			fmt.Printf("MX\t%d\t%s\n", mx.Pref, mx.Host)
		}
	case "CNAME":
		cname, err := net.LookupCNAME(domain)
		if err != nil {
			logger.Errorf("CNAME lookup failed: %v", err)
			return
		}
		fmt.Printf("CNAME\t%s\n", cname)
	case "TXT":
		txts, err := net.LookupTXT(domain)
		if err != nil {
			logger.Errorf("TXT lookup failed: %v", err)
			return
		}
		for _, txt := range txts {
			fmt.Printf("TXT\t%s\n", txt)
		}
	case "NS":
		nss, err := net.LookupNS(domain)
		if err != nil {
			logger.Errorf("NS lookup failed: %v", err)
			return
		}
		for _, ns := range nss {
			fmt.Printf("NS\t%s\n", ns.Host)
		}
	default:
		logger.Fatalf("unsupported record type: %s", rtype)
	}
}

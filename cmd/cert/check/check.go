package check

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/cert"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	expiryOnly bool
	insecure   bool
)

var Cmd = &cobra.Command{
	Use:   "check <host:port>",
	Short: "Inspect TLS certificate of a remote host",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		address := args[0]
		if !strings.Contains(address, ":") {
			address += ":443"
		}

		host := strings.Split(address, ":")[0]

		conf := &tls.Config{
			InsecureSkipVerify: insecure,
			ServerName:         host,
		}

		conn, err := tls.DialWithDialer(
			&net.Dialer{Timeout: 10 * time.Second},
			"tcp", address, conf,
		)
		if err != nil {
			logger.Fatalf("TLS connection failed: %v", err)
		}
		defer conn.Close()

		certs := conn.ConnectionState().PeerCertificates
		if len(certs) == 0 {
			logger.Fatal("no certificates returned")
		}

		cert := certs[0]
		daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)

		if expiryOnly {
			fmt.Println(daysLeft)
			return
		}

		fmt.Printf("Subject:     %s\n", cert.Subject.CommonName)
		fmt.Printf("Issuer:      %s\n", cert.Issuer.CommonName)
		fmt.Printf("Not Before:  %s\n", cert.NotBefore.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("Not After:   %s\n", cert.NotAfter.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("Days Left:   %d\n", daysLeft)
		if len(cert.DNSNames) > 0 {
			fmt.Printf("SANs:        %s\n", strings.Join(cert.DNSNames, ", "))
		}
	},
}

func init() {
	Cmd.Flags().BoolVar(&expiryOnly, "expiry-only", false, "Only print days until expiry")
	Cmd.Flags().BoolVar(&insecure, "insecure", false, "Skip certificate verification (for self-signed certs)")
	parent_cmd.Cmd.AddCommand(Cmd)
}

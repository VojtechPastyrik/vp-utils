package parse

import (
	"crypto/tls"
	"crypto/x509"
	_ "crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	tls_cmd "github.com/VojtechPastyrik/vp-utils/cmd/tls"
	"github.com/VojtechPastyrik/vp-utils/pkg/logger"
	"github.com/spf13/cobra"
)

var FlagAddr string
var FlagServerName string
var FlagOutputFile string

var Cmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse TLS Certificate data from server (connection)",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		printCertificateFromServer(FlagAddr, FlagServerName, FlagOutputFile)
	},
}

func init() {
	tls_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagAddr,
		"address",
		"a",
		"",
		"Address (eg.: google.com:443)",
	)
	Cmd.MarkFlagRequired("address")
	Cmd.Flags().StringVarP(
		&FlagServerName,
		"server-name",
		"n",
		"",
		"ServerName (SNI)",
	)
	Cmd.Flags().StringVarP(
		&FlagOutputFile,
		"output-file",
		"o",
		"",
		"File to save the certificate (PEM format)",
	)
}

func printCertificateFromServer(addr, serverName, outputFile string) {
	var conf *tls.Config
	if serverName == "" {
		conf = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		conf = &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         serverName,
		}
	}

	conn, err := tls.Dial("tcp", addr, conf)
	if err != nil {
		logger.Info("error in Dial", err)
		return
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	for i, cert := range certs {
		fmt.Printf("Subject Name: %s\n", cert.Subject)
		fmt.Printf("Subject Common Name: %s\n", cert.Subject.CommonName)
		fmt.Printf("Issuer Name: %s\n", cert.Issuer)
		fmt.Printf("Issuer Common Name: %s \n", cert.Issuer.CommonName)
		fmt.Printf("Created: %s \n", cert.NotBefore.Format(time.RFC3339))
		fmt.Printf("Expiry: %s \n", cert.NotAfter.Format(time.RFC3339))
		fmt.Printf("DNS Names: %v\n", cert.DNSNames)
		fmt.Printf("IP Addresses: %v\n", cert.IPAddresses)
		fmt.Println()

		if outputFile != "" {
			uniqueFileName := fmt.Sprintf("%s_%d.pem", outputFile, i+1)
			saveCertificateToFile(cert, uniqueFileName)
		}
	}
}

func saveCertificateToFile(cert *x509.Certificate, outputFile string) {
	file, err := os.Create(outputFile)
	if err != nil {
		logger.Fatalf("failed to create file %s: %v", outputFile, err)
	}
	defer file.Close()

	err = pem.Encode(file, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	if err != nil {
		logger.Fatalf("failed to write certificate to file %s: %v", outputFile, err)
	}

	logger.Infof("certificate saved to %s", outputFile)
	fmt.Println()
}

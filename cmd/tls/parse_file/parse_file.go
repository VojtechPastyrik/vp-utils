package parse_file

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	tls_cmd "github.com/VojtechPastyrik/vpd/cmd/tls"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/VojtechPastyrik/vpd/pkg/tlsutil"
	"github.com/spf13/cobra"
	"software.sslmate.com/src/go-pkcs12"
)

var CmdFlagCertFile string
var CmdFlagKeyFile string
var CmdFlagP12Password string

var Cmd = &cobra.Command{
	Use:     "parse-file",
	Short:   "Parse TLS Certificate data from file (cert, key, or P12)",
	Aliases: []string{"pf"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		printCertificateFromFile(CmdFlagCertFile, CmdFlagKeyFile, CmdFlagP12Password)
	},
}

func init() {
	tls_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&CmdFlagCertFile,
		"cert",
		"c",
		"",
		"Cert file (PEM or P12 format)",
	)
	Cmd.MarkFlagRequired("cert")
	Cmd.Flags().StringVarP(
		&CmdFlagKeyFile,
		"key",
		"k",
		"",
		"Key file (optional, PEM format)",
	)
	Cmd.Flags().StringVarP(
		&CmdFlagP12Password,
		"password",
		"p",
		"",
		"Password for P12 file (optional)",
	)
}

func printCertificateFromFile(certFile, keyFile, p12Password string) {
	if strings.HasSuffix(certFile, ".p12") {
		loadCertificateFromP12(certFile, p12Password)
	} else {
		loadCertificateFromPEM(certFile, keyFile)
	}
}

func loadCertificateFromP12(certFile, password string) {
	data, err := os.ReadFile(certFile)
	if err != nil {
		logger.Fatalf("failed to read P12 file: %v", err)
	}

	privateKey, cert, caCerts, err := pkcs12.DecodeChain(data, password)
	if err != nil {
		logger.Fatalf("failed to decode P12 file: %v", err)
	}

	tlsutil.PrintCertInfo(cert)

	if privateKey != nil {
		fmt.Printf("Private Key Type: %T\n", privateKey)
		fmt.Println("Private key is present in the P12 file.")
	}

	if len(caCerts) > 0 {
		fmt.Printf("Number of CA certificates: %d\n", len(caCerts))
		for i, caCert := range caCerts {
			fmt.Printf("CA Certificate %d Subject: %s\n", i+1, caCert.Subject)
			fmt.Printf("CA Certificate %d Issuer: %s\n", i+1, caCert.Issuer)
		}
	}
}

func loadCertificateFromPEM(certFile, keyFile string) {
	certData, err := os.ReadFile(certFile)
	if err != nil {
		logger.Fatalf("failed to read cert file: %v", err)
	}

	var certs []*x509.Certificate
	if keyFile == "" {
		block, _ := pem.Decode(certData)
		if block == nil || block.Type != "CERTIFICATE" {
			logger.Fatal("failed to decode PEM certificate")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			logger.Fatalf("failed to parse certificate: %v", err)
		}
		certs = append(certs, cert)
	} else {
		tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			logger.Fatalf("failed to load key pair: %v", err)
		}
		certs, err = x509.ParseCertificates(tlsCert.Certificate[0])
		if err != nil {
			logger.Fatalf("failed to parse certificates: %v", err)
		}
	}

	for _, cert := range certs {
		tlsutil.PrintCertInfo(cert)
	}
}

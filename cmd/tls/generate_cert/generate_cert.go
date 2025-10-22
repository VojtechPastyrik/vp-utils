package generate_cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	tls_cmd "github.com/VojtechPastyrik/vp-utils/cmd/tls"
	"github.com/VojtechPastyrik/vp-utils/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	FlagOutCert string
	FlagOutKey  string
	FlagHost    string
	FlagDays    int
)

var Cmd = &cobra.Command{
	Use:     "generate-cert",
	Short:   "Generate self-signed certificate for testing",
	Aliases: []string{"gc"},
	Run: func(cmd *cobra.Command, args []string) {
		generateCert(FlagOutCert, FlagOutKey, FlagHost, FlagDays)
	},
}

func init() {
	tls_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagOutCert, "cert", "c", "cert.pem", "Output certificate file")
	Cmd.Flags().StringVarP(&FlagOutKey, "key", "k", "key.pem", "Output private key file")
	Cmd.Flags().StringVarP(&FlagHost, "host", "H", "localhost", "Hostname or IP address")
	Cmd.Flags().IntVarP(&FlagDays, "days", "d", 365, "Certificate validity in days")
}

func generateCert(outCert, outKey, host string, days int) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		logger.Fatalf("error generating key: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(days) * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		logger.Fatalf("error generating serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
			CommonName:   host,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		logger.Fatalf("error generating certificate: %v", err)
	}

	// Save certificate
	certOut, err := os.Create(outCert)
	if err != nil {
		logger.Fatalf("error creating certificate file: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		logger.Fatalf("error writing certificate: %v", err)
	}
	certOut.Close()

	// Save key
	keyOut, err := os.Create(outKey)
	if err != nil {
		logger.Fatalf("error creating key file: %v", err)
	}

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		logger.Fatalf("error converting key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
		logger.Fatalf("error writing key: %v", err)
	}
	keyOut.Close()

	fmt.Printf("Certificate saved to: %s\n", outCert)
	fmt.Printf("Private key saved to: %s\n", outKey)
}

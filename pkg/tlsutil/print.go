package tlsutil

import (
	"crypto/x509"
	"fmt"
	"time"
)

func PrintCertInfo(cert *x509.Certificate) {
	fmt.Printf("Subject Name: %s\n", cert.Subject)
	fmt.Printf("Subject Common Name: %s\n", cert.Subject.CommonName)
	fmt.Printf("Issuer Name: %s\n", cert.Issuer)
	fmt.Printf("Issuer Common Name: %s\n", cert.Issuer.CommonName)
	fmt.Printf("Created: %s\n", cert.NotBefore.Format(time.RFC3339))
	fmt.Printf("Expiry: %s\n", cert.NotAfter.Format(time.RFC3339))
	fmt.Printf("DNS Names: %v\n", cert.DNSNames)
	fmt.Printf("IP Addresses: %v\n", cert.IPAddresses)
	fmt.Println()
}

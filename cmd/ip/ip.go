package ip

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "ip",
	Short: "Displays geolocation information about your public IP address",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := fetchIPInfo()
		if err != nil {
			fmt.Println("Error fetching IP info:", err)
			return
		}

		fmt.Printf("🌐 IP Address: %s\n", info.IP)

		ipv4, ipv6, err := getLocalIPs()
		if ipv4 != "" {
			fmt.Printf("🌐 Local IPv4 Address: %s\n", ipv4)
		}
		if ipv6 != "" {
			fmt.Printf("🌐 Local IPv6 Address: %s\n", ipv6)
		}
		fmt.Printf("🌍 Country: %s\n", info.Country)
		fmt.Printf("🏙️ City: %s\n", info.City)
		fmt.Printf("🏢 ISP: %s\n", info.Org)
	},
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}

type IPInfo struct {
	IP      string `json:"ip"`
	City    string `json:"city"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

func fetchIPInfo() (*IPInfo, error) {
	resp, err := http.Get("https://ipinfo.io/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info IPInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func getLocalIPs() (ipv4, ipv6 string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && ipv4 == "" {
				ipv4 = ipnet.IP.String()
			} else if ipnet.IP.To16() != nil && ipnet.IP.To4() == nil && ipv6 == "" {
				ipv6 = ipnet.IP.String()
			}
		}
	}
	return
}

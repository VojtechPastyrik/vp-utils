package ip

import (
	"encoding/json"
	"fmt"
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	"github.com/spf13/cobra"
	"io"
	"net/http"
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

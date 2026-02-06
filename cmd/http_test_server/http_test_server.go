package ip

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/VojtechPastyrik/vpd/version"
	"github.com/spf13/cobra"
)

var FlagPort string
var FlagLogHeaders bool
var FlagLoBody bool

var Cmd = &cobra.Command{
	Use:     "http-test-server",
	Aliases: []string{"hts"},
	Short:   "Starts a simple HTTP test server",
	Run: func(cmd *cobra.Command, args []string) {
		runHttpTestServer(FlagPort, FlagLogHeaders, FlagLoBody)
	},
}

func init() {
	root.RootCmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagPort,
		"port",
		"p",
		"8000",
		"Port to use for the HTTP request (default is 8000)")
	Cmd.Flags().BoolVarP(
		&FlagLogHeaders,
		"log-headers",
		"l",
		false,
		"Log request headers")
	Cmd.Flags().BoolVarP(
		&FlagLoBody,
		"log-body",
		"b",
		false,
		"Log request body")

}

func runHttpTestServer(port string, logHeaders, logBody bool) {
	hostname, _ := os.Hostname()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("RemoteAddr=%s Context=%s ", r.RemoteAddr, r.RequestURI)
		if logHeaders {
			headers := ""
			for key, values := range r.Header {
				headers += fmt.Sprintf("%s:%v,", key, values)
			}
			if len(headers) > 0 {
				headers = headers[:len(headers)-1]
			}
			fmt.Printf("Headers=[%s] ", headers)
		}
		if logBody {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				fmt.Println("Error reading body:", err)
				http.Error(w, "Unable to read request body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()
			bodyStr := strings.ReplaceAll(string(body), "\n", "")
			fmt.Printf("Body=%s", bodyStr)
		}

		fmt.Fprintf(w, "[vpd "+version.Version+"] Test HTTP Server! %s %s \n", hostname, port)
		fmt.Println()
	})

	fmt.Println("[vpd "+version.Version+"] Server started on 0.0.0.0:"+port+", see http://127.0.0.1:"+port, ", logging headers:", logHeaders, "logging body:", logBody)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

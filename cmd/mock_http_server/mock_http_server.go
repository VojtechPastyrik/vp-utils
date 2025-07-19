package mock_http_server

import (
	"bytes"
	"fmt"
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	"github.com/VojtechPastyrik/vp-utils/version"
	"github.com/dop251/goja"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"k8s.io/apimachinery/pkg/util/uuid"
	"log"
	"net/http"
	"os"
	"strconv"
)

var FlagConfigPath string
var FlagFetExampleConfig bool
var FlagPort int

var Cmd = &cobra.Command{
	Use:   "mock_http_server",
	Short: "Mock HTTP Server",
	Long:  "A simple mock HTTP server for testing purposes.It can be used to simulate API responses and test client applications without needing a real backend.",

	Aliases: []string{"mock-server", "mhs"},
	Run: func(cmd *cobra.Command, args []string) {
		if FlagFetExampleConfig {
			generateConfigExample()
			return
		}
		runMockHTTPServer(FlagPort, FlagConfigPath)
	},
}

func init() {
	root.RootCmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagConfigPath, "config", "c", "", "Path to the configuration file for the mock HTTP server")
	Cmd.Flags().BoolVarP(&FlagFetExampleConfig, "example-config", "e", false, "Fetch example configuration for the mock HTTP server")
	Cmd.Flags().IntVarP(&FlagPort, "port", "p", 8080, "Port on which the mock HTTP server will run")
}

type Config struct {
	Routes []Route `yaml:"routes"`
}

type Route struct {
	Path                string `yaml:"path"`
	Method              string `yaml:"method"`
	Response            string `yaml:"response"`
	ResponseContentType string `yaml:"responseContentType"`
	ResponseCode        int    `yaml:"responseCode"`
	Script              string `yaml:"script"`
}

type Ctx struct {
	Request  *http.Request
	Response http.ResponseWriter
}

func (c *Ctx) GetPayload() (string, error) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil

}

func (c *Ctx) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Ctx) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

func (c *Ctx) SetResponse(status int, body string) {
	c.Response.WriteHeader(status)
	c.Response.Write([]byte(body))
}

func generateConfigExample() {
	exampleConfig := Config{
		Routes: []Route{
			{Path: "/api/v1/resource", Method: "GET", Response: `{"message": "Hello, World!"}`, ResponseCode: 200},
			{
				Path:   "/api/v1/resource",
				Method: "POST",
				Script: `
var payloadString = ctx.getPayload();
log(payloadString);
var payload =JSON.parse(payloadString);
ctx.setResponse(200, JSON.stringify({
message: "Received payload",
payload: payload
}));`,
			},
		},
	}

	yamlData, err := yaml.Marshal(&exampleConfig)
	if err != nil {
		fmt.Printf("Error marshalling to YAML: %v\n", err)
		return
	}

	fmt.Println(string(yamlData))
}

func runMockHTTPServer(port int, configPath string) {
	if FlagConfigPath == "" {
		log.Fatalf("Flag config file path is required")
	}
	portStr := strconv.Itoa(port)
	config, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		log.Fatalf("cannot parse yaml: %v", err)
	}

	r := mux.NewRouter()
	for _, route := range cfg.Routes {
		log.Println("Setting up route:", route.Path, "with method:", route.Method)
		r.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			tracingId := uuid.NewUUID()
			if r.Method != route.Method {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			if route.ResponseContentType != "" {
				w.Header().Set("Content-Type", route.ResponseContentType)
			} else {
				w.Header().Set("Content-Type", "application/json")
			}

			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Obnoví r.Body pro další čtení
			body := string(bodyBytes)
			log.Println("[", tracingId, "] Received request:", r.Method, r.URL.Path, "with body:", body, "from", r.RemoteAddr)

			if route.Script == "" {
				w.WriteHeader(route.ResponseCode)
				w.Write([]byte(route.Response))
				log.Println("[", tracingId, "] Response", route.Path, "with code", route.ResponseCode, "and body", route.Response)
				return
			}

			ctx := &Ctx{Request: r, Response: w}
			vm := goja.New()
			vm.Set("ctx", map[string]interface{}{
				"getPayload": func() string {
					payload, err := ctx.GetPayload()
					if err != nil {
						log.Printf("Error reading payload: %v", err)
						return ""
					}
					log.Println("[", tracingId, "] Payload:", payload)
					return payload
				},
				"setHeader":   ctx.SetHeader,
				"setResponse": ctx.SetResponse,
				"getHeader":   ctx.GetHeader,
			})
			vm.Set("log", func(msg string) {
				log.Println("[", tracingId, "] JS log: ", msg)
			})

			_, err := vm.RunString(route.Script)
			if err != nil {
				log.Printf("[ %s ] Error executing script for route %s: %v", tracingId, route.Path, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			log.Println("[", tracingId, "] Response", route.Path, "with code", route.ResponseCode, "and body", route.Response)
		}).Methods(route.Method)

	}

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "[vp-utils "+version.Version+"] Hello from Mock server! You can access the configured routes.\n")
	})

	log.Println("Mock server listening on :" + portStr)
	log.Fatal(http.ListenAndServe(":"+portStr, r))
}

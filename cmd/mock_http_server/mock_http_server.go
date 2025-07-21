package mock_http_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	"github.com/VojtechPastyrik/vp-utils/version"
	jwtlib "github.com/dgrijalva/jwt-go"
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
	"strings"
	"time"
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

type AuthType string

const (
	AuthTypeBearer AuthType = "oauth2"
	AuthTypeBasic  AuthType = "basic"
)

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Auth struct {
	Type   AuthType          `yaml:"type"`
	Users  []User            `yaml:"users"`
	Claims map[string]string `yaml:"claims"`
}

type Route struct {
	Path                string `yaml:"path"`
	Method              string `yaml:"method"`
	Response            string `yaml:"response"`
	ResponseContentType string `yaml:"responseContentType"`
	ResponseCode        int    `yaml:"responseCode"`
	Script              string `yaml:"script"`
	Auth                Auth   `yaml:"auth"`
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
			{Path: "/api/v1/resource", Method: "GET", Response: `{"message": "Hello, World!"}`, ResponseCode: 200, Auth: Auth{Type: AuthTypeBearer, Claims: map[string]string{
				"sub":  "1234567890",
				"name": "John Doe",
			}}},
			{
				Path:   "/api/v1/resource",
				Method: "POST",
				Auth: Auth{
					Type: AuthTypeBasic,
					Users: []User{
						{Username: "user1", Password: "password1"},
						{Username: "user2", Password: "password2"},
					},
				},
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
	r.Use(AuthMiddleware(cfg.Routes))

	for _, route := range cfg.Routes {
		log.Println("Setting up route:", route.Path, "with method:", route.Method)

		if route.Auth.Type != AuthTypeBearer && route.Auth.Type != AuthTypeBasic {
			log.Fatalf("Invalid auth type: %v. Possible values are: [%s , %s]", route.Auth.Type, AuthTypeBearer, AuthTypeBasic)
		}

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
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the body so it can be read again later
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

func AuthMiddleware(routes []Route) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, route := range routes {
				if route.Path == r.URL.Path && route.Method == r.Method {
					if route.Auth.Type == AuthTypeBearer {
						token := r.Header.Get("Authorization")
						if token == "" || !validateJwtToken(token, route.Auth.Claims) {
							http.Error(w, "Unauthorized", http.StatusUnauthorized)
							return
						}
					} else if route.Auth.Type == AuthTypeBasic {
						username, password, ok := r.BasicAuth()
						if !ok || !validateBasicAuth(username, password, route.Auth.Users) {
							http.Error(w, "Unauthorized", http.StatusUnauthorized)
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func validateBasicAuth(username, password string, users []User) bool {
	for _, user := range users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

func validateJwtToken(token string, claims map[string]string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		log.Printf("Invalid JWT token: expected 3 parts, but found %d", len(parts))
		return false
	}

	claimsJSON, err := jwtlib.DecodeSegment(parts[1])
	if err != nil {
		log.Printf("Error decoding Claims: %v", err)
		return false
	}

	var claimsMap map[string]interface{}
	if err := json.Unmarshal(claimsJSON, &claimsMap); err != nil {
		log.Printf("Error unmarshalling claims JSON: %v", err)
		return false
	}

	// Check if the 'exp' claim exists and is a valid timestamp
	if exp, ok := claimsMap["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			log.Printf("JWT token has expired. Exp: %d, Now: %d", int64(exp), time.Now().Unix())
			return false
		}
	} else {
		log.Printf("JWT token does not contain 'exp' claim")
		return false
	}

	// Check if the claims map contains the required claims
	for key, value := range claims {
		if claimsMap[key] != value {
			log.Printf("JWT token does not contain required claim: %s with value: %s", key, value)
			return false
		}
	}

	return true
}

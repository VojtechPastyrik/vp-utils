package mock_http_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/VojtechPastyrik/vpd/version"
	"github.com/dop251/goja"
	jwtlib "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/uuid"
)

var FlagConfigPath string
var FlagFetExampleConfig bool
var FlagPort int

var Cmd = &cobra.Command{
	Use:   "mock-http-server",
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
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
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

func (c *Ctx) GetURLParam(key string) string {
	vars := mux.Vars(c.Request)
	return vars[key]
}

func (c *Ctx) SetResponse(status int, body string) {
	c.Response.WriteHeader(status)
	c.Response.Write([]byte(body))
}

// Prometheus metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed, labeled by route and status code",
		},
		[]string{"route", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route"},
	)
)

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
		logger.Fatal("flag config file path is required")
	}
	portStr := strconv.Itoa(port)
	config, err := os.ReadFile(configPath)
	if err != nil {
		logger.Fatalf("error reading config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		logger.Fatalf("cannot parse yaml: %v", err)
	}

	r := mux.NewRouter()
	// Auth middleware
	r.Use(AuthMiddleware(cfg.Routes))
	// Metrics middleware
	r.Use(MetricsMiddleware)

	for _, route := range cfg.Routes {
		logger.Info("setting up route:", route.Path, "with method:", route.Method)

		if route.Auth.Type != "" && route.Auth.Type != AuthTypeBearer && route.Auth.Type != AuthTypeBasic {
			logger.Fatalf("invalid auth type: %v. Possible values are: [%s , %s]", route.Auth.Type, AuthTypeBearer, AuthTypeBasic)
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
			logger.Info("[", tracingId, "] received request:", r.Method, r.URL.Path, "with body:", body, "from", r.RemoteAddr)

			if route.Script == "" {
				w.WriteHeader(route.ResponseCode)
				w.Write([]byte(route.Response))
				logger.Info("[", tracingId, "] response", route.Path, "with code", route.ResponseCode, "and body", route.Response)
				return
			}

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			ctx := &Ctx{Request: r, Response: rw}
			vm := goja.New()
			vm.Set("ctx", map[string]interface{}{
				"getPayload": func() string {
					payload, err := ctx.GetPayload()
					if err != nil {
						logger.Infof("error reading payload: %v", err)
						return ""
					}
					logger.Info("[", tracingId, "] payload:", payload)
					return payload
				},
				"setHeader":   ctx.SetHeader,
				"setResponse": ctx.SetResponse,
				"getHeader":   ctx.GetHeader,
				"getURLParam": ctx.GetURLParam,
			})
			vm.Set("log", func(msg string) {
				logger.Info("[", tracingId, "] JS log: ", msg)
			})

			_, err := vm.RunString(route.Script)
			if err != nil {
				logger.Infof("[ %s ] error executing script for route %s: %v", tracingId, route.Path, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			logger.Info("[", tracingId, "] response", route.Path, "with code", rw.statusCode, "and body", rw.body.String())
		}).Methods(route.Method)

	}

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "[vpd "+version.Version+"] Hello from Mock server! You can access the configured routes.\n")
	})

	go func() {
		logger.Info("prometheus metrics available at /metrics")
		if err := http.ListenAndServe(":8090", promhttp.Handler()); err != nil {
			logger.Fatalf("%v", err)
		}
	}()

	logger.Info("mock server listening on :" + portStr)
	if err := http.ListenAndServe(":"+portStr, r); err != nil {
		logger.Fatalf("%v", err)
	}
}

func AuthMiddleware(routes []Route) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, route := range routes {
				router := mux.NewRouter()
				router.HandleFunc(route.Path, func(http.ResponseWriter, *http.Request) {}).Methods(route.Method)
				match := router.Match(r, &mux.RouteMatch{})
				if match {
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
		logger.Infof("invalid JWT token: expected 3 parts, but found %d", len(parts))
		return false
	}

	claimsJSON, err := jwtlib.DecodeSegment(parts[1])
	if err != nil {
		logger.Infof("error decoding claims: %v", err)
		return false
	}

	var claimsMap map[string]interface{}
	if err := json.Unmarshal(claimsJSON, &claimsMap); err != nil {
		logger.Infof("error unmarshalling claims JSON: %v", err)
		return false
	}

	// Check if the 'exp' claim exists and is a valid timestamp
	if exp, ok := claimsMap["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			logger.Infof("JWT token has expired. Exp: %d, Now: %d", int64(exp), time.Now().Unix())
			return false
		}
	} else {
		logger.Info("JWT token does not contain 'exp' claim")
		return false
	}

	// Check if the claims map contains the required claims
	for key, value := range claims {
		if claimsMap[key] != value {
			logger.Infof("JWT token does not contain required claim: %s with value: %s", key, value)
			return false
		}
	}

	return true
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Get the current route and its path
		route := mux.CurrentRoute(r)
		routePath, _ := route.GetPathTemplate()

		// Call the next handler in the chain
		next.ServeHTTP(rw, r)

		// Measure the duration of the request
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(routePath).Observe(duration)

		// Log request total with route path and status code
		httpRequestsTotal.WithLabelValues(routePath, strconv.Itoa(rw.statusCode)).Inc()
	})
}

// ResponseWriter is a custom http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

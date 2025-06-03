package load

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/api"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	FlagURL             string
	FlagMethod          string
	FlagConcurrency     int
	FlagRequests        int
	FlagDuration        int
	FlagContentType     string
	FlagData            string
	FlagDataFile        string
	FlagHeaders         []string
	FlagAuthType        string
	FlagUsername        string
	FlagPassword        string
	FlagOAuthClientID   string
	FlagOAuthSecret     string
	FlagOAuthTokenURL   string
	FlagOAuthScopes     []string
	FlagBearerToken     string
	FlagInsecure        bool
	FlagVerbose         bool
	FlagOutputFormat    string
	FlagOutputFile      string
	FlagTimeoutSeconds  int
	FlagRampUpSeconds   int
	FlagThinkTimeMillis int
)

var Cmd = &cobra.Command{
	Use:     "load",
	Short:   "Execute a load test for REST or SOAP API",
	Aliases: []string{"l", "load"},
	Long: `Tool for load testing API endpoints with support for various protocols and authentication methods.
Supports REST and SOAP, multiple HTTP methods, authentication via Basic Auth, OAuth2, or Bearer token.
Allows defining the number of concurrent requests, total requests, or test duration.`,
	Example: `  # Simple GET request with 10 concurrent users, total 1000 requests
  vp-utils api load-test --url https://api.example.com/endpoint --concurrency 10 --requests 1000

  # POST request with JSON body and OAuth2 authentication
  vp-utils api load-test --url https://api.example.com/endpoint --method POST --content-type "application/json" \
    --data '{"key": "value"}' --auth-type oauth --oauth-client-id "client_id" --oauth-secret "secret" \
    --oauth-token-url "https://auth.example.com/token"

  # SOAP request with Basic authentication
  vp-utils api load-test --url https://api.example.com/soap --method POST --content-type "text/xml" \
    --data-file request.xml --auth-type basic --username "user" --password "pass"`,
	Run: func(cmd *cobra.Command, args []string) {
		runLoadTest()
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)

	// Basic request parameters
	Cmd.Flags().StringVarP(&FlagURL, "url", "u", "", "URL of the endpoint to test")
	Cmd.MarkFlagRequired("url")
	Cmd.Flags().StringVarP(&FlagMethod, "method", "m", "GET", "HTTP method (GET, POST, PUT, DELETE, etc.)")
	Cmd.Flags().StringVarP(&FlagContentType, "content-type", "c", "application/json", "Content-Type header")
	Cmd.Flags().StringVarP(&FlagData, "data", "d", "", "Request data (body)")
	Cmd.Flags().StringVarP(&FlagDataFile, "data-file", "f", "", "File with request data")
	Cmd.Flags().StringArrayVarP(&FlagHeaders, "header", "H", []string{}, "Custom HTTP headers (format: 'Key:Value')")

	// Load test parameters
	Cmd.Flags().IntVarP(&FlagConcurrency, "concurrency", "n", 1, "Number of concurrent users")
	Cmd.Flags().IntVarP(&FlagRequests, "requests", "r", 0, "Total number of requests (0 = unlimited)")
	Cmd.Flags().IntVarP(&FlagDuration, "duration", "t", 0, "Duration of the test in seconds (0 = until all requests complete)")
	Cmd.Flags().IntVar(&FlagTimeoutSeconds, "timeout", 30, "Request timeout in seconds")
	Cmd.Flags().IntVar(&FlagRampUpSeconds, "ramp-up", 0, "Ramp-up time for users in seconds")
	Cmd.Flags().IntVar(&FlagThinkTimeMillis, "think-time", 0, "Delay between requests in milliseconds")

	// Authentication
	Cmd.Flags().StringVar(&FlagAuthType, "auth-type", "", "Authentication type (basic, oauth, bearer)")
	Cmd.Flags().StringVar(&FlagUsername, "username", "", "Username for Basic Auth")
	Cmd.Flags().StringVar(&FlagPassword, "password", "", "Password for Basic Auth")
	Cmd.Flags().StringVar(&FlagOAuthClientID, "oauth-client-id", "", "OAuth2 Client ID")
	Cmd.Flags().StringVar(&FlagOAuthSecret, "oauth-secret", "", "OAuth2 Client Secret")
	Cmd.Flags().StringVar(&FlagOAuthTokenURL, "oauth-token-url", "", "OAuth2 Token URL")
	Cmd.Flags().StringArrayVar(&FlagOAuthScopes, "oauth-scope", []string{}, "OAuth2 scopes")
	Cmd.Flags().StringVar(&FlagBearerToken, "bearer-token", "", "Bearer token for authentication")

	// Other options
	Cmd.Flags().BoolVar(&FlagInsecure, "insecure", false, "Skip SSL certificate verification")
	Cmd.Flags().BoolVarP(&FlagVerbose, "verbose", "v", false, "Show detailed output")
	Cmd.Flags().StringVar(&FlagOutputFormat, "output", "text", "Output format (text, json, csv)")
	Cmd.Flags().StringVar(&FlagOutputFile, "output-file", "", "File to save results")
}

type TestResult struct {
	TotalRequests      int           `json:"totalRequests"`
	SuccessfulRequests int           `json:"successfulRequests"`
	FailedRequests     int           `json:"failedRequests"`
	TotalDuration      time.Duration `json:"totalDuration"`
	MinResponseTime    time.Duration `json:"minResponseTime"`
	MaxResponseTime    time.Duration `json:"maxResponseTime"`
	AvgResponseTime    time.Duration `json:"avgResponseTime"`
	RPS                float64       `json:"requestsPerSecond"`
	StatusCodes        map[int]int   `json:"statusCodes"`
	Errors             []string      `json:"errors"`
}

type RequestResult struct {
	Duration    time.Duration
	StatusCode  int
	Error       error
	ContentSize int64
}

func runLoadTest() {
	fmt.Println("Starting load test...")
	fmt.Printf("URL: %s\n", FlagURL)
	fmt.Printf("Method: %s\n", FlagMethod)
	fmt.Printf("Concurrent users: %d\n", FlagConcurrency)

	// Create request body from file or string
	var requestBody []byte
	if FlagDataFile != "" {
		var err error
		requestBody, err = os.ReadFile(FlagDataFile)
		if err != nil {
			fmt.Printf("Error reading data file: %v\n", err)
			return
		}
	} else if FlagData != "" {
		requestBody = []byte(FlagData)
	}

	// Create HTTP client
	client := createHTTPClient()

	// Initialize results
	results := TestResult{
		MinResponseTime: time.Hour, // High initial value to ensure any real response time is lower
		StatusCodes:     make(map[int]int),
		Errors:          make([]string, 0),
	}

	// Channel for results from workers
	resultsChan := make(chan RequestResult, FlagConcurrency*10)

	// Create worker pool
	var wg sync.WaitGroup
	requestsCounter := int32(0)
	done := make(chan struct{})

	// Graceful shutdown na signály
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		fmt.Println("\nSignal caught, stopping the test...")
		close(done)
	}()

	startTime := time.Now()

	// Start worker goroutines
	for i := 0; i < FlagConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Ramp-up delay if specified
			if FlagRampUpSeconds > 0 {
				delay := time.Duration(FlagRampUpSeconds) * time.Second / time.Duration(FlagConcurrency) * time.Duration(workerID)
				time.Sleep(delay)
			}

			for {
				select {
				case <-done:
					return // do not start new requests
				default:
					// Check request count limit
					if FlagRequests > 0 && int(atomic.AddInt32(&requestsCounter, 1)) > FlagRequests {
						return
					}
					result := executeRequest(client, requestBody, done)
					resultsChan <- result
					if FlagThinkTimeMillis > 0 {
						time.Sleep(time.Duration(FlagThinkTimeMillis) * time.Millisecond)
					}
				}
			}
		}(i)
	}

	// Timer for test duration
	if FlagDuration > 0 {
		go func() {
			time.Sleep(time.Duration(FlagDuration) * time.Second)
			close(done)
		}()
	}

	// Ticker for verbose output
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if FlagVerbose {
					fmt.Printf("Progress: %d requests sent\n", atomic.LoadInt32(&requestsCounter))
				}
			case <-done:
				return
			}
		}
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Results collection
	totalResponseTime := time.Duration(0)

	for result := range resultsChan {
		// If the result is a context error (canceled or deadline exceeded), skip it
		if result.Error != nil && (errors.Is(result.Error, context.Canceled) || errors.Is(result.Error, context.DeadlineExceeded)) {
			continue
		}

		results.TotalRequests++

		if result.Error != nil {
			results.FailedRequests++
			results.Errors = append(results.Errors, result.Error.Error())
			if FlagVerbose {
				fmt.Printf("Error: %v\n", result.Error)
			}
		} else {
			results.StatusCodes[result.StatusCode]++
			if result.StatusCode >= 400 && result.StatusCode <= 599 {
				results.FailedRequests++
				results.Errors = append(results.Errors, fmt.Sprintf("HTTP %d", result.StatusCode))
				if FlagVerbose {
					fmt.Printf("Error: HTTP %d\n", result.StatusCode)
				}
			} else {
				results.SuccessfulRequests++
				if result.Duration < results.MinResponseTime {
					results.MinResponseTime = result.Duration
				}
				if result.Duration > results.MaxResponseTime {
					results.MaxResponseTime = result.Duration
				}
				totalResponseTime += result.Duration
			}
		}
	}

	// Final statistics
	results.TotalDuration = time.Since(startTime)

	if results.SuccessfulRequests > 0 {
		results.AvgResponseTime = totalResponseTime / time.Duration(results.SuccessfulRequests)
		results.RPS = float64(results.SuccessfulRequests) / results.TotalDuration.Seconds()
	}

	// Print results
	printResults(results)

	// Save results to file
	if FlagOutputFile != "" {
		saveResults(results)
	}
}

func createHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: time.Duration(FlagTimeoutSeconds) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        FlagConcurrency,
			MaxIdleConnsPerHost: FlagConcurrency,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	switch FlagAuthType {
	case "oauth":
		if FlagOAuthClientID != "" && FlagOAuthSecret != "" && FlagOAuthTokenURL != "" {
			config := clientcredentials.Config{
				ClientID:     FlagOAuthClientID,
				ClientSecret: FlagOAuthSecret,
				TokenURL:     FlagOAuthTokenURL,
				Scopes:       FlagOAuthScopes,
			}
			ctx := context.Background()
			client = config.Client(ctx)
			client.Timeout = time.Duration(FlagTimeoutSeconds) * time.Second
		} else {
			fmt.Println("Missing parameters for OAuth2 authentication")
		}
	}

	return client
}

func executeRequest(client *http.Client, requestBody []byte, done <-chan struct{}) RequestResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(FlagTimeoutSeconds)*time.Second)
	defer cancel()

	// Handle cancellation on done channel
	go func() {
		select {
		case <-done:
			cancel()
		case <-ctx.Done():
		}
	}()

	req, err := http.NewRequestWithContext(ctx, FlagMethod, FlagURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return RequestResult{Error: err}
	}

	// Set headers
	req.Header.Set("Content-Type", FlagContentType)
	for _, header := range FlagHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// Authentication
	switch FlagAuthType {
	case "basic":
		if FlagUsername != "" {
			req.SetBasicAuth(FlagUsername, FlagPassword)
		}
	case "bearer":
		if FlagBearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+FlagBearerToken)
		}
	}

	// Measure time and execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return RequestResult{
			Duration: duration,
			Error:    err,
		}
	}
	defer resp.Body.Close()

	// Read response body to get size
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return RequestResult{
			Duration:   duration,
			StatusCode: resp.StatusCode,
			Error:      err,
		}
	}

	return RequestResult{
		Duration:    duration,
		StatusCode:  resp.StatusCode,
		ContentSize: int64(len(body)),
	}
}

func printResults(results TestResult) {
	fmt.Println("\n\n--- Load Test Results ---")
	fmt.Printf("Total requests: %d\n", results.TotalRequests)
	fmt.Printf("Successful: %d\n", results.SuccessfulRequests)
	fmt.Printf("Failed: %d\n", results.FailedRequests)
	fmt.Printf("Total test duration: %.2f seconds\n", results.TotalDuration.Seconds())

	if results.SuccessfulRequests > 0 {
		fmt.Printf("Min response time: %.2f ms\n", float64(results.MinResponseTime.Microseconds())/1000)
		fmt.Printf("Max response time: %.2f ms\n", float64(results.MaxResponseTime.Microseconds())/1000)
		fmt.Printf("Average response time: %.2f ms\n", float64(results.AvgResponseTime.Microseconds())/1000)
		fmt.Printf("Requests per second: %.2f\n", results.RPS)
	}

	fmt.Println("\nStatus codes:")
	for code, count := range results.StatusCodes {
		fmt.Printf("  %d: %d\n", code, count)
	}

	if len(results.Errors) > 0 {
		fmt.Println("\nSample errors:")
		for i, err := range results.Errors {
			if i >= 5 {
				fmt.Printf("... and %d more errors\n", len(results.Errors)-5)
				break
			}
			fmt.Printf("  - %s\n", err)
		}
	}
}

func saveResults(results TestResult) {
	var data []byte
	var err error

	switch FlagOutputFormat {
	case "json":
		data, err = json.MarshalIndent(results, "", "  ")
	case "csv":
		var buffer bytes.Buffer
		buffer.WriteString("Total Requests,Successful,Failed,Total Duration (s),Min Response (ms),Max Response (ms),Avg Response (ms),RPS\n")
		buffer.WriteString(fmt.Sprintf("%d,%d,%d,%.2f,%.2f,%.2f,%.2f,%.2f\n",
			results.TotalRequests,
			results.SuccessfulRequests,
			results.FailedRequests,
			results.TotalDuration.Seconds(),
			float64(results.MinResponseTime.Microseconds())/1000,
			float64(results.MaxResponseTime.Microseconds())/1000,
			float64(results.AvgResponseTime.Microseconds())/1000,
			results.RPS))

		buffer.WriteString("\nStatus Codes\n")
		for code, count := range results.StatusCodes {
			buffer.WriteString(fmt.Sprintf("%d,%d\n", code, count))
		}

		data = buffer.Bytes()
	default:
		var buffer bytes.Buffer
		buffer.WriteString("--- Load Test Results ---\n")
		buffer.WriteString(fmt.Sprintf("Total requests: %d\n", results.TotalRequests))
		buffer.WriteString(fmt.Sprintf("Successful: %d\n", results.SuccessfulRequests))
		buffer.WriteString(fmt.Sprintf("Failed: %d\n", results.FailedRequests))
		buffer.WriteString(fmt.Sprintf("Total test duration: %.2f seconds\n", results.TotalDuration.Seconds()))

		if results.SuccessfulRequests > 0 {
			buffer.WriteString(fmt.Sprintf("Min response time: %.2f ms\n", float64(results.MinResponseTime.Microseconds())/1000))
			buffer.WriteString(fmt.Sprintf("Max response time: %.2f ms\n", float64(results.MaxResponseTime.Microseconds())/1000))
			buffer.WriteString(fmt.Sprintf("Average response time: %.2f ms\n", float64(results.AvgResponseTime.Microseconds())/1000))
			buffer.WriteString(fmt.Sprintf("Requests per second: %.2f\n", results.RPS))
		}

		buffer.WriteString("\nStatus codes:\n")
		for code, count := range results.StatusCodes {
			buffer.WriteString(fmt.Sprintf("  %d: %d\n", code, count))
		}

		data = buffer.Bytes()
	}

	if err != nil {
		fmt.Printf("Error serializing results: %v\n", err)
		return
	}

	err = os.WriteFile(FlagOutputFile, data, 0644)
	if err != nil {
		fmt.Printf("Error saving results to file: %v\n", err)
		return
	}

	fmt.Printf("Results saved to file: %s\n", FlagOutputFile)
}

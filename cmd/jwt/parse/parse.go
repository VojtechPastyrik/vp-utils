package parse

import (
	"bufio"
	"encoding/json"
	"fmt"
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/jwt"
	jwtlib "github.com/golang-jwt/jwt/v4"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"time"
)

var Cmd = &cobra.Command{
	Use:   "parse <jwt>",
	Short: "Parse JWT from stdin into JSON list of 3 objects (Header, Payload, Signature)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jwtToken := args[0]
		if jwtToken == "-" {
			jwtToken = readFromPipe()
		}
		parseJWT(jwtToken)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

func parseJWT(jwtToken string) {
	// JWT token should be in the format: Header.Payload.Signature
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		log.Fatalf("Neplatný JWT token: očekávány 3 části, ale nalezeno %d", len(parts))
	}

	// Decoding
	headerJSON, err := jwtlib.DecodeSegment(parts[0])
	if err != nil {
		log.Fatalf("Chyba při dekódování Header: %v", err)
	}

	claimsJSON, err := jwtlib.DecodeSegment(parts[1])
	if err != nil {
		log.Fatalf("Chyba při dekódování Claims: %v", err)
	}

	signature := parts[2]

	// Result creation as a slice of interfaces
	result := []interface{}{
		decodeJSON(headerJSON),
		decodeJSON(claimsJSON),
		signature,
	}

	// Result output as JSON
	outputJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal("Chyba při serializaci výsledku do JSON: ", err)
	}
	fmt.Println(string(outputJSON))
}

func readFromPipe() string {
	// Ensure input is from a pipe
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeNamedPipe == 0 {
		log.Fatalln("No input from pipe.")
	}

	// Read the input from stdin (pipe)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}

	// Handle errors during scanning
	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading from stdin: ", err)
	}

	return ""
}

func decodeJSON(data []byte) interface{} {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		log.Fatal("Error unmarshalling JSON: ", err)
	}

	// Object contains timestamps, format them
	for key, value := range obj {
		if timestamp, ok := value.(float64); ok && isTimestamp(timestamp) {
			obj[key] = formatTimestamp(timestamp)
		}
	}

	return obj
}

// Check if the value is a valid timestamp
func isTimestamp(value float64) bool {
	// Timestamp by měl být větší než 0 a menší než aktuální čas + 100 let
	now := time.Now().Unix()
	return value > 0 && value < float64(now+100*365*24*60*60)
}

// Format timestampt to reader-friendly string
func formatTimestamp(timestamp float64) string {
	loc := time.Now().Location()
	return time.Unix(int64(timestamp), 0).In(loc).Format("2006-01-02 15:04:05")
}

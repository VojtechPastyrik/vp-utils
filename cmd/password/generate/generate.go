package generate

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/password"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var (
	length    int
	count     int
	noSpecial bool
	noUpper   bool
	noDigits  bool
	noCopy    bool
	easyCopy  bool
)

var Cmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate cryptographically secure random passwords",
	Run: func(cmd *cobra.Command, args []string) {
		charset := buildCharset()
		if len(charset) == 0 {
			logger.Fatal("no characters available with the given flags")
		}
		var passwords []string
		for i := 0; i < count; i++ {
			pw, err := generatePassword(charset, length)
			if err != nil {
				logger.Fatalf("error generating password: %v", err)
			}
			passwords = append(passwords, pw)
			fmt.Println(pw)
		}
		if !noCopy {
			text := strings.Join(passwords, "\n")
			if err := clipboard.WriteAll(text); err != nil {
				// silently skip clipboard if unavailable (e.g. headless server)
			} else if count == 1 {
				fmt.Println("(copied to clipboard)")
			} else {
				fmt.Printf("(all %d passwords copied to clipboard)\n", count)
			}
		}
	},
}

func init() {
	Cmd.Flags().IntVarP(&length, "length", "l", 16, "Password length")
	Cmd.Flags().IntVarP(&count, "count", "c", 1, "Number of passwords to generate")
	Cmd.Flags().BoolVar(&noSpecial, "no-special", false, "Exclude special characters")
	Cmd.Flags().BoolVar(&noUpper, "no-upper", false, "Exclude uppercase letters")
	Cmd.Flags().BoolVar(&noDigits, "no-digits", false, "Exclude digits")
	Cmd.Flags().BoolVarP(&easyCopy, "easy-copy", "e", false, "Use only double-click friendly chars (letters, digits, underscore)")
	Cmd.Flags().BoolVar(&noCopy, "no-copy", false, "Don't copy to clipboard")
	parent_cmd.Cmd.AddCommand(Cmd)
}

func buildCharset() string {
	if easyCopy {
		// Only characters selectable by double-click: letters, digits, underscore
		var b strings.Builder
		b.WriteString("abcdefghijklmnopqrstuvwxyz")
		b.WriteString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		b.WriteString("0123456789")
		b.WriteString("_")
		return b.String()
	}

	var b strings.Builder
	b.WriteString("abcdefghijklmnopqrstuvwxyz")
	if !noUpper {
		b.WriteString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	}
	if !noDigits {
		b.WriteString("0123456789")
	}
	if !noSpecial {
		b.WriteString("!@#$%^&*()-_=+[]{}|;:,.<>?")
	}
	return b.String()
}

func generatePassword(charset string, length int) (string, error) {
	result := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

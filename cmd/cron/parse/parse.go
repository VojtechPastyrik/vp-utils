package parse

import (
	"fmt"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/cron"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "parse <expression>",
	Short: "Show human-readable description of a cron expression",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		expr := args[0]
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, err := parser.Parse(expr)
		if err != nil {
			logger.Fatalf("invalid cron expression: %v", err)
		}

		fmt.Printf("Expression: %s\n", expr)
		fmt.Printf("Meaning:    %s\n", describeCron(expr))
		_ = schedule
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
}

func describeCron(expr string) string {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return "non-standard cron expression"
	}

	minute, hour, dom, month, dow := fields[0], fields[1], fields[2], fields[3], fields[4]

	var parts []string

	// Minute
	switch {
	case minute == "*":
		parts = append(parts, "every minute")
	case strings.HasPrefix(minute, "*/"):
		parts = append(parts, fmt.Sprintf("every %s minutes", minute[2:]))
	default:
		parts = append(parts, fmt.Sprintf("at minute %s", minute))
	}

	// Hour
	switch {
	case hour == "*":
		// already implied by minute
	case strings.HasPrefix(hour, "*/"):
		parts = append(parts, fmt.Sprintf("every %s hours", hour[2:]))
	default:
		parts = append(parts, fmt.Sprintf("at hour %s", hour))
	}

	// Day of month
	if dom != "*" {
		parts = append(parts, fmt.Sprintf("on day %s of the month", dom))
	}

	// Month
	if month != "*" {
		parts = append(parts, fmt.Sprintf("in month %s", month))
	}

	// Day of week
	if dow != "*" {
		parts = append(parts, fmt.Sprintf("on %s", dow))
	}

	return strings.Join(parts, ", ")
}

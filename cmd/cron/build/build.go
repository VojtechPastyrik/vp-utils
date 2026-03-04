package build

import (
	"fmt"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/cron"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var (
	minute     string
	hour       string
	dayOfMonth string
	month      string
	dayOfWeek  string
	copy       bool
)

var presets = map[string]string{
	"every-minute":      "* * * * *",
	"hourly":            "0 * * * *",
	"daily":             "0 0 * * *",
	"weekly":            "0 0 * * 0",
	"monthly":           "0 0 1 * *",
	"yearly":            "0 0 1 1 *",
	"weekdays-9am":      "0 9 * * 1-5",
	"every-5min":        "*/5 * * * *",
	"every-15min":       "*/15 * * * *",
	"every-30min":       "*/30 * * * *",
	"every-2hours":      "0 */2 * * *",
	"every-6hours":      "0 */6 * * *",
	"midnight":          "0 0 * * *",
	"business-hours":    "0 9-17 * * 1-5",
	"twice-daily":       "0 0,12 * * *",
	"weekdays-morning":  "0 8 * * 1-5",
	"weekdays-evening":  "0 18 * * 1-5",
	"first-of-month":    "0 0 1 * *",
	"last-day-of-month": "0 0 28 * *",
	"quarterly":         "0 0 1 1,4,7,10 *",
}

var Cmd = &cobra.Command{
	Use:   "build [preset]",
	Short: "Build a cron expression from flags or common presets",
	Long: `Build a cron expression from flags or use a preset name.

Presets: every-minute, hourly, daily, weekly, monthly, yearly,
  weekdays-9am, every-5min, every-15min, every-30min,
  every-2hours, every-6hours, midnight, business-hours,
  twice-daily, weekdays-morning, weekdays-evening,
  first-of-month, last-day-of-month, quarterly

Examples:
  vpd cron build weekdays-9am
  vpd cron build --minute "*/5" --hour "9-17" --dow "1-5"
  vpd cron build daily --copy

Field reference:
  Minute:       0-59, *, */N, N-M
  Hour:         0-23, *, */N, N-M
  Day of month: 1-31, *, */N, N-M
  Month:        1-12 or JAN-DEC, *, */N, N-M
  Day of week:  0-7 (0 and 7 = Sunday) or SUN-SAT, *, N-M`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var expr string

		if len(args) == 1 {
			preset, ok := presets[args[0]]
			if !ok {
				fmt.Printf("Unknown preset %q. Available presets:\n", args[0])
				printPresets()
				return
			}
			expr = preset
		} else if flagsSet(cmd) {
			expr = strings.Join([]string{minute, hour, dayOfMonth, month, dayOfWeek}, " ")
		} else {
			fmt.Println("Available presets:")
			printPresets()
			fmt.Println("\nOr build custom: vpd cron build --minute '*/5' --hour '9-17' --dow '1-5'")
			return
		}

		fmt.Println(expr)

		if copy {
			if err := clipboard.WriteAll(expr); err != nil {
				fmt.Printf("(could not copy to clipboard: %v)\n", err)
			} else {
				fmt.Println("(copied to clipboard)")
			}
		}
	},
}

func init() {
	Cmd.Flags().StringVar(&minute, "minute", "*", "Minute field (0-59)")
	Cmd.Flags().StringVar(&hour, "hour", "*", "Hour field (0-23)")
	Cmd.Flags().StringVar(&dayOfMonth, "dom", "*", "Day of month field (1-31)")
	Cmd.Flags().StringVar(&month, "month", "*", "Month field (1-12 or JAN-DEC)")
	Cmd.Flags().StringVar(&dayOfWeek, "dow", "*", "Day of week field (0-7 or SUN-SAT)")
	Cmd.Flags().BoolVarP(&copy, "copy", "c", false, "Copy result to clipboard")
	parent_cmd.Cmd.AddCommand(Cmd)
}

func flagsSet(cmd *cobra.Command) bool {
	for _, name := range []string{"minute", "hour", "dom", "month", "dow"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func printPresets() {
	order := []string{
		"every-minute", "every-5min", "every-15min", "every-30min",
		"hourly", "every-2hours", "every-6hours",
		"daily", "midnight", "twice-daily",
		"weekly", "monthly", "quarterly", "yearly",
		"weekdays-9am", "weekdays-morning", "weekdays-evening",
		"business-hours", "first-of-month", "last-day-of-month",
	}
	for _, name := range order {
		fmt.Printf("  %-20s %s\n", name, presets[name])
	}
}

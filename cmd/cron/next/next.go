package next

import (
	"fmt"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/cron"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var count int

var Cmd = &cobra.Command{
	Use:   "next <expression>",
	Short: "Show next scheduled execution times for a cron expression",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		expr := args[0]
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, err := parser.Parse(expr)
		if err != nil {
			logger.Fatalf("invalid cron expression: %v", err)
		}

		now := time.Now()
		fmt.Printf("Next %d execution times for %q:\n", count, expr)
		for i := 0; i < count; i++ {
			now = schedule.Next(now)
			fmt.Printf("  %d. %s\n", i+1, now.Format("2006-01-02 15:04:05 MST"))
		}
	},
}

func init() {
	Cmd.Flags().IntVarP(&count, "count", "c", 5, "Number of next execution times to show")
	parent_cmd.Cmd.AddCommand(Cmd)
}

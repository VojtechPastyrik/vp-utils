package cron

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cron",
	Short: "Cron expression utilities",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}

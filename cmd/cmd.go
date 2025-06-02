package cmd

import (
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/vault"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/vault/init_unseal"
	"github.com/spf13/cobra"
)

func Execute() {
	cobra.CheckErr(root.RootCmd.Execute())
}

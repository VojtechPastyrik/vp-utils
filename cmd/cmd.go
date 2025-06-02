package cmd

import (
	_ "github.com/VojtechPastyrik/vp-utils/cmd/jwt"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/jwt/parse"
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/vault"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/vault/init_unseal"
	"github.com/spf13/cobra"
)

func Execute() {
	cobra.CheckErr(root.RootCmd.Execute())
}

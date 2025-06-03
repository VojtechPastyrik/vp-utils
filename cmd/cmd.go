package cmd

import (
	_ "github.com/VojtechPastyrik/vp-utils/cmd/base64"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/base64/interactive_clipboard"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/git"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/git/clean_branches"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/jwt"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/jwt/parse"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/k8s"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/k8s/add_context"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/k8s/remove_context"
	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/tls"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/tls/parse"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/tls/parse_file"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/vault"
	_ "github.com/VojtechPastyrik/vp-utils/cmd/vault/init_unseal"
	"github.com/spf13/cobra"
)

func Execute() {
	cobra.CheckErr(root.RootCmd.Execute())
}

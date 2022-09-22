package cmd

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
	"github.com/zbysir/hollow/internal/bblog/editor"
	"github.com/zbysir/hollow/internal/pkg/config"
	"github.com/zbysir/hollow/internal/pkg/log"
	"github.com/zbysir/hollow/internal/pkg/signal"
)

type EditorParams struct {
	Address       string `json:"address"`
	Source        string `json:"source"`
	PreviewDomain string `json:"preview_domain"`
	Secret        string `json:"secret"`
}

var Editor = &cobra.Command{
	Use:   "editor",
	Short: "editor start a web service",
	//Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.AutomaticEnv()
		p, err := config.Get[EditorParams]()
		if err != nil {
			return err
		}

		if p.Secret == "" {
			p.Secret = funk.RandomString(8)
		}

		//gin.SetMode(gin.ReleaseMode)
		log.Infof("config: %+v", p)

		e := editor.NewEditor(func(pid int64) (billy.Filesystem, error) {
			return osfs.New(p.Source), nil
		}, editor.Config{
			PreviewDomain: p.PreviewDomain,
			Secret:        p.Secret,
		})

		ctx, c := signal.NewContext()
		defer c()
		err = e.Run(ctx, p.Address)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	config.DeclareFlag(Editor, "address", "a", ":9432", "service listen address")
	config.DeclareFlag(Editor, "source", "s", "source", "source file dir")
	config.DeclareFlag(Editor, "preview_domain", "p", "", "preview website with the domain ")
	config.DeclareFlag(Editor, "secret", "c", "", "secret for web ui")
}

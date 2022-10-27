package cmd

import (
	"errors"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zbysir/hollow/internal/hollow"
	"github.com/zbysir/hollow/internal/pkg/config"
	"github.com/zbysir/hollow/internal/pkg/log"
	"github.com/zbysir/hollow/internal/pkg/signal"
	"net/http"
)

type ServerParams struct {
	Address string `json:"address"`
	Source  string `json:"source"`
	Theme   string `json:"theme"`
}

var Server = &cobra.Command{
	Use:   "server",
	Short: "preview your website",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.AutomaticEnv()
		p, err := config.Get[ServerParams]()
		if err != nil {
			return err
		}
		ctx, c := signal.NewContext()
		defer c()

		//gin.SetMode(gin.ReleaseMode)
		log.Infof("config: %+v", p)

		b, err := hollow.NewHollow(hollow.Option{
			Fs: osfs.New(p.Source),
		})
		if err != nil {
			return err
		}

		addr := p.Address
		log.Infof("listening %v", addr)
		theme := p.Theme
		err = b.Service(ctx, hollow.ExecOption{
			Log:   nil,
			IsDev: true,
			Theme: theme,
		}, addr)
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return err
		}

		return nil
	},
}

func init() {
	config.DeclareFlag(Server, "address", "a", ":9400", "server listen address")
	config.DeclareFlag(Server, "source", "s", "source", "source file dir")
	config.DeclareFlag(Server, "theme", "t", "", "Specify Theme")
}

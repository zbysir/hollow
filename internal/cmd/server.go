package cmd

import (
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zbysir/hollow/internal/hollow"
	"github.com/zbysir/hollow/internal/pkg/config"
	"github.com/zbysir/hollow/internal/pkg/log"
	"github.com/zbysir/hollow/internal/pkg/signal"
	"sync"
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

		h, err := hollow.NewHollow(hollow.Option{
			SourceFs:   osfs.New(p.Source),
			FixedTheme: p.Theme,
		})
		if err != nil {
			panic(err)
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()

			addr := p.Address
			log.Infof("listening %v", addr)
			err = h.Service(ctx, hollow.ExecOption{
				Log:   nil,
				IsDev: true,
			}, addr)

			if err != nil {

				panic(err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			err = h.DevService(ctx)
			if err != nil {
				panic(err)
			}
		}()

		wg.Wait()
		return nil
	},
}

func init() {
	config.DeclareFlag(Server, "address", "a", ":9400", "server listen address")
	config.DeclareFlag(Server, "source", "s", "source", "source file dir")
	config.DeclareFlag(Server, "theme", "t", "", "Specify Theme")
}

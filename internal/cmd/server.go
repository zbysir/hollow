package cmd

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
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
	Cache   string `json:"cache"`
}

func Server() *cobra.Command {
	v := viper.New()
	v.AutomaticEnv()

	cmd := &cobra.Command{
		Use:   "server",
		Short: "preview your website",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.Get[ServerParams](v)
			if err != nil {
				return err
			}
			ctx, c := signal.NewContext()
			defer c()

			//gin.SetMode(gin.ReleaseMode)
			log.Infof("config: %+v", p)

			var cacheFs billy.Filesystem
			switch p.Cache {
			case "memory":
				cacheFs = memfs.New()
			default:
				cacheFs = osfs.New(p.Cache)
			}

			h, err := hollow.NewHollow(hollow.Option{
				SourceFs:   osfs.New(p.Source),
				FixedTheme: p.Theme,
				CacheFs:    cacheFs,
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

	config.DeclareFlag(v, cmd, "address", "a", ":9400", "server listen address")
	config.DeclareFlag(v, cmd, "source", "s", ".", "source file dir")
	config.DeclareFlag(v, cmd, "theme", "t", "", "specify theme")
	config.DeclareFlag(v, cmd, "cache", "c", "memory", "cache file path, default in memory")
	return cmd
}

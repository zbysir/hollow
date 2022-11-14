package cmd

import (
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zbysir/hollow/internal/hollow"
	"github.com/zbysir/hollow/internal/pkg/config"
	"github.com/zbysir/hollow/internal/pkg/log"
	"github.com/zbysir/hollow/internal/pkg/signal"
)

type BuildParams struct {
	Output string `json:"output"`
	Source string `json:"source"`
}

var Build = &cobra.Command{
	Use:   "build",
	Short: "build your website",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.AutomaticEnv()
		p, err := config.Get[BuildParams]()
		if err != nil {
			return err
		}
		ctx, c := signal.NewContext()
		defer c()

		//gin.SetMode(gin.ReleaseMode)
		log.Infof("config: %+v", p)

		b, err := hollow.NewHollow(hollow.Option{
			SourceFs: osfs.New(p.Source),
		})
		if err != nil {
			return err
		}

		err = b.Build(ctx, p.Output, hollow.ExecOption{IsDev: true})
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	config.DeclareFlag(Build, "output", "o", "./dist", "output dir")
	config.DeclareFlag(Build, "source", "s", ".", "source file dir")
}

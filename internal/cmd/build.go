package cmd

import (
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zbysir/hollow/internal/hollow"
	"github.com/zbysir/hollow/internal/pkg/config"
	"github.com/zbysir/hollow/internal/pkg/log"
)

type BuildParams struct {
	Output string `json:"output"`
	Source string `json:"source2"`
}

func Build() *cobra.Command {
	v := viper.New()
	v.AutomaticEnv()

	cmd := &cobra.Command{
		Use:   "build",
		Short: "build your website",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.Get[BuildParams](v)
			if err != nil {
				return err
			}
			//ctx, c := signal.NewContext()
			//defer c()

			//gin.SetMode(gin.ReleaseMode)
			log.Infof("config: %+v", p)

			ho, err := hollow.NewHollow(hollow.Option{
				SourceFs: osfs.New(p.Source),
			})
			if err != nil {
				return err
			}

			err = ho.Build(hollow.NewRenderContext(), p.Output, hollow.ExecOption{IsDev: true})
			if err != nil {
				return err
			}

			return nil
		},
	}

	config.DeclareFlag(v, cmd, "output", "o", "./dist", "output dir")
	config.DeclareFlag(v, cmd, "source", "s", ".", "source file dir")
	return cmd
}

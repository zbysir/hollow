package cmd

import (
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zbysir/hollow/internal/hollow"
	"github.com/zbysir/hollow/internal/pkg/log"
	"github.com/zbysir/hollow/internal/pkg/oss/qiniu"
	"os"
)

func AssetsUpload() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "upload file to OSS",
		//Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := viper.GetString("d")
			b, err := hollow.NewHollow(hollow.Option{
				SourceFs: osfs.New(dir),
				//ThemeFs: nil,
			})
			if err != nil {
				return err
			}

			conf, err := b.LoadConfig(hollow.NewRenderContext())
			if err != nil {
				return err
			}
			q := qiniu.NewQiniu(conf.Hollow.Oss.AccessKey, conf.Hollow.Oss.SecretKey)
			err = q.Uploader().UploadFs(log.Logger(), conf.Hollow.Oss.Bucket, "img", os.DirFS("./fe"))
			if err != nil {
				return err
			}

			//c := viper.GetString("config")
			s := viper.AllSettings()
			//log.Infof("config :%+v", c)
			log.Infof("123 %+v %v", args, s)
			return nil
		},
	}
	return cmd
}

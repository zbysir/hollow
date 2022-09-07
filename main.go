package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zbysir/blog/internal/cmd"
	"github.com/zbysir/blog/internal/pkg/log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "hollow",
	Short: "hollow",
	Long:  `hollow`,
}

func init() {
	rootCmd.PersistentFlags().StringP("d", "d", ".", "workspace")
	viper.BindPFlag("d", rootCmd.PersistentFlags().Lookup("d"))
}

func init() {
	var upload = &cobra.Command{
		Use:   "upload file to hollow service",
		Short: "u",
		Long:  `upload`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	rootCmd.AddCommand(upload)

	// 资源管理
	var assets = &cobra.Command{
		Use:   "assets",
		Short: "a",
		Long:  `assets`,
	}
	{

		assets.AddCommand(cmd.AssetsUpload)
	}
	rootCmd.AddCommand(assets)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Errorf("%+v", err)
		os.Exit(1)
	}
}

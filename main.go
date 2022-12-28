package main

import (
	"github.com/spf13/cobra"
	"github.com/zbysir/hollow/internal/cmd"
	"github.com/zbysir/hollow/internal/pkg/log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "hollow",
	Short: "hollow",
	Long:  `hollow`,
}

func init() {
	//rootCmd.PersistentFlags().StringP("d", "d", ".", "workspace")
	//viper.BindPFlag("d", rootCmd.PersistentFlags().Lookup("d"))
}

func init() {
	rootCmd.AddCommand(cmd.Editor)
	rootCmd.AddCommand(cmd.Server)
	rootCmd.AddCommand(cmd.Build)
	rootCmd.AddCommand(cmd.Version("v0.2.1"))
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Errorf("%+v", err)
		os.Exit(1)
	}
}

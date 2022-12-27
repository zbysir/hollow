package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Version = &cobra.Command{
	Use:   "version",
	Short: "print version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s\n", "v0.0.5")
		return nil
	},
}

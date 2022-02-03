package cmd

import (
	"github.com/lateralusd/bloggy/config"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.NewConfig("./cfg.yaml")
		if err := cfg.Generate(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(genCmd)
}

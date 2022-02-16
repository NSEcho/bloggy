package cmd

import (
	"github.com/lateralusd/bloggy/config"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate",
	RunE: func(cmd *cobra.Command, args []string) error {
		filename, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		cfg := config.NewConfig(filename)
		if err := cfg.Generate(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(genCmd)
	genCmd.Flags().StringP("config", "c", "cfg.yaml", "config filename")
}

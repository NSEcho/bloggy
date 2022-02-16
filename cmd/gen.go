package cmd

import (
	"fmt"

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
		count, err := cfg.Generate()
		if err != nil {
			return err
		}
		fmt.Printf("Generated %d posts\n", count)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(genCmd)
	genCmd.Flags().StringP("config", "c", "cfg.yaml", "config filename")
}

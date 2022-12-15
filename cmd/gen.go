package cmd

import (
	"fmt"

	"github.com/lateralusd/bloggy/config"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate site",
	RunE: func(cmd *cobra.Command, args []string) error {
		filename, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		cfg := config.NewConfig(filename, embedded)
		posts, pages, err := cfg.Generate()
		if err != nil {
			return err
		}
		fmt.Printf("Generated %d posts and %d pages at %q\n", posts, pages,
			cfg.OutDir())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().StringP("config", "c", "cfg.yaml", "config filename")
}

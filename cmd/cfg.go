package cmd

import (
	"fmt"

	"github.com/nsecho/bloggy/config"
	"github.com/spf13/cobra"
)

var cfgCmd = &cobra.Command{
	Use:   "cfg",
	Short: "Create sample config",
	RunE: func(cmd *cobra.Command, args []string) error {
		filename, err := cmd.Flags().GetString("out")
		if err != nil {
			return err
		}
		if err := config.SaveConfig(filename); err != nil {
			return err
		}

		fmt.Printf("New config \"%s\" created\n", filename)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cfgCmd)
	cfgCmd.Flags().StringP("out", "o", "cfg.yaml", "config filename")
}

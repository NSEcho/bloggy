package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "bloggy",
	Short: "small static site generator",
}

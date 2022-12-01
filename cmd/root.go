package cmd

import (
	"embed"
	"github.com/spf13/cobra"
)

var embedded embed.FS

var rootCmd = &cobra.Command{
	Use:   "bloggy",
	Short: "small static site generator",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute(content embed.FS) error {
	embedded = content
	return rootCmd.Execute()
}

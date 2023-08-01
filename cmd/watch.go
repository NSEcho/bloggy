package cmd

import (
	"fmt"
	"github.com/nsecho/bloggy/config"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for changes and generate on change",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgFilename, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		drafts, err := cmd.Flags().GetBool("drafts")
		if err != nil {
			return err
		}

		watchedDirs := []string{
			"custom",
			"images",
			"pages",
			"posts",
		}

		watchedFiles := make(map[string]int64)
		fs, err := os.Stat("cfg.yaml")
		if err != nil {
			return err
		}
		watchedFiles["cfg.yaml"] = fs.ModTime().Unix()

		for _, dir := range watchedDirs {
			files, err := os.ReadDir(dir)
			if err != nil {
				return err
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}

				filePath := filepath.Join(dir, file.Name())
				fs, err := os.Stat(filePath)
				if err != nil {
					return err
				}
				watchedFiles[filePath] = fs.ModTime().Unix()
			}
		}

		fmt.Printf("[%s] Started monitoring %d files\n",
			time.Now().Format("2006-01-02 15:04:05"), len(watchedFiles))

		ticker := time.NewTicker(5 * time.Second)

		for now := range ticker.C {
			if name, changed := fileStats(watchedFiles); changed {
				fmt.Printf("[%s] File %s is changed. Recompiling!\n", now.Format("2006-01-02 15:04:05"), name)
				cfg := config.NewConfig(cfgFilename, embedded)
				if _, _, err := cfg.Generate(drafts); err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func init() {
	watchCmd.Flags().BoolP("drafts", "d", true, "generate with drafts")
	watchCmd.Flags().StringP("config", "c", "cfg.yaml", "config filename")
	rootCmd.AddCommand(watchCmd)
}

func fileStats(stats map[string]int64) (string, bool) {
	m := make(map[string]int64, len(stats))

	for k, v := range stats {
		m[k] = v
	}

	name := ""
	changed := false
	for file, change := range m {
		fs, err := os.Stat(file)
		if err != nil {
			return name, false
		}
		if fs.ModTime().Unix() > change {
			m[file] = fs.ModTime().Unix()
			name = file
			changed = true
		}
	}

	for k, v := range m {
		stats[k] = v
	}

	return name, changed
}

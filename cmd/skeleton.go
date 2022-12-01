package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/lateralusd/bloggy/config"
	"github.com/lateralusd/bloggy/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var skeletonCmd = &cobra.Command{
	Use:   "skeleton [outputDirectory]",
	Short: "Create skeleton directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("missing directory argument")
		}

		dirName := args[0]
		if _, err := os.Stat(dirName); os.IsNotExist(err) {
			if err := os.Mkdir(dirName, os.ModePerm); err != nil {
				return err
			}
		}

		var dirs = []string{
			"custom",
			"posts",
			"pages",
		}

		for _, dir := range dirs {
			outDir := filepath.Join(dirName, dir)
			if err := os.Mkdir(outDir, os.ModePerm); err != nil {
				return err
			}
		}

		if err := config.SaveConfig(filepath.Join(dirName, "cfg.yaml")); err != nil {
			return err
		}

		postName := "example.md"
		postPath := filepath.Join(dirName, "posts", postName)

		p := models.PostMetadata{
			Title:       "Test post",
			Description: "This is short description",
			Date:        time.Now(),
			WithToC:     false,
			References: []string{
				"https://www.google.com",
				"https://www.facebook.com",
			},
		}

		f, err := os.Create(postPath)
		if err != nil {
			return err
		}
		defer f.Close()

		buf := new(bytes.Buffer)

		if err := yaml.NewEncoder(buf).Encode(&p); err != nil {
			return err
		}

		out := fmt.Sprintf(output, buf.String(), content)
		io.Copy(f, strings.NewReader(out))

		pageName := "example_page.md"
		pagePath := filepath.Join(dirName, "pages", pageName)

		page := models.PageMetadata{
			Title:    "Test page",
			Subtitle: "This is subtitle",
		}

		pageF, err := os.Create(pagePath)
		if err != nil {
			return err
		}
		defer f.Close()

		pageBuf := new(bytes.Buffer)

		if err := yaml.NewEncoder(pageBuf).Encode(&page); err != nil {
			return err
		}

		out = fmt.Sprintf(output, buf.String(), body)
		io.Copy(pageF, strings.NewReader(out))

		ccss := filepath.Join(dirName, "custom", "custom.css")
		cssF, err := os.Create(ccss)
		if err != nil {
			return err
		}
		cssF.Close()

		fmt.Printf("Created bloggy skeleton at \"%s\"\n", dirName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(skeletonCmd)
}

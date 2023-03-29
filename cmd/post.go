package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nsecho/bloggy/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var content = `# Introduction

Here comes the content.

## How to

Just write markdown and when you want to reference the image, place the image inside the static/images and reference it in url ../images/nameoftheimage.png.

![Image](../images/sample.png)
`

var output = `---
%s---

%s
`

var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Create new post",
	RunE: func(cmd *cobra.Command, args []string) error {
		joined := strings.Join(args, "_")
		name := fmt.Sprintf("%s.md", joined)
		filename := filepath.Join("./posts", name)

		p := models.PostMetadata{
			Title:       "Test post",
			Description: "This is short description",
			Date:        time.Now(),
			WithToC:     false,
			Tags: []string{
				"reverse-engineering",
				"frida",
				"lldb",
			},
			References: []string{
				"https://www.google.com",
				"https://www.facebook.com",
			},
			Draft: true,
		}

		f, err := os.Create(filename)
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

		fmt.Printf("New post %s created\n", filename)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(postCmd)
}

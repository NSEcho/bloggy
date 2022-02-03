package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lateralusd/bloggy/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var content = `# Introduction

Here comes the content.`

var output = `---
%s---

%s
`

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new post",
	RunE: func(cmd *cobra.Command, args []string) error {
		joined := strings.Join(args, "_")
		name := fmt.Sprintf("%s.md", joined)
		filename := filepath.Join("./posts", name)

		p := models.PostMetadata{
			Title:       "Test post",
			Description: "This is short description",
			Date:        time.Now(),
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
	RootCmd.AddCommand(newCmd)
}

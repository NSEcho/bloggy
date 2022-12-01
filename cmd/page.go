package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lateralusd/bloggy/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var body = `# Contact ways
You can contact me here:  
* foo
* bar
`

var wholeFile = `---
%s---

%s
`

var pageCmd = &cobra.Command{
	Use:   "page",
	Short: "Create new page",
	RunE: func(cmd *cobra.Command, args []string) error {
		joined := strings.Join(args, "_")
		name := fmt.Sprintf("%s.md", joined)
		filename := filepath.Join("./pages", name)

		p := models.PageMetadata{
			Title:    "Test page",
			Subtitle: "This is subtitle",
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

		out := fmt.Sprintf(output, buf.String(), body)
		io.Copy(f, strings.NewReader(out))

		fmt.Printf("New post %s created\n", filename)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pageCmd)
}

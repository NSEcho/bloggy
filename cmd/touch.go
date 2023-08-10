package cmd

import (
	"bytes"
	"errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var touchCmd = &cobra.Command{
	Use:   "touch post_file.md",
	Short: "Update timestamp of the post",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("missing post file")
		}

		oldContent, err := readContent(args[0])
		if err != nil {
			return err
		}

		timeNow := func() []byte {
			buf := new(bytes.Buffer)
			yaml.NewEncoder(buf).Encode(struct {
				Date time.Time
			}{
				Date: time.Now(),
			})
			return bytes.TrimSpace(buf.Bytes())
		}

		re := regexp.MustCompile(`date:.*`)
		newContent := re.ReplaceAll(oldContent, timeNow())

		f, err := os.OpenFile(filepath.Join("./posts", args[0]),
			os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.Write(newContent); err != nil {
			return err
		}
		
		return nil
	},
}

func readContent(postPath string) ([]byte, error) {
	f, err := os.Open(filepath.Join("./posts", postPath))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

func init() {
	rootCmd.AddCommand(touchCmd)
}

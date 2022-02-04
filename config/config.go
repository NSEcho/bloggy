package config

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/lateralusd/bloggy/models"
	"gopkg.in/yaml.v3"
)

type data struct {
	BlogTitle   string `yaml:"title"`
	CurrentYear string
	TwitterLink string `yaml:"twitter"`
	GithubLink  string `yaml:"github"`
	Mail        string `yaml:"mail"`
	Author      string `yaml:"author"`
	About       string `yaml:"about"`
	Outdir      string `yaml:"outdir"`
	AboutMD     template.HTML
	Posts       []models.Post
}

type Config struct {
	cfgPath string
	outDir  string
}

func NewConfig(cfgPath string) *Config {
	return &Config{
		cfgPath: cfgPath,
	}
}

func (c *Config) Generate() error {
	var data data
	f, err := os.Open(c.cfgPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&data); err != nil {
		return err
	}

	c.outDir = data.Outdir

	md := markdown.ToHTML([]byte(data.About), nil, nil)
	data.AboutMD = template.HTML(string(md))

	data.CurrentYear = time.Now().Format("2006")

	files, err := ioutil.ReadDir("./posts")
	if err != nil {
		return err
	}

	for _, file := range files {
		postPath := filepath.Join("./posts", file.Name())
		post, err := postFromFile(postPath)
		if err != nil {
			return err
		}
		post.Name = getOutName(file.Name())
		data.Posts = append(data.Posts, *post)
	}

	sort.Slice(data.Posts, func(i, j int) bool {
		return data.Posts[i].Date.After(data.Posts[j].Date)
	})

	if err := copyDirs("./static", c.outDir); err != nil {
		return err
	}

	t, err := template.New("").Funcs(template.FuncMap{
		"printDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"checkField": func(name string, data interface{}) bool {
			v := reflect.ValueOf(data)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Kind() != reflect.Struct {
				return false
			}
			return v.FieldByName(name).IsValid()
		},
	}).ParseFiles(getTplFiles()...)
	if err != nil {
		return err
	}

	basicTpls := map[string]string{
		"index": "index.html",
		"about": "about.html",
	}

	for tname, out := range basicTpls {
		fpath := filepath.Join(c.outDir, out)
		f, err := os.Create(fpath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := t.ExecuteTemplate(f, tname, &data); err != nil {
			return err
		}
	}

	for _, post := range data.Posts {
		post.Author = data.Author
		if err := c.postToHTML(&data, &post, t); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) postToHTML(dt *data, post *models.Post, t *template.Template) error {
	if err := createIfNotExists(c.outDir+"/posts/", 0755); err != nil {
		return err
	}

	outName := c.outDir + "/posts/" + getOutName(post.Name)
	f, err := os.Create(outName)
	if err != nil {
		return err
	}
	defer f.Close()

	d := struct {
		Data *data
		Post *models.Post
	}{
		Data: dt,
		Post: post,
	}

	return t.ExecuteTemplate(f, "post", d)
}

func postFromFile(filepath string) (*models.Post, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, f)

	splitted := strings.Split(buf.String(), "\n")

	ct := 0
	idx := 0
	for id, line := range splitted {
		if strings.Contains(line, "---") {
			ct++
		}
		if ct == 2 {
			idx = id
			break
		}
	}

	cfg := strings.NewReader(strings.Join(splitted[1:idx], "\n"))

	var post models.Post
	var p models.PostMetadata
	if err := yaml.NewDecoder(cfg).Decode(&p); err != nil {
		return nil, err
	}

	post.Content = strings.Join(splitted[idx+2:], "\n")
	md := markdown.ToHTML([]byte(post.Content), nil, nil)
	post.ContentMD = template.HTML(string(md))
	post.PostMetadata = p
	return &post, nil
}

func getTplFiles() []string {
	files, _ := filepath.Glob("./templates/*.gohtml")
	return files
}

func getOutName(filename string) string {
	splitted := strings.Split(filename, ".")
	base := splitted[0]
	return base + ".html"
}

func copyDirs(sourceDir, destDir string) error {
	entries, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := copyDirs(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := copy(sourcePath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

func createIfNotExists(dir string, perm os.FileMode) error {
	if exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return err
	}

	return nil
}

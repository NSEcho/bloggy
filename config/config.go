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
	"github.com/gorilla/feeds"
	"github.com/lateralusd/bloggy/models"
	"gopkg.in/yaml.v3"
)

type outcfg struct {
	URL         string `yaml:"url"`
	BlogTitle   string `yaml:"title"`
	TwitterLink string `yaml:"twitter"`
	GithubLink  string `yaml:"github"`
	Mail        string `yaml:"mail"`
	Author      string `yaml:"author"`
	About       string `yaml:"about"`
	Outdir      string `yaml:"outdir"`
}

type data struct {
	URL         string `yaml:"url"`
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
	Pages       []models.Page
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

func SaveConfig(filename string) error {
	cfg := outcfg{
		URL:         "https://username.github.io/",
		BlogTitle:   "sample blog",
		TwitterLink: "https://twitter.com/user",
		GithubLink:  "https://github.com/user",
		Mail:        "someone@something.com",
		Author:      "Haxor",
		About:       "About page section",
		Outdir:      "public",
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewEncoder(f).Encode(&cfg)
}

func (c *Config) Generate() (int, int, error) {
	var data data
	f, err := os.Open(c.cfgPath)
	if err != nil {
		return -1, -1, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&data); err != nil {
		return -1, -1, err
	}

	c.outDir = data.Outdir

	md := markdown.ToHTML([]byte(data.About), nil, nil)
	data.AboutMD = template.HTML(string(md))

	data.CurrentYear = time.Now().Format("2006")

	posts, err := ioutil.ReadDir("./posts")
	if err != nil {
		return -1, -1, err
	}

	for _, file := range posts {
		postPath := filepath.Join("./posts", file.Name())
		post, err := postFromFile(postPath)
		if err != nil {
			return -1, -1, err
		}
		post.Name = getOutName(file.Name())
		data.Posts = append(data.Posts, *post)
	}

	pages, err := ioutil.ReadDir("./pages")
	if err != nil {
		return -1, -1, err
	}

	for _, file := range pages {
		pagePath := filepath.Join("./pages", file.Name())
		page, err := pageFromFile(pagePath)
		if err != nil {
			return -1, -1, err
		}
		page.Name = getOutName(file.Name())
		data.Pages = append(data.Pages, *page)
	}

	sort.Slice(data.Posts, func(i, j int) bool {
		return data.Posts[i].Date.After(data.Posts[j].Date)
	})

	if err := copyDirs("./static", c.outDir); err != nil {
		return -1, -1, err
	}

	tplFiles, err := getGlobFiles("./templates/*.gohtml")
	if err != nil {
		return -1, -1, err
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
	}).ParseFiles(tplFiles...)
	if err != nil {
		return -1, -1, err
	}

	basicTpls := map[string]string{
		"index": "index.html",
		"about": "about.html",
	}

	for tname, out := range basicTpls {
		fpath := filepath.Join(c.outDir, out)
		f, err := os.Create(fpath)
		if err != nil {
			return -1, -1, err
		}
		defer f.Close()
		if err := t.ExecuteTemplate(f, tname, &data); err != nil {
			return -1, -1, err
		}
	}

	for _, post := range data.Posts {
		post.Author = data.Author
		if err := c.postToHTML(&data, &post, t); err != nil {
			return -1, -1, err
		}
	}

	for _, page := range data.Pages {
		if err := c.pageToHTML(&data, &page, t); err != nil {
			return -1, -1, err
		}
	}

	if data.URL != "" {
		if err := c.generateRSS(&data); err != nil {
			return -1, -1, err
		}
	}

	return len(data.Posts), len(data.Pages), nil
}

func getGlobFiles(dirPath string) ([]string, error) {
	return filepath.Glob(dirPath)
}

func (c *Config) generateRSS(dt *data) error {
	feed := &feeds.Feed{
		Title:       dt.BlogTitle,
		Link:        &feeds.Link{Href: dt.URL},
		Description: "custom blog",
		Author:      &feeds.Author{Name: dt.Author, Email: dt.Mail},
		Created:     time.Now(),
	}

	for _, post := range dt.Posts {
		item := feeds.Item{
			Title: post.Title,
			Link: &feeds.Link{
				Href: dt.URL + post.Name,
			},
			Description: post.Description,
			Author:      &feeds.Author{Name: dt.Author, Email: dt.Mail},
			Created:     post.Date,
		}
		feed.Items = append(feed.Items, &item)
	}

	rss, err := feed.ToRss()
	if err != nil {
		return err
	}

	outPath := filepath.Join(c.outDir, "index.xml")
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, strings.NewReader(rss))
	return err
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

func (c *Config) pageToHTML(dt *data, page *models.Page, t *template.Template) error {
	if err := createIfNotExists(c.outDir+"/pages/", 0755); err != nil {
		return err
	}

	outName := c.outDir + "/pages/" + getOutName(page.Name)
	f, err := os.Create(outName)
	if err != nil {
		return err
	}
	defer f.Close()

	d := struct {
		Data *data
		Page *models.Page
	}{
		Data: dt,
		Page: page,
	}

	return t.ExecuteTemplate(f, "page", d)
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

func pageFromFile(filepath string) (*models.Page, error) {
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

	var page models.Page
	var p models.PageMetadata
	if err := yaml.NewDecoder(cfg).Decode(&p); err != nil {
		return nil, err
	}

	content := strings.Join(splitted[idx+2:], "\n")
	md := markdown.ToHTML([]byte(content), nil, nil)
	page.PageContent = template.HTML(string(md))
	page.PageMetadata = p
	return &page, nil
}

func getOutName(filename string) string {
	splitted := strings.Split(filename, ".")
	base := strings.ToLower(splitted[0])
	return strings.Replace(base, " ", "_", -1) + ".html"
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

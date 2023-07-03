package config

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/feeds"
	"github.com/nsecho/bloggy/models"
)

var (
	gistRe    = regexp.MustCompile(`<p>gist:<a\shref="(.*?)".*?</p>`)
	headersRe = regexp.MustCompile(`##?\s(.*)`)
	embedRe   = regexp.MustCompile(`embed:(.*):([^\s]+)`)
)

type outcfg struct {
	URL          string `yaml:"url"`
	BlogTitle    string `yaml:"title"`
	TwitterLink  string `yaml:"twitter"`
	GithubLink   string `yaml:"github"`
	Mail         string `yaml:"mail"`
	Author       string `yaml:"author"`
	About        string `yaml:"about"`
	Outdir       string `yaml:"outdir"`
	CNAME        string `yaml:"cname"`
	PostsPerPage int    `yaml:"posts_per_page"`
	WithRSS      bool   `yaml:"rss"`
}

type data struct {
	URL          string `yaml:"url"`
	BlogTitle    string `yaml:"title"`
	CurrentYear  string
	TwitterLink  string `yaml:"twitter"`
	GithubLink   string `yaml:"github"`
	Mail         string `yaml:"mail"`
	Author       string `yaml:"author"`
	About        string `yaml:"about"`
	Outdir       string `yaml:"outdir"`
	DiffBlog     string `yaml:"diffblog"`
	CNAME        string `yaml:"cname"`
	PostsPerPage int    `yaml:"posts_per_page"`
	WithRSS      bool   `yaml:"rss"`
	AboutMD      template.HTML
	HasCustomCSS bool
	Posts        []models.Post
	Pages        []models.Page
	Tags         map[string][]TagData
}

type TagData struct {
	Name string
	Path string
}

type Config struct {
	embedded embed.FS
	cfgPath  string
	outDir   string
}

func NewConfig(cfgPath string, embedded embed.FS) *Config {
	return &Config{
		embedded: embedded,
		cfgPath:  cfgPath,
	}
}

func (c *Config) OutDir() string {
	return c.outDir
}

func SaveConfig(filename string) error {
	cfg := outcfg{
		URL:          "https://username.github.io/",
		BlogTitle:    "sample blog",
		TwitterLink:  "https://twitter.com/user",
		GithubLink:   "https://github.com/user",
		Mail:         "someone@something.com",
		Author:       "Haxor",
		About:        "About page section",
		Outdir:       "public",
		CNAME:        "www.example.com",
		PostsPerPage: 5,
		WithRSS:      true,
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewEncoder(f).Encode(&cfg)
}

func writeBasicTemplate(buf *bytes.Buffer, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, buf)
	return err
}

func (c *Config) Generate(genDrafts bool) (int, int, error) {
	// Read and parse config file
	var dt data
	dt.Tags = make(map[string][]TagData)
	f, err := os.Open(c.cfgPath)
	if err != nil {
		return -1, -1, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&dt); err != nil {
		return -1, -1, err
	}

	c.outDir = dt.Outdir

	md := markdown.ToHTML([]byte(dt.About), nil, nil)
	dt.AboutMD = template.HTML(string(md))

	dt.CurrentYear = time.Now().Format("2006")

	posts, err := filepath.Glob("./posts/" + "*.md")
	if err != nil {
		return -1, -1, err
	}

	for _, file := range posts {
		name := strings.Split(file, "/")[1]
		post, err := postFromFile(file)
		if err != nil {
			return -1, -1, err
		}
		if post.Draft == true {
			if genDrafts {
				post.Name = getOutName(name)
				dt.Posts = append(dt.Posts, *post)

				for _, tag := range post.Tags {
					dt.Tags[tag] = append(dt.Tags[tag], TagData{
						Name: post.Title,
						Path: post.Name,
					})
				}
			}
		} else {
			post.Name = getOutName(name)
			dt.Posts = append(dt.Posts, *post)

			for _, tag := range post.Tags {
				dt.Tags[tag] = append(dt.Tags[tag], TagData{
					Name: post.Title,
					Path: post.Name,
				})
			}
		}
	}

	pages, err := os.ReadDir("./pages")
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
		dt.Pages = append(dt.Pages, *page)
	}

	sort.Slice(dt.Posts, func(i, j int) bool {
		return dt.Posts[i].Date.After(dt.Posts[j].Date)
	})

	customCssPath := filepath.Join("./custom", "custom.css")
	if exists(customCssPath) {
		dt.HasCustomCSS = true
	}

	if err := c.copyDirs("static", c.outDir); err != nil {
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
		"inc": func(value int) int {
			return value + 1
		},
		"dec": func(value int) int {
			return value - 1
		},
	}).ParseFS(c.embedded, "templates/*")
	if err != nil {
		return -1, -1, err
	}

	basicTpls := map[string]string{
		"index": "index.html",
		"about": "about.html",
		"tags":  "tags.html",
	}

	perPage := 5
	if dt.PostsPerPage != 0 {
		perPage = dt.PostsPerPage
	}

	pagePosts := make(map[int][]models.Post)
	ctr := 1

	for i, post := range dt.Posts {
		if i%perPage == 0 && i != 0 {
			ctr++
		}
		pagePosts[ctr] = append(pagePosts[ctr], post)
	}

	originalPosts := make([]models.Post, len(dt.Posts))
	copy(originalPosts, dt.Posts)

	for k, v := range pagePosts {
		buf := new(bytes.Buffer)
		dt.Posts = make([]models.Post, len(v))
		copy(dt.Posts, v)
		tplData := struct {
			Data        data
			CurrentPage int
			TotalPages  int
		}{
			Data:        dt,
			CurrentPage: k,
			TotalPages:  len(pagePosts),
		}
		if k == 1 {
			for tpname, out := range basicTpls {
				outPath := filepath.Join(c.outDir, out)

				if tpname == "index" {
					if err := t.ExecuteTemplate(buf, tpname, tplData); err != nil {
						return -1, -1, err
					}
				} else {
					if err := t.ExecuteTemplate(buf, tpname, &dt); err != nil {
						return -1, -1, err
					}
				}

				if err := writeBasicTemplate(buf, outPath); err != nil {
					return -1, -1, err
				}
			}
		} else {
			dirName := filepath.Join(c.outDir, "pgs", strconv.Itoa(k))
			os.MkdirAll(dirName, os.ModePerm)
			outPath := filepath.Join(dirName, "index.html")

			if err := t.ExecuteTemplate(buf, "pgs", &tplData); err != nil {
				return -1, -1, err
			}

			if err := writeBasicTemplate(buf, outPath); err != nil {
				return -1, -1, err
			}
		}
	}

	dt.Posts = make([]models.Post, len(originalPosts))

	copy(dt.Posts, originalPosts)

	for _, post := range dt.Posts {
		post.Author = dt.Author
		if len(post.References) > 0 {
			post.RealRefs = make(map[string]string, len(post.References))
			for _, ref := range post.References {
				splitted := strings.Split(ref, " => ")
				if len(splitted) == 1 {
					post.RealRefs[ref] = ref
				} else {
					post.RealRefs[splitted[0]] = splitted[1]
				}
			}
		}

		if err := c.postToHTML(&dt, &post, t); err != nil {
			return -1, -1, err
		}
	}

	for _, page := range dt.Pages {
		if err := c.pageToHTML(&dt, &page, t); err != nil {
			return -1, -1, err
		}
	}

	for name, tag := range dt.Tags {
		if err := c.tagToHTML(&dt, name, tag, t); err != nil {
			return -1, -1, err
		}
	}

	// copy custom bgs
	if exists("./custom") {
		copyCustom(&dt)
	}

	if exists("./images") {
		copyImages(&dt)
	}

	if dt.URL != "" && dt.WithRSS {
		if err := c.generateRSS(&dt); err != nil {
			return -1, -1, err
		}
	}

	if dt.CNAME != "" {
		cnameFile, err := os.Create(filepath.Join(dt.Outdir, "CNAME"))
		if err != nil {
			return -1, -1, err
		}
		defer cnameFile.Close()

		cnameFile.WriteString(dt.CNAME)
	}

	return len(dt.Posts), len(dt.Pages), nil
}

func copyImages(dt *data) error {
	images, err := os.ReadDir("./images")
	if err != nil {
		return err
	}

	outDir := filepath.Join(dt.Outdir, "images")
	if !exists(filepath.Join(outDir)) {
		if err := os.Mkdir(outDir, os.ModePerm); err != nil {
			return err
		}
	}

	for _, image := range images {
		imgPath := filepath.Join("./images", image.Name())
		err := copySimpleFile(outDir, imgPath, image.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

func copyCustom(dt *data) error {
	var bgs = []string{
		"home-bg.jpg",
		"about-bg.jpg",
		"post-bg.jpg",
	}

	for _, bg := range bgs {
		fp := filepath.Join("./custom", bg)
		if exists(fp) {
			out := filepath.Join(dt.Outdir, "assets", "img")
			err := copySimpleFile(out, fp, bg)
			if err != nil {
				return err
			}
		}
	}

	if dt.HasCustomCSS {
		out := filepath.Join(dt.Outdir, "css")
		err := copySimpleFile(out, filepath.Join("./custom", "custom.css"), "custom.css")
		if err != nil {
			return err
		}
	}

	return nil
}

func copySimpleFile(outDir, fullPath, name string) error {
	in, err := os.Open(fullPath)
	if err != nil {
		return nil
	}

	outf := filepath.Join(outDir, name)
	out, err := os.Create(outf)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	return err
}

func (c *Config) getTemplateFiles() ([]string, error) {
	filesFS, err := c.embedded.ReadDir("templates")
	if err != nil {
		return nil, err
	}
	files := make([]string, len(filesFS))
	for i, f := range filesFS {
		files[i] = f.Name()
	}
	return files, nil
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
				Href: dt.URL + "posts/" + post.Name,
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

func (c *Config) tagToHTML(dt *data, name string, tags []TagData, t *template.Template) error {
	if err := createIfNotExists(c.outDir+"/tags/", 0755); err != nil {
		return err
	}

	outName := c.outDir + "/tags/" + getOutName(name)
	f, err := os.Create(outName)
	if err != nil {
		return err
	}
	defer f.Close()

	d := struct {
		Data *data
		Name string
		Tags []TagData
	}{
		Data: dt,
		Name: name,
		Tags: tags,
	}

	return t.ExecuteTemplate(f, "tagpage", d)
}

func parsePostImages(content string) (string, error) {
	newContent := content
	imageRe := regexp.MustCompile(`<img src="(.*?)"`)
	matches := imageRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return "", nil
	}
	for _, match := range matches {
		imagePath := match[1]
		stripped := strings.TrimLeft(imagePath, "../")
		mime, rawBase, err := func(img string) (string, string, error) {
			f, err := os.Open(img)
			if err != nil {
				return "", "", err
			}
			defer f.Close()

			imageData, err := io.ReadAll(f)
			if err != nil {
				return "", "", err
			}

			encoded := base64.StdEncoding.EncodeToString(imageData)
			mimeType := http.DetectContentType(imageData)

			return mimeType, encoded, nil
		}(stripped)
		if err != nil {
			return "", err
		}
		// 	replacedWithGists := gistRe.ReplaceAllString(string(md), `<script src="$1"></script>`)
		newImageSrc := fmt.Sprintf("<img src=\"data:%s;base64,%s\"", mime, rawBase)
		newContent = strings.Replace(newContent, match[0], newImageSrc, -1)
	}
	return newContent, nil
}

func postFromFile(filepath string) (*models.Post, error) {
	f, err := os.OpenFile(filepath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, f)

	rawContent := buf.String()
	matches := embedRe.FindAllStringSubmatch(rawContent, -1)
	if len(matches) > 0 {
		newContent := parseEmbeddedFiles(rawContent)
		f.Truncate(0)
		f.Write([]byte(newContent))
		return postFromFile(filepath)
	}

	splitted := strings.Split(rawContent, "\n")

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

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	content := strings.Join(splitted[idx+2:], "\n")
	if p.WithToC {
		hasRef := false
		if len(p.References) > 0 {
			hasRef = true
		}
		content = prependToC(p.Title, content, hasRef)
	}

	md := markdown.ToHTML([]byte(content), parser, nil)
	replacedWithGists := gistRe.ReplaceAllString(string(md), `<script src="$1"></script>`)
	newContent, err := parsePostImages(replacedWithGists)
	if err != nil {
		return nil, err
	}
	post.ContentMD = template.HTML(newContent)
	post.PostMetadata = p
	return &post, nil
}

// prependToC generates table of contents markdown
func prependToC(title, oldContent string, hasReferences bool) string {
	matches := headersRe.FindAllStringSubmatch(oldContent, -1)
	var withToCContent = ""
	if len(matches) > 0 {
		withToCContent += "# Table of Contents\n"
		for _, match := range matches {
			// remove whitespace
			ln := strings.Replace(match[1], " ", "-", -1)
			// convert to lower
			ln = strings.ToLower(ln)
			withToCContent += fmt.Sprintf("* [%s](#%s)\n", match[1], ln)
		}
	}

	if hasReferences {
		withToCContent += fmt.Sprintf("* [References](#references)\n")
	}

	return withToCContent + "\n" + oldContent
}

func parseEmbeddedFiles(postContent string) string {
	parsed := embedRe.ReplaceAllFunc([]byte(postContent), func(bt []byte) []byte {
		splitted := strings.Split(string(bt), ":")
		if len(splitted) != 3 {
			return bt
		}

		buf := new(bytes.Buffer)
		hdr := []byte{'`', '`', '`'}
		ftr := hdr

		hdr = append(hdr, []byte(splitted[2])...)
		hdr = append(hdr, '\n')
		f, err := os.Open(splitted[1])
		if err != nil {
			return bt
		}
		defer f.Close()

		buf.Write(hdr)
		io.Copy(buf, f)
		buf.Write([]byte{'\n'})
		buf.Write(ftr)

		return buf.Bytes()
	})
	return string(parsed)
}

// pageFromFile generates *models.Page from raw markdown file
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

func (c *Config) copyDirs(sourceDir, destDir string) error {
	entries, err := c.embedded.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		switch entry.Type() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := c.copyDirs(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := c.copyFile(sourcePath, destPath); err != nil {
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

func (c *Config) copyFile(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := c.embedded.ReadFile(srcFile)
	if err != nil {
		return err
	}

	return os.WriteFile(dstFile, in, os.ModePerm)
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

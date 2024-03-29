package models

import (
	"html/template"
	"time"
)

type References []string
type Tags []string

type PostMetadata struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Date        time.Time `yaml:"date"`
	WithToC     bool      `yaml:"toc"`
	Tags        `yaml:"tags"`
	References  `yaml:"refs"`
	Draft       bool `yaml:"draft"`
}

type Post struct {
	PostMetadata
	Name      string
	Content   string
	ContentMD template.HTML
	Author    string
	RealRefs  map[string]string
}

type PageMetadata struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"description"`
}

type Page struct {
	PageMetadata
	Name        string
	PageContent template.HTML
}

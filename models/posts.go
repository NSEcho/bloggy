package models

import (
	"html/template"
	"time"
)

type References []string

type PostMetadata struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Date        time.Time `yaml:"date"`
	WithToC     bool      `yaml:"toc"`
	References  `yaml:"refs"`
}

type Post struct {
	PostMetadata
	Name      string
	Content   string
	ContentMD template.HTML
	Author    string
}

type PageMetadata struct {
	Title      string `yaml:"title"`
	Subtitle   string `yaml:"description"`
	References `yaml:"refs"`
}

type Page struct {
	PageMetadata
	Name        string
	PageContent template.HTML
}

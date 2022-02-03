package models

import (
	"html/template"
	"time"
)

type PostMetadata struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Date        time.Time `yaml:"date"`
}

type Post struct {
	PostMetadata
	Name      string
	Content   string
	ContentMD template.HTML
	Author    string
}

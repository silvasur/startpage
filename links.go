package main

import (
	"html/template"
)

type Link struct {
	Title string
	URL   template.URL
}

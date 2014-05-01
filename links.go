package main

import (
	"errors"
	"html/template"
)

type Link struct {
	Title string
	URL   template.URL
}

var links = []Link{}

func addLinkCmd(params []string) error {
	if len(params) != 2 {
		return errors.New("add-link needs 2 parameters: title url")
	}

	links = append(links, Link{params[0], template.URL(params[1])})
	return nil
}

package main

import (
	"bufio"
	"html/template"
	"log"
	"os"
	"strings"
)

type Link struct {
	Title string
	URL   template.URL
}

func GetLinks() (links []Link) {
	fh, err := os.Open(os.ExpandEnv("$HOME/.startpage-urls"))
	if err != nil {
		log.Printf("Couldn't read links: %s", err)
		return
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "->", 2)
		links = append(links, Link{
			strings.TrimSpace(parts[0]),
			template.URL(strings.TrimSpace(parts[1])),
		})
	}

	return
}

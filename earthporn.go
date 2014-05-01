package main

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"strings"
)

type redditList struct {
	Data struct {
		Children []struct {
			Data *EarthPorn `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type EarthPorn struct {
	Title     string `json:"title"`
	URL       string `json:"url,omitempty"`
	Permalink string `json:"permalink"`
	Domain    string `json:"domain"`
}

const earthPornURL = "http://www.reddit.com/r/EarthPorn.json"

func GetEarthPorn() (EarthPorn, error) {
	resp, err := http.Get(earthPornURL)
	if err != nil {
		return EarthPorn{}, err
	}
	defer resp.Body.Close()

	var list redditList
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&list); err != nil {
		return EarthPorn{}, err
	}

	for _, el := range list.Data.Children {
		p := el.Data

		if p.Domain == "self.EarthPorn" {
			continue
		}

		if p.URL == "" {
			continue
		}

		if (&p).getImageURL() {
			return p, nil
		}
	}

	return EarthPorn{}, errors.New("Could not get EarthPorn: No image could be extracted")
}

func (p *EarthPorn) getImageURL() bool {
	// TODO: We can do further processing here (e.g. if we get a link to flickr, extract the image).
	// For now, we will simply test, if the URL points to an image.

	resp, err := http.Head(p.URL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	t := resp.Header.Get("Content-Type")
	if t == "" {
		return false
	}
	mt, _, err := mime.ParseMediaType(t)
	if err != nil {
		return false
	}

	return (strings.Split(mt, "/")[0] == "image")
}

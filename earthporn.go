package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
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
	Title          string `json:"title"`
	URL            string `json:"url,omitempty"`
	Permalink      string `json:"permalink"`
	Domain         string `json:"domain"`
	Saved          bool
	data, origdata []byte
	mediatype      string
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

		if p.fetch() {
			return *p, nil
		}
	}

	return EarthPorn{}, errors.New("Could not get EarthPorn: No image could be extracted")
}

func (p *EarthPorn) fetch() bool {
	// TODO: We can do further processing here (e.g. if we get a link to flickr, extract the image).
	// For now, we will simply test, if the URL points to an image.

	resp, err := http.Get(p.URL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	t := resp.Header.Get("Content-Type")
	if t == "" {
		log.Printf("could not get image of %s: no Content-Type", p.Permalink)
		return false
	}
	mt, _, err := mime.ParseMediaType(t)
	if err != nil {
		log.Printf("could not get image of %s: %s", p.Permalink, err)
		return false
	}

	if strings.Split(mt, "/")[0] != "image" {
		log.Printf("could not get image of %s: not an image", p.Permalink)
		return false
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		log.Printf("could not get image of %s: %s", p.Permalink, err)
		return false
	}

	p.mediatype = t
	p.origdata = buf.Bytes()
	p.data = p.origdata
	p.resize()
	return true
}

var maxdim = 2500

func setMaxdimCmd(params []string) error {
	if len(params) != 1 {
		return errors.New("set-maxdim needs one parameter, which must be a positive number")
	}

	newmaxdim, err := strconv.ParseUint(params[0], 10, 32)
	if err != nil {
		return fmt.Errorf("1st parameter of set-maxdim could not be parsed as a number: %s", err)
	}
	maxdim = int(newmaxdim)

	return nil
}

// resize resizes image, if it's very large
func (p *EarthPorn) resize() {
	im, _, err := image.Decode(bytes.NewReader(p.origdata))
	if err != nil {
		log.Printf("Failed decoding in resize(): %s", err)
		return
	}

	size := im.Bounds().Size()
	if !(size.X > maxdim || size.Y > maxdim) {
		return
	}

	var w, h int
	if size.X > size.Y {
		h = maxdim * (size.Y / size.X)
		w = maxdim
	} else {
		w = maxdim * (size.X / size.Y)
		h = maxdim
	}

	im = resize.Resize(uint(w), uint(h), im, resize.Bicubic)

	buf := new(bytes.Buffer)

	if err := jpeg.Encode(buf, im, &jpeg.Options{Quality: 90}); err != nil {
		log.Printf("Failed encoding in resize(): %s", err)
		return
	}

	p.data = buf.Bytes()
	p.mediatype = "image/jpeg"
}

var extensions = map[string]string{
	"image/png":      "png",
	"image/jpeg":     "jpg",
	"image/gif":      "gif",
	"image/x-ms-bmp": "bmp",
	"image/x-bmp":    "bmp",
	"image/bmp":      "bmp",
	"image/tiff":     "tiff",
	"image/tiff-fx":  "tiff",
	"image/x-targa":  "tga",
	"image/x-tga":    "tga",
	"image/webp":     "webp",
}

var savepath = ""

func setSavepathCmd(params []string) error {
	if len(params) != 1 {
		return errors.New("set-earthporn-savepath needs one parameter")
	}

	savepath = params[0]
	return nil
}

const maxTitleLenInFilename = 100

func (p *EarthPorn) save() error {
	ext := extensions[p.mediatype]
	pp := strings.Split(p.Permalink, "/")
	threadid := pp[len(pp)-3]

	title := strings.Replace(p.Title, "/", "-", -1)
	tRunes := []rune(title)
	if len(tRunes) > maxTitleLenInFilename {
		title = string(tRunes[0:maxTitleLenInFilename])
	}

	f, err := os.Create(path.Join(savepath, threadid+" - "+title+"."+ext))
	if err != nil {
		return fmt.Errorf("Could not save earthporn: %s", err)
	}
	defer f.Close()

	if _, err := f.Write(p.origdata); err != nil {
		return fmt.Errorf("Could not save earthporn: %s", err)
	}

	p.Saved = true

	return nil
}

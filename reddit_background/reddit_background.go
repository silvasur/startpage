package reddit_background

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"github.com/silvasur/startpage/http_getter"
	"github.com/silvasur/startpage/interval"
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
	"strings"
	"time"
)

type redditList struct {
	Data struct {
		Children []struct {
			Data *RedditImage `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type RedditImage struct {
	Title     string `json:"title"`
	URL       string `json:"url,omitempty"`
	Permalink string `json:"permalink"`
	Domain    string `json:"domain"`
	Saved     bool   `json:"-"`
	Data      []byte `json:"-"`
	origdata  []byte `json:"-"`
	Mediatype string `json:"-"`
}

func GetRedditImage(maxsize int, subreddit string) (*RedditImage, error) {
	subredditUrl := fmt.Sprintf("https://www.reddit.com/r/%s.json", subreddit)

	resp, err := http_getter.Get(subredditUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var list redditList
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&list); err != nil {
		return nil, err
	}

	for _, el := range list.Data.Children {
		ri := el.Data

		if ri.Domain == fmt.Sprintf("self.%s", subreddit) {
			continue
		}

		if ri.URL == "" {
			continue
		}

		if ri.fetch(maxsize) {
			return ri, nil
		}
	}

	return nil, errors.New("Could not get RedditImage: No image could be extracted")
}

func (ri *RedditImage) fetch(maxsize int) bool {
	// TODO: We can do further processing here (e.g. if we get a link to flickr, extract the image).
	// For now, we will simply test, if the URL points to an image.

	resp, err := http.Get(ri.URL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	t := resp.Header.Get("Content-Type")
	if t == "" {
		log.Printf("could not get image of %s: no Content-Type", ri.Permalink)
		return false
	}
	mt, _, err := mime.ParseMediaType(t)
	if err != nil {
		log.Printf("could not get image of %s: %s", ri.Permalink, err)
		return false
	}

	if strings.Split(mt, "/")[0] != "image" {
		log.Printf("could not get image of %s: not an image", ri.Permalink)
		return false
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		log.Printf("could not get image of %s: %s", ri.Permalink, err)
		return false
	}

	ri.Mediatype = t
	ri.origdata = buf.Bytes()
	ri.Data = ri.origdata
	ri.resize(maxsize)
	return true
}

// resize resizes image, if it's very large
func (ri *RedditImage) resize(maxdim int) {
	im, _, err := image.Decode(bytes.NewReader(ri.origdata))
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

	ri.Data = buf.Bytes()
	ri.Mediatype = "image/jpeg"
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

const maxTitleLenInFilename = 100

func (p *RedditImage) Save(savepath string) error {
	ext := extensions[p.Mediatype]
	pp := strings.Split(p.Permalink, "/")
	threadid := pp[len(pp)-3]

	title := strings.Replace(p.Title, "/", "-", -1)
	tRunes := []rune(title)
	if len(tRunes) > maxTitleLenInFilename {
		title = string(tRunes[0:maxTitleLenInFilename])
	}

	f, err := os.Create(path.Join(savepath, threadid+" - "+title+"."+ext))
	if err != nil {
		return fmt.Errorf("Could not save image: %s", err)
	}
	defer f.Close()

	if _, err := f.Write(p.origdata); err != nil {
		return fmt.Errorf("Could not save image: %s", err)
	}

	p.Saved = true

	return nil
}

const (
	UPDATE_INTERVAL = 30 * time.Minute
	RETRY_INTERVAL  = 1 * time.Minute
)

type RedditImageProvider struct {
	intervalRunner *interval.IntervalRunner
	maxsize        int
	subreddit      string
	image          *RedditImage
}

func NewRedditImageProvider(maxsize int, subreddit string) *RedditImageProvider {
	return &RedditImageProvider{
		intervalRunner: interval.NewIntervalRunner(UPDATE_INTERVAL, RETRY_INTERVAL),
		maxsize:        maxsize,
		subreddit:      subreddit,
	}
}

func (rip *RedditImageProvider) Image() *RedditImage {
	rip.intervalRunner.Run(func() bool {
		log.Printf("Getting new RedditImage")

		var err error
		rip.image, err = GetRedditImage(rip.maxsize, rip.subreddit)

		if err == nil {
			log.Printf("Successfully loaded RedditImage")
		} else {
			log.Printf("Failed loading RedditImage: %s", err)
		}

		return err == nil
	})

	return rip.image
}

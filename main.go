package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/silvasur/startpage/reddit_background"
	"github.com/silvasur/startpage/weather"
)

var tpl *template.Template

func loadTemplateFromDir(templateDir string) error {
	var err error
	tpl, err = template.ParseFiles(path.Join(templateDir, "template.html"))

	if err != nil {
		return fmt.Errorf("Could not load template from '%s': %w", templateDir, err)
	}
	return nil
}

func loadTemplateFromGoPath() error {
	gopaths := strings.Split(os.Getenv("GOPATH"), ":")
	for _, p := range gopaths {
		var err error
		tpl, err = template.ParseFiles(
			path.Join(p, "src", "github.com", "silvasur", "startpage", "templates", "template.html"),
		)
		if err == nil {
			return nil
		}
	}

	return errors.New("could not find template in $GOPATH/src/github.com/silvasur/startpage")
}

func loadTemplate(templateDir string) error {
	if templateDir != "" {
		return loadTemplateFromDir(templateDir)
	} else {
		return loadTemplateFromGoPath()
	}
}

func buildWeatherProvider(config Config) *weather.WeatherProvider {
	if config.WeatherCoords.Lat == "" || config.WeatherCoords.Lon == "" {
		return nil
	}

	return weather.NewWeatherProvider(config.WeatherCoords.Lat, config.WeatherCoords.Lon)
}

func buildRedditImageProvider(config Config) *reddit_background.RedditImageProvider {
	subreddit := config.ImageSubreddit
	if subreddit == "" {
		subreddit = "EarthPorn"
	}

	return reddit_background.NewRedditImageProvider(config.GetBackgroundMaxdim(), subreddit)
}

func main() {
	laddr := flag.String("laddr", ":25145", "Listen on this port")
	templateDir := flag.String("template-dir", "", "Search for templates in this directory instead of default")

	flag.Parse()

	if err := loadTemplate(*templateDir); err != nil {
		log.Fatalf("Failed loading template: %s", err)
	}

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed loading config: %s", err)
	}

	redditImageProvider := buildRedditImageProvider(*config)

	http.HandleFunc("/", startpage(*config, redditImageProvider))
	http.HandleFunc("/bgimg", bgimg(redditImageProvider))
	http.HandleFunc("/update-bgimg", updateBgimg(redditImageProvider))

	if config.BackgroundSavepath != "" {
		http.HandleFunc("/savebg", savebg(redditImageProvider, config.BackgroundSavepath))
	}

	log.Fatal(http.ListenAndServe(*laddr, nil))
}

type TplWeather struct {
	Temp int
}

func convertWeather(ts *weather.TimeseriesEntry) *TplWeather {
	if ts == nil {
		return nil
	}

	return &TplWeather{
		Temp: int(math.Round(ts.Temperature())),
	}
}

type TplData struct {
	BgImage   *reddit_background.RedditImageForAjax
	Weather   *TplWeather
	Links     []Link
	CanSaveBg bool
}

func startpage(config Config, redditImageProvider *reddit_background.RedditImageProvider) http.HandlerFunc {
	weatherProvider := buildWeatherProvider(config)

	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		var curWeather *weather.TimeseriesEntry = nil
		if weatherProvider != nil {
			var err error
			if curWeather, err = weatherProvider.CurrentWeather(); err != nil {
				log.Printf("Failed getting weather: %s", err)
			}
		}

		if err := tpl.Execute(rw, &TplData{
			redditImageProvider.Image().ForAjax(),
			convertWeather(curWeather),
			config.Links,
			config.BackgroundSavepath != "",
		}); err != nil {
			log.Printf("Failed executing template: %s\n", err)
		}
	}
}

func bgimg(redditImageProvider *reddit_background.RedditImageProvider) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		image := redditImageProvider.Image()

		if image == nil || len(image.Data) == 0 {
			rw.WriteHeader(http.StatusNotFound)
		}

		rw.Header().Add("Content-Type", image.Mediatype)
		if _, err := rw.Write(image.Data); err != nil {
			log.Printf("Failed serving background: %s", err)
		}
	}
}

func updateBgimg(redditImageProvider *reddit_background.RedditImageProvider) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		updated := redditImageProvider.UpdateImage()

		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(struct {
			Updated bool
			Image   *reddit_background.RedditImageForAjax
		}{
			updated,
			redditImageProvider.Image().ForAjax(),
		})
	}
}

func savebg(redditImageProvider *reddit_background.RedditImageProvider, savepath string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		image := redditImageProvider.Image()

		if image == nil || len(image.Data) == 0 {
			fmt.Fprintln(rw, "No background image available")
			return
		}

		if err := image.Save(savepath); err != nil {
			log.Println(err)
			fmt.Fprintln(rw, err)
		}

		rw.Header().Add("Location", "/")
		rw.WriteHeader(http.StatusFound)
	}
}

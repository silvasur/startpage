package main

import (
	"encoding/json"
	"os"

	"github.com/adrg/xdg"
)

// Config contains all configuration options that are read from .config/startpage/config.json
type Config struct {
	// The place for which to get the weather data. If omitted, no weather will be shown
	WeatherCoords struct {
		Lat, Lon string
	}

	// A list of links to show
	Links []Link

	// If set, background images can be saved here
	BackgroundSavepath string

	// If set, this limits the background image size, the default is DEFAULT_BACKGROUND_MAXDIM
	BackgroundMaxdim *int

	// Get background images from this subreddit. Defaults to "EarthPorn"
	ImageSubreddit string
}

func LoadConfig() (*Config, error) {
	path, err := xdg.ConfigFile("startpage/config.json")
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	switch {
	case err == nil:
		// All OK, we can continue
	case os.IsNotExist(err):
		return &Config{}, nil
	default:
		return nil, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	if err == nil {
		return &config, nil
	} else {
		return nil, err
	}
}

const DEFAULT_BACKGROUND_MAXDIM = 2500

func (c Config) GetBackgroundMaxdim() int {
	if c.BackgroundMaxdim == nil {
		return DEFAULT_BACKGROUND_MAXDIM
	} else {
		return *c.BackgroundMaxdim
	}
}

// Package weather provides functions to retrieve eweather forecast data from yr.no
package weather

import (
	"encoding/xml"
	"github.com/silvasur/startpage/http_getter"
	"github.com/silvasur/startpage/interval"
	"log"
	"time"
)

func toTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05", s)
	return t
}

type Temperature struct {
	Value int    `xml:"value,attr"`
	Unit  string `xml:"unit,attr"`
}

type Weather struct {
	Temp   Temperature `xml:"temperature"`
	Symbol struct {
		Var string `xml:"var,attr"`
	} `xml:"symbol"`
	From string `xml:"from,attr"`
	URL  string
	Icon string
}

func (w *Weather) prepIcon() {
	w.Icon = "http://symbol.yr.no/grafikk/sym/b100/" + w.Symbol.Var + ".png"
}

type weatherdata struct {
	Forecast []*Weather `xml:"forecast>tabular>time"`
}

func CurrentWeather(place string) (*Weather, error) {
	url := "http://www.yr.no/place/" + place + "/forecast_hour_by_hour.xml"

	resp, err := http_getter.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var wd weatherdata
	dec := xml.NewDecoder(resp.Body)
	if err := dec.Decode(&wd); err != nil {
		return nil, err
	}

	w := wd.Forecast[0]
	w.URL = "http://www.yr.no/place/" + place
	w.prepIcon()

	return w, nil
}

type WeatherProvider struct {
	place          string
	intervalRunner *interval.IntervalRunner
	weather        *Weather
	err            error
}

const (
	UPDATE_INTERVAL = 30 * time.Minute
	RETRY_INTERVAL  = 1 * time.Minute
)

func NewWeatherProvider(place string) *WeatherProvider {
	return &WeatherProvider{
		place:          place,
		intervalRunner: interval.NewIntervalRunner(UPDATE_INTERVAL, RETRY_INTERVAL),
	}
}

func (wp *WeatherProvider) CurrentWeather() (*Weather, error) {
	wp.intervalRunner.Run(func() bool {
		log.Printf("Getting new weather data")
		wp.weather, wp.err = CurrentWeather(wp.place)

		if wp.err == nil {
			log.Printf("Successfully loaded weather data")
		} else {
			log.Printf("Failed loading weather data: %s", wp.err)
		}

		return wp.err == nil
	})

	return wp.weather, wp.err
}

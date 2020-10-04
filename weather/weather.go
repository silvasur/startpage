// Package weather provides functions to retrieve eweather forecast data from yr.no
package weather

import (
	"encoding/xml"
	"net/http"
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

func CurrentWeather(place string) (Weather, error) {
	url := "http://www.yr.no/place/" + place + "/forecast_hour_by_hour.xml"

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Weather{}, err
	}

	req.Header.Add("User-Agent", "github.com/slivasur/startpage/weather")

	resp, err := client.Do(req)
	if err != nil {
		return Weather{}, err
	}
	defer resp.Body.Close()

	var wd weatherdata
	dec := xml.NewDecoder(resp.Body)
	if err := dec.Decode(&wd); err != nil {
		return Weather{}, err
	}

	w := wd.Forecast[0]
	w.URL = "http://www.yr.no/place/" + place
	w.prepIcon()

	return *w, nil
}

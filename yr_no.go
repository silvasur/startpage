package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func toTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05", s)
	return t
}

type Weather struct {
	Temp   Temperature `xml:"temperature"`
	Symbol struct {
		Number int `xml:"number,attr"`
	} `xml:"symbol"`
	From string `xml:"from,attr"`
	URL  string
	Icon string
}

func (w *Weather) prepIcon(sun Sun) {
	rise := toTime(sun.Rise)
	set := toTime(sun.Set)
	t := toTime(w.From)

	night := t.Before(rise) || t.After(set)
	format := "http://symbol.yr.no/grafikk/sym/b100/%02d"
	switch w.Symbol.Number {
	case 1, 2, 3, 5, 6, 7, 8, 20, 21:
		if night {
			format += "n"
		} else {
			format += "d"
		}
	}
	format += ".png"

	w.Icon = fmt.Sprintf(format, w.Symbol.Number)
}

type Temperature struct {
	Value int    `xml:"value,attr"`
	Unit  string `xml:"unit,attr"`
}

type Sun struct {
	Rise string `xml:"rise,attr"`
	Set  string `xml:"set,attr"`
}

type weatherdata struct {
	Sun      Sun        `xml:"sun"`
	Forecast []*Weather `xml:"forecast>tabular>time"`
}

func getPlace() string {
	fh, err := os.Open(os.ExpandEnv("$HOME/.startpage-weather"))
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, fh); err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(buf.Bytes()))
}

var place = getPlace()

func CurrentWeather() (Weather, Sun, error) {
	url := "http://www.yr.no/place/" + place + "/forecast_hour_by_hour.xml"
	resp, err := http.Get(url)
	if err != nil {
		return Weather{}, Sun{}, err
	}
	defer resp.Body.Close()

	var wd weatherdata
	dec := xml.NewDecoder(resp.Body)
	if err := dec.Decode(&wd); err != nil {
		return Weather{}, Sun{}, err
	}

	w := wd.Forecast[0]
	w.URL = "http://www.yr.no/place/" + place
	w.prepIcon(wd.Sun)

	return *w, wd.Sun, nil
}

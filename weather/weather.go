// Package weather provides functions to retrieve eweather forecast data from yr.no
package weather

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/silvasur/startpage/http_getter"
	"github.com/silvasur/startpage/interval"
)

type TimeseriesEntry struct {
	Time time.Time `json:"time"`
	Data struct {
		Instant struct {
			Details struct {
				AirTemperature float64 `json:"air_temperature"`
			} `json:"details"`
		} `json:"instant"`
		Next1Hours struct {
			Summary struct {
				SymbolCode string `json:"symbol_code"`
			} `json:"summary"`
		} `json:"next_1_hours"`
	} `json:"data"`
}

func (te TimeseriesEntry) Temperature() float64 {
	return te.Data.Instant.Details.AirTemperature
}

func (te TimeseriesEntry) SymbolCode() string {
	return te.Data.Next1Hours.Summary.SymbolCode
}

type Timeseries []TimeseriesEntry

func (ts Timeseries) Current(now time.Time) *TimeseriesEntry {
	if ts == nil {
		return nil
	}

	found := false
	var out TimeseriesEntry

	for _, entry := range ts {
		if entry.Time.After(now) {
			// Is in the future, not interested
			continue
		}

		if !found || now.Sub(out.Time) < now.Sub(entry.Time) {
			out = entry
			found = true
		}
	}

	if found {
		return &out
	} else {
		return nil
	}
}

type ApiResponse struct {
	Properties struct {
		Timeseries Timeseries `json:"timeseries"`
	} `json:"properties"`
}

func truncateCoord(coord string) string {
	parts := strings.SplitN(coord, ".", 2)
	if len(parts) == 1 {
		return parts[0] + ".0"
	}
	tail := parts[1]
	if len(tail) > 4 {
		tail = tail[0:4]
	}
	return parts[0] + "." + tail
}

type WeatherProvider struct {
	lat, lon       string
	intervalRunner *interval.IntervalRunner
	timeseries     Timeseries
	expires        time.Time
	lastModified   time.Time
	err            error
}

const (
	UPDATE_INTERVAL = 30 * time.Minute
	RETRY_INTERVAL  = 1 * time.Minute
)

func NewWeatherProvider(lat, lon string) *WeatherProvider {
	return &WeatherProvider{
		lat:            lat,
		lon:            lon,
		intervalRunner: interval.NewIntervalRunner(UPDATE_INTERVAL, RETRY_INTERVAL),
	}
}

func parseTimeFromHeader(header http.Header, key string) (*time.Time, error) {
	raw := header.Get(key)
	if raw == "" {
		return nil, fmt.Errorf("Could not parse time from header %s: Not set or empty", key)
	}

	t, err := http.ParseTime(raw)
	if err != nil {
		return nil, fmt.Errorf("Could not parse time from header %s: %w", key, err)
	}

	return &t, nil
}

func updateTimeFromHeaderIfOK(header http.Header, key string, time *time.Time) {
	newTime, err := parseTimeFromHeader(header, key)
	if err != nil {
		log.Printf("Will not update time for key=%s: %s", key, err)
		return
	}

	*time = *newTime
}

func (wp *WeatherProvider) update() error {
	if time.Now().Before(wp.expires) {
		log.Printf("Will not update weather yet, as it's not yet expired (expires=%s)", wp.expires)
		return nil
	}

	url := "https://api.met.no/weatherapi/locationforecast/2.0/compact?lat=" + truncateCoord(wp.lat) + "&lon=" + truncateCoord(wp.lon)

	req, err := http_getter.BuildGetRequest(url)
	if err != nil {
		return err
	}

	if wp.timeseries != nil {
		req.Header.Add("If-Modified-Since", wp.lastModified.Format(http.TimeFormat))
	}

	resp, err := http_getter.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 304 {
		log.Println("Weather was not modified yet (got a 304)")
		return nil
	} else if resp.StatusCode != 200 {
		log.Println("Warning: Got a non-200 response from the weather API: %d", resp.StatusCode)
	}

	updateTimeFromHeaderIfOK(resp.Header, "Expires", &wp.expires)
	updateTimeFromHeaderIfOK(resp.Header, "Last-Modified", &wp.lastModified)

	var apiResponse ApiResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&apiResponse); err != nil {
		return fmt.Errorf("Failed decoding weather API response: %w", err)
	}

	wp.timeseries = apiResponse.Properties.Timeseries

	return nil
}

func (wp *WeatherProvider) CurrentWeather() (*TimeseriesEntry, error) {
	wp.intervalRunner.Run(func() bool {
		wp.err = wp.update()

		if wp.err == nil {
			log.Printf("Successfully updated weather data")
		} else {
			log.Printf("Failed updating weather data: %s", wp.err)
		}

		return wp.err == nil
	})

	if wp.err != nil {
		return nil, wp.err
	}

	return wp.timeseries.Current(time.Now()), nil
}

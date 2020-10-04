package http_getter

import (
	"net/http"
)

// Get is like http.Get, but we're sending our own user agent.
func Get(url string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "github.com/slivasur/startpage")

	return client.Do(req)
}

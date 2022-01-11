package http_getter

import (
	"net/http"
)

func BuildGetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "github.com/slivasur/startpage")
	return req, nil
}

func Do(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

// Get is like http.Get, but we're sending our own user agent.
func Get(url string) (*http.Response, error) {
	req, err := BuildGetRequest(url)
	if err != nil {
		return nil, err
	}

	return Do(req)
}

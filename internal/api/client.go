package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"dochandler/internal/config"
)

func Get(path string, queryParams map[string]string) ([]byte, error) {
	u, err := url.Parse(config.BaseURL + path)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d für %s", resp.StatusCode, u.String())
	}

	return io.ReadAll(resp.Body)
}

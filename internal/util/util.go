package util

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"net/url"
	"os"
)

// LoadYamlFile is refactored to use a generic type T.
// T must be a type that can be unmarshaled from JSON.
func LoadYamlFile[T any](pathOrURL string) (*T, error) {
	var data []byte
	var err error

	if isValidURL(pathOrURL) {
		resp, err := http.Get(pathOrURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("failed to fetch config from URL")
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = os.ReadFile(pathOrURL)
		if err != nil {
			return nil, err
		}
	}

	var config T
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func isValidURL(toTest string) bool {
	u, err := url.Parse(toTest)
	return err == nil && u.Scheme != "" && u.Host != ""
}
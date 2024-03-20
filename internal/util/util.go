package util

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"net/url"
	"os"
)

// LoadYamlFile attempts to load and parse a YAML file into a Go struct. The input can be a filepath,
// a URL, or an environment variable name containing the YAML content.
func LoadYamlFile[T any](file string) (*T, error) {
	var data []byte
	var err error

	// Check if file is an environment variable containing YAML data
	if raw, isInENV := os.LookupEnv(file); isInENV {
		data = []byte(raw)
	} else {
		// Load data from URL or local file
		if IsValidURL(file) {
			data, err = ReadFileFromURL(file)
		} else {
			data, err = os.ReadFile(file)
		}
		if err != nil {
			return nil, err
		}
	}

	// Parse YAML data into the specified struct type
	return ParseYamlFile[T](data)
}

// ParseYamlFile parses YAML data into a struct of type T.
func ParseYamlFile[T any](data []byte) (*T, error) {
	var config T
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ReadFileFromURL fetches the content of a file from a URL.
func ReadFileFromURL(pathOrURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, pathOrURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch config from URL: status code " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// IsValidURL checks if the given string is a well-formed URL.
func IsValidURL(toTest string) bool {
	u, err := url.Parse(toTest)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}

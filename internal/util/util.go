package util

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// LoadJSONFile attempts to load and parse a JSON file into a Go struct. The input can be a filepath,
// a URL, or an environment variable name containing the JSON content.
func LoadJSONFile[T any](file string) (*T, error) {
	var data []byte
	var err error

	// Check if file is an environment variable containing JSON data
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

	// Parse JSON data into the specified struct type
	return ParseJSONlFile[T](data)
}

// ParseJSONlFile parses JSON data into a struct of type T.
func ParseJSONlFile[T any](data []byte) (*T, error) {
	var config T
	if err := json.Unmarshal(data, &config); err != nil {
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

// DurationUnmarshalled is a wrapper around time.Duration to handle JSON unmarshalling.
type DurationUnmarshalled time.Duration

// UnmarshalJSON converts a JSON string to a DurationUnmarshalled.
func (d *DurationUnmarshalled) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = DurationUnmarshalled(time.Duration(value))
	case string:
		var err error
		duration, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = DurationUnmarshalled(duration)
	default:
		return errors.New("invalid duration")
	}

	return nil
}

package server_test

import (
	"net/url"
	"testing"

	"github.com/runatlantis/atlantis/server"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNormalizeBaseURL_Valid(t *testing.T) {
	t.Log("When given a valid base URL, NormalizeBaseURL returns such URLs unchanged.")
	examples := []string{
		"https://example.com",
		"https://example.com/some/path",
		"http://example.com:8080",
	}
	for _, example := range examples {
		url, err := url.Parse(example)
		Ok(t, err)
		normalized, err := server.NormalizeBaseURL(url)
		Ok(t, err)
		Equals(t, url, normalized)
	}
}

func TestNormalizeBaseURL_Relative(t *testing.T) {
	t.Log("We do not allow relative URLs as base URLs.")
	_, err := server.NormalizeBaseURL(&url.URL{Path: "hi"})
	Assert(t, err != nil, "should be an error")
	Equals(t, "Base URLs must be absolute.", err.Error())
}

func TestNormalizeBaseURL_NonHTTP(t *testing.T) {
	t.Log("Base URLs must be http or https.")
	_, err := server.NormalizeBaseURL(&url.URL{Scheme: "ftp", Host: "example", Path: "hi"})
	Assert(t, err != nil, "should be an error")
	Equals(t, "Base URLs must be HTTP or HTTPS.", err.Error())
}

func TestNormalizeBaseURL_TrailingSlashes(t *testing.T) {
	t.Log("We strip off any trailing slashes from the base URL.")
	examples := []struct {
		input  string
		output string
	}{
		{"https://example.com/", "https://example.com"},
		{"https://example.com/some/path/", "https://example.com/some/path"},
		{"http://example.com:8080/", "http://example.com:8080"},
		{"https://example.com//", "https://example.com"},
		{"https://example.com/path///", "https://example.com/path"},
	}
	for _, example := range examples {
		inputURL, err := url.Parse(example.input)
		Ok(t, err)
		outputURL, err := url.Parse(example.output)
		Ok(t, err)
		normalized, err := server.NormalizeBaseURL(inputURL)
		Ok(t, err)
		Equals(t, outputURL, normalized)
	}
}

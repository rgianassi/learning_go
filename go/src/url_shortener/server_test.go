package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddURL(t *testing.T) {
	sut := URLShortener{
		port: 9090,

		expanderRoute:   "/",
		shortenRoute:    "/shorten/",
		statisticsRoute: "/statistics",

		mappings: make(map[string]string),

		statistics: NewStatsJSON(),
	}

	tests := []struct {
		longURL          string
		shortURL         string
		expectedTotalURL int64
	}{
		{"", "", 1},
		{"a", "a", 2},
		{"b", "b", 3},
		{"a", "a", 3},
	}

	for _, test := range tests {
		sut.addURL(test.longURL, test.shortURL)
		if sut.statistics.ServerStats.TotalURL != test.expectedTotalURL {
			t.Errorf("Incorrect total URL value, got: %v, wanted: %v.", sut.statistics.ServerStats.TotalURL, test.expectedTotalURL)
		}
	}
}

func TestGetURL(t *testing.T) {
	sut := URLShortener{
		port: 9090,

		expanderRoute:   "/",
		shortenRoute:    "/shorten/",
		statisticsRoute: "/statistics",

		mappings: make(map[string]string),

		statistics: NewStatsJSON(),
	}

	tests := []struct {
		shortURL        string
		longURL         string
		expectedLongURL string
		expectedError   bool
	}{
		{"", "", "", false},
		{"q", "q", "q", false},
		{"w", "w", "w", false},
		{"e", "r", "r", true},
	}

	sut.addURL("", "")
	sut.addURL("q", "q")
	sut.addURL("w", "w")

	for _, test := range tests {
		longURL, err := sut.getURL(test.shortURL)

		if !test.expectedError && err != nil {
			t.Errorf("Unexpected error but got: %s.", err)

			if longURL != test.expectedLongURL {
				t.Errorf("Incorrect long URL value, got: %s, wanted: %s.", longURL, test.expectedLongURL)
			}
		}

		if test.expectedError && err == nil {
			t.Error("Expected error but got nil.")
		}
	}
}

func TestExpanderHandler(t *testing.T) {
	sut := URLShortener{
		port: 9090,

		expanderRoute:   "/",
		shortenRoute:    "/shorten/",
		statisticsRoute: "/statistics",

		mappings: make(map[string]string),

		statistics: NewStatsJSON(),
	}

	longURL := "https:/github.com/develersrl/powersoft-hmi"
	shortURL := "f63377"
	sut.addURL(longURL, shortURL)

	request := httptest.NewRequest("GET", "/f63377", nil)
	responseRecorder := httptest.NewRecorder()

	sut.expanderHandler(responseRecorder, request)

	response := responseRecorder.Result()

	if response.StatusCode != http.StatusSeeOther {
		t.Errorf("Unexpected status code, got: %v, wanted: %v.", response.StatusCode, http.StatusSeeOther)
	}

	if responseRecorder.HeaderMap.Get("Location") != longURL {
		t.Errorf("Unexpected location, got: %s, wanted: %s.", responseRecorder.HeaderMap.Get("Location"), longURL)
	}

	request = httptest.NewRequest("GET", "/123456", nil)
	responseRecorder = httptest.NewRecorder()

	sut.expanderHandler(responseRecorder, request)

	response = responseRecorder.Result()

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Unexpected status code, got: %v, wanted: %v.", response.StatusCode, http.StatusNotFound)
	}
}

func TestStatisticsHandler(t *testing.T) {
	sut := URLShortener{
		port: 9090,

		expanderRoute:   "/",
		shortenRoute:    "/shorten/",
		statisticsRoute: "/statistics",

		mappings: make(map[string]string),

		statistics: NewStatsJSON(),
	}

	request := httptest.NewRequest("GET", "/statistics", nil)
	responseRecorder := httptest.NewRecorder()

	sut.statisticsHandler(responseRecorder, request)

	response := responseRecorder.Result()

	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code, got: %v, wanted: %v.", response.StatusCode, http.StatusOK)
	}

	if response.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("Unexpected location, got: %s, wanted: %s.", response.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	}

	request = httptest.NewRequest("GET", "/statistics?format=json", nil)
	responseRecorder = httptest.NewRecorder()

	sut.statisticsHandler(responseRecorder, request)

	response = responseRecorder.Result()

	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code, got: %v, wanted: %v.", response.StatusCode, http.StatusOK)
	}

	body, _ := ioutil.ReadAll(response.Body)
	var stats StatsJSON
	json.Unmarshal(body, &stats)

	if stats.ServerStats.TotalURL != 0 {
		t.Errorf("Incorrect TotalURL, got: %v, wanted: %v.", stats.ServerStats.TotalURL, 0)
	}

	if stats.ServerStats.Redirects.Success != 1 {
		t.Errorf("Incorrect success, got: %v, wanted: %v.", stats.ServerStats.Redirects.Success, 1)
	}
}

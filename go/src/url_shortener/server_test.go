package main

import "testing"

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

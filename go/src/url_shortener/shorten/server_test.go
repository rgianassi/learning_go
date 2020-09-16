package shorten

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newURLShortener() URLShortener {
	return URLShortener{
		expanderRoute:   "/",
		shortenRoute:    "/shorten/",
		statisticsRoute: "/statistics",

		mappings: make(map[string]string),

		statistics: NewStatsJSON(),
	}
}

func TestAddURL(t *testing.T) {
	sut := newURLShortener()

	tests := []struct {
		longURL      string
		shortURL     string
		wantTotalURL int64
	}{
		{"", "", 1},
		{"a", "a", 2},
		{"b", "b", 3},
		{"a", "a", 3},
	}

	for _, test := range tests {
		sut.addURL(test.longURL, test.shortURL)
		if sut.statistics.ServerStats.TotalURL != test.wantTotalURL {
			t.Errorf("Incorrect total URL value, got: %v, want: %v.", sut.statistics.ServerStats.TotalURL, test.wantTotalURL)
		}
	}
}

func TestGetURL(t *testing.T) {
	sut := newURLShortener()

	tests := []struct {
		shortURL    string
		longURL     string
		wantLongURL string
		wantErr     bool
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

		if !test.wantErr && err != nil {
			t.Errorf("Unexpected error but got: %s.", err)

			if longURL != test.wantLongURL {
				t.Errorf("Incorrect long URL value, got: %s, want: %s.", longURL, test.wantLongURL)
			}
		}

		if test.wantErr && err == nil {
			t.Error("Expected error but got nil.")
		}
	}
}

func TestExpanderHandler(t *testing.T) {
	sut := newURLShortener()

	longURL := "https:/github.com/develersrl/powersoft-hmi"
	shortURL := "f63377"
	sut.addURL(longURL, shortURL)

	request := httptest.NewRequest("GET", "/f63377", nil)
	responseRecorder := httptest.NewRecorder()

	sut.expanderHandler(responseRecorder, request)

	response := responseRecorder.Result()

	if response.StatusCode != http.StatusSeeOther {
		t.Errorf("Unexpected status code, got: %v, want: %v.", response.StatusCode, http.StatusSeeOther)
	}

	if responseRecorder.HeaderMap.Get("Location") != longURL {
		t.Errorf("Unexpected location, got: %s, want: %s.", responseRecorder.HeaderMap.Get("Location"), longURL)
	}

	request = httptest.NewRequest("GET", "/123456", nil)
	responseRecorder = httptest.NewRecorder()

	sut.expanderHandler(responseRecorder, request)

	response = responseRecorder.Result()

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Unexpected status code, got: %v, want: %v.", response.StatusCode, http.StatusNotFound)
	}
}

func TestStatisticsHandler(t *testing.T) {
	sut := newURLShortener()

	request := httptest.NewRequest("GET", "/statistics", nil)
	responseRecorder := httptest.NewRecorder()

	sut.statisticsHandler(responseRecorder, request)

	response := responseRecorder.Result()

	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code, got: %v, want: %v.", response.StatusCode, http.StatusOK)
	}

	if response.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("Unexpected content type, got: %s, want: %s.", response.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	}

	request = httptest.NewRequest("GET", "/statistics?format=json", nil)
	responseRecorder = httptest.NewRecorder()

	sut.statisticsHandler(responseRecorder, request)

	response = responseRecorder.Result()

	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code, got: %v, want: %v.", response.StatusCode, http.StatusOK)
	}

	body, _ := ioutil.ReadAll(response.Body)
	var stats StatsJSON
	json.Unmarshal(body, &stats)

	if stats.ServerStats.TotalURL != 0 {
		t.Errorf("Incorrect TotalURL, got: %v, want: %v.", stats.ServerStats.TotalURL, 0)
	}

	if stats.ServerStats.Redirects.Success != 1 {
		t.Errorf("Incorrect success, got: %v, want: %v.", stats.ServerStats.Redirects.Success, 1)
	}
}

func TestShortenHandler(t *testing.T) {
	sut := newURLShortener()

	request := httptest.NewRequest("GET", "/shorten/https:/github.com/develersrl/powersoft-hmi", nil)
	request.Host = *address
	responseRecorder := httptest.NewRecorder()

	sut.shortenHandler(responseRecorder, request)

	response := responseRecorder.Result()

	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code, got: %v, want: %v.", response.StatusCode, http.StatusOK)
	}

	if response.Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Unexpected content type, got: %s, want: %s.", response.Header.Get("Content-Type"), "text/html; charset=utf-8")
	}

	body, _ := ioutil.ReadAll(response.Body)

	if string(body) != "<a href=\"http://localhost:9090/f63377\">f63377 -> https:/github.com/develersrl/powersoft-hmi</a>" {
		t.Errorf("Incorrect body, got: %s, want: %s.", body, "<a href=\"http://localhost:9090/f63377\">f63377 -> https:/github.com/develersrl/powersoft-hmi</a>")
	}
}

func TestPersistTo(t *testing.T) {
	sut := newURLShortener()

	const longURL = "https:/github.com/develersrl/powersoft-hmi"
	const shortURL = "f63377"
	var want = fmt.Sprintf(`{"%s":"%s"}`, shortURL, longURL)
	var builder strings.Builder

	sut.addURL(longURL, shortURL)

	if err := sut.persistTo(&builder); err != nil {
		t.Fatalf("Unexpected error but got: %s.", err)
	}

	got := strings.TrimSpace(builder.String())
	if got != want {
		t.Errorf("Incorrect persisted JSON, got: %s, want: %s", got, want)
	}
}

func TestUnpersistFrom(t *testing.T) {
	sut := newURLShortener()

	const longURL = "https:/github.com/develersrl/powersoft-hmi"
	const shortURL = "f63377"
	var data = fmt.Sprintf(`{"%s": "%s"}`, shortURL, longURL)

	if err := sut.unpersistFrom(strings.NewReader(data)); err != nil {
		t.Fatalf("Unexpected error but got: %s.", err)
	}

	if longURL != sut.mappings[shortURL] {
		t.Errorf("Incorrect long URL value, got: %s, want: %s.", sut.mappings[shortURL], longURL)
	}

	if sut.statistics.ServerStats.TotalURL != 1 {
		t.Errorf("Incorrect total URL value, got: %v, want: %v.", sut.statistics.ServerStats.TotalURL, 1)
	}
}

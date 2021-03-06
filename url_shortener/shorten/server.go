package shorten

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// URLShortener URL shortener server data structure
type URLShortener struct {
	expanderRoute   string
	shortenRoute    string
	statisticsRoute string

	mappings map[string]string

	statistics StatsJSON

	mux sync.Mutex
}

// NewURLShortener a URLShortener constructor
func NewURLShortener() *URLShortener {
	urlShortener := URLShortener{}

	urlShortener.expanderRoute = "/"
	urlShortener.shortenRoute = "/shorten"
	urlShortener.statisticsRoute = "/statistics"

	urlShortener.mappings = make(map[string]string)

	urlShortener.statistics = NewStatsJSON()

	return &urlShortener
}

// UnpersistFrom function reads and decodes a JSON from the reader passed in
// and then updates the URL mappings
func (c *URLShortener) UnpersistFrom(r io.Reader) error {
	decoder := json.NewDecoder(r)

	if err := decoder.Decode(&c.mappings); err != nil {
		return err
	}

	c.statistics.updateTotalURL(int64(len(c.mappings)))
	return nil
}

// PersistTo function encodes the URL mappings in a JSON written to the writer
// passed in
func (c *URLShortener) PersistTo(w io.Writer) error {
	encoder := json.NewEncoder(w)

	if err := encoder.Encode(c.mappings); err != nil {
		return err
	}

	return nil
}

// SetupHandlerFunctions setups handler functions
func (c *URLShortener) SetupHandlerFunctions() {
	http.HandleFunc(c.shortenRoute, c.shortenHandler)
	http.HandleFunc(c.statisticsRoute, c.statisticsHandler)
	http.HandleFunc(c.expanderRoute, c.expanderHandler)
}

func (c *URLShortener) addURL(longURL, shortURL string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.mappings[shortURL] = longURL

	c.statistics.updateTotalURL(int64(len(c.mappings)))
}

// GetURL returns the complete URL corresponding to the shortened URL
func (c *URLShortener) GetURL(shortURL string) (string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	longURL, ok := c.mappings[shortURL]

	if !ok {
		return "", fmt.Errorf("short URL not found: %s", shortURL)
	}

	return longURL, nil
}

func (c *URLShortener) shortenHandler(w http.ResponseWriter, r *http.Request) {
	serverAddress := r.Host
	url := r.URL
	query := url.Query()
	longURL := query.Get("url")

	shortURL := Shorten(longURL)

	c.addURL(longURL, shortURL)

	linkAddress := fmt.Sprintf("http://%s", serverAddress)
	hrefAddress := fmt.Sprintf("%s/%s", linkAddress, shortURL)
	hrefText := fmt.Sprintf("%s -> %s", shortURL, longURL)

	fmt.Fprintf(w, "<a href=\"%s\">%s</a>", hrefAddress, hrefText)
	c.statistics.incrementHandlerCounter(ShortenHandlerIndex, true)
}

func (c *URLShortener) statisticsHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL
	query := url.Query()
	format := query.Get("format")

	if f := strings.ToLower(format); f == "json" {
		jsonCandidate, err := json.Marshal(&c.statistics)

		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			c.statistics.incrementHandlerCounter(StatisticsHandlerIndex, false)
			return
		}

		fmt.Fprintf(w, "%s", jsonCandidate)
		c.statistics.incrementHandlerCounter(StatisticsHandlerIndex, true)
		return
	}

	fmt.Fprintf(w, "%s", &c.statistics)
	c.statistics.incrementHandlerCounter(StatisticsHandlerIndex, true)
}

func (c *URLShortener) expanderHandler(w http.ResponseWriter, r *http.Request) {
	shortURLCandidate := r.URL.Path[len(c.expanderRoute):]

	redirectURL, err := c.GetURL(shortURLCandidate)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		c.statistics.incrementHandlerCounter(ExpanderHandlerIndex, false)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	c.statistics.incrementHandlerCounter(ExpanderHandlerIndex, true)
}

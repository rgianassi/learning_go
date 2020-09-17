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
	ExpanderRoute   string
	ShortenRoute    string
	StatisticsRoute string

	Mappings map[string]string

	Statistics StatsJSON

	mux sync.Mutex
}

// UnpersistFrom function reads and decodes a JSON from the reader passed in
// and then updates the URL mappings
func (c *URLShortener) UnpersistFrom(r io.Reader) error {
	decoder := json.NewDecoder(r)

	if err := decoder.Decode(&c.Mappings); err != nil {
		return err
	}

	c.Statistics.updateTotalURL(int64(len(c.Mappings)))
	return nil
}

// PersistTo function encodes the URL mappings in a JSON written to the writer
// passed in
func (c *URLShortener) PersistTo(w io.Writer) error {
	encoder := json.NewEncoder(w)

	if err := encoder.Encode(c.Mappings); err != nil {
		return err
	}

	return nil
}

func (c *URLShortener) addURL(longURL, shortURL string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.Mappings[shortURL] = longURL

	c.Statistics.updateTotalURL(int64(len(c.Mappings)))
}

func (c *URLShortener) getURL(shortURL string) (string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	longURL, ok := c.Mappings[shortURL]

	if !ok {
		return "", fmt.Errorf("short URL not found: %s", shortURL)
	}

	return longURL, nil
}

// ShortenHandler is the shorten handler
func (c *URLShortener) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	serverAddress := r.Host
	longURL := r.URL.Path[len(c.ShortenRoute):]
	shortURL := Shorten(longURL)

	c.addURL(longURL, shortURL)

	linkAddress := fmt.Sprintf("http://%s", serverAddress)
	hrefAddress := fmt.Sprintf("%s/%s", linkAddress, shortURL)
	hrefText := fmt.Sprintf("%s -> %s", shortURL, longURL)

	fmt.Fprintf(w, "<a href=\"%s\">%s</a>", hrefAddress, hrefText)
	c.Statistics.incrementHandlerCounter(ShortenHandlerIndex, true)
}

// StatisticsHandler is the statistics handler
func (c *URLShortener) StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL
	query := url.Query()
	format := query.Get("format")

	if f := strings.ToLower(format); f == "json" {
		jsonCandidate, err := json.Marshal(&c.Statistics)

		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			c.Statistics.incrementHandlerCounter(StatisticsHandlerIndex, false)
			return
		}

		fmt.Fprintf(w, "%s", jsonCandidate)
		c.Statistics.incrementHandlerCounter(StatisticsHandlerIndex, true)
		return
	}

	fmt.Fprintf(w, "%s", &c.Statistics)
	c.Statistics.incrementHandlerCounter(StatisticsHandlerIndex, true)
}

// ExpanderHandler is the expander handler
func (c *URLShortener) ExpanderHandler(w http.ResponseWriter, r *http.Request) {
	shortURLCandidate := r.URL.Path[len(c.ExpanderRoute):]

	redirectURL, err := c.getURL(shortURLCandidate)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		c.Statistics.incrementHandlerCounter(ExpanderHandlerIndex, false)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	c.Statistics.incrementHandlerCounter(ExpanderHandlerIndex, true)
}

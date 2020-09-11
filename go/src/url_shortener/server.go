package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// URLShortener URL shortener server data structure
type URLShortener struct {
	port int

	expanderRoute   string
	shortenRoute    string
	statisticsRoute string

	mappings map[string]string

	statistics StatsJSON

	mux sync.Mutex
}

func (c *URLShortener) addURL(longURL, shortURL string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.mappings[shortURL] = longURL

	c.statistics.updateTotalURL(len(c.mappings))
}

func (c *URLShortener) getURL(shortURL string) (string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	longURL, ok := c.mappings[shortURL]

	if !ok {
		return "", fmt.Errorf("short URL not found: %s", shortURL)
	}

	return longURL, nil
}

func (c *URLShortener) shortenHandler(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Path[len(c.shortenRoute):]
	shortURL := shorten(longURL)

	c.addURL(longURL, shortURL)

	linkAddress := fmt.Sprintf("http://localhost:%v", c.port)
	hrefAddress := fmt.Sprintf("%s/%s", linkAddress, shortURL)
	hrefText := fmt.Sprintf("%s -> %s", shortURL, longURL)

	fmt.Fprintf(w, "<a href=\"%s\">%s</a>", hrefAddress, hrefText)
	c.statistics.incrementHandlerCounter(c.shortenRoute, true)
}

func (c *URLShortener) statisticsHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL
	query := url.Query()
	format := query.Get("format")

	if f := strings.ToLower(format); f == "json" {
		jsonCandidate, err := json.Marshal(&c.statistics)

		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			c.statistics.incrementHandlerCounter(c.statisticsRoute, false)
			return
		}

		fmt.Fprintf(w, "%s", jsonCandidate)
		c.statistics.incrementHandlerCounter(c.statisticsRoute, true)
		return
	}

	fmt.Fprintf(w, "%s", &c.statistics)
	c.statistics.incrementHandlerCounter(c.statisticsRoute, true)
}

func (c *URLShortener) expanderHandler(w http.ResponseWriter, r *http.Request) {
	shortURLCandidate := r.URL.Path[len(c.expanderRoute):]

	redirectURL, err := c.getURL(shortURLCandidate)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		c.statistics.incrementHandlerCounter(c.expanderRoute, false)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	c.statistics.incrementHandlerCounter(c.expanderRoute, true)
}

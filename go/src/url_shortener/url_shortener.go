package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

// helpers

func shorten(longURL string) string {
	hasher := sha1.New()

	hasher.Write([]byte(longURL))
	sum := hasher.Sum(nil)

	shortURL := fmt.Sprintf("%x", sum)[:6]

	return shortURL
}

// URLShortener URL shortener server data structure
type URLShortener struct {
	port int

	expanderRoute   string
	shortenRoute    string
	statisticsRoute string

	mux sync.Mutex

	mappings map[string]string

	statistics urlStatistics
}

type urlStatistics struct {
	succeededRedirects int
	failedRedirects    int
	handlerCalls       map[string]int
}

func (c *URLShortener) addURL(longURL, shortURL string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.mappings[shortURL] = longURL
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

func (c *URLShortener) incrementHandlerCounter(handler string, succeeded bool) {
	c.mux.Lock()
	defer c.mux.Unlock()

	statistics := c.statistics

	statistics.handlerCalls[handler]++

	if succeeded {
		statistics.succeededRedirects++
	} else {
		statistics.failedRedirects++
	}
}

func (c *URLShortener) getStatistics() string {
	c.mux.Lock()
	defer c.mux.Unlock()

	statistics := c.statistics
	pairsInCache := fmt.Sprintf("Number of long/short URL pairs: %v", len(c.mappings))
	succeededRedirects := fmt.Sprintf("Succeeded redirects: %v", statistics.succeededRedirects)
	failedRedirects := fmt.Sprintf("Failed redirects: %v", statistics.failedRedirects)

	handlerCalls := make([]string, 0, len(statistics.handlerCalls))

	for handlerURL, counter := range statistics.handlerCalls {
		handlerCalls = append(handlerCalls, fmt.Sprintf("Handler %s called %v time(s)", handlerURL, counter))
	}

	statisticsOut := fmt.Sprintf("Some statistics:\n\n%s\n%s\n%s\n%s", pairsInCache, succeededRedirects, failedRedirects, strings.Join(handlerCalls, "\n"))

	return statisticsOut
}

func (c *URLShortener) computeStatistics(stats *statsJSON) {
	c.mux.Lock()
	defer c.mux.Unlock()

	s := c.statistics

	serverStats := stats.serverStats

	serverStats.totalURL = len(c.mappings)
	serverStats.redirects = redirectsJSON{s.succeededRedirects, s.failedRedirects}

	serverStats.handlers = make([]handlerJSON, 0, len(s.handlerCalls))
	for handlerURL, counter := range s.handlerCalls {
		serverStats.handlers = append(serverStats.handlers, handlerJSON{handlerURL, counter})
	}
}

// handlers

func (c *URLShortener) shortenHandler(w http.ResponseWriter, r *http.Request) {
	c.incrementHandlerCounter(c.shortenRoute, true)

	longURL := r.URL.Path[len(c.shortenRoute):]
	shortURL := shorten(longURL)

	c.addURL(longURL, shortURL)

	linkAddress := fmt.Sprintf("http://localhost:%v", c.port)
	hrefAddress := fmt.Sprintf("%s/%s", linkAddress, shortURL)
	hrefText := fmt.Sprintf("%s -> %s", shortURL, longURL)

	fmt.Fprintf(w, "<a href=\"%s\">%s</a>", hrefAddress, hrefText)
}

type redirectsJSON struct {
	success int
	failed  int
}

type handlerJSON struct {
	name  string
	count int
}

type serverStatsJSON struct {
	totalURL  int
	redirects redirectsJSON
	handlers  []handlerJSON
}

type statsJSON struct {
	serverStats serverStatsJSON
}

func (c *URLShortener) statisticsHandler(w http.ResponseWriter, r *http.Request) {
	c.incrementHandlerCounter(c.statisticsRoute, true)

	url := r.URL
	query := url.Query()
	format := query["format"]

	statsJSON := statsJSON{}

	for i := 0; i < len(format); i++ {
		if f := strings.ToLower(format[i]); f == "json" {
			c.computeStatistics(&statsJSON)
			// TODO send json
			return
		}
	}

	fmt.Fprintf(w, "%s", c.getStatistics())
}

func (c *URLShortener) expanderHandler(w http.ResponseWriter, r *http.Request) {
	shortURLCandidate := r.URL.Path[len(c.expanderRoute):]

	redirectURL, err := c.getURL(shortURLCandidate)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		c.incrementHandlerCounter(c.expanderRoute, false)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	c.incrementHandlerCounter(c.expanderRoute, true)
}

func main() {
	cache := URLShortener{
		port: 9090,

		expanderRoute:   "/",
		shortenRoute:    "/shorten/",
		statisticsRoute: "/statistics",

		mappings: make(map[string]string),

		statistics: urlStatistics{
			handlerCalls: make(map[string]int),
		},
	}

	http.HandleFunc(cache.shortenRoute, cache.shortenHandler)
	http.HandleFunc(cache.statisticsRoute, cache.statisticsHandler)
	http.HandleFunc(cache.expanderRoute, cache.expanderHandler)

	listenAddress := fmt.Sprintf(":%v", cache.port)

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

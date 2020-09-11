package main

import (
	"crypto/sha1"
	"encoding/json"
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

	statistics statsJSON
}

func (c *URLShortener) addURL(longURL, shortURL string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.mappings[shortURL] = longURL

	c.statistics.ServerStats.TotalURL = len(c.mappings)
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

	if c.statistics.ServerStats.Handlers == nil {
		c.statistics.ServerStats.Handlers = make([]handlerJSON, 0)
	}

	found := false
	for i := range c.statistics.ServerStats.Handlers {
		name := c.statistics.ServerStats.Handlers[i].Name
		if name != handler {
			continue
		}

		c.statistics.ServerStats.Handlers[i].Count++
		found = true
		break
	}

	if !found {
		c.statistics.ServerStats.Handlers = append(c.statistics.ServerStats.Handlers, handlerJSON{handler, 1})
	}

	if succeeded {
		c.statistics.ServerStats.Redirects.Success++
	} else {
		c.statistics.ServerStats.Redirects.Failed++
	}
}

func (s statsJSON) String() string {
	stats := s.ServerStats

	pairsInCache := fmt.Sprintf("Number of long/short URL pairs: %v", stats.TotalURL)
	succeededRedirects := fmt.Sprintf("Succeeded redirects: %v", stats.Redirects.Success)
	failedRedirects := fmt.Sprintf("Failed redirects: %v", stats.Redirects.Failed)

	handlerCalls := make([]string, 0, len(stats.Handlers))

	for _, handler := range stats.Handlers {
		handlerCalls = append(handlerCalls, fmt.Sprintf("Handler %s called %v time(s)", handler.Name, handler.Count))
	}

	statsBody := fmt.Sprintf("Some statistics:\n\n%s\n%s\n%s\n%s", pairsInCache, succeededRedirects, failedRedirects, strings.Join(handlerCalls, "\n"))
	return statsBody
}

func (c *URLShortener) computeStatistics() statsJSON {
	c.mux.Lock()
	defer c.mux.Unlock()

	stats := statsJSON{}

	stats.ServerStats.TotalURL = len(c.mappings)

	return stats
}

func (c *URLShortener) shortenHandler(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Path[len(c.shortenRoute):]
	shortURL := shorten(longURL)

	c.addURL(longURL, shortURL)

	linkAddress := fmt.Sprintf("http://localhost:%v", c.port)
	hrefAddress := fmt.Sprintf("%s/%s", linkAddress, shortURL)
	hrefText := fmt.Sprintf("%s -> %s", shortURL, longURL)

	fmt.Fprintf(w, "<a href=\"%s\">%s</a>", hrefAddress, hrefText)
	c.incrementHandlerCounter(c.shortenRoute, true)
}

type redirectsJSON struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

type handlerJSON struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type serverStatsJSON struct {
	TotalURL  int           `json:"total_url"`
	Redirects redirectsJSON `json:"redirects"`
	Handlers  []handlerJSON `json:"handlers"`
}

type statsJSON struct {
	ServerStats serverStatsJSON `json:"server_stats"`
}

func (c *URLShortener) statisticsHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL
	query := url.Query()
	format := query.Get("format")

	if f := strings.ToLower(format); f == "json" {
		jsonCandidate, err := json.Marshal(c.statistics)

		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			c.incrementHandlerCounter(c.statisticsRoute, false)
			return
		}

		fmt.Fprintf(w, "%s", jsonCandidate)
		c.incrementHandlerCounter(c.statisticsRoute, true)
		return
	}

	fmt.Fprintf(w, "%s", c.statistics)
	c.incrementHandlerCounter(c.statisticsRoute, true)
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

		statistics: statsJSON{},
	}

	http.HandleFunc(cache.shortenRoute, cache.shortenHandler)
	http.HandleFunc(cache.statisticsRoute, cache.statisticsHandler)
	http.HandleFunc(cache.expanderRoute, cache.expanderHandler)

	listenAddress := fmt.Sprintf(":%v", cache.port)

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

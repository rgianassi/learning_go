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

func (cache *URLShortener) addURL(longURL string, shortURL string) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	cache.mappings[shortURL] = longURL
}

func (cache *URLShortener) getURL(shortURL string) (string, error) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	if longURL, ok := cache.mappings[shortURL]; ok {
		return longURL, nil
	}

	return "", fmt.Errorf("short URL not found: %s", shortURL)
}

func (cache *URLShortener) incrementHandlerCounter(handler string, succeeded bool) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	statistics := cache.statistics

	statistics.handlerCalls[handler]++

	if succeeded {
		statistics.succeededRedirects++
	} else {
		statistics.failedRedirects++
	}
}

func (cache *URLShortener) getStatistics() string {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	statistics := cache.statistics
	pairsInCache := fmt.Sprintf("Number of long/short URL pairs: %v", len(cache.mappings))
	succeededRedirects := fmt.Sprintf("Succeeded redirects: %v", statistics.succeededRedirects)
	failedRedirects := fmt.Sprintf("Failed redirects: %v", statistics.failedRedirects)

	handlerCalls := make([]string, len(statistics.handlerCalls))
	i := 0

	for handlerURL, counter := range statistics.handlerCalls {
		handlerCalls[i] = fmt.Sprintf("Handler %s called %v time(s)", handlerURL, counter)

		i++
	}

	statisticsOut := fmt.Sprintf("Some statistics:\n\n%s\n%s\n%s\n%s", pairsInCache, succeededRedirects, failedRedirects, strings.Join(handlerCalls, "\n"))

	return statisticsOut
}

// handlers

func (cache *URLShortener) shortenHandler(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Path[len(cache.shortenRoute):]
	shortURL := shorten(longURL)

	cache.addURL(longURL, shortURL)

	linkAddress := fmt.Sprintf("http://localhost:%v", cache.port)
	hrefAddress := fmt.Sprintf("%s/%s", linkAddress, shortURL)
	hrefText := fmt.Sprintf("%s -> %s", shortURL, longURL)

	fmt.Fprintf(w, "<a href=\"%s\">%s</a>", hrefAddress, hrefText)
	cache.incrementHandlerCounter(cache.shortenRoute, true)
}

func (cache *URLShortener) statisticsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", cache.getStatistics())
	cache.incrementHandlerCounter(cache.statisticsRoute, true)
}

func (cache *URLShortener) expanderHandler(w http.ResponseWriter, r *http.Request) {
	shortURLCandidate := r.URL.Path[len(cache.expanderRoute):]

	redirectURL, err := cache.getURL(shortURLCandidate)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		cache.incrementHandlerCounter(cache.expanderRoute, false)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	cache.incrementHandlerCounter(cache.expanderRoute, true)
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

package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

var (
	cache URLShortener = URLShortener{shorts: make(map[string]*urlInfo)}
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
	mux    sync.Mutex
	shorts map[string]*urlInfo
}

type urlInfo struct {
	longURL string

	// statistics
	shortURL string
	count    int
}

func (cache *URLShortener) addURL(longURL string, shortURL string) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	cache.shorts[shortURL] = &urlInfo{longURL, shortURL, 0}
}

func (cache *URLShortener) getURL(shortURL string) (string, error) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	if info, ok := cache.shorts[shortURL]; ok {
		return info.longURL, nil
	}

	return "", fmt.Errorf("short URL not found: %s", shortURL)
}

func (cache *URLShortener) incrementURLCounter(shortURL string) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	cache.shorts[shortURL].count++
}

func (cache *URLShortener) getStatistics() string {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	visits := 0
	stats := make([]string, len(cache.shorts))
	i := 0

	for shortURL, info := range cache.shorts {
		counter := info.count
		longURL := info.longURL

		visits += counter
		stats[i] = fmt.Sprintf("URL: [%s] %s visited %v time(s)", shortURL, longURL, counter)

		i++
	}

	statistics := fmt.Sprintf("Some statistics:\n\n%s\n\nTotal visits: %v", strings.Join(stats, "\n"), visits)

	return statistics
}

// handlers

func shortenerHandler(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Path[len("/shorten/"):]
	shortURL := shorten(longURL)

	cache.addURL(longURL, shortURL)

	fmt.Fprintf(w, "/%s", shortURL)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := cache.getStatistics()
	fmt.Fprintf(w, "%s", stats)
}

func expanderHandler(w http.ResponseWriter, r *http.Request) {
	shortURLCandidate := r.URL.Path[len("/"):]

	redirectURL, err := cache.getURL(shortURLCandidate)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cache.incrementURLCounter(shortURLCandidate)

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/shorten/", shortenerHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/", expanderHandler)

	log.Fatal(http.ListenAndServe(":9090", nil))
}

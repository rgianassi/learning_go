package main

import (
	"fmt"
	"strings"
	"sync"
)

// StatsJSON Statistic data ready for JSON serialization
type StatsJSON struct {
	ServerStats serverStatsJSON `json:"server_stats"`

	mux sync.Mutex
}

type serverStatsJSON struct {
	TotalURL  int           `json:"total_url"`
	Redirects redirectsJSON `json:"redirects"`
	Handlers  []handlerJSON `json:"handlers"`
}

type redirectsJSON struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

type handlerJSON struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (s *StatsJSON) String() string {
	s.mux.Lock()
	defer s.mux.Unlock()

	stats := s.ServerStats

	pairsInCache := fmt.Sprintf("Number of long/short URL pairs: %v", stats.TotalURL)

	redirects := stats.Redirects
	succeededRedirects := fmt.Sprintf("Succeeded redirects: %v", redirects.Success)
	failedRedirects := fmt.Sprintf("Failed redirects: %v", redirects.Failed)

	handlers := &stats.Handlers
	handlerCalls := make([]string, 0, len(*handlers))

	for _, handler := range *handlers {
		handlerCalls = append(handlerCalls, fmt.Sprintf("Handler %s called %v time(s)", handler.Name, handler.Count))
	}

	statsBody := fmt.Sprintf("Some statistics:\n\n%s\n%s\n%s\n%s", pairsInCache, succeededRedirects, failedRedirects, strings.Join(handlerCalls, "\n"))
	return statsBody
}

func (s *StatsJSON) updateTotalURL(totalURL int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	stats := &s.ServerStats

	stats.TotalURL = totalURL
}

func (s *StatsJSON) incrementHandlerCounter(handlerURL string, succeeded bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	stats := &s.ServerStats

	handlers := &stats.Handlers
	if handlers == nil {
		*handlers = make([]handlerJSON, 0)
	}

	found := false
	for i := range *handlers {
		handler := &(*handlers)[i]
		name := handler.Name
		if name != handlerURL {
			continue
		}

		handler.Count++
		found = true
		break
	}

	if !found {
		*handlers = append(*handlers, handlerJSON{handlerURL, 1})
	}

	redirects := &stats.Redirects
	if succeeded {
		redirects.Success++
	} else {
		redirects.Failed++
	}
}

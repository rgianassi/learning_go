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

	statsBody := &strings.Builder{}

	stats := s.ServerStats

	statsBody.WriteString("Some statistics:\n\n")
	fmt.Fprintf(statsBody, "Number of long/short URL pairs: %v\n", stats.TotalURL)

	redirects := stats.Redirects
	fmt.Fprintf(statsBody, "Succeeded redirects: %v\n", redirects.Success)
	fmt.Fprintf(statsBody, "Failed redirects: %v\n", redirects.Failed)

	handlers := &stats.Handlers
	for _, handler := range *handlers {
		name := handler.Name
		count := handler.Count
		fmt.Fprintf(statsBody, "Handler %s called %v time(s)\n", name, count)
	}

	return statsBody.String()
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

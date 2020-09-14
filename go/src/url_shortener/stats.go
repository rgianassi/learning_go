package main

import (
	"fmt"
	"strings"
	"sync"
)

// HandlerIndex an index for handlers
type HandlerIndex int

// Indexes for all handlers
const (
	ShortenHandlerIndex HandlerIndex = iota
	StatisticsHandlerIndex
	ExpanderHandlerIndex
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
	index HandlerIndex
}

// NewStatsJSON a StatsJSON constructor
func NewStatsJSON() StatsJSON {
	statsJSON := StatsJSON{}

	stats := &statsJSON.ServerStats

	stats.TotalURL = 0

	redirects := &stats.Redirects
	redirects.Success = 0
	redirects.Failed = 0

	handlers := &stats.Handlers
	*handlers = make([]handlerJSON, 0)
	*handlers = append(*handlers, handlerJSON{"/shorten/", 0, ShortenHandlerIndex})
	*handlers = append(*handlers, handlerJSON{"/statistics", 0, StatisticsHandlerIndex})
	*handlers = append(*handlers, handlerJSON{"/", 0, ExpanderHandlerIndex})

	return statsJSON
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

func (s *StatsJSON) incrementHandlerCounter(handlerIndex HandlerIndex, succeeded bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	stats := &s.ServerStats

	handlers := &stats.Handlers
	for i := range *handlers {
		handler := &(*handlers)[i]
		index := handler.index
		if index != handlerIndex {
			continue
		}

		handler.Count++
		break
	}

	redirects := &stats.Redirects
	if succeeded {
		redirects.Success++
	} else {
		redirects.Failed++
	}
}

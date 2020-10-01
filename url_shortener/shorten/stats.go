package shorten

import (
	"fmt"
	"strings"
	"sync/atomic"
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
}

type serverStatsJSON struct {
	TotalURL  int64         `json:"total_url"`
	Redirects redirectsJSON `json:"redirects"`
	Handlers  []handlerJSON `json:"handlers"`
}

type redirectsJSON struct {
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
}

type handlerJSON struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
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
	statsBody := &strings.Builder{}

	stats := s.ServerStats

	statsBody.WriteString("Some statistics:\n\n")
	totalURL := atomic.LoadInt64(&stats.TotalURL)
	fmt.Fprintf(statsBody, "Number of long/short URL pairs: %v\n", totalURL)

	redirects := stats.Redirects
	succeeded := atomic.LoadInt64(&redirects.Success)
	fmt.Fprintf(statsBody, "Succeeded redirects: %v\n", succeeded)
	failed := atomic.LoadInt64(&redirects.Failed)
	fmt.Fprintf(statsBody, "Failed redirects: %v\n", failed)

	handlers := &stats.Handlers
	for _, handler := range *handlers {
		name := handler.Name
		count := atomic.LoadInt64(&handler.Count)
		fmt.Fprintf(statsBody, "Handler %s called %v time(s)\n", name, count)
	}

	return statsBody.String()
}

func (s *StatsJSON) updateTotalURL(totalURL int64) {
	stats := &s.ServerStats

	atomic.StoreInt64(&stats.TotalURL, totalURL)
}

func (s *StatsJSON) incrementHandlerCounter(handlerIndex HandlerIndex, succeeded bool) {
	stats := &s.ServerStats

	handlers := &stats.Handlers
	for i := range *handlers {
		handler := &(*handlers)[i]
		index := handler.index
		if index != handlerIndex {
			continue
		}

		atomic.AddInt64(&handler.Count, 1)
		break
	}

	redirects := &stats.Redirects
	if succeeded {
		atomic.AddInt64(&redirects.Success, 1)
	} else {
		atomic.AddInt64(&redirects.Failed, 1)
	}
}

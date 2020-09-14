package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	cache := URLShortener{
		port: 9090,

		expanderRoute:   "/",
		shortenRoute:    "/shorten/",
		statisticsRoute: "/statistics",

		mappings: make(map[string]string),

		statistics: NewStatsJSON(),
	}

	http.HandleFunc(cache.shortenRoute, cache.shortenHandler)
	http.HandleFunc(cache.statisticsRoute, cache.statisticsHandler)
	http.HandleFunc(cache.expanderRoute, cache.expanderHandler)

	listenAddress := fmt.Sprintf(":%v", cache.port)

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

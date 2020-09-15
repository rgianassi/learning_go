package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	address     = flag.String("addr", "localhost:9090", "server listen address")
	persistence = flag.String("load", "persistence.json", "persistence JSON file for URLs")
)

func main() {
	flag.Parse()

	cache := URLShortener{
		expanderRoute:   "/",
		shortenRoute:    "/shorten/",
		statisticsRoute: "/statistics",

		mappings: make(map[string]string),

		statistics: NewStatsJSON(),
	}

	http.HandleFunc(cache.shortenRoute, cache.shortenHandler)
	http.HandleFunc(cache.statisticsRoute, cache.statisticsHandler)
	http.HandleFunc(cache.expanderRoute, cache.expanderHandler)

	listenAddress := fmt.Sprintf("%s", *address)

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

/**
Implement server graceful shutdown. The server should handle SIGINT and SIGKILL
 by saving all its URLs in a file in the current directory, this file should have
  the same format that the one the -load option accepts.
*/

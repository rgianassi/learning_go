package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	address = flag.String("addr", "localhost:9090", "server listen address")
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
1.First the program should now accept some command line flags:

-addr localhost:9090 should set the server to listen to that address
-load FILE should load a JSON file containing the short/long URL pairs
 that the server will immediately be able to redirect

Statistics don't need to be persisted, but the statistic that indicates
 the number of URLs in the server should reflect the reality.

Implement server graceful shutdown. The server should handle SIGINT and SIGKILL
 by saving all its URLs in a file in the current directory, this file should have
  the same format that the one the -load option accepts.
*/

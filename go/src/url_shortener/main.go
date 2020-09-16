package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	log.Println("loading persistence data from:", *persistence)
	cache.loadPersistenceFile(*persistence)

	var server http.Server

	idleConnectionsClosed := make(chan struct{})
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGKILL)
		<-signalChannel

		// We received an interrupt signal, shut down.
		log.Println("shutting down...")
		if err := server.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}

		log.Println("storing persistence data to:", *persistence)
		cache.storePersistenceFile(*persistence)

		close(idleConnectionsClosed)
	}()

	server.Addr = fmt.Sprintf("%s", *address)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnectionsClosed
	log.Println("shutdown completed")
}

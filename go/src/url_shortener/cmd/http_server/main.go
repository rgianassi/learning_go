package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rgianassi/learning/go/src/url_shortener/shorten"
)

var (
	address     = flag.String("addr", "localhost:9090", "server listen address")
	persistence = flag.String("load", "persistence.json", "persistence JSON file for URLs")
)

func unpersist(cache *shorten.URLShortener) {
	log.Println("loading persistence data from:", *persistence)

	f, err := os.Open(*persistence)
	if err != nil {
		log.Println("error unpersisting:", err)
	} else {
		defer f.Close()
	}

	reader := bufio.NewReader(f)

	cache.UnpersistFrom(reader)
}

func persist(cache *shorten.URLShortener) {
	log.Println("storing persistence data to:", *persistence)

	f, err := os.Open(*persistence)
	if err != nil {
		log.Fatalln("error persisting:", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	cache.PersistTo(writer)
}

func setupHTTPServerShutdown(cache *shorten.URLShortener, server *http.Server, idleConnectionsClosed chan struct{}) {
	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGKILL)

	<-signalChannel

	log.Println("shutting down...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server Shutdown error: %v", err)
	}

	persist(cache)

	close(idleConnectionsClosed)
}

func launchHTTPServer(server *http.Server) {
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe error: %v", err)
	}
}

func main() {
	flag.Parse()

	idleConnectionsClosed := make(chan struct{})

	var server http.Server
	server.Addr = fmt.Sprintf("%s", *address)

	cache := shorten.NewURLShortener()

	cache.SetupHandlerFunctions()
	unpersist(cache)

	go setupHTTPServerShutdown(cache, &server, idleConnectionsClosed)

	launchHTTPServer(&server)

	<-idleConnectionsClosed
	log.Println("shutdown completed")
}

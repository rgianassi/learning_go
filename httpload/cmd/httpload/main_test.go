package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNumberOfRequests(t *testing.T) {
	want := 100
	counter := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		w.WriteHeader(200)
		w.Write([]byte("Success!"))
	}))
	defer srv.Close()

	flags := flag.NewFlagSet("httpload test", flag.ExitOnError)
	flags.Usage = func() {
		progName := os.Args[0]
		fmt.Fprintf(flags.Output(), "Usage: %s [options...] URL\n", progName)
		flags.PrintDefaults()
	}

	args := []string{"-w", "10", "-n", "100", srv.URL}

	trueMain(flags, args)

	if counter != want {
		t.Fatal("failed to fulfill requests, want:", want, "got:", counter)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/rgianassi/learning/go/httpload/pkg/loader"
)

const (
	exitCodeOk    = 0
	exitCodeError = 1
)

func setupGracefulShutdown() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, os.Kill)
		<-signalChannel

		cancel()
	}()

	return ctx
}

func trueMain(flags *flag.FlagSet, args []string) int {
	ctx := setupGracefulShutdown()

	config := loader.NewConfigFromFlags(flags)

	if err := config.Parse(flags, args); err != nil {
		fmt.Println("main: error during arguments parsing. Error:", err)
		flags.Usage()
		return exitCodeError
	}

	if err := config.CheckFlags(); err != nil {
		log.Println("main: error checking flags. Error:", err)
		return exitCodeError
	}

	loadTester := loader.NewLoadTesterFromConfig(config)
	err := loadTester.RunLoaderPipeline(ctx)
	if err != nil {
		log.Println("main: error during load test. Error:", err)
		os.Exit(exitCodeError)
	}

	outBuilder := &strings.Builder{}
	loadTester.WriteResults(outBuilder)
	fmt.Println(outBuilder.String())

	return exitCodeOk
}

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.Usage = func() {
		progName := os.Args[0]
		fmt.Fprintf(flags.Output(), "Usage: %s [options...] URL\n", progName)
		flags.PrintDefaults()
	}

	exitCode := trueMain(flags, os.Args[1:])

	os.Exit(exitCode)
}

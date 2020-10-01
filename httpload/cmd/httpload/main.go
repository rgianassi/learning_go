package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rgianassi/learning/go/httpload/pkg/loader"
)

const (
	exitCodeOk    = 0
	exitCodeError = 1
)

// Results is a list of Result
type Results []Result

// Result represents a result from a request made by the load tester
type Result struct {
	status int
	timing time.Duration
}

// LoadTester represents an instance of the load testing logic
type LoadTester struct {
	results Results
	config  *loader.Config
}

func newLoadTesterFromConfig(config *loader.Config) LoadTester {
	loadTester := LoadTester{}
	loadTester.config = config
	return loadTester
}

func (lt *LoadTester) writeResults(w io.Writer) {
	fmt.Fprintf(w, "\n%s\n\n", "Summary:")
	lt.dumpTimings(w)

	fmt.Fprintf(w, "\n%s\n\n", "Status code distribution:")
	lt.dumpStatuses(w)
}

func (lt *LoadTester) genLoadRequest(ctx context.Context, in <-chan string) (<-chan Result, <-chan error, error) {
	out := make(chan Result)
	errc := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errc)

		for url := range in {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				errc <- err
				return
			}

			req = req.WithContext(ctx)

			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errc <- err
				return
			}
			defer resp.Body.Close()

			elapsed := time.Since(start)
			code := resp.StatusCode

			select {
			case out <- Result{code, elapsed}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, errc, nil
}

func (lt *LoadTester) genLoadRequests(ctx context.Context, in <-chan string) (<-chan Result, <-chan error, error) {
	out := make(chan Result)
	errc := make(chan error, 1)

	var wg sync.WaitGroup
	numWorkers := lt.config.NWorkers()
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			resultc, errcr, err := lt.genLoadRequest(ctx, in)
			if err != nil {
				errc <- err
				return
			}

			for result := range resultc {
				select {
				case out <- result:
				case <-ctx.Done():
					return
				}
			}
			for err := range errcr {
				select {
				case errc <- err:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()

		close(out)
		close(errc)
	}()

	return out, errc, nil
}

func (lt *LoadTester) genResults(ctx context.Context, in <-chan Result) (<-chan error, error) {
	errc := make(chan error, 1) // signals goroutine completion

	go func() {
		defer close(errc)

		for result := range in {
			lt.results = append(lt.results, result)
		}
	}()

	return errc, nil
}

func (lt *LoadTester) waitForPipeline(errs ...<-chan error) error {
	errc := lt.mergeErrors(errs...)
	for err := range errc {
		if err != nil {
			return err
		}
	}
	return nil
}

func (lt *LoadTester) mergeErrors(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup

	// We must ensure that the output channel has the capacity to
	// hold as many errors as there are error channels.
	// This will ensure that it never blocks, even if WaitForPipeline returns early.
	out := make(chan error, len(cs))

	// Start an output goroutine for each input channel in cs. output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are done.
	// This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (lt *LoadTester) run(done chan bool) error {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if lt.config.AppDuration() > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), lt.config.AppDuration())
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	go func() {
		<-done
		cancel()
	}()

	var errcList []<-chan error
	urlc, errc, err := lt.config.RequestsSource(ctx)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	resultc, errc, err := lt.genLoadRequests(ctx, urlc)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	errc, err = lt.genResults(ctx, resultc)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	err = lt.waitForPipeline(errcList...)

	done <- true

	return err
}

func (lt *LoadTester) dumpStatuses(w io.Writer) {
	statuses := make(map[int]int)

	for _, result := range lt.results {
		statuses[result.status]++
	}

	keys := make([]int, 0, len(statuses))

	for key := range statuses {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	for _, key := range keys {
		code := float64(key)
		count := float64(statuses[key])
		fmt.Fprintf(w, "%s\n", fmt.Sprintf("[%3.f] %12.f response(s)", code, count))
	}
}

func (lt *LoadTester) dumpTimings(w io.Writer) {
	totalTiming := float64(0)
	maxTiming := float64(0)
	minTiming := float64(0)
	averageTiming := float64(0)
	requestsPerSecond := float64(0)

	n := len(lt.results)
	noResults := (n == 0)

	if noResults {
		return
	}

	timing0 := lt.results[0].timing.Seconds()
	totalTiming = timing0
	maxTiming = timing0
	minTiming = timing0

	for i := 1; i < n; i++ {
		timing := lt.results[i].timing.Seconds()

		totalTiming += timing

		if timing > maxTiming {
			maxTiming = timing
		}

		if timing < minTiming {
			minTiming = timing
		}
	}

	averageTiming = totalTiming / float64(n)
	requestsPerSecond = float64(1) / averageTiming

	fmt.Fprintf(w, "%s\n", fmt.Sprintf("Total:        %12.4f secs", totalTiming))
	fmt.Fprintf(w, "%s\n", fmt.Sprintf("Slowest:      %12.4f secs", maxTiming))
	fmt.Fprintf(w, "%s\n", fmt.Sprintf("Fastest:      %12.4f secs", minTiming))
	fmt.Fprintf(w, "%s\n", fmt.Sprintf("Average:      %12.4f secs", averageTiming))
	fmt.Fprintf(w, "%s\n", fmt.Sprintf("Requests/sec: %12.4f", requestsPerSecond))
}

func setupGracefulShutdown(done chan bool) {
	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, os.Interrupt, os.Kill)

	<-signalChannel

	done <- true
}

func trueMain(flags *flag.FlagSet, args []string) int {
	done := make(chan bool)
	defer close(done)
	go setupGracefulShutdown(done)

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

	loadTester := newLoadTesterFromConfig(config)

	go func() {
		err := loadTester.run(done)

		if errors.Is(err, context.DeadlineExceeded) {
			return
		}

		if err != nil {
			log.Println("main: error during load test. Error:", err)
			os.Exit(exitCodeError)
		}
	}()

	<-done

	outBuilder := &strings.Builder{}
	loadTester.writeResults(outBuilder)
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

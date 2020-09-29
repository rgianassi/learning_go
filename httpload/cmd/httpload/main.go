package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	exitCodeOk    = 0
	exitCodeError = 1

	usageMessage = `Usage: httpload [options...] URL

	Options:
		-w int		number of concurrent workers running (default: 50)
		-n int		number of requests to run (default: 200): must be equal or greater than -w
		-z duration	application duration to send requests (default: unlimited)
`
)

// Results is a list of Result
type Results []Result

// Result represents a result from a request made by the load tester
type Result struct {
	status int
	timing time.Duration
}

// Config manages the configuration parameters passed on the command line
type Config struct {
	nWorkers    int
	nRequests   int
	appDuration time.Duration
	url         string
}

func newConfigFromFlags(flags *flag.FlagSet) Config {
	config := Config{}
	flags.IntVar(&config.nWorkers, "w", 50, "number of concurrent workers running")
	flags.IntVar(&config.nRequests, "n", 200, "number of requests to run")
	flags.DurationVar(&config.appDuration, "z", 0, "application duration to send requests")
	return config
}

func (c *Config) checkFlags() (err error) {
	if c.nRequests < c.nWorkers {
		err = fmt.Errorf("the number of requests to run (%v) cannot be less than the number of workers (%v)", c.nRequests, c.nWorkers)
	}
	return err
}

func (c *Config) parse(flags *flag.FlagSet, args []string) (err error) {
	flags.Parse(args)

	if flags.NArg() != 1 {
		return fmt.Errorf("Wrong number of arguments")
	}

	c.url = flags.Arg(0)
	return err
}

func (c *Config) genRequests(ctx context.Context) (<-chan string, <-chan error, error) {
	theRequest := c.url

	if theRequest == "" {
		return nil, nil, errors.Errorf("no URL provided")
	}

	out := make(chan string)
	errc := make(chan error, 1) // unused

	go func() {
		defer close(out)
		defer close(errc)

		proceed := func(i int) bool {
			// if the duration is defined we emit requests indefinitely
			isDurationDefined := (c.appDuration > 0)
			// if no duration is defined, emit -n requests
			isLoopUnfinished := (i < c.nRequests)

			isUnfinished := isDurationDefined || isLoopUnfinished
			return isUnfinished
		}

		for i := 0; proceed(i); i++ {
			select {
			case out <- theRequest:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, errc, nil
}

// LoadTester represents an instance of the load testing logic
type LoadTester struct {
	results Results
}

func newLoadTesterFromConfig(config Config) LoadTester {
	loadTester := LoadTester{}
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

func (lt *LoadTester) genLoadRequests(ctx context.Context, config Config, in <-chan string) (<-chan Result, <-chan error, error) {
	out := make(chan Result)
	errc := make(chan error, 1)

	var wg sync.WaitGroup
	numWorkers := config.nWorkers
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

func (lt *LoadTester) run(config Config) error {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if config.appDuration > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), config.appDuration)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	var errcList []<-chan error
	urlc, errc, err := config.genRequests(ctx)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	resultc, errc, err := lt.genLoadRequests(ctx, config, urlc)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	errc, err = lt.genResults(ctx, resultc)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	return lt.waitForPipeline(errcList...)
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

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.Usage = func() {
		progName := os.Args[0]
		fmt.Fprintf(flags.Output(), "Usage: %s [options...] URL\n", progName)
		flags.PrintDefaults()
	}

	config := newConfigFromFlags(flags)

	if err := config.parse(flags, os.Args[1:]); err != nil {
		fmt.Println("main: error during arguments parsing. Error:", err)
		flags.Usage()
		os.Exit(exitCodeError)
	}

	if err := config.checkFlags(); err != nil {
		log.Println("main: error checking flags. Error:", err)
		os.Exit(exitCodeError)
	}

	loadTester := newLoadTesterFromConfig(config)

	if err := loadTester.run(config); err != nil {
		log.Println("main: error during load test. Error:", err)
		os.Exit(exitCodeError)
	}

	outBuilder := &strings.Builder{}
	loadTester.writeResults(outBuilder)
	fmt.Println(outBuilder.String())

	os.Exit(exitCodeOk)
}

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
	"time"
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
	flag.IntVar(&config.nWorkers, "w", 50, "number of concurrent workers running")
	flag.IntVar(&config.nRequests, "n", 200, "number of requests to run")
	flag.DurationVar(&config.appDuration, "z", 0, "application duration to send requests")
	return config
}

func (c *Config) checkFlags() (err error) {
	if c.nRequests < c.nWorkers {
		err = fmt.Errorf("the number of requests to run (%v) cannot be less than the number of workers (%v)", c.nRequests, c.nWorkers)
	}
	return err
}

func (c *Config) parse(flags *flag.FlagSet) (err error) {
	flags.Parse(os.Args[1:])

	if flags.NArg() != 1 {
		return fmt.Errorf("Wrong number of arguments")
	}

	c.url = flags.Arg(0)
	return err
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

	return lt.makeLoadRequest(ctx, config.url)
}

func (lt *LoadTester) makeLoadRequest(ctx context.Context, query string) error {
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return err
	}

	start := time.Now()
	err = lt.httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		elapsed := time.Since(start)
		code := resp.StatusCode

		lt.results = append(lt.results, Result{code, elapsed})

		return nil
	})

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

func (lt *LoadTester) httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	c := make(chan error, 1)
	req = req.WithContext(ctx)
	go func() { c <- f(http.DefaultClient.Do(req)) }()
	select {
	case <-ctx.Done():
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func main() {
	var flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.SetOutput(nil)
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage of %s:\n", os.Args[0])
		flags.PrintDefaults()
	}

	config := newConfigFromFlags(flags)

	if err := config.parse(flags); err != nil {
		fmt.Println("main: error during arguments parsing. Error:", err)
		flags.PrintDefaults()
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

/*

## description

Write a load tester for web servers, that is, a program that simulates
some load for programs that responds to HTTP requests.

## options:

The program should have the following command-line interface and accept
the following options:

```
Usage: httpload [options...] URL

Options:

-w int        number of workers to run concurrently. default:50.
-n int        number of requests to run. default:200.
-z string   duration of application to send requests. default:unlimited.
```

Where:

- `URL` is the URL of the server to load-test
- the total number of requests cannot be smaller than the number of
concurrent workers
- If `-z` is given then `-n` is ignored.
- if `-z` is given then the application stops and exits after the
specified duration.
- example of valid duration: `-z 10m` , `-z 3s`

## final report

When the program exits (either by itself if `-z` was given) or
in response of a CTRL-C, the program should dump on standard output
the following information:

A summary about request lengths:

```
Summary:

  Total:        0.0326 secs
  Slowest:      0.0295 secs
  Fastest:      0.0003 secs
  Average:      0.0054 secs
  Requests/sec: 6132.8565
```

A distribution of the status codes:

```
Status code distribution:

  [200] 198 responses
  [404] 2 responses
```

## testing

Write one unit test that verifies that the program respects the specified
number of requests (`-n`)

## workflow

It's ok and advised to split the work into multiple atomic, self-contained PR.

## hints:

- https://golang.org/pkg/time/#ParseDuration
- https://blog.golang.org/pipelines
- https://blog.golang.org/context
- https://golang.org/pkg/net/http/httptest/#NewServer for testing

*/

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

type results []result

type result struct {
	status int
	timing time.Duration
}

var (
	nWorkers    = flag.Int("w", 50, "number of concurrent workers running (default: 50)")
	nRequests   = flag.Int("n", 200, "number of requests to run (default: 200)")
	appDuration = flag.Duration("z", 0, "application duration to send requests (default: unlimited)")
)

func dumpStatuses(w io.Writer, results results) {
	statuses := make(map[int]int)

	for _, result := range results {
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

func dumpTimings(w io.Writer, results results) {
	totalTiming := float64(0)
	maxTiming := float64(0)
	minTiming := float64(0)
	averageTiming := float64(0)
	requestsPerSecond := float64(0)

	n := len(results)
	noResults := (n == 0)

	if noResults {
		return
	}

	timing0 := results[0].timing.Seconds()
	totalTiming = timing0
	maxTiming = timing0
	minTiming = timing0

	for i := 1; i < n; i++ {
		timing := results[i].timing.Seconds()

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

func checkFlags(nWorkers int, nRequests int, appDuration time.Duration) (err error) {
	if nRequests < nWorkers {
		err = fmt.Errorf("the number of requests to run (%v) cannot be less than the number of workers (%v)", nRequests, nWorkers)
	}
	return err
}

func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
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

func makeLoadRequest(ctx context.Context, query string, results results) (results, error) {
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return results, err
	}

	start := time.Now()
	err = httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		elapsed := time.Since(start)
		code := resp.StatusCode

		results = append(results, result{code, elapsed})

		return nil
	})

	return results, err
}

func loadServer(nWorkers int, nRequests int, appDuration time.Duration, url string, results results) (results, error) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if appDuration > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), appDuration)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	return makeLoadRequest(ctx, url, results)
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println(usageMessage)
		os.Exit(exitCodeError)
	}
	theURL := flag.Arg(0)

	if err := checkFlags(*nWorkers, *nRequests, *appDuration); err != nil {
		log.Println("main: error checking flags. Error:", err)
		os.Exit(exitCodeError)
	}

	var results results
	results, err := loadServer(*nWorkers, *nRequests, *appDuration, theURL, results)
	if err != nil {
		log.Println("main: error during load test. Error:", err)
		os.Exit(exitCodeError)
	}

	outBuilder := &strings.Builder{}
	fmt.Fprintf(outBuilder, "\n%s\n\n", "Summary:")
	dumpTimings(outBuilder, results)

	fmt.Fprintf(outBuilder, "\n%s\n\n", "Status code distribution:")
	dumpStatuses(outBuilder, results)

	fmt.Println(outBuilder.String())

	fmt.Println("Arguments:", *nWorkers, *nRequests, *appDuration, theURL)

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

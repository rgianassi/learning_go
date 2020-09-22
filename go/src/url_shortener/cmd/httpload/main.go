package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

const (
	exitCodeOk    = 0
	exitCodeError = 1
)

func computeStatuses() (statuses map[int]int, err error) {
	statuses = make(map[int]int)

	statuses[200] = 198
	statuses[404] = 2

	return statuses, err
}

func dumpStatuses(statuses map[int]int) string {
	dumpBuilder := &strings.Builder{}

	keys := make([]int, 0, len(statuses))

	for key := range statuses {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	for _, key := range keys {
		code := float64(key)
		count := float64(statuses[key])
		fmt.Fprintf(dumpBuilder, "%s\n", fmt.Sprintf("[%3.f] %12.f response(s)", code, count))
	}

	return dumpBuilder.String()
}

func computeTimings() (timings []float64, err error) {
	timings = make([]float64, 0, 512)

	timings = append(timings, 0.0003)
	timings = append(timings, 0.0054)
	timings = append(timings, 0.0024)
	timings = append(timings, 0.0295)

	return timings, err
}

func dumpTimings(timings []float64) string {
	dumpBuilder := &strings.Builder{}

	totalTiming := float64(0)
	slowestTiming := float64(0)
	fastestTiming := float64(0)
	averageTiming := float64(0)
	requestsPerSecond := float64(0)

	n := len(timings)

	if n > 0 {
		totalTiming = timings[0]
		slowestTiming = timings[0]
		fastestTiming = timings[0]

		for i := 1; i < n; i++ {
			timing := timings[i]

			totalTiming += timing

			// the following conditions are counter-intuitive,
			// but a bigger timing means a slower execution and
			// a tinier timing means a faster execution
			// this is the reason I prefer to measure speed
			// instead of time :)
			if timing > slowestTiming {
				slowestTiming = timing
			}

			if timing < fastestTiming {
				fastestTiming = timing
			}
		}

		averageTiming = totalTiming / float64(n)
		requestsPerSecond = float64(1) / averageTiming
	}

	fmt.Fprintf(dumpBuilder, "%s\n", fmt.Sprintf("Total:        %12.4f secs", totalTiming))
	fmt.Fprintf(dumpBuilder, "%s\n", fmt.Sprintf("Slowest:      %12.4f secs", slowestTiming))
	fmt.Fprintf(dumpBuilder, "%s\n", fmt.Sprintf("Fastest:      %12.4f secs", fastestTiming))
	fmt.Fprintf(dumpBuilder, "%s\n", fmt.Sprintf("Average:      %12.4f secs", averageTiming))
	fmt.Fprintf(dumpBuilder, "%s\n", fmt.Sprintf("Requests/sec: %12.4f", requestsPerSecond))

	return dumpBuilder.String()
}

func main() {
	timings, err := computeTimings()

	if err != nil {
		log.Println("main: error on computing request timings. Error:", err)
		os.Exit(exitCodeError)
	}

	statuses, err := computeStatuses()

	if err != nil {
		log.Println("main: error on computing status codes. Error:", err)
		os.Exit(exitCodeError)
	}

	fmt.Println("Summary:")
	fmt.Println("")
	timingsDump := dumpTimings(timings)
	fmt.Println(timingsDump)

	fmt.Println("Status code distribution:")
	fmt.Println("")
	statusesDump := dumpStatuses(statuses)
	fmt.Println(statusesDump)

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

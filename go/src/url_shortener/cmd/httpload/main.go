package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

/*

## description

Write a load tester for web servers, that is, a program that simulates some load for programs that responds to HTTP requests.

## options:

The program should have the following command-line interface and accept the following options:

```
Usage: httpload [options...] URL

Options:

  -w int        number of workers to run concurrently. default:50.
  -n int        number of requests to run. default:200.
  -z string   duration of application to send requests. default:unlimited.
```

Where:

 - `URL` is the URL of the server to load-test
 - the total number of requests cannot be smaller than the number of concurrent workers
 - If `-z` is given then `-n` is ignored.
 - if `-z` is given then the application stops and exits after the specified duration.
 - example of valid duration: `-z 10m` , `-z 3s`

## final report

When the program exits (either by itself if `-z` was given) or in response of a CTRL-C, the program should dump on standard output the following information:

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
Write one unit test that verifies that the program respects the specified number of requests (`-n`)

## workflow

It's ok and advised to split the work into multiple atomic, self-contained PR.

## hints:

 - https://golang.org/pkg/time/#ParseDuration
 - https://blog.golang.org/pipelines
 - https://blog.golang.org/context
 - https://golang.org/pkg/net/http/httptest/#NewServer for testing

*/

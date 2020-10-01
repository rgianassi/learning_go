package loader

import (
	"context"
	"flag"
	"fmt"
	"time"
)

// Config manages the configuration parameters passed on the command line
type Config struct {
	nWorkers    int
	nRequests   int
	appDuration time.Duration
	url         string
}

// NewConfigFromFlags constructs a config binding flags to its fields
func NewConfigFromFlags(flags *flag.FlagSet) *Config {
	config := &Config{}
	flags.IntVar(&config.nWorkers, "w", 50, "number of concurrent workers running")
	flags.IntVar(&config.nRequests, "n", 200, "number of requests to run")
	flags.DurationVar(&config.appDuration, "z", 0, "application duration to send requests")
	return config
}

// NWorkers returns the number of workers
func (c *Config) NWorkers() int {
	return c.nWorkers
}

// AppDuration returns the duration for running the load test
func (c *Config) AppDuration() time.Duration {
	return c.appDuration
}

// CheckFlags checks if flags passed to the program are correct
func (c *Config) CheckFlags() (err error) {
	if c.nRequests < c.nWorkers {
		err = fmt.Errorf("the number of requests to run (%v) cannot be less than the number of workers (%v)", c.nRequests, c.nWorkers)
	}
	return err
}

// Parse parses arguments and saves URL in the config
func (c *Config) Parse(flags *flag.FlagSet, args []string) error {
	flags.Parse(args)

	if flags.NArg() != 1 {
		return fmt.Errorf("Wrong number of arguments")
	}

	c.url = flags.Arg(0)
	return nil
}

// RequestsSource is a source for requests, it generates URLs to be processed later
// see: https://medium.com/statuscode/pipeline-patterns-in-go-a37bb3a7e61d
func (c *Config) RequestsSource(ctx context.Context) (<-chan string, <-chan error, error) {
	theRequest := c.url

	if theRequest == "" {
		return nil, nil, fmt.Errorf("no URL provided")
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

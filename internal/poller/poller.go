// Package poller provides bounded status polling for SDK packages.
package poller

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Options configures asynchronous status polling.
type Options struct {
	MaxAttempts   int
	Interval      time.Duration
	BackoffFactor float64
	MaxInterval   time.Duration
	Timeout       time.Duration
}

// DefaultOptions returns the SDK polling defaults.
func DefaultOptions() Options {
	return Options{
		MaxAttempts:   10,
		Interval:      time.Second,
		BackoffFactor: 1.5,
		MaxInterval:   10 * time.Second,
		Timeout:       10 * time.Minute,
	}
}

// Poll calls fetch until done succeeds, failed returns an error, or the limits
// in options are reached.
func Poll[T any](ctx context.Context, options Options, fetch func(context.Context) (T, error), done func(T) bool, failed func(T) error) (T, error) {
	var zero T
	options, err := normalize(options)
	if err != nil {
		return zero, err
	}
	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	interval := options.Interval

	for attempt := 1; attempt <= options.MaxAttempts; attempt++ {
		result, err := fetch(ctx)
		if err != nil {
			return zero, err
		}
		if failed != nil {
			if err := failed(result); err != nil {
				return zero, err
			}
		}
		if done(result) {
			return result, nil
		}
		if attempt == options.MaxAttempts {
			break
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return zero, fmt.Errorf("status polling timed out: %w", ctx.Err())
		case <-timer.C:
		}
		interval = min(time.Duration(float64(interval)*options.BackoffFactor), options.MaxInterval)
	}
	return zero, fmt.Errorf("maximum polling attempts (%d) exceeded", options.MaxAttempts)
}

func normalize(options Options) (Options, error) {
	defaults := DefaultOptions()
	if options.MaxAttempts == 0 {
		options.MaxAttempts = defaults.MaxAttempts
	}
	if options.Interval == 0 {
		options.Interval = defaults.Interval
	}
	if options.BackoffFactor == 0 {
		options.BackoffFactor = defaults.BackoffFactor
	}
	if options.MaxInterval == 0 {
		options.MaxInterval = defaults.MaxInterval
	}
	if options.Timeout == 0 {
		options.Timeout = defaults.Timeout
	}
	if options.MaxAttempts < 1 || options.Interval < 0 || options.BackoffFactor < 1 || options.MaxInterval < 0 || options.Timeout <= 0 {
		return Options{}, errors.New("invalid poller options")
	}
	return options, nil
}

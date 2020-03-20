package bpool

import (
	"time"
)

// Options holds the optional BufferPool parameters.
type Options struct {
	// Maximum concurrent buffer can get from pool
	MaxPoolSize int

	// The duration for waiting in the queue if buffer pool reaches its target size
	InitialInterval time.Duration

	// RandomizationFactor sets factor to backoff when buffer pool reaches target size
	RandomizationFactor float64

	// MaxElapsedTime sets maximum elapsed time to wait during backoff
	MaxElapsedTime time.Duration

	// WriteBackOff to turn on Backoff for buffer writes
	WriteBackOff bool
}

func (src *Options) copyWithDefaults() *Options {
	opts := Options{}
	if src != nil {
		opts = *src
	}

	if opts.MaxPoolSize == 0 {
		opts.MaxPoolSize = maxPoolSize
	}

	if opts.InitialInterval == 0 {
		opts.InitialInterval = DefaultInitialInterval
	}

	if opts.RandomizationFactor == 0 {
		opts.RandomizationFactor = DefaultRandomizationFactor
	}

	if opts.MaxElapsedTime == 0 {
		opts.MaxElapsedTime = DefaultMaxElapsedTime
	}

	return &opts
}

package bpool

import (
	"time"
)

// Options holds the optional BufferPool parameters.
type Options struct {
	// Maximum concurrent buffer can get from pool
	MaxPoolSize int

	// The duration for waiting in the queue due to system memory surge operations
	InitialInterval     time.Duration
	RandomizationFactor float64
	MaxElapsedTime      time.Duration
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

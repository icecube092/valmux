package valmux

import (
	"context"
	"math"
	"time"
	"sync/atomic"
	"errors"
)

const DefaultTimeout = time.Second

var ErrMaxCount = errors.New("count exceeded")
var ErrIntegerOverflow = errors.New("integer overflow")

type ValMux struct {
	current  atomic.Uint64
	maxCount uint64
	timeout  time.Duration

	mustWait      bool
	checkInterval time.Duration
}

func New(maxCount uint64, opts ...Option) *ValMux {
	s := &ValMux{
		maxCount: maxCount,
		timeout:  DefaultTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// IncAutoDec increments value and then decrements it in background on context done.
func (v *ValMux) IncAutoDec(ctx context.Context) error {
	return v.AddAutoSub(ctx, 1)
}

// AddAutoSub adds value and then subtracts it in background on context done.
func (v *ValMux) AddAutoSub(ctx context.Context, value uint64) error {
	if err := v.AddCtx(ctx, value); err != nil {
		return err
	}

	context.AfterFunc(
		ctx, func() {
			v.Sub(value)
		},
	)

	return nil
}

// IncCtx tries to increment until context alive.
func (v *ValMux) IncCtx(ctx context.Context) error {
	return v.AddCtx(ctx, 1)
}

// AddCtx tries to add value until context alive.
func (v *ValMux) AddCtx(ctx context.Context, value uint64) error {
	return v.add(ctx, value)
}

// Inc increments value.
func (v *ValMux) Inc() error {
	return v.Add(1)
}

// Add adds value.
func (v *ValMux) Add(value uint64) error {
	return v.add(context.Background(), value)
}

// Dec decrements value.
func (v *ValMux) Dec() {
	v.Sub(1)
}

// Sub subtracts value.
func (v *ValMux) Sub(value uint64) {
	var swapped bool
	for !swapped {
		cur := v.current.Load()
		if cur < value {
			swapped = v.current.CompareAndSwap(cur, 0)
		} else {
			swapped = v.current.CompareAndSwap(cur, cur-value)
		}
	}
}

// Reset sets value to 0.
func (v *ValMux) Reset() {
	var swapped bool
	for !swapped {
		swapped = v.current.CompareAndSwap(v.current.Load(), 0)
	}
}

// Max returns maximum available value.
func (v *ValMux) Max() uint64 {
	return v.maxCount
}

// Current returns current value.
func (v *ValMux) Current() uint64 {
	return v.current.Load()
}

func (v *ValMux) add(ctx context.Context, value uint64) error {
	ctx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	var ticker *time.Ticker
	if v.mustWait {
		ticker = time.NewTicker(v.checkInterval)
		defer ticker.Stop()
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		cur := v.current.Load()

		if value > v.maxCount {
			return ErrMaxCount
		} else if math.MaxUint64-cur < value {
			if !v.mustWait {
				return ErrIntegerOverflow
			}
		} else if cur+value > v.maxCount {
			if !v.mustWait {
				return ErrMaxCount
			}
		} else {
			if swapped := v.current.CompareAndSwap(cur, cur+value); swapped {
				return nil
			}
		}

		if v.mustWait {
			<-ticker.C
		}
	}
}

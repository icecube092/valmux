package valmux

import (
	"context"
	"math"
	"time"
	"sync/atomic"
)

type Single struct {
	current  atomic.Uint64
	maxCount uint64

	mustWait      bool
	checkInterval time.Duration
}

func NewSingle(maxCount uint64, opts ...SingleOption) *Single {
	s := &Single{
		maxCount: maxCount,
		mustWait: false,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// IncAutoDec increments value and then decrements it in background on context done.
func (v *Single) IncAutoDec(ctx context.Context) error {
	return v.AddAutoSub(ctx, 1)
}

// AddAutoSub adds value and then subtracts it in background on context done.
func (v *Single) AddAutoSub(ctx context.Context, value uint64) error {
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

// IncCtx tries to increment until context alive in case of waiting mode.
// Otherwise, it returns immediately.
func (v *Single) IncCtx(ctx context.Context) error {
	return v.AddCtx(ctx, 1)
}

// AddCtx tries to add value until context alive in case of waiting mode.
// Otherwise, it returns immediately.
func (v *Single) AddCtx(ctx context.Context, value uint64) error {
	return v.add(ctx, value)
}

// Inc increments value.
func (v *Single) Inc() error {
	return v.Add(1)
}

// Add adds value.
func (v *Single) Add(value uint64) error {
	return v.add(context.Background(), value)
}

// Dec decrements value.
func (v *Single) Dec() {
	v.Sub(1)
}

// Sub subtracts value.
func (v *Single) Sub(value uint64) {
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
func (v *Single) Reset() {
	var swapped bool
	for !swapped {
		swapped = v.current.CompareAndSwap(v.current.Load(), 0)
	}
}

// Max returns maximum available value.
func (v *Single) Max() uint64 {
	return v.maxCount
}

// Current returns current value.
func (v *Single) Current() uint64 {
	return v.current.Load()
}

func (v *Single) add(ctx context.Context, value uint64) error {
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

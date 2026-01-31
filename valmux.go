package valmux

import (
	"context"
	"math"
	"time"
	"sync/atomic"
	"errors"
)

var ErrMaxCount = errors.New("count exceeded")
var ErrIntegerOverflow = errors.New("integer overflow")

// ValMux represents semaphore-like structure with max limit.
// See docs for the details.
type ValMux struct {
	current  atomic.Uint64
	maxCount atomic.Uint64
	timeout  time.Duration

	mustWait      bool
	checkInterval time.Duration
}

// New returns new instance of ValMux.
func New(maxCount uint64, opts ...Option) *ValMux {
	s := &ValMux{
		maxCount: atomic.Uint64{},
		timeout:  DefaultTimeout,
	}
	s.maxCount.Store(maxCount)

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
//
// If ctx has timeout then this timeout is used.
// If ctx has no timeout then internal timeout is used.
// If completely no timeout then Sub occurs immediately after Add.
func (v *ValMux) AddAutoSub(ctx context.Context, value uint64) error {
	cancel := func() {}
	_, externalTimeout := ctx.Deadline()
	internalTimeout := v.timeout != NoTimeout

	if !externalTimeout && internalTimeout {
		ctx, cancel = context.WithTimeout(ctx, v.timeout)
		defer cancel()
	}

	if err := v.AddCtx(ctx, value); err != nil {
		return err
	}

	if _, timeoutExists := ctx.Deadline(); !timeoutExists {
		v.Sub(value)
		return nil
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
//
// If ctx has timeout then this timeout is used.
// if ctx has no timeout then internal timeout is used.
// If completely no timeout then function waits until success.
func (v *ValMux) AddCtx(ctx context.Context, value uint64) error {
	cancel := func() {}
	_, externalTimeout := ctx.Deadline()
	internalTimeout := v.timeout != NoTimeout

	if !externalTimeout && internalTimeout {
		ctx, cancel = context.WithTimeout(ctx, v.timeout)
		defer cancel()
	}

	if err := v.add(ctx, value); err != nil {
		return err
	}

	return nil
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

func (v *ValMux) SetOpts(opts ...Option) *ValMux {
	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Max returns maximum available value.
func (v *ValMux) Max() uint64 {
	return v.maxCount.Load()
}

// Current returns current value.
func (v *ValMux) Current() uint64 {
	return v.current.Load()
}

// Timeout returns maximum timeout for -Ctx methods.
func (v *ValMux) Timeout() time.Duration {
	return v.timeout
}

// WaitingMode returns flag is instance in Waiting mode or not.
func (v *ValMux) WaitingMode() bool {
	return v.mustWait
}

func (v *ValMux) add(ctx context.Context, value uint64) error {
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
		mx := v.maxCount.Load()

		if value > mx {
			return ErrMaxCount
		} else if math.MaxUint64-cur < value {
			if !v.mustWait {
				return ErrIntegerOverflow
			}
		} else if cur+value > mx {
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

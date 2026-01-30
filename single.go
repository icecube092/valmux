package valmux

import (
	"context"
	"fmt"
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

func (v *Single) IncCtx(ctx context.Context) error {
	return v.AddCtx(ctx, 1)
}

func (v *Single) AddCtx(ctx context.Context, value uint64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("ctx.Err: %w", err)
	}
	if err := v.add(ctx, value); err != nil {
		return err
	}

	context.AfterFunc(
		ctx, func() {
			v.Sub(value)
		},
	)

	return nil
}

func (v *Single) Inc() error {
	return v.Add(1)
}

func (v *Single) Add(value uint64) error {
	return v.add(context.Background(), value)
}

func (v *Single) Dec() {
	v.Sub(1)
}

func (v *Single) Sub(value uint64) {
	if cur := v.current.Load(); cur < value {
		v.current.Add(-cur)
	} else {
		v.current.Add(-value)
	}
}

func (v *Single) Reset() {
	v.current.Add(-v.current.Load())
}

func (v *Single) Max() uint64 {
	return v.maxCount
}

func (v *Single) Current() uint64 {
	return v.current.Load()
}

func (v *Single) add(ctx context.Context, value uint64) error {
	var err error

	cur := v.current.Load()
	if value > v.maxCount {
		return ErrMaxCount
	} else if cur+value > v.maxCount {
		err = ErrMaxCount
	} else {
		if swapped := v.current.CompareAndSwap(cur, cur+value); swapped {
			return nil
		}
		return v.add(ctx, value)
	}

	if !v.mustWait {
		return err
	}

	ticker := time.NewTicker(v.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			cur = v.current.Load()
			if math.MaxUint64-cur < value {
				continue
			}
			if cur+value > v.maxCount {
				continue
			}
			if swapped := v.current.CompareAndSwap(cur, cur+value); swapped {
				return nil
			}
		}
	}
}

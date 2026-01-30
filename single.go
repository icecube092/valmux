package valmux

import (
	"context"
	"fmt"
	"math"
	"sync"
)

type Single struct {
	mux sync.RWMutex

	current  uint64
	maxCount uint64
}

func NewSingle(maxCount uint64) *Single {
	return &Single{
		mux:      sync.RWMutex{},
		current:  0,
		maxCount: maxCount,
	}
}

func (v *Single) IncCtx(ctx context.Context) error {
	return v.AddCtx(ctx, 1)
}

func (v *Single) AddCtx(ctx context.Context, value uint64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("ctx.Err: %w", err)
	}
	if err := v.Add(value); err != nil {
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
	v.mux.Lock()
	defer v.mux.Unlock()

	if math.MaxUint64-v.current < value {
		return ErrIntegerOverflow
	}
	if v.current+value > v.maxCount {
		return ErrMaxCount
	}

	v.current += value

	return nil
}

func (v *Single) Dec() {
	v.Sub(1)
}

func (v *Single) Sub(value uint64) {
	v.mux.Lock()
	defer v.mux.Unlock()

	if v.current < value {
		v.current = 0
	} else {
		v.current -= value
	}
}

func (v *Single) Reset() {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.current = 0
}

func (v *Single) Max() uint64 {
	return v.maxCount
}

func (v *Single) Current() uint64 {
	return v.current
}

package valmux

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"math"
)

var ErrMaxCount = errors.New("count exceeded")
var ErrIntegerOverflow = errors.New("integer overflow")

type Store[T comparable] struct {
	subjects map[T]uint64
	mux      sync.RWMutex

	maxCount uint64
}

func NewStore[T comparable](maxCount uint64) *Store[T] {
	return &Store[T]{
		subjects: make(map[T]uint64),
		mux:      sync.RWMutex{},
		maxCount: maxCount,
	}
}

func (v *Store[T]) IncCtx(ctx context.Context, subject T) error {
	return v.AddCtx(ctx, subject, 1)
}

func (v *Store[T]) AddCtx(ctx context.Context, subject T, value uint64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("ctx.Err: %w", err)
	}
	if err := v.Add(subject, value); err != nil {
		return err
	}

	context.AfterFunc(
		ctx, func() {
			v.Sub(subject, value)
		},
	)

	return nil
}

func (v *Store[T]) Inc(subject T) error {
	return v.Add(subject, 1)
}

func (v *Store[T]) Add(subject T, value uint64) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	current := v.subjects[subject]

	if math.MaxUint64-current < value {
		return ErrIntegerOverflow
	}
	if current+value > v.maxCount {
		return ErrMaxCount
	}

	v.subjects[subject] += value

	return nil
}

func (v *Store[T]) Dec(subject T) {
	v.Sub(subject, 1)
}

func (v *Store[T]) Sub(subject T, value uint64) {
	v.mux.Lock()
	defer v.mux.Unlock()

	if v.subjects[subject] < value {
		v.subjects[subject] = 0
	} else {
		v.subjects[subject] -= value
	}

	if v.subjects[subject] == 0 {
		delete(v.subjects, subject)
	}
}

func (v *Store[T]) Reset(subject T) {
	v.mux.Lock()
	defer v.mux.Unlock()

	delete(v.subjects, subject)
}

func (v *Store[T]) ResetAll() {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.subjects = make(map[T]uint64)
}

func (v *Store[T]) Max() uint64 {
	return v.maxCount
}

func (v *Store[T]) Current(subject T) uint64 {
	v.mux.Lock()
	defer v.mux.Unlock()

	return v.subjects[subject]
}

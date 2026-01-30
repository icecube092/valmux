package suspender

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/exp/constraints"
	"math"
)

var ErrCountOverflow = errors.New("count overflow")

type ValMux[T constraints.Ordered] interface {
	Inc(value T) error
	IncCtx(ctx context.Context, value T) error
	Dec(value T) error
}

type Config struct {
	Count uint64
}

type valMux[T constraints.Ordered] struct {
	m   map[T]uint64
	mux sync.RWMutex

	count uint64
}

const defaultCount uint64 = 1

func NewWithCfg[T constraints.Ordered](cfg Config) *valMux[T] {
	s := New[T]()

	if cfg.Count > defaultCount {
		s.count = cfg.Count
	}

	return s
}

func New[T constraints.Ordered]() *valMux[T] {
	return &valMux[T]{
		m:     make(map[T]uint64),
		mux:   sync.RWMutex{},
		count: defaultCount,
	}
}

func (s *valMux[T]) IncCtx(ctx context.Context, value T) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("ctx.Err: %w", err)
	}
	if err := s.Inc(value); err != nil {
		return err
	}

	context.AfterFunc(
		ctx, func() {
			s.Dec(value)
		},
	)

	return nil
}

func (s *valMux[T]) Inc(value T) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	v := s.m[value]
	if v >= s.count {
		return ErrCountOverflow
	}
	if v == math.MaxUint64 {
		return ErrCountOverflow
	}

	s.m[value]++

	return nil
}

func (s *valMux[T]) Dec(value T) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.m[value] > 0 {
		s.m[value]--
	}

	return nil
}

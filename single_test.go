package valmux

import (
	"context"
	"testing"
	"time"
	"errors"
)

func TestSingle(t *testing.T) {
	tests := []struct {
		name   string
		valmux *Single
		do     func(v *Single)
	}{
		{
			name:   "Inc",
			valmux: NewSingle(1),
			do: func(v *Single) {
				var err error
				if v.Max() != 1 {
					t.Fatal(err)
				}

				err = v.Inc()
				if err != nil {
					t.Fatal(err)
				}
				err = v.Inc()
				if err == nil {
					t.Fatal(err)
				}
				if v.Current() != 1 {
					t.Fatal(err)
				}

				v.Dec()
				v.Dec()
				if v.Current() != 0 {
					t.Fatal(err)
				}

				err = v.Inc()
				if err != nil {
					t.Fatal(err)
				}

				v.Reset()
				if v.Current() != 0 {
					t.Fatal(err)
				}
			},
		},
		{
			name:   "IncCtx",
			valmux: NewSingle(2),
			do: func(v *Single) {
				var err error
				if v.Max() != 2 {
					t.Fatal(err)
				}

				ctx, cancel := context.WithTimeout(
					context.Background(),
					10*time.Millisecond,
				)
				defer cancel()

				err = v.IncCtx(ctx)
				if err != nil {
					t.Fatal(err)
				}
				err = v.IncCtx(ctx)
				if err != nil {
					t.Fatal(err)
				}
				v.Dec()

				time.Sleep(100 * time.Millisecond)

				if v.Current() != 0 {
					t.Fatal(err)
				}
			},
		},
		{
			name:   "IncCtx WithWaiting timeout exceeded",
			valmux: NewSingle(1, WithWaiting(time.Millisecond)),
			do: func(v *Single) {
				const timeout = 10 * time.Millisecond
				var err error
				if v.Max() != 1 {
					t.Fatal(err)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				err = v.IncCtx(ctx)
				if err != nil {
					t.Fatal(err)
				}
				err = v.IncCtx(ctx)
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}

				time.Sleep(timeout) // needs some time after context done

				if v.Current() != 0 {
					t.Fatal()
				}
			},
		},
		{
			name:   "AddCtx WithWaiting",
			valmux: NewSingle(2, WithWaiting(time.Millisecond)),
			do: func(v *Single) {
				const timeout = 10 * time.Millisecond
				var err error
				if v.Max() != 2 {
					t.Fatal(err)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				biggerCtx, cancel2 := context.WithTimeout(context.Background(), timeout*2)
				defer cancel2()

				err = v.AddCtx(ctx, 1)
				if err != nil {
					t.Fatal(err)
					return
				}

				waitCh := make(chan struct{})
				go func() {
					err = v.AddCtx(biggerCtx, 2)
					if err != nil {
						t.Error(err)
					}
					waitCh <- struct{}{}
				}()

				<-waitCh

				if v.Current() != 2 {
					t.Fatal(err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				test.do(test.valmux)
			},
		)
	}
}

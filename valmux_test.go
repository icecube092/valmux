package valmux

import (
	"context"
	"testing"
	"time"
	"errors"
)

func TestValMux(t *testing.T) {
	tests := []struct {
		name   string
		valmux *ValMux
		do     func(v *ValMux)
	}{
		{
			name:   "Inc",
			valmux: New(1),
			do: func(v *ValMux) {
				var err error
				if v.Max() != 1 {
					t.Error(err)
				}

				err = v.Inc()
				if err != nil {
					t.Error(err)
				}
				err = v.Inc()
				if err == nil {
					t.Error(err)
				}
				if v.Current() != 1 {
					t.Error(err)
				}

				v.Dec()
				v.Dec()
				if v.Current() != 0 {
					t.Error(err)
				}

				err = v.Inc()
				if err != nil {
					t.Error(err)
				}

				v.Reset()
				if v.Current() != 0 {
					t.Error(err)
				}
			},
		},
		{
			name:   "Inc is not affected by timeout",
			valmux: New(1, WithTimeout(time.Nanosecond)),
			do: func(v *ValMux) {
				var err error
				if v.Max() != 1 {
					t.Error(err)
				}

				err = v.Inc()
				if err != nil {
					t.Error(err)
				}
				err = v.Inc()
				if err == nil || errors.Is(err, context.DeadlineExceeded) {
					t.Error()
				}
				if v.Current() != 1 {
					t.Error(v.Current())
				}
			},
		},
		{
			name:   "IncAutoDec",
			valmux: New(2),
			do: func(v *ValMux) {
				var err error
				if v.Max() != 2 {
					t.Error(err)
				}

				ctx, cancel := context.WithTimeout(
					context.Background(),
					10*time.Millisecond,
				)
				defer cancel()

				err = v.IncAutoDec(ctx)
				if err != nil {
					t.Error(err)
				}
				err = v.IncAutoDec(ctx)
				if err != nil {
					t.Error(err)
				}
				v.Dec()

				time.Sleep(100 * time.Millisecond)

				if v.Current() != 0 {
					t.Error(v.Current())
				}
			},
		},
		{
			name:   "IncAutoDec WithWaiting timeout exceeded",
			valmux: New(1, WithWaiting(time.Millisecond)),
			do: func(v *ValMux) {
				const timeout = 10 * time.Millisecond
				var err error
				if v.Max() != 1 {
					t.Error(err)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				err = v.IncAutoDec(ctx)
				if err != nil {
					t.Error(err)
				}
				err = v.IncAutoDec(ctx)
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Error()
				}

				time.Sleep(timeout) // needs some time after context done

				if c := v.Current(); c != 0 {
					t.Error(v.Current())
				}
			},
		},
		{
			name:   "AddAutoSub WithWaiting",
			valmux: New(2, WithWaiting(time.Millisecond)),
			do: func(v *ValMux) {
				const timeout = 10 * time.Millisecond
				var err error
				if v.Max() != 2 {
					t.Error(err)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				biggerCtx, cancel2 := context.WithTimeout(context.Background(), timeout*2)
				defer cancel2()

				err = v.AddAutoSub(ctx, 1)
				if err != nil {
					t.Error(err)
					return
				}

				waitCh := make(chan struct{})
				go func() {
					err = v.AddAutoSub(biggerCtx, 2)
					if err != nil {
						t.Error(err)
					}
					waitCh <- struct{}{}
				}()

				<-waitCh

				if v.Current() != 2 {
					t.Error(v.Current())
				}
			},
		},
		{
			name: "AddAutoSub WithWaiting WithTimeout",
			valmux: New(
				2,
				WithWaiting(time.Millisecond),
				WithTimeout(10*time.Millisecond),
			),
			do: func(v *ValMux) {
				var err error
				if v.Max() != 2 {
					t.Error(err)
				}

				ctx, cancel := context.WithTimeout(
					context.Background(),
					10*time.Millisecond,
				)
				defer cancel()
				lessCtx, cancel2 := context.WithTimeout(
					context.Background(),
					10*time.Millisecond,
				)
				defer cancel2()

				waitCh := make(chan struct{}, 1)
				go func() {
					waitCh <- struct{}{}
					<-waitCh

					err = v.AddAutoSub(lessCtx, 2)
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Error()
					}
				}()

				<-waitCh
				err = v.AddAutoSub(ctx, 1)
				if err != nil {
					t.Error(err)
					return
				}
				waitCh <- struct{}{}

				time.Sleep(10 * time.Millisecond * 2) // needs a little more time than timeout to be subtracted

				if v.Current() != 0 {
					t.Error(v.Current())
				}
			},
		},
		{
			name:   "AddAutoSub WithTimeout",
			valmux: New(2, WithTimeout(10*time.Millisecond)),
			do: func(v *ValMux) {
				const timeout = 10 * time.Millisecond
				var err error
				if v.Max() != 2 {
					t.Error(err)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				err = v.AddAutoSub(ctx, 1)
				if err != nil {
					t.Error(err)
				}

				time.Sleep(timeout * 2) // needs a little more time than timeout to be subtracted

				if v.Current() != 0 {
					t.Error(v.Current())
				}
			},
		},
		{
			name:   "AddAutoSub NoTimeout with timeoutless context",
			valmux: New(1, WithNoTimeout()),
			do: func(v *ValMux) {
				if v.Max() != 1 {
					t.Error()
				}

				var err error

				now := time.Now()

				ctx := context.Background()
				err = v.AddAutoSub(ctx, 1)
				if err != nil {
					t.Error(err)
				}
				err = v.AddAutoSub(ctx, 1)
				if err != nil {
					t.Error(err)
				}

				if time.Since(now) > 10*time.Millisecond {
					t.Error()
				}

				if v.Current() != 0 {
					t.Error(v.Current())
				}
			},
		},
		{
			name:   "AddAutoSub WithWaiting applies external timeout",
			valmux: New(1, WithWaiting(time.Millisecond)),
			do: func(v *ValMux) {
				if v.Max() != 1 {
					t.Error()
				}

				var err error
				firstTimeout := 100 * time.Millisecond
				secondTimeout := 110 * time.Millisecond

				fctx, cancel := context.WithTimeout(
					context.Background(),
					firstTimeout,
				)
				defer cancel()
				sctx, cancel2 := context.WithTimeout(
					context.Background(),
					secondTimeout,
				)
				defer cancel2()

				err = v.AddAutoSub(fctx, 1)
				if err != nil {
					t.Error(err)
				}

				now := time.Now()

				err = v.AddAutoSub(sctx, 1)
				if err != nil {
					t.Error(err)
				}

				if time.Since(now) > secondTimeout {
					t.Error()
				}
			},
		},
		{
			name: "AddAutoSub WithWaiting applies internal timeout",
			valmux: New(
				1,
				WithWaiting(time.Millisecond),
				WithTimeout(10*time.Millisecond),
			),
			do: func(v *ValMux) {
				if v.Max() != 1 {
					t.Error()
				}

				var err error

				ctx := context.Background()

				err = v.AddAutoSub(ctx, 1)
				if err != nil {
					t.Error(err)
				}

				now := time.Now()

				err = v.AddAutoSub(ctx, 1)
				if err != nil {
					t.Error(err)
				}

				if time.Since(now) > v.Timeout() {
					t.Error()
				}

				time.Sleep(v.Timeout())

				if v.Current() != 0 {
					t.Error(v.Current())
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

package valmux

import (
	"context"
	"testing"
	"time"
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
					t.Fail()
				}

				err = v.Inc()
				if err != nil {
					t.Fail()
				}
				err = v.Inc()
				if err == nil {
					t.Fail()
				}
				if v.Current() != 1 {
					t.Fail()
				}

				v.Dec()
				v.Dec()
				if v.Current() != 0 {
					t.Fail()
				}

				err = v.Inc()
				if err != nil {
					t.Fail()
				}

				v.Reset()
				if v.Current() != 0 {
					t.Fail()
				}
			},
		},
		{
			name:   "IncCtx",
			valmux: NewSingle(2),
			do: func(v *Single) {
				var err error
				if v.Max() != 2 {
					t.Fail()
				}

				ctx, cancel := context.WithTimeout(
					context.Background(),
					10*time.Millisecond,
				)
				defer cancel()

				err = v.IncCtx(ctx)
				if err != nil {
					t.Fail()
				}
				err = v.IncCtx(ctx)
				if err != nil {
					t.Fail()
				}
				v.Dec()

				time.Sleep(100 * time.Millisecond)

				if v.Current() != 0 {
					t.Fail()
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

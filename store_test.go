package valmux

import (
	"context"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	tests := []struct {
		name   string
		valmux *Store[string]
		do     func(v *Store[string])
	}{
		{
			name:   "Inc",
			valmux: NewStore[string](1),
			do: func(v *Store[string]) {
				var err error
				if v.Max() != 1 {
					t.Fail()
				}

				err = v.Inc("s")
				if err != nil {
					t.Fail()
				}
				err = v.Inc("s")
				if err == nil {
					t.Fail()
				}
				if v.Current("s") != 1 {
					t.Fail()
				}

				v.Dec("s")
				v.Dec("s")
				if v.Current("s") != 0 {
					t.Fail()
				}

				err = v.Inc("s")
				if err != nil {
					t.Fail()
				}

				v.Reset("s")
				if v.Current("s") != 0 {
					t.Fail()
				}
			},
		},
		{
			name:   "IncAutoDec",
			valmux: NewStore[string](2),
			do: func(v *Store[string]) {
				var err error
				if v.Max() != 2 {
					t.Fail()
				}

				ctx, cancel := context.WithTimeout(
					context.Background(),
					10*time.Millisecond,
				)
				defer cancel()

				err = v.IncCtx(ctx, "s")
				if err != nil {
					t.Fail()
				}
				err = v.IncCtx(ctx, "s")
				if err != nil {
					t.Fail()
				}
				v.Dec("s")

				time.Sleep(100 * time.Millisecond)

				if v.Current("s") != 0 {
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

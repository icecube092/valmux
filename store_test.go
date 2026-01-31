package valmux

import (
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	tests := []struct {
		name  string
		store *Store[string]
		do    func(s *Store[string])
	}{
		{
			name:  "Default",
			store: NewStore[string](),
			do: func(s *Store[string]) {
				if len(s.current) != 0 {
					t.Error()
				}

				vm := s.Get("1")
				if len(s.current) != 1 {
					t.Error()
				}
				if vm.Max() != DefaultMax {
					t.Error()
				}
				if vm.Timeout() != DefaultTimeout {
					t.Error()
				}

				s.Drop("1")
				if len(s.current) != 0 {
					t.Error()
				}
			},
		},
		{
			name: "Override Opts",
			store: NewStore[string](
				WithNoTimeout(),
				WithWaiting(time.Millisecond),
				WithMax(2),
			),
			do: func(s *Store[string]) {
				if len(s.current) != 0 {
					t.Error()
				}

				vm := s.Get("1")
				if len(s.current) != 1 {
					t.Error()
				}
				if vm.Max() != 2 {
					t.Error()
				}
				if vm.Timeout() != NoTimeout {
					t.Error()
				}
				if !vm.WaitingMode() {
					t.Error()
				}

				vm.SetOpts(
					WithTimeout(time.Millisecond),
					WithMax(3),
					WithNoWaiting(),
				)

				vm = s.Get("1")
				if vm.Max() != 3 {
					t.Error()
				}
				if vm.Timeout() != time.Millisecond {
					t.Error()
				}
				if vm.WaitingMode() {
					t.Error()
				}
			},
		},
		{
			name:  "GetAll",
			store: NewStore[string](),
			do: func(s *Store[string]) {
				if len(s.current) != 0 {
					t.Error()
				}

				s.Get("1")
				s.Get("2")

				vms := s.GetAll()
				if len(vms) != 2 {
					t.Error()
				}

				s.Clear()
				vms = s.GetAll()
				if len(vms) != 0 {
					t.Error()
				}
			},
		},
		{
			name:  "Set",
			store: NewStore[string](WithNoTimeout()),
			do: func(s *Store[string]) {
				if len(s.current) != 0 {
					t.Error()
				}

				vm := s.Get("1")
				if len(s.current) != 1 {
					t.Error()
				}
				if vm.Max() != DefaultMax {
					t.Error()
				}
				if vm.Timeout() != NoTimeout {
					t.Error()
				}

				vm = New(1, WithTimeout(time.Millisecond))
				s.Set("1", vm)

				vm = s.Get("1")
				if vm.Timeout() != time.Millisecond {
					t.Error()
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				test.do(test.store)
			},
		)
	}
}

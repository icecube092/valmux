package valmux

import "time"

type SingleOption func(v *Single)

// WithWaiting force Single to wait until it become possible to add value.
// checkInterval specifies how often it needs to check counter during the waiting.
func WithWaiting(checkInterval time.Duration) SingleOption {
	return func(v *Single) {
		v.mustWait = true
		v.checkInterval = checkInterval
	}
}

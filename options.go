package valmux

import "time"

type Option func(v *ValMux)

// WithWaiting force ValMux to wait until it become possible to add value.
// checkInterval specifies how often it needs to check counter during the waiting.
func WithWaiting(checkInterval time.Duration) Option {
	return func(v *ValMux) {
		v.mustWait = true
		v.checkInterval = checkInterval
	}
}

// WithTimeout specifies maximum timeout on waiting for increment.
func WithTimeout(timeout time.Duration) Option {
	return func(v *ValMux) {
		v.timeout = timeout
	}
}

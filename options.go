package valmux

import "time"

const (
	// DefaultTimeout is a default timeout for waiting operations.
	DefaultTimeout = time.Second
	// NoTimeout specifies there is no timeout for waiting operations. Use with care - possible memory leaks.
	NoTimeout time.Duration = -1
)

// DefaultMax is a default maximum value for ValMux.
const DefaultMax = 1

type Option func(v *ValMux)

// WithWaiting force ValMux to wait until it become possible to add value.
// checkInterval specifies how often it needs to check counter during the waiting.
func WithWaiting(checkInterval time.Duration) Option {
	return func(v *ValMux) {
		v.mustWait = true
		v.checkInterval = checkInterval
	}
}

// WithNoWaiting removed waiting mode.
func WithNoWaiting() Option {
	return func(v *ValMux) {
		v.mustWait = false
	}
}

// WithTimeout specifies maximum timeout on waiting for increment.
// It also influences on background Sub time for Auto-methods.
func WithTimeout(timeout time.Duration) Option {
	return func(v *ValMux) {
		v.timeout = timeout
	}
}

// WithNoTimeout removes maximum timeout for waiting operations.
func WithNoTimeout() Option {
	return func(v *ValMux) {
		v.timeout = NoTimeout
	}
}

// WithMax specifies new limit for ValMux.
func WithMax(maxCount uint64) Option {
	return func(v *ValMux) {
		v.maxCount.Store(maxCount)
	}
}

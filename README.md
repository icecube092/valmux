# Value-based mutex library

[![Tests](https://github.com/icecube092/valmux/actions/workflows/tests.yml/badge.svg)](https://github.com/icecube092/valmux/actions/workflows/tests.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/icecube092/valmux.svg)](https://pkg.go.dev/github.com/icecube092/valmux)

Semaphore-like structure with maximum counter limitation and non-blocking decrement
operations.
Might wait or not on increment depends on configuration.

Operations:
- Increment, Add 
  - With waiting for unlock
  - With auto-decrement
- Decrement, Sub
- Reset

### Simple

```go
s := New(1)

err := s.Inc() // nil
err = s.Inc() // error: already locked

s.Dec()
err = s.Inc() // nil

```

### With context

```go
s := New(1)

ctx, cancel := context.WithCancel(context.Background())
err := s.IncCtx(ctx) // nil
err = s.IncCtx(ctx) // error: already locked

cancel()

err = s.IncCtx(ctx) // error: already locked
```

### With auto-decrement

```go
s := New(1)

ctx, cancel := context.WithCancel(context.Background())
err := s.IncAutoDec(ctx) // nil
err = s.IncAutoDec(ctx) // error: already locked

cancel()

err = s.IncAutoDec(ctx) // nil
```

### Store

```go
s := NewStore()

key := "1"
key2 := "2"

err := s.Inc(key) // nil
err = s.Inc(key)  // error: already locked
err = s.Inc(key2) // nil

s.Dec()
err = s.Inc() // nil

```
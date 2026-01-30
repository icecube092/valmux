# Value-based mutex library

## Example

### Single

```go
s := NewSingle(1)

err := s.Inc() // nil
err = s.Inc() // error: already locked

err = s.Inc() // nil

s.Dec()
err = s.Inc() // nil
```

### Single with context

#### Be careful with timeoutless contexts: possible memory leak

```go
s := NewSingle(1)

ctx, cancel := context.WithCancel(context.Background())
err := s.IncCtx(ctx) // nil
err = s.IncCtx(ctx) // error: already locked

cancel()

err = s.IncCtx(ctx) // nil
```
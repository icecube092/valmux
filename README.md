# Value-based mutex library

### Single

```go
s := NewSingle(1)

err := s.Inc() // nil
err = s.Inc() // error: already locked

err = s.Inc() // nil

s.Dec()
err = s.Inc() // nil
```

#### Single with context

```go
s := NewSingle(1)

ctx, cancel := context.WithCancel(context.Background())
err := s.IncCtx(ctx) // nil
err = s.IncCtx(ctx) // error: already locked

cancel()

err = s.IncCtx(ctx) // nil
```

### Store

```go
s := NewStore(1)

id := "1"
id2 := "2"

err := s.Inc(id) // nil
err = s.Inc(id) // error: already locked
err = s.Inc(id2) // nil

s.Dec(id)
err = s.Inc(id) // nil
```

#### Store with context

```go
s := NewSingle(1)

id := "1"

ctx, cancel := context.WithCancel(context.Background())
err := s.IncCtx(ctx, id) // nil
err = s.IncCtx(ctx, id) // error: already locked

cancel()

err = s.IncCtx(ctx, id) // nil
```

### Possible problems

- Be careful with long-live contexts on using -Ctx methods: possible memory leak in case
  when context lives longer when ValMux
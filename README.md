# Weighted sliding window log rate limiter

A thread-safe rate limiter library for Golang inspired but unsatisfied by https://github.com/Narasimha1997/ratelimiter and https://github.com/uber-go/ratelimit

This library can be used in your codebase to rate-limit any API. 

### Installation:
The package can be installed as a Go 1.22 module.

```
go get github.com/filinvadim/ratelimiter
```

### Using the library:
There are two types of rate limiters used.

### Examples: 
The generic rate limiter:

```go
    l := NewLimiter(8, time.Second * 5)
    defer l.Close()
	
    for i := 0; i < 10; i++ {
        l.Limit(12, func() {
            // some API request with weight (if requests are differentiated)
        })
```
Keyed rate limiter
```go
    l := NewKeyLimiter()
    defer l.DeleteKeys() 
    l.RegisterKey("test", uint32(limit), window)
    // or l.DeleteKeys("test")
    for i := 0; i < 10; i++ {
        l.LimitKey("test", 6, func() {
            // some 'test' keyed API request with weight (if requests are differentiated)
        })
    }
```

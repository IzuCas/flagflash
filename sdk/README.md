# FlagFlash Go SDK

Go SDK for the [FlagFlash](https://github.com/IzuCas/flagflash) feature flag platform.

## Installation

```bash
go get github.com/IzuCas/flagflash/sdk
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"

    sdk "github.com/IzuCas/flagflash/sdk"
)

func main() {
    // Create the client
    client := sdk.New(
        "ff_your_api_key_here",
        "http://localhost:9001",
        sdk.WithTimeout(3*time.Second),
    )

    ctx := context.Background()

    // Connect: bootstraps cache + opens WebSocket
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("connect error: %v", err)
    }
    defer client.Close()

    // Quick boolean check (zero network latency after Connect)
    if client.IsEnabled(ctx, "dark_mode") {
        // Feature is enabled
    }
}
```

## Features

- **Real-time updates**: WebSocket connection keeps flags in sync
- **Local caching**: Zero-latency flag evaluation after initial connection
- **Targeting support**: Server-side evaluation with user context
- **Type-safe values**: Methods for bool, string, float64, int, and JSON

## API Reference

### Creating a Client

```go
client := sdk.New(apiKey, serverURL, options...)
```

Options:
- `sdk.WithTimeout(d time.Duration)` - HTTP request timeout
- `sdk.WithHTTPClient(hc *http.Client)` - Custom HTTP client

### Connecting

```go
err := client.Connect(ctx)
defer client.Close()
```

### Evaluating Flags

```go
// Simple boolean check
if client.IsEnabled(ctx, "my-feature") {
    // feature is on
}

// Full evaluation result
result, err := client.Evaluate(ctx, "my-feature", nil)
fmt.Println(result.Enabled, result.Value)

// With targeting context (server-side evaluation)
result, err := client.Evaluate(ctx, "my-feature", sdk.EvaluationContext{
    "user_id": "usr-123",
    "plan":    "pro",
})

// Evaluate all flags
allFlags, err := client.EvaluateAll(ctx, nil)
for key, result := range allFlags {
    fmt.Printf("%s: %v\n", key, result.Enabled)
}

// Get raw flag descriptors
flags, err := client.GetFlags(ctx)
```

### Typed Value Access

```go
result, _ := client.Evaluate(ctx, "my-flag", nil)

// Get typed values with defaults
boolVal := result.BoolValue(false)
stringVal := result.StringValue("default")
floatVal := result.Float64Value(0.0)
intVal := result.IntValue(0)

// Unmarshal JSON value
var config MyConfig
err := result.JSONValue(&config)
```

## Examples

See the [example/](example/) directory for complete usage examples.

## License

MIT

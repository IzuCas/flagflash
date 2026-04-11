// Package sdk provides a Go SDK for the FlagFlash feature flag platform.
//
// The SDK connects to the server via WebSocket and maintains a local in-memory
// cache of all feature flags. Flag evaluations are served from the cache with
// no network round-trip; the cache is kept up-to-date by the WebSocket stream.
//
// # Quick Start
//
//	client := sdk.New("ff_live_xxxxx", "http://your-flagflash-server")
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	if client.IsEnabled(ctx, "dark-mode") {
//	    // feature is on
//	}
//
// # Evaluation Methods
//
// The SDK supports multiple ways to evaluate flags:
//
//   - [Client.IsEnabled]: Quick boolean check
//   - [Client.Evaluate]: Full result with metadata
//   - [Client.EvaluateAll]: Bulk evaluation of all flags
//   - [Client.GetFlags]: Raw flag descriptors
//
// # Caching
//
// After Connect(), all flag reads come from the local in-memory cache.
// The WebSocket goroutine keeps the cache in sync in the background.
// When targeting rules are needed (EvaluationContext provided), the SDK
// makes a server-side evaluation.
//
// # Targeting
//
// For per-user targeting, provide an EvaluationContext:
//
//	result, err := client.Evaluate(ctx, "premium-feature", sdk.EvaluationContext{
//	    "user_id": "usr-42",
//	    "plan":    "pro",
//	    "country": "BR",
//	})
package sdk

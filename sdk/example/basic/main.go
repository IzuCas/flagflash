// Package main demonstrates how to use the FlagFlash Go SDK.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	sdk "github.com/IzuCas/flagflash/sdk"
)

var client *sdk.Client

func main() {
	// ── 1. Create the client ────────────────────────────────────────
	client = sdk.New(
		"ff_94260720d275b1fc0df988ff74bdcc58f9fa396595b9df60e190b776c113d63f",
		"http://localhost:9001",
		sdk.WithTimeout(3*time.Second),
	)

	ctx := context.Background()

	// ── 2. Connect: bootstraps cache + opens WebSocket ──────────────
	//
	// After Connect() all flag reads come from the local in-memory cache.
	// The WebSocket goroutine keeps the cache in sync in the background.
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("connect error: %v", err)
	}
	defer client.Close()

	fmt.Println("SDK connected, flags loaded from cache.")

	// ── 3. Quick boolean check (zero network latency after Connect) ─
	if client.IsEnabled(ctx, "dark_mode") {
		fmt.Println("dark mode is ON")
	} else {
		fmt.Println("dark mode is OFF")
	}

	// ── 4. Full EvalResult from cache ───────────────────────────────
	//
	// Pass nil EvaluationContext → served from cache.
	result, err := client.Evaluate(ctx, "dark_mode", nil)
	if err != nil {
		log.Printf("evaluate error: %v", err)
	} else {
		fmt.Printf("[dark_mode] enabled=%v  value=%s  version=%d  fromCache=%v\n",
			result.Enabled, result.Value, result.Version, result.FromCache)
	}

	// ── 5. Targeting rules (always hits server) ─────────────────────
	//
	// Provide an EvaluationContext when you need per-user targeting.
	targeted, err := client.Evaluate(ctx, "new-checkout", sdk.EvaluationContext{
		"user_id": "usr-42",
		"plan":    "pro",
		"country": "BR",
	})
	if err != nil {
		log.Printf("targeted evaluate error: %v", err)
	} else {
		fmt.Printf("[new-checkout] enabled=%v (targeting applied)\n", targeted.Enabled)
	}

	// ── 6. All flags from cache ─────────────────────────────────────
	allFlags, err := client.EvaluateAll(ctx, nil)
	if err != nil {
		log.Fatalf("evaluate-all error: %v", err)
	}

	fmt.Println("\n── All flags (from cache) ───────────────────────────")
	for key, r := range allFlags {
		fmt.Printf("  %-30s enabled=%-5v  value=%s\n", key, r.Enabled, r.Value)
	}

	theme := allFlags.Get("ui-theme").StringValue("light")
	fmt.Printf("\nUI theme: %s\n", theme)

	// ── 7. Raw flag descriptors (always HTTP) ───────────────────────
	flags, err := client.GetFlags(ctx)
	if err != nil {
		log.Fatalf("get-flags error: %v", err)
	}

	fmt.Println("\n── Raw flags ────────────────────────────────────────")
	for _, f := range flags {
		fmt.Printf("  %-30s type=%-10s enabled=%v\n", f.Key, f.Type, f.Enabled)
	}

	// ── 8. Start HTTP server with flags endpoint ────────────────────
	http.HandleFunc("/flags", handleFlags)
	http.HandleFunc("/flags/", handleFlagByKey)

	fmt.Println("\n── HTTP Server ──────────────────────────────────────")
	fmt.Println("Endpoints:")
	fmt.Println("  GET /flags       - List all feature flags")
	fmt.Println("  GET /flags/{key} - Get a specific feature flag")
	fmt.Println("\nServer listening on :8080...")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleFlags returns all feature flags as JSON.
func handleFlags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	allFlags, err := client.EvaluateAll(ctx, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error evaluating flags: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allFlags)
}

// handleFlagByKey returns a specific feature flag by key.
func handleFlagByKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract flag key from URL path: /flags/{key}
	flagKey := r.URL.Path[len("/flags/"):]
	if flagKey == "" {
		http.Error(w, "Flag key is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	result, err := client.Evaluate(ctx, flagKey, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error evaluating flag: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"flag_key":   result.FlagKey,
		"enabled":    result.Enabled,
		"value":      result.Value,
		"version":    result.Version,
		"from_cache": result.FromCache,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

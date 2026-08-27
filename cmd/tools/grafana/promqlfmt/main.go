// Package main formats PromQL queries in bulk, mirroring `promtool promql format`.
//
// promtool formats a single query per invocation, so checking every dashboard
// expression that way costs ~2000 process spawns. This helper does the same work
// in one process: it reads ["query", ...] as JSON on stdin and writes
// [{"formatted":"..."}, ...] as JSON on stdout, in the same order.
//
// It lives in its own module so the root module does not have to vendor
// github.com/prometheus/prometheus.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/prometheus/prometheus/promql/parser"
)

type result struct {
	Formatted string `json:"formatted"`
	Err       string `json:"err,omitempty"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var queries []string
	if err := json.NewDecoder(os.Stdin).Decode(&queries); err != nil {
		return fmt.Errorf("decode queries: %w", err)
	}

	// parser.Options{} matches `promtool promql format`, which passes no
	// --enable-feature flags. Enabling any of them here would silently accept
	// expressions that real Prometheus rejects.
	p := parser.NewParser(parser.Options{})

	results := make([]result, len(queries))
	for i, q := range queries {
		expr, err := p.ParseExpr(q)
		if err != nil {
			results[i] = result{Err: err.Error()}
			continue
		}
		results[i] = result{Formatted: expr.Pretty(0)}
	}

	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		return fmt.Errorf("encode results: %w", err)
	}
	return nil
}

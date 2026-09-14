package rules

import (
	"log/slog"
)

// discardLogger returns a logger that writes nowhere, so tests do not emit the
// rule manager's operational logging.
func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

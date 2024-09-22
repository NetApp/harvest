package logging

import (
	"context"
	"errors"
	"log/slog"
	"slices"
)

// MultiHandler distributes records to multiple slog.Handler
func MultiHandler(handlers ...slog.Handler) slog.Handler {
	return &MultiWriters{
		handlers: handlers,
	}
}

type MultiWriters struct {
	handlers []slog.Handler
}

func (h *MultiWriters) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

func (h *MultiWriters) Handle(ctx context.Context, r slog.Record) error {
	var errs []error

	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err := h.handlers[i].Handle(ctx, r.Clone())
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func (h *MultiWriters) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlers[i] = h.handlers[i].WithAttrs(slices.Clone(attrs))
	}
	return MultiHandler(handlers...)
}

func (h *MultiWriters) WithGroup(name string) slog.Handler {
	// The WithGroup API says that if the name is empty, the handler should be returned as is
	if name == "" {
		return h
	}

	handlers := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlers[i] = h.handlers[i].WithGroup(name)
	}

	return MultiHandler(handlers...)
}

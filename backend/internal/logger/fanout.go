package logger

import (
	"context"
	"log/slog"
)

type FanoutHandler struct {
	handlers []slog.Handler
}

func NewFanout(handlers ...slog.Handler) *FanoutHandler {
	return &FanoutHandler{handlers: handlers}
}

func (f *FanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range f.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range f.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		// Clone is mandatory: child handlers may mutate the attrs slice and would
		// race with each other under load otherwise.
		if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (f *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	children := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		children[i] = h.WithAttrs(attrs)
	}
	return &FanoutHandler{handlers: children}
}

func (f *FanoutHandler) WithGroup(name string) slog.Handler {
	children := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		children[i] = h.WithGroup(name)
	}
	return &FanoutHandler{handlers: children}
}

// SPDX-FileCopyrightText: Copyright 2025 Krishna Iyer (www.krishnaiyer.tech)
// SPDX-License-Identifier: Apache-2.0

// Package logger adds and retrieves a logger from context.
package logger

import (
	"context"
	"io"
	"log/slog"
)

type loggerKeyType string

var loggerKey loggerKeyType = "logger"

// NewContext returns a new context with a logger.
// Call this function at the start of the program and use this as the base context.
func NewContext(w io.Writer, level slog.Level) context.Context {
	ctx := context.Background()
	logger := slog.New(
		slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: level,
		}),
	)
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves a logger from a context and panics if there isn't one.
func FromContext(ctx context.Context) *slog.Logger {
	val := ctx.Value(loggerKey)
	logger, ok := val.(*slog.Logger)
	if !ok {
		panic("No logger in context")
	}
	return logger
}

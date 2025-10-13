// SPDX-FileCopyrightText: Copyright 2025 Krishna Iyer (www.krishnaiyer.tech)
// SPDX-License-Identifier: Apache-2.0

// Package wrappers provides wrapper implementation.
package wrappers

import (
	"context"
	"log/slog"
)

// SLog warps slog.
type SLog struct {
	logger *slog.Logger
}

// New returns a new logger.
func New() Log {
	return SLog{
		logger: slog.With(),
	}
}

// Info implements log.
func (l SLog) Info(string) {

}

// Debug implements log.
func (l SLog) Debug(string) {

}

// Warn implements log.
func (l SLog) Warn(string) {}

// Error implements log.
func (l SLog) Error(string) {}

// Fatal implements log.
func (l SLog) Fatal(string) {}

// Shutdown implements log. Not used for Slog.
func (l SLog) Shutdown(context.Context) error {
	return nil
}

// SPDX-FileCopyrightText: Copyright 2025 Krishna Iyer (www.krishnaiyer.tech)
// SPDX-License-Identifier: Apache-2.0

// Package wrappers provides wrapper implementation.
package wrappers

import "context"

// Log provides logging functions.
type Log interface {
	Info(string)
	Debug(string)
	Warn(string)
	Error(string)
	Fatal(string)

	Shutdown(context.Context) error
}

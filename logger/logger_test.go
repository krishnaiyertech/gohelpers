// SPDX-FileCopyrightText: Copyright 2025 Krishna Iyer (www.krishnaiyer.tech)
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	for _, test := range []struct {
		Name  string
		Level slog.Level
	}{
		{
			Name:  "ReturnsContextWithLogger",
			Level: slog.LevelWarn,
		},
	} {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			ctx := NewContext(&bytes.Buffer{}, test.Level)

			assert.NotNil(t, ctx)

			value := ctx.Value(loggerKey)
			logger, ok := value.(*slog.Logger)
			assert.True(t, ok)
			assert.NotNil(t, logger)

			retrieved := FromContext(ctx)
			assert.Same(t, logger, retrieved)
		})
	}
}

func TestFromContextPanics(t *testing.T) {
	for _, test := range []struct {
		Name    string
		Context func() context.Context
	}{
		{
			Name: "MissingLoggerPanics",
			Context: func() context.Context {
				return context.Background()
			},
		},
		{
			Name: "WrongTypePanics",
			Context: func() context.Context {
				return context.WithValue(context.Background(), loggerKey, "unexpected")
			},
		},
	} {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			assert.PanicsWithValue(t, "No logger in context", func() {
				FromContext(test.Context())
			})
		})
	}
}

// SPDX-FileCopyrightText: Copyright 2025 Krishna Iyer (www.krishnaiyer.tech)
// SPDX-License-Identifier: Apache-2.0

// Package logger provides a highly opinionated thin logger wrapper.
package logger

import (
	"context"
	"fmt"

	"krishnaiyer.tech/golang/gohelpers/logger/wrappers"
)

type loggerKeyType string

var loggerKey loggerKeyType = "logger"

// Level is the log level.
// Default is Info.
type Level uint

// Type is the type of the underlying logger.
// Default is slog.
type Type uint

const (
	LevelError Level = 0
	LevelWarn  Level = 1
	LevelInfo  Level = 2
	LevelDebug Level = 3

	defaultLevel = LevelInfo

	TypeSLog Type = 0
	TypeZap  Type = 1

	defaultType = TypeSLog
)

// Option is a configuration option.
type Option func(*Logger)

// Tag is a key value pair optionally attached to a logger or a message.
type Tag struct {
	Key   string
	Value any
}

// Logger abstracts the underlying logger implementation.
type Logger struct {
	level Level
	typ   Type

	impl wrappers.Log
	tags []Tag
}

// New creates a new logger.
// Call `Shutdown` when done, usually in a `defer` immediately after New.
func New(ctx context.Context, opts ...Option) *Logger {
	logger := &Logger{
		level: defaultLevel,
		typ:   defaultType,
		tags:  make([]Tag, 0),
	}

	// Apply the options.
	for _, opt := range opts {
		opt(logger)
	}

	switch logger.typ {
	case TypeSLog:
	case TypeZap:
	default:
		panic(fmt.Sprintf("unreachable type: %d", logger.typ))
	}
	return logger
}

// Shutdown attempts to gracefully shut down the logger.
func (l *Logger) Shutdown(ctx context.Context) error {
	return l.impl.Shutdown(ctx)
}

// WithError configures the logger to only print error messages.
func WithError() Option {
	return Option(func(l *Logger) {
		l.level = LevelDebug
	})
}

// WithWarn configures the logger to only print warning messages.
func WithWarn() Option {
	return Option(func(l *Logger) {
		l.level = LevelDebug
	})
}

// WithDebug allows the logger to print debug messages.
func WithDebug() Option {
	return Option(func(l *Logger) {
		l.level = LevelDebug
	})
}

// WithCustomLogger adds a custom logger. This is primarily used for testing.
func WithCustomLogger(logger wrappers.Log) Option {
	return Option(func(l *Logger) {
		l.impl = logger
	})
}

// WithType sets the type of the underlying implementation.
// Undefined types are ignored and the default is used.
func WithType(typ Type) Option {
	return Option(func(l *Logger) {
		if typ <= TypeZap {
			l.typ = typ
		}
	})
}

// WithTag returns a new logger with the tag.
func (l *Logger) WithTag(key string, val any) *Logger {
	logger := &Logger{
		impl:  l.impl,
		level: l.level,
		tags:  make([]Tag, len(l.tags)),
	}
	copy(logger.tags, l.tags)
	logger.tags = append(logger.tags, Tag{
		Key:   key,
		Value: val,
	})
	return logger
}

// WithTags returns a new logger with the tags.
// The number of arguments must be even and the first value of each pair should be a string.
func (l *Logger) WithTags(args ...any) (*Logger, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("odd number of arguments")
	}
	logger := &Logger{
		impl:  l.impl,
		level: l.level,
		tags:  make([]Tag, len(l.tags)),
	}
	copy(logger.tags, l.tags)
	// Read the args in pairs and create tags.
	for i := 0; i < len(args)-1; i = i + 2 {
		key, ok := args[i].(string)
		if !ok {
			return nil, fmt.Errorf("argument %d is not a string", i)
		}
		logger.tags = append(logger.tags, Tag{
			Key:   key,
			Value: args[i+1],
		})
	}
	return logger, nil
}

// Info logs informational messages.
func (l *Logger) Info(msg string) {
	if l.level >= LevelInfo {
		l.impl.Info(msg)
	}
}

// Debug logs debugging messages.
func (l *Logger) Debug(msg string) {
	if l.level >= LevelDebug {
		l.impl.Debug(msg)
	}
}

// Warn logs warnings.
func (l *Logger) Warn(msg string) {
	if l.level >= LevelWarn {
		l.impl.Warn(msg)
	}
}

// Error logs errors.
func (l *Logger) Error(msg string) {
	// In theory since LevelError is 0, it should always be printed.
	// This check is just for consistency.
	if l.level >= LevelError {
		l.impl.Error(msg)
	}
}

// Fatal logs fatal messages.
func (l *Logger) Fatal(msg string) {
	l.impl.Fatal(msg)
}

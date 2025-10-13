// SPDX-FileCopyrightText: Copyright 2025 Krishna Iyer (www.krishnaiyer.tech)
// SPDX-License-Identifier: Apache-2.0

// Package logger provides a highly opinionated thin logger wrapper.
package logger

import (
	"context"
	"fmt"
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

// Log provides logging functions.
type Log interface {
	Info(string)
	Debug(string)
	Warn(string)
	Error(string)
	Fatal(string)

	Shutdown(context.Context) error
}

// Option is a configuration option.
type Option func(*Logger)

// Logger abstracts the underlying logger implementation.
type Logger struct {
	level Level
	typ   Type

	impl Log
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

// WithLogLevel sets the level of the logger. Undefined levels are ignored.
// Fatal messages are always logged.
func WithLogLevel(level Level) Option {
	return Option(func(l *Logger) {
		if level <= LevelDebug {
			l.level = level
		}
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
		tags:  l.tags,
	}
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
		tags:  l.tags,
	}
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

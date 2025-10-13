// SPDX-FileCopyrightText: Copyright 2025 Krishna Iyer (www.krishnaiyer.tech)
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"context"
	"errors"
	"testing"
)

type stubLog struct {
	calls        []string
	shutdownErr  error
	shutdownArgs []context.Context
}

func (s *stubLog) Info(msg string) {
	s.calls = append(s.calls, "Info:"+msg)
}

func (s *stubLog) Debug(msg string) {
	s.calls = append(s.calls, "Debug:"+msg)
}

func (s *stubLog) Warn(msg string) {
	s.calls = append(s.calls, "Warn:"+msg)
}

func (s *stubLog) Error(msg string) {
	s.calls = append(s.calls, "Error:"+msg)
}

func (s *stubLog) Fatal(msg string) {
	s.calls = append(s.calls, "Fatal:"+msg)
}

func (s *stubLog) Shutdown(ctx context.Context) error {
	s.shutdownArgs = append(s.shutdownArgs, ctx)
	return s.shutdownErr
}

func TestNewDefaults(t *testing.T) {
	ctx := context.Background()
	stub := &stubLog{}
	logger := New(ctx, WithCustomLogger(stub))
	t.Cleanup(func() {
		_ = logger.Shutdown(ctx)
	})
	if logger.level != defaultLevel {
		t.Fatalf("expected level %d, got %d", defaultLevel, logger.level)
	}
	if logger.typ != defaultType {
		t.Fatalf("expected type %d, got %d", defaultType, logger.typ)
	}
	if len(logger.tags) != 0 {
		t.Fatalf("expected no tags, got %d", len(logger.tags))
	}
}

func TestNewPanicOnInvalidType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for invalid type")
		}
	}()
	New(context.Background(), WithCustomLogger(&stubLog{}), func(l *Logger) {
		l.typ = Type(99)
	})
}

func TestWithLogLevel(t *testing.T) {
	for _, tt := range []struct {
		name         string
		level        Level
		want         Level
		originalWant Level
	}{
		{name: "ValidError", level: LevelError, want: LevelError, originalWant: defaultLevel},
		{name: "ValidDebug", level: LevelDebug, want: LevelDebug, originalWant: defaultLevel},
		{name: "InvalidHigh", level: Level(99), want: defaultLevel, originalWant: defaultLevel},
	} {
		t.Run(tt.name, func(t *testing.T) {
			logger := &Logger{level: tt.originalWant}
			WithLogLevel(tt.level)(logger)
			if logger.level != tt.want {
				t.Fatalf("expected level %d, got %d", tt.want, logger.level)
			}
		})
	}
}

func TestWithType(t *testing.T) {
	for _, tt := range []struct {
		name  string
		typ   Type
		want  Type
		start Type
	}{
		{name: "ValidZap", typ: TypeZap, want: TypeZap, start: defaultType},
		{name: "InvalidHigh", typ: Type(99), want: defaultType, start: defaultType},
	} {
		t.Run(tt.name, func(t *testing.T) {
			logger := &Logger{typ: tt.start}
			WithType(tt.typ)(logger)
			if logger.typ != tt.want {
				t.Fatalf("expected type %d, got %d", tt.want, logger.typ)
			}
		})
	}
}

func TestLoggerWithTag(t *testing.T) {
	tagKey := "key"
	tagValue := 42
	ctx := context.Background()
	stub := &stubLog{}
	base := New(ctx, WithCustomLogger(stub))
	base.tags = []Tag{{Key: "existing", Value: "tag"}}
	derived := base.WithTag(tagKey, tagValue)
	t.Cleanup(func() {
		_ = derived.Shutdown(ctx)
	})
	if len(derived.tags) != 2 {
		t.Fatalf("expected two tags, got %d", len(derived.tags))
	}
	if derived.tags[1].Key != tagKey || derived.tags[1].Value != tagValue {
		t.Fatalf("unexpected tag: %+v", derived.tags[1])
	}
	if len(base.tags) != 1 {
		t.Fatalf("base logger tags mutated: %d", len(base.tags))
	}
	if derived.impl != base.impl {
		t.Fatalf("expected impl to be shared")
	}
}

func TestLoggerWithTags(t *testing.T) {
	ctx := context.Background()
	stub := &stubLog{}
	base := New(ctx, WithCustomLogger(stub))
	base.tags = []Tag{{Key: "existing", Value: "tag"}}
	derived, err := base.WithTags("key1", "value1", "key2", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		_ = derived.Shutdown(ctx)
	})
	if len(derived.tags) != 3 {
		t.Fatalf("expected three tags, got %d", len(derived.tags))
	}
	if derived.tags[1].Key != "key1" || derived.tags[1].Value != "value1" {
		t.Fatalf("unexpected tag pair: %+v", derived.tags[1])
	}
	if derived.tags[2].Key != "key2" || derived.tags[2].Value != 123 {
		t.Fatalf("unexpected tag pair: %+v", derived.tags[2])
	}
	if len(base.tags) != 1 {
		t.Fatalf("base logger tags mutated: %d", len(base.tags))
	}
}

func TestLoggerWithTagsErrors(t *testing.T) {
	ctx := context.Background()
	base := New(ctx, WithCustomLogger(&stubLog{}))
	t.Cleanup(func() {
		_ = base.Shutdown(ctx)
	})

	if _, err := base.WithTags("onlyKey"); err == nil {
		t.Fatalf("expected error for odd arg count")
	}

	if _, err := base.WithTags(100, "value"); err == nil {
		t.Fatalf("expected error for non-string key")
	}
}

func TestLoggerLogMethods(t *testing.T) {
	for _, tt := range []struct {
		name    string
		invoke  func(*Logger, string)
		message string
		want    string
	}{
		{name: "Info", invoke: (*Logger).Info, message: "info message", want: "Info:info message"},
		{name: "Debug", invoke: (*Logger).Debug, message: "debug message", want: "Debug:debug message"},
		{name: "Warn", invoke: (*Logger).Warn, message: "warn message", want: "Warn:warn message"},
		{name: "Error", invoke: (*Logger).Error, message: "error message", want: "Error:error message"},
		{name: "Fatal", invoke: (*Logger).Fatal, message: "fatal message", want: "Fatal:fatal message"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubLog{}
			ctx := context.Background()
			logger := New(ctx, WithCustomLogger(stub))
			t.Cleanup(func() {
				_ = logger.Shutdown(ctx)
			})
			tt.invoke(logger, tt.message)
			if len(stub.calls) != 1 {
				t.Fatalf("expected 1 call, got %d", len(stub.calls))
			}
			if stub.calls[0] != tt.want {
				t.Fatalf("expected call %q, got %q", tt.want, stub.calls[0])
			}
		})
	}
}

func TestLoggerShutdown(t *testing.T) {
	expectedErr := errors.New("shutdown error")
	stub := &stubLog{shutdownErr: expectedErr}
	ctx := context.WithValue(context.Background(), loggerKey, "value")
	logger := New(ctx, WithCustomLogger(stub))
	err := logger.Shutdown(ctx)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if len(stub.shutdownArgs) != 1 {
		t.Fatalf("expected shutdown to be called once, got %d", len(stub.shutdownArgs))
	}
	if stub.shutdownArgs[0] != ctx {
		t.Fatalf("expected shutdown context to be passed through")
	}
}

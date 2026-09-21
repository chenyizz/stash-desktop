package logger

import (
	"context"
	"fmt"
	"log/slog"
)

// SlogLogger 是 LoggerImpl 的 slog 实现
type SlogLogger struct {
	logger *slog.Logger
}

var _ LoggerImpl = &SlogLogger{}

// NewSlogLogger 用给定的 Handler 创建日志实现
func NewSlogLogger(handler slog.Handler) *SlogLogger {
	return &SlogLogger{
		logger: slog.New(handler),
	}
}

func (l *SlogLogger) Progressf(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...), "type", "progress")
}

func (l *SlogLogger) Trace(args ...interface{}) {
	l.logger.Log(context.Background(), slog.LevelDebug-4, fmt.Sprint(args...))
}

func (l *SlogLogger) Tracef(format string, args ...interface{}) {
	l.logger.Log(context.Background(), slog.LevelDebug-4, fmt.Sprintf(format, args...))
}

func (l *SlogLogger) TraceFunc(fn func() (string, []interface{})) {
	format, args := fn()
	l.Tracef(format, args...)
}

func (l *SlogLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *SlogLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l *SlogLogger) DebugFunc(fn func() (string, []interface{})) {
	format, args := fn()
	l.Debugf(format, args...)
}

func (l *SlogLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *SlogLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *SlogLogger) InfoFunc(fn func() (string, []interface{})) {
	format, args := fn()
	l.Infof(format, args...)
}

func (l *SlogLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *SlogLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *SlogLogger) WarnFunc(fn func() (string, []interface{})) {
	format, args := fn()
	l.Warnf(format, args...)
}

func (l *SlogLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *SlogLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *SlogLogger) ErrorFunc(fn func() (string, []interface{})) {
	format, args := fn()
	l.Errorf(format, args...)
}

func (l *SlogLogger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
	// Fatal 不直接 os.Exit，交给调用方决定
}

func (l *SlogLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

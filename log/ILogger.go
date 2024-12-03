package log

import "log/slog"

type ILogger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)

	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})

	Close()
	DoSPrintf(level slog.Level, a []interface{})
	FormatHeader(buf *Buffer, level slog.Level, callDepth int)

	Trace(msg string, args ...any)
	Stack(msg string, args ...any)
	Dump(msg string, args ...any)
}

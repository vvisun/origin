package log

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var OpenConsole bool
var LogSize int64
var LogChannelCap int
var LogPath string
var LogLevel = LevelTrace

type Logger struct {
	SLogger  *slog.Logger
	ioWriter IoWriter
	sBuff    Buffer
}

func (logger *Logger) setLogChannel(logChannel int) (err error) {
	return logger.ioWriter.setLogChannel(logChannel)
}

// Close It's dangerous to call the method on logging
func (logger *Logger) Close() {
	logger.ioWriter.Close()
}

func (logger *Logger) Trace(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelTrace, msg, args...)
}

func (logger *Logger) Stack(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelStack, msg, args...)
}

func (logger *Logger) Dump(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelDump, msg, args...)
}

func (logger *Logger) Debug(msg string, args ...any) {

	logger.SLogger.Log(context.Background(), LevelDebug, msg, args...)
}

func (logger *Logger) Info(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelInfo, msg, args...)
}

func (logger *Logger) Warn(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelWarning, msg, args...)
}

func (logger *Logger) Error(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelError, msg, args...)
}

func (logger *Logger) Fatal(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelFatal, msg, args...)
	os.Exit(1)
}

func (logger *Logger) Debugf(msg string, args ...any) {

	logger.SLogger.Log(context.Background(), LevelDebug, msg, args...)
}

func (logger *Logger) Infof(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelInfo, msg, args...)
}

func (logger *Logger) Warnf(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelWarning, msg, args...)
}

func (logger *Logger) Errorf(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelError, msg, args...)
}

func (logger *Logger) Fatalf(msg string, args ...any) {
	logger.SLogger.Log(context.Background(), LevelFatal, msg, args...)
	os.Exit(1)
}

func (logger *Logger) DoSPrintf(level slog.Level, a []interface{}) {
	if !logger.SLogger.Enabled(context.Background(), level) {
		return
	}

	logger.SLogger.Handler().(IOriginHandler).Lock()
	defer logger.SLogger.Handler().(IOriginHandler).UnLock()

	logger.sBuff.Reset()

	logger.FormatHeader(&logger.sBuff, level, 3)

	for _, s := range a {
		logger.sBuff.AppendString(slog.AnyValue(s).String())
	}
	logger.sBuff.AppendString("\"\n")
	logger.ioWriter.Write(logger.sBuff.Bytes())
}

func (logger *Logger) FormatHeader(buf *Buffer, level slog.Level, callDepth int) {
	t := time.Now()
	var file string
	var line int

	// Release lock while getting caller info - it's expensive.
	var ok bool
	_, file, line, ok = runtime.Caller(callDepth)
	if !ok {
		file = "???"
		line = 0
	}
	file = filepath.Base(file)

	buf.AppendString("time=\"")
	buf.AppendString(t.Format("2006/01/02 15:04:05"))
	buf.AppendString("\"")
	logger.sBuff.AppendString(" level=")
	logger.sBuff.AppendString(getStrLevel(level))
	logger.sBuff.AppendString(" source=")

	buf.AppendString(file)
	buf.AppendByte(':')
	buf.AppendInt(int64(line))
	buf.AppendString(" msg=\"")
}

func (logger *Logger) STrace(a ...interface{}) {
	logger.DoSPrintf(LevelTrace, a)
}

func (logger *Logger) SDebug(a ...interface{}) {
	logger.DoSPrintf(LevelDebug, a)
}

func (logger *Logger) SInfo(a ...interface{}) {
	logger.DoSPrintf(LevelInfo, a)
}

func (logger *Logger) SWarning(a ...interface{}) {
	logger.DoSPrintf(LevelWarning, a)
}

func (logger *Logger) SError(a ...interface{}) {
	logger.DoSPrintf(LevelError, a)
}

//-----------------------------------------------------------

func NewTextLogger(level slog.Level, pathName string, filePrefix string, addSource bool, logChannelCap int) (ILogger, error) {
	var logger Logger
	logger.ioWriter.filePath = pathName
	logger.ioWriter.filePrefix = filePrefix

	logger.SLogger = slog.New(NewOriginTextHandler(level, &logger.ioWriter, addSource, defaultReplaceAttr))
	logger.setLogChannel(logChannelCap)
	err := logger.ioWriter.switchFile()
	if err != nil {
		return nil, err
	}

	return &logger, nil
}

func NewJsonLogger(level slog.Level, pathName string, filePrefix string, addSource bool, logChannelCap int) (ILogger, error) {
	var logger Logger
	logger.ioWriter.filePath = pathName
	logger.ioWriter.filePrefix = filePrefix

	logger.SLogger = slog.New(NewOriginJsonHandler(level, &logger.ioWriter, true, defaultReplaceAttr))
	logger.setLogChannel(logChannelCap)
	err := logger.ioWriter.switchFile()
	if err != nil {
		return nil, err
	}

	return &logger, nil
}

//-----------------------------------------------------------

func ErrorAttr(key string, value error) slog.Attr {
	if value == nil {
		return slog.Attr{Key: key, Value: slog.StringValue("nil")}
	}
	return slog.Attr{Key: key, Value: slog.StringValue(value.Error())}
}

func String(key, value string) slog.Attr {
	return slog.Attr{Key: key, Value: slog.StringValue(value)}
}

func Int(key string, value int) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Int64Value(int64(value))}
}

func Int64(key string, value int64) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Int64Value(value)}
}

func Int32(key string, value int32) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Int64Value(int64(value))}
}

func Int16(key string, value int16) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Int64Value(int64(value))}
}

func Int8(key string, value int8) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Int64Value(int64(value))}
}

func Uint(key string, value uint) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Uint64Value(uint64(value))}
}

func Uint64(key string, v uint64) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Uint64Value(v)}
}

func Uint32(key string, value uint32) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Uint64Value(uint64(value))}
}

func Uint16(key string, value uint16) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Uint64Value(uint64(value))}
}

func Uint8(key string, value uint8) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Uint64Value(uint64(value))}
}

func Float64(key string, v float64) slog.Attr {
	return slog.Attr{Key: key, Value: slog.Float64Value(v)}
}

func Bool(key string, v bool) slog.Attr {
	return slog.Attr{Key: key, Value: slog.BoolValue(v)}
}

func Time(key string, v time.Time) slog.Attr {
	return slog.Attr{Key: key, Value: slog.TimeValue(v)}
}

func Duration(key string, v time.Duration) slog.Attr {
	return slog.Attr{Key: key, Value: slog.DurationValue(v)}
}

func Any(key string, value any) slog.Attr {
	return slog.Attr{Key: key, Value: slog.AnyValue(value)}
}

func Group(key string, args ...any) slog.Attr {
	return slog.Group(key, args...)
}

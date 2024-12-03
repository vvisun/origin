package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

var isSetLogger bool
var gLogger = NewDefaultLogger()

type Logger struct {
	*zap.Logger
	stack bool

	OpenConsole   *bool
	LogPath       string
	FileName      string
	LogLevel      zapcore.Level
	Encoder       zapcore.Encoder
	LogConfig     *lumberjack.Logger
	sugaredLogger *zap.SugaredLogger
}

func SetLogger(logger *Logger) {
	if logger != nil && isSetLogger == false {
		gLogger = logger
		isSetLogger = true
	}
}

func GetLogger() *Logger {
	return gLogger
}

func (logger *Logger) SetEncoder(encoder zapcore.Encoder) {
	logger.Encoder = encoder
}

func GetJsonEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

func GetTxtEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogConfig() *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   "",
		MaxSize:    2048,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
	}
}

func NewDefaultLogger() *Logger {
	logger := Logger{}
	logger.Encoder = GetJsonEncoder()
	logger.LogConfig = getLogConfig()
	logger.LogConfig.LocalTime = true

	logger.Init()
	return &logger
}

func (logger *Logger) SetLogLevel(level zapcore.Level) {
	logger.LogLevel = level
}

func (logger *Logger) Enabled(zapcore.Level) bool {
	return logger.stack
}

func (logger *Logger) Init() {
	var coreList []zapcore.Core

	if logger.OpenConsole == nil || *logger.OpenConsole {
		core := zapcore.NewCore(logger.Encoder, zapcore.AddSync(os.Stdout), logger.LogLevel)
		coreList = append(coreList, core)
	}

	if logger.LogPath != "" {
		writeSyncer := zapcore.AddSync(logger.LogConfig)
		core := zapcore.NewCore(logger.Encoder, writeSyncer, logger.LogLevel)
		coreList = append(coreList, core)
	}

	core := zapcore.NewTee(coreList...)
	logger.Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(logger), zap.AddCallerSkip(1))
	logger.sugaredLogger = logger.Logger.Sugar()
}

func (logger *Logger) Debug(msg string, fields ...zap.Field) {
	logger.Logger.Debug(msg, fields...)
}

func (logger *Logger) Info(msg string, fields ...zap.Field) {
	logger.Logger.Info(msg, fields...)
}

func (logger *Logger) Warn(msg string, fields ...zap.Field) {
	logger.Logger.Warn(msg, fields...)
}

func (logger *Logger) Error(msg string, fields ...zap.Field) {
	logger.Logger.Error(msg, fields...)
}

func (logger *Logger) StackError(msg string, args ...zap.Field) {
	logger.stack = true
	logger.Logger.Log(zapcore.ErrorLevel, msg, args...)
	logger.stack = false
}

func (logger *Logger) Fatal(msg string, fields ...zap.Field) {
	gLogger.stack = true
	logger.Logger.Fatal(msg, fields...)
	gLogger.stack = false
}

func Debug(msg string, fields ...zap.Field) {
	gLogger.Logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	gLogger.Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	gLogger.Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	gLogger.Logger.Error(msg, fields...)
}

func StackError(msg string, fields ...zap.Field) {
	gLogger.stack = true
	gLogger.Logger.Error(msg, fields...)
	gLogger.stack = false
}

func Fatal(msg string, fields ...zap.Field) {
	gLogger.stack = true
	gLogger.Logger.Fatal(msg, fields...)
	gLogger.stack = false
}

func Debugf(template string, args ...any) {
	gLogger.sugaredLogger.Debugf(template, args...)
}

func Infof(template string, args ...any) {
	gLogger.sugaredLogger.Infof(template, args...)
}

func Warnf(template string, args ...any) {
	gLogger.sugaredLogger.Warnf(template, args...)
}

func Errorf(template string, args ...any) {
	gLogger.sugaredLogger.Errorf(template, args...)
}

func StackErrorf(template string, args ...any) {
	gLogger.stack = true
	gLogger.sugaredLogger.Errorf(template, args...)
	gLogger.stack = false
}

func Fatalf(template string, args ...any) {
	gLogger.sugaredLogger.Fatalf(template, args...)
}

func (logger *Logger) SDebug(args ...interface{}) {
	logger.sugaredLogger.Debugln(args...)
}

func (logger *Logger) SInfo(args ...interface{}) {
	logger.sugaredLogger.Infoln(args...)
}

func (logger *Logger) SWarn(args ...interface{}) {
	logger.sugaredLogger.Warnln(args...)
}

func (logger *Logger) SError(args ...interface{}) {
	logger.sugaredLogger.Errorln(args...)
}

func (logger *Logger) SStackError(args ...interface{}) {
	gLogger.stack = true
	logger.sugaredLogger.Errorln(args...)
	gLogger.stack = false
}

func (logger *Logger) SFatal(args ...interface{}) {
	gLogger.stack = true
	logger.sugaredLogger.Fatalln(args...)
	gLogger.stack = false
}

func SDebug(args ...interface{}) {
	gLogger.sugaredLogger.Debugln(args...)
}

func SInfo(args ...interface{}) {
	gLogger.sugaredLogger.Infoln(args...)
}

func SWarn(args ...interface{}) {
	gLogger.sugaredLogger.Warnln(args...)
}

func SError(args ...interface{}) {
	gLogger.sugaredLogger.Errorln(args...)
}

func SStackError(args ...interface{}) {
	gLogger.stack = true
	gLogger.sugaredLogger.Errorln(args...)
	gLogger.stack = false
}

func SFatal(args ...interface{}) {
	gLogger.stack = true
	gLogger.sugaredLogger.Fatalln(args...)
	gLogger.stack = false
}

func ErrorField(key string, value error) zap.Field {
	if value == nil {
		return zap.String(key, "nil")
	}
	return zap.String(key, value.Error())
}

func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

func Int32(key string, value int32) zap.Field {
	return zap.Int32(key, value)
}

func Int16(key string, value int16) zap.Field {
	return zap.Int16(key, value)
}

func Int8(key string, value int8) zap.Field {
	return zap.Int8(key, value)
}

func Uint(key string, value uint) zap.Field {
	return zap.Uint(key, value)
}

func Uint64(key string, v uint64) zap.Field {
	return zap.Uint64(key, v)
}

func Uint32(key string, value uint32) zap.Field {
	return zap.Uint32(key, value)
}

func Uint16(key string, value uint16) zap.Field {
	return zap.Uint16(key, value)
}

func Uint8(key string, value uint8) zap.Field {
	return zap.Uint8(key, value)
}

func Float64(key string, v float64) zap.Field {
	return zap.Float64(key, v)
}

func Bool(key string, v bool) zap.Field {
	return zap.Bool(key, v)
}

func Bools(key string, v []bool) zap.Field {
	return zap.Bools(key, v)
}

func Time(key string, v time.Time) zap.Field {
	return zap.Time(key, v)
}

func Duration(key string, v time.Duration) zap.Field {
	return zap.Duration(key, v)
}

func Durations(key string, v []time.Duration) zap.Field {
	return zap.Durations(key, v)
}

func Any(key string, value any) zap.Field {
	return zap.Any(key, value)
}

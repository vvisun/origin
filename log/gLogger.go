package log

var gLogger, _ = NewTextLogger(LevelDebug, "", "", true, LogChannelCap)
var isSetLogger bool

// SetLogger It's non-thread-safe
func SetLogger(logger ILogger) {
	if logger != nil && !isSetLogger {
		gLogger = logger
		isSetLogger = true
	}
}

func GetLogger() ILogger {
	return gLogger
}

func GetDefaultHandler() IOriginHandler {
	return gLogger.(*Logger).SLogger.Handler().(IOriginHandler)
}

func Trace(msg string, args ...any) {
	gLogger.Trace(msg, args...)
}

func Debug(msg string, args ...any) {
	gLogger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	gLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	gLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	gLogger.Error(msg, args...)
}

func Stack(msg string, args ...any) {
	gLogger.Stack(msg, args...)
}

func Dump(dump string, args ...any) {
	gLogger.Dump(dump, args...)
}

func Fatal(msg string, args ...any) {
	gLogger.Fatal(msg, args...)
}

func Close() {
	gLogger.Close()
}

func STrace(a ...interface{}) {
	gLogger.DoSPrintf(LevelTrace, a)
}

func SDebug(a ...interface{}) {
	gLogger.DoSPrintf(LevelDebug, a)
}

func SInfo(a ...interface{}) {
	gLogger.DoSPrintf(LevelInfo, a)
}

func SWarning(a ...interface{}) {
	gLogger.DoSPrintf(LevelWarning, a)
}

func SError(a ...interface{}) {
	gLogger.DoSPrintf(LevelError, a)
}

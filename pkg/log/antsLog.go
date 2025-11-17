package log

import "go.uber.org/zap"

type AntsLogger struct {
	baseLogger *zap.SugaredLogger
}

func (a *AntsLogger) Printf(format string, args ...any) {
	a.baseLogger.Infof(format, args...)
}

func NewAntsLogger(baseLogger *zap.SugaredLogger) *AntsLogger {
	return &AntsLogger{baseLogger: baseLogger}
}

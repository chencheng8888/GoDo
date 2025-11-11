package log

import "go.uber.org/zap"

type CronLogger struct {
	baseLogger *zap.SugaredLogger
}

func NewCronLogger(baseLogger *zap.SugaredLogger) *CronLogger {
	return &CronLogger{baseLogger: baseLogger}
}

func (c *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	c.baseLogger.Infow(msg, keysAndValues...)
}

func (c *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	newKV := append([]interface{}{"error", err}, keysAndValues...)
	c.baseLogger.Errorw(msg, newKV...)
}

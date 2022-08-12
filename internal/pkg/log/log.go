package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func Debugf(template string, args ...interface{}) { logger.Debugf(template, args...) }

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{})  { logger.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { logger.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { logger.Errorf(template, args...) }
func Fatalf(template string, args ...interface{}) { logger.Fatalf(template, args...) }
func Panicf(template string, args ...interface{}) { logger.Panicf(template, args...) }

var logger *zap.SugaredLogger

func Logger() *zap.SugaredLogger {
	return logger
}

func init() {
	SetDev(true)
}

// SetDev 会影响颜色 和 最低等级
func SetDev(logDebug bool) {
	logger = New(logDebug)
}

func New(isDev bool) *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.StampMilli)
	logger, _ := config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	return logger.Sugar()
}

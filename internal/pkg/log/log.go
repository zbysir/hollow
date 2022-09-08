package log

import (
	"bytes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func Debugf(template string, args ...interface{}) { innerLogger.Debugf(template, args...) }

// Infof uses fmt.Sprintf To log a templated message.
func Infof(template string, args ...interface{})  { innerLogger.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { innerLogger.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { innerLogger.Errorf(template, args...) }
func Fatalf(template string, args ...interface{}) { innerLogger.Fatalf(template, args...) }
func Panicf(template string, args ...interface{}) { innerLogger.Panicf(template, args...) }

var innerLogger *zap.SugaredLogger
var StdLogger *zap.SugaredLogger

func Logger() *zap.SugaredLogger {
	return innerLogger
}

func init() {
	SetDev(false)
	StdLogger = New(Options{IsDev: false})
}

// SetDev 会影响颜色 和 最低等级
func SetDev(logDebug bool) {
	innerLogger = New(Options{IsDev: logDebug, CallerSkip: 1})
}

type BuffSink struct {
	buf bytes.Buffer
}

func (b *BuffSink) Write(p []byte) (n int, err error) {
	return b.buf.Write(p)
}

func (b *BuffSink) Sync() error {
	return nil
}

func (b *BuffSink) Close() error {
	return nil
}

type Options struct {
	IsDev         bool
	To            zapcore.WriteSyncer
	DisableCaller bool
	CallerSkip    int
	Name          string
}

func New(o Options) *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	if !o.IsDev {
		config.Level.SetLevel(zap.InfoLevel)
	}

	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.StampMilli)

	var ops []zap.Option

	ops = append(ops)
	if o.CallerSkip != 0 {
		ops = append(ops, zap.AddCallerSkip(o.CallerSkip))
	}

	sink := o.To
	if sink == nil {
		var err error
		var closeOut func()
		sink, closeOut, err = zap.Open(config.OutputPaths...)
		if err != nil {
			panic(err)

		}
		errSink, _, err := zap.Open(config.ErrorOutputPaths...)
		if err != nil {
			closeOut()
			panic(err)

		}

		ops = append(ops, zap.ErrorOutput(errSink))
	}
	if !o.DisableCaller {
		ops = append(ops, zap.AddCaller())
	}
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(config.EncoderConfig), sink, config.Level), ops...)

	sugar := logger.Sugar()
	if o.Name != "" {
		sugar = sugar.Named(o.Name)
	}
	return sugar
}

package log

import (
	"bytes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/url"
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
	SetDev(true)
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
	To            zap.Sink
	DisableCaller bool
	CallerSkip    int
	Name          string
}

func New(o Options) *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	if o.To != nil {
		_ = zap.RegisterSink("custom", func(u *url.URL) (zap.Sink, error) {
			return o.To, nil
		})

		config.OutputPaths = []string{"custom://"}
	}

	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.StampMilli)
	var ops []zap.Option

	config.DisableCaller = o.DisableCaller

	if o.CallerSkip != 0 {
		ops = append(ops, zap.AddCallerSkip(o.CallerSkip))
	}
	logger, _ := config.Build(ops...)
	sugar := logger.Sugar()
	if o.Name != "" {
		sugar = sugar.Named(o.Name)
	}
	return sugar
}

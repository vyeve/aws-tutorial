package logger

import (
	"aws-tutorial/core/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"
)

const (
	// logTimeFormat represents time format in log messages
	logTimeFormat = "2006-01-02 15:04:05.99"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	aws.Logger
}

type logger struct {
	*logrus.Logger
}

func New(conf *config.Configuration) Logger {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = logTimeFormat
	customFormatter.FullTimestamp = true
	l := logger{}
	l.Logger = logrus.New()

	l.SetFormatter(customFormatter)
	if conf.LogDebug {
		l.SetLevel(logrus.DebugLevel)
	} else {
		l.SetLevel(logrus.InfoLevel)
	}
	return l
}

func (lo logger) Log(args ...interface{}) {
	lo.Logger.Debug(args...)
}

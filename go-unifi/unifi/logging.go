package unifi

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Trace(format string)
	Debug(format string)
	Info(format string)
	Error(format string)
	Warn(format string)
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

type LoggingLevel int

const (
	DisabledLevel LoggingLevel = iota
	TraceLevel
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

func NewDefaultLogger(level LoggingLevel) Logger {
	l := logrus.New()
	var logrusLevel logrus.Level
	switch level {
	case DisabledLevel:
		return &noopLogger{}
	case TraceLevel:
		logrusLevel = logrus.TraceLevel
	case DebugLevel:
		logrusLevel = logrus.DebugLevel
	case InfoLevel:
		logrusLevel = logrus.InfoLevel
	case WarnLevel:
		logrusLevel = logrus.WarnLevel
	case ErrorLevel:
		logrusLevel = logrus.ErrorLevel
	default:
		logrusLevel = logrus.InfoLevel
	}
	l.SetLevel(logrusLevel)
	l.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		FullTimestamp:          false,
		ForceColors:            true,
	})
	return &defaultLogger{l}
}

type noopLogger struct{}

func (l *noopLogger) Trace(msg string)                          {}
func (l *noopLogger) Debug(msg string)                          {}
func (l *noopLogger) Info(msg string)                           {}
func (l *noopLogger) Error(msg string)                          {}
func (l *noopLogger) Warn(msg string)                           {}
func (l *noopLogger) Tracef(format string, args ...interface{}) {}
func (l *noopLogger) Debugf(format string, args ...interface{}) {}
func (l *noopLogger) Infof(format string, args ...interface{})  {}
func (l *noopLogger) Errorf(format string, args ...interface{}) {}
func (l *noopLogger) Warnf(format string, args ...interface{})  {}

type defaultLogger struct {
	*logrus.Logger
}

func (l *defaultLogger) Trace(msg string) {
	l.Logger.Trace(msg)
}

func (l *defaultLogger) Debug(msg string) {
	l.Logger.Debug(msg)
}

func (l *defaultLogger) Info(msg string) {
	l.Logger.Info(msg)
}

func (l *defaultLogger) Error(msg string) {
	l.Logger.Error(msg)
}

func (l *defaultLogger) Warn(msg string) {
	l.Logger.Warn(msg)
}

func (l *defaultLogger) Tracef(format string, args ...interface{}) {
	l.Logger.Tracef(format, args...)
}

func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

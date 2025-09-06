package logger

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	DebugLevel   = logrus.DebugLevel
	WarningLevel = logrus.WarnLevel
	ErrorLevel   = logrus.ErrorLevel
	FatalLevel   = logrus.FatalLevel
	InfoLevel    = logrus.InfoLevel
)

type LoggerWriter interface {
	Printf(format string, args ...any)
}

type Logger interface {
	LoggerWriter
	Debug(ctx context.Context, args ...any)
	Debugf(ctx context.Context, format string, args ...any)
	Info(ctx context.Context, args ...any)
	Infof(ctx context.Context, format string, args ...any)
	Warning(ctx context.Context, args ...any)
	Warningf(ctx context.Context, format string, args ...any)
	Error(ctx context.Context, args ...any)
	Errorf(ctx context.Context, format string, args ...any)
	Fatal(ctx context.Context, args ...any)
	Fatalf(ctx context.Context, format string, args ...any)
}

func NewLoggerFactory(label, level string) LoggerFactory {
	var logLevel logrus.Level
	switch level {
	case "debug":
		logLevel = logrus.DebugLevel
	case "warning":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	case "fatal":
		logLevel = logrus.FatalLevel
	default:
		logLevel = logrus.InfoLevel
	}
	return LoggerFactory{
		level:    logLevel,
		lvString: strings.ToUpper(level),
		label:    label,
	}
}

type LoggerFactory struct {
	level    logrus.Level
	lvString string
	label    string
}

func (lf *LoggerFactory) NewLogger(module string) Logger {
	logger := &logrus.Logger{
		Out:   os.Stdout,
		Level: lf.level,
		Formatter: &formatter{
			label:  lf.label,
			level:  lf.lvString,
			module: module,
		},
	}
	return &loggerImpl{
		logger: logger,
	}
}

type loggerImpl struct {
	logger *logrus.Logger
}

func (l *loggerImpl) Debug(ctx context.Context, args ...any) {
	l.logger.WithContext(ctx).Debug(args...)
}

func (l *loggerImpl) Debugf(ctx context.Context, format string, args ...any) {
	l.logger.WithContext(ctx).Debugf(format, args...)
}

func (l *loggerImpl) Info(ctx context.Context, args ...any) {
	l.logger.WithContext(ctx).Info(args...)
}

func (l *loggerImpl) Infof(ctx context.Context, format string, args ...any) {
	l.logger.WithContext(ctx).Infof(format, args...)
}

func (l *loggerImpl) Warning(ctx context.Context, args ...any) {
	l.logger.WithContext(ctx).Warning(args...)
}

func (l *loggerImpl) Warningf(ctx context.Context, format string, args ...any) {
	l.logger.WithContext(ctx).Warningf(format, args...)
}

func (l *loggerImpl) Error(ctx context.Context, args ...any) {
	l.logger.WithContext(ctx).Error(args...)
}

func (l *loggerImpl) Errorf(ctx context.Context, format string, args ...any) {
	l.logger.WithContext(ctx).Errorf(format, args...)
}

func (l *loggerImpl) Fatal(ctx context.Context, args ...any) {
	l.logger.WithContext(ctx).Fatal(args...)
}

func (l *loggerImpl) Fatalf(ctx context.Context, format string, args ...any) {
	l.logger.WithContext(ctx).Fatalf(format, args...)
}

func (l *loggerImpl) Printf(format string, args ...any) {
	l.logger.Printf(format, args...)
}

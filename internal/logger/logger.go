package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zlog zerolog.Logger
}

func New(level, format, outputFile string) (*Logger, error) {
	var output io.Writer = os.Stdout

	if outputFile != "" {
		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		output = file
	}

	if format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
	}

	logLevel := parseLevel(level)
	zlog := zerolog.New(output).Level(logLevel).With().Timestamp().Logger()

	return &Logger{zlog: zlog}, nil
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *Logger) Debug(msg string) {
	l.zlog.Debug().Msg(msg)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.zlog.Debug().Msgf(format, args...)
}

func (l *Logger) Info(msg string) {
	l.zlog.Info().Msg(msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.zlog.Info().Msgf(format, args...)
}

func (l *Logger) Warn(msg string) {
	l.zlog.Warn().Msg(msg)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zlog.Warn().Msgf(format, args...)
}

func (l *Logger) Error(msg string) {
	l.zlog.Error().Msg(msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.zlog.Error().Msgf(format, args...)
}

func (l *Logger) Fatal(msg string) {
	l.zlog.Fatal().Msg(msg)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.zlog.Fatal().Msgf(format, args...)
}

func (l *Logger) With(key string, value interface{}) *Logger {
	newLogger := l.zlog.With().Interface(key, value).Logger()
	return &Logger{zlog: newLogger}
}

func (l *Logger) WithError(err error) *Logger {
	newLogger := l.zlog.With().Err(err).Logger()
	return &Logger{zlog: newLogger}
}
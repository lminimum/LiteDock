package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// Interface defines the logging operations supported by the application.
type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

// Logger wraps zerolog to implement the Interface for application logging.
type Logger struct {
	logger *zerolog.Logger
}

var _ Interface = (*Logger)(nil)

// New creates a new Logger instance with the specified log level.
func New(level string) *Logger {
	var l zerolog.Level

	switch strings.ToLower(level) {
	case "error":
		l = zerolog.ErrorLevel
	case "warn":
		l = zerolog.WarnLevel
	case "info":
		l = zerolog.InfoLevel
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	skipFrameCount := 3
	logger := zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	return &Logger{
		logger: &logger,
	}
}

// Debug logs a message at debug level.
func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg(zerolog.DebugLevel, message, args...)
}

// Info logs a message at info level.
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(zerolog.InfoLevel, message, args...)
}

// Warn logs a message at warn level.
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(zerolog.WarnLevel, message, args...)
}

// Error logs a message at error level.
func (l *Logger) Error(message interface{}, args ...interface{}) {
	l.msg(zerolog.ErrorLevel, message, args...)
}

// Fatal logs a message at fatal level and exits the program.
func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg(zerolog.FatalLevel, message, args...)

	os.Exit(1)
}

func (l *Logger) log(level zerolog.Level, message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.WithLevel(level).Msg(message)
	} else {
		l.logger.WithLevel(level).Msgf(message, args...)
	}
}

func (l *Logger) msg(level zerolog.Level, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(level, msg.Error(), args...)
	case string:
		l.log(level, msg, args...)
	default:
		l.log(level, fmt.Sprintf("%s message %v has unknown type %v", level, message, msg), args...)
	}
}

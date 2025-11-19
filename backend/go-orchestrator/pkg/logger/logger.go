package logger

import (
	"github.com/rs/zerolog"
	"os"
)

type Interface interface {
	Debug(msg string, err error, fields ...map[string]interface{})
	Info(msg string, err error, fields ...map[string]interface{})
	Warn(msg string, err error, fields ...map[string]interface{})
	Error(msg string, err error, fields ...map[string]interface{})
	Fatal(msg string, err error, fields ...map[string]interface{})
}

type Logger struct {
	logger *zerolog.Logger
}

func New(level string) *Logger {
	var l zerolog.Level

	switch level {
	case "debug":
		l = zerolog.DebugLevel
	case "info":
		l = zerolog.InfoLevel
	case "warn":
		l = zerolog.WarnLevel
	case "error":
		l = zerolog.ErrorLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{logger: &logger}
}

func (l *Logger) Info(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Info()
	if len(fields) > 0 {
		event = event.Fields(fields[0])
	}
	if err != nil {
		event = event.Err(err)
	}
	event.Msg(msg)
}

func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Error()
	if len(fields) > 0 {
		event = event.Fields(fields[0])
	}
	if err != nil {
		event = event.Err(err)
	}
	event.Msg(msg)
}

func (l *Logger) Debug(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Debug()
	if len(fields) > 0 {
		event = event.Fields(fields[0])
	}
	if err != nil {
		event = event.Err(err)
	}
	event.Msg(msg)
}

func (l *Logger) Warn(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Warn()
	if len(fields) > 0 {
		event = event.Fields(fields[0])
	}
	if err != nil {
		event = event.Err(err)
	}
	event.Msg(msg)
}

func (l *Logger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Fatal()
	if len(fields) > 0 {
		event = event.Fields(fields[0])
	}
	if err != nil {
		event = event.Err(err)
	}
	event.Msg(msg)
}

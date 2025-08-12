package logger

import (
	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"os"
	"screenshorter/config"
)

type Logger struct {
	zerolog.Logger
}

var (
	debugOut = zerolog.ConsoleWriter{Out: os.Stdout}
	errorOut = zerolog.ConsoleWriter{Out: os.Stderr}
)

// logWriter реализует zerolog.LevelWriter и разделяет логи по уровням
type logWriter struct {
	useJSON bool // если true, пишем в JSON, иначе в консольный формат
}

// Write (используется, когда уровень не указан)
func (l logWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

// WriteLevel определяет, куда писать логи в зависимости от уровня
func (l logWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if l.useJSON {
		// JSON-режим: Error и выше → stderr, остальное → stdout
		if level > zerolog.WarnLevel {
			return os.Stderr.Write(p)
		}
		return os.Stdout.Write(p)
	}
	// Консольный режим: используем разные форматеры
	if level > zerolog.WarnLevel {
		return errorOut.Write(p)
	}
	return debugOut.Write(p)
}

// SentryHook - хук для zerolog, отправляющий ошибки в Sentry
type SentryHook struct{}

func (h SentryHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level >= zerolog.ErrorLevel { // Отправляем только ошибки и выше (Fatal, Panic)
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentryLevel(level))
			scope.SetExtra("message", msg)
			sentry.CaptureMessage(msg)
		})
	}
}

// Конвертирует уровень zerolog в уровень Sentry
func sentryLevel(level zerolog.Level) sentry.Level {
	switch level {
	case zerolog.ErrorLevel:
		return sentry.LevelError
	case zerolog.FatalLevel, zerolog.PanicLevel:
		return sentry.LevelFatal
	default:
		return sentry.LevelInfo
	}
}

func NewLogger(cfg *config.Config) *Logger {

	zerolog.SetGlobalLevel(zerolog.Level(cfg.LogLevel))

	// Создаем базовый логгер с общими настройками
	createLogger := func(useJSON bool, addSentryHook bool) *Logger {
		logger := zerolog.New(logWriter{useJSON: useJSON}).
			With().
			Timestamp().
			Logger()

		if addSentryHook {
			logger = logger.Hook(SentryHook{})
		}

		return &Logger{logger}
	}

	// Для JSON формата
	if cfg.LogFormat == "json" {
		return createLogger(true, cfg.LogTarget == "sentry")
	}

	// Для консольного вывода
	return createLogger(false, false)
}

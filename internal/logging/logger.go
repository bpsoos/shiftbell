package logging

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

type Config struct {
	Level   slog.Level
	Handler Handler
}

type Handler uint8

const (
	HandlerJSON Handler = iota
	HandlerConsole
	HandlerDiscard
)

var (
	once      sync.Once
	singleton *slog.Logger
)

func Configure(config Config) *slog.Logger {
	once.Do(func() {
		singleton = newLogger(config, os.Stdout)
		slog.SetDefault(singleton)
	})

	return singleton
}

func Default() *slog.Logger {
	if singleton == nil {
		return Configure(Config{})
	}

	return singleton
}

func newLogger(config Config, output io.Writer) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{Level: config.Level}
	switch config.Handler {
	case HandlerConsole:
		return slog.New(slog.NewTextHandler(output, handlerOptions))
	case HandlerDiscard:
		return slog.New(slog.DiscardHandler)
	default:
		return slog.New(slog.NewJSONHandler(output, handlerOptions))
	}
}

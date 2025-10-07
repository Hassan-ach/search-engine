package utils

import (
	"context"
	"log"
	"log/slog"
	"os"
)

var Log *Logger = NewMultiLogger("appLog.json")

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiLogger(fileName string) *Logger {
	txtHandler := slog.NewTextHandler(os.Stdout, nil)

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	jsonHandler := slog.NewJSONHandler(file, nil)

	logger := slog.New(&MultiHandler{[]slog.Handler{txtHandler, jsonHandler}})
	return &Logger{logger, file}
}

func (m MultiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	//
	for _, h := range m.handlers {
		if !h.Enabled(ctx, l) {
			return false
		}
	}
	return true
}

func (m MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		err := h.Handle(ctx, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{newHandlers}
}

func (m MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{newHandlers}
}

type Logger struct {
	*slog.Logger
	file *os.File
}

func (l *Logger) Network() *slog.Logger {
	return l.With("Context", "Netwok")
}

func (l *Logger) DB() *slog.Logger {
	return l.With("Context", "DB")
}

func (l *Logger) Cache() *slog.Logger {
	return l.With("Context", "Cache")
}

func (l *Logger) IO() *slog.Logger {
	return l.With("Context", "IO")
}

func (l *Logger) Parsing() *slog.Logger {
	return l.With("Context", "Parsing")
}

func (l *Logger) General() *slog.Logger {
	return l.With("Context", "General")
}

func (l *Logger) Close() {
	l.file.Close()
}

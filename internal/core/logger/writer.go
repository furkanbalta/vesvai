package logger

import (
	"io"
	"strings"
	"sync"
)

const maxWriterBuffer = 64 << 10

func (l *Logger) Writer(level Level, label string) io.Writer {
	if l == nil {
		return io.Discard
	}
	return &lineWriter{l: l, level: level, label: label}
}

type lineWriter struct {
	l     *Logger
	level Level
	label string

	mu  sync.Mutex
	buf []byte
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf = append(w.buf, p...)
	start := 0
	for i, b := range w.buf {
		if b == '\n' {
			w.emit(string(w.buf[start:i]))
			start = i + 1
		}
	}
	if start > 0 {
		rest := make([]byte, len(w.buf)-start)
		copy(rest, w.buf[start:])
		w.buf = rest
	}
	if len(w.buf) > maxWriterBuffer {
		w.emit(string(w.buf))
		w.buf = w.buf[:0]
	}
	return len(p), nil
}

func (w *lineWriter) emit(line string) {
	line = strings.TrimRight(line, "\r")
	if line == "" {
		return
	}
	w.l.log(w.level, w.label+line)
}

var _ io.Writer = (*lineWriter)(nil)

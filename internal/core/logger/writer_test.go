package logger

import (
	"strings"
	"testing"
)

type captureHandler struct {
	lines []string
}

func (c *captureHandler) Write(rec Record) error {
	c.lines = append(c.lines, rec.Message)
	return nil
}
func (c *captureHandler) Close() error { return nil }

func TestWriterForwardsCompleteLines(t *testing.T) {
	h := &captureHandler{}
	l := New(LevelDebug, h)
	w := l.Writer(LevelDebug, "srv: ")

	if _, err := w.Write([]byte("hello\nwor")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := w.Write([]byte("ld\r\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	if len(h.lines) != 2 {
		t.Fatalf("lines = %v, want 2", h.lines)
	}
	if h.lines[0] != "srv: hello" || h.lines[1] != "srv: world" {
		t.Errorf("lines = %v", h.lines)
	}
}

func TestNilLoggerWriterDiscards(t *testing.T) {
	var l *Logger
	w := l.Writer(LevelInfo, "x: ")
	if w == nil {
		t.Fatal("writer is nil")
	}
	if _, err := w.Write([]byte("ignored\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestWriterFlushesLongPartialLine(t *testing.T) {
	h := &captureHandler{}
	l := New(LevelDebug, h)
	w := l.Writer(LevelDebug, "")

	_, _ = w.Write([]byte(strings.Repeat("a", maxWriterBuffer+10)))
	if len(h.lines) != 1 {
		t.Fatalf("lines = %d, want 1 (long partial line flushed)", len(h.lines))
	}
}

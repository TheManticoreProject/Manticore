package logger_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/logger"
)

var tsLine = regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2} \d{2}h\d{2}m\d{2}s[^]]*\] `)

func TestFormatAndTimestamp(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.WithOutput(&buf), logger.WithNoColors(true))
	l.Info("hello")
	got := buf.String()
	if !tsLine.MatchString(got) {
		t.Errorf("missing/malformed timestamp prefix: %q", got)
	}
	if !strings.Contains(got, "INFO: hello\n") {
		t.Errorf("got %q, want it to contain \"INFO: hello\"", got)
	}
}

func TestPlainHasNoTimestamp(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.WithOutput(&buf), logger.WithNoColors(true))
	l.WithoutTimestamp().Info("x")
	got := buf.String()
	if tsLine.MatchString(got) {
		t.Errorf("plain logger should not emit a timestamp: %q", got)
	}
	if got != "INFO: x\n" {
		t.Errorf("got %q, want \"INFO: x\\n\"", got)
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.WithOutput(&buf), logger.WithNoColors(true), logger.WithLevel(logger.LevelWarn))
	l.Debug("d")
	l.Info("i")
	l.Warn("w")
	l.Error("e")
	got := buf.String()
	for _, no := range []string{"DEBUG", "INFO: i"} {
		if strings.Contains(got, no) {
			t.Errorf("level Warn should suppress %q; got %q", no, got)
		}
	}
	if !strings.Contains(got, "WARN: w") || !strings.Contains(got, "ERROR: e") {
		t.Errorf("WARN/ERROR should pass; got %q", got)
	}
}

func TestColorsOnAndOff(t *testing.T) {
	var on, off bytes.Buffer
	// ColorAlways forces colour even though a buffer is not a terminal.
	logger.New(logger.WithOutput(&on), logger.WithColorMode(logger.ColorAlways)).Error("boom")
	if !strings.Contains(on.String(), "\x1b[") {
		t.Errorf("colours on: expected ANSI in %q", on.String())
	}
	// noColors strips ANSI, including sequences embedded in the message.
	logger.New(logger.WithOutput(&off), logger.WithNoColors(true)).Info("a\x1b[31mred\x1b[0mb")
	if strings.Contains(off.String(), "\x1b") {
		t.Errorf("noColors: expected no ANSI in %q", off.String())
	}
	if !strings.Contains(off.String(), "INFO: aredb\n") {
		t.Errorf("noColors: message ANSI should be stripped; got %q", off.String())
	}
}

func TestColorAutoDisablesOnNonTTY(t *testing.T) {
	var buf bytes.Buffer
	// Default mode is ColorAuto; a bytes.Buffer is not a terminal, so no colour and any
	// embedded ANSI is stripped.
	logger.New(logger.WithOutput(&buf)).Error("x\x1b[31my\x1b[0m")
	if strings.Contains(buf.String(), "\x1b") {
		t.Errorf("ColorAuto on a non-terminal should not emit/keep ANSI: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "ERROR: xy\n") {
		t.Errorf("got %q, want stripped \"ERROR: xy\"", buf.String())
	}
}

func TestPrintIsRawAndUnfiltered(t *testing.T) {
	var buf bytes.Buffer
	// Even with the level raised to Error, Print must emit, and with no level tag.
	l := logger.New(logger.WithOutput(&buf), logger.WithNoColors(true), logger.WithLevel(logger.LevelError))
	l.WithoutTimestamp().Print("just text")
	if buf.String() != "just text\n" {
		t.Errorf("got %q, want \"just text\\n\"", buf.String())
	}
}

func TestPrintfFormatting(t *testing.T) {
	var buf bytes.Buffer
	logger.New(logger.WithOutput(&buf), logger.WithNoColors(true)).WithoutTimestamp().Infof("%s=%d", "x", 7)
	if buf.String() != "INFO: x=7\n" {
		t.Errorf("got %q, want \"INFO: x=7\\n\"", buf.String())
	}
}

func TestLogToFileTeesAndStrips(t *testing.T) {
	var term bytes.Buffer
	// Use the package-level facade; restore global state afterwards.
	logger.SetOutput(&term)
	logger.SetColorMode(logger.ColorAlways) // force colour even though term is a buffer
	defer func() {
		logger.CloseLogFile()
		logger.SetOutput(os.Stderr)
		logger.SetColorMode(logger.ColorAuto)
		logger.SetLevel(logger.LevelDebug)
	}()

	path := filepath.Join(t.TempDir(), "out.log")
	if err := logger.LogToFile(path); err != nil {
		t.Fatalf("LogToFile: %v", err)
	}
	logger.Plain.Error("disk\x1b[33mmsg\x1b[0m")

	// Terminal keeps colour (the error tag is coloured).
	if !strings.Contains(term.String(), "\x1b[") {
		t.Errorf("terminal output should keep colour: %q", term.String())
	}
	logger.CloseLogFile() // flush/close before reading
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if strings.Contains(string(data), "\x1b") {
		t.Errorf("file must be ANSI-free: %q", string(data))
	}
	if !strings.Contains(string(data), "ERROR: diskmsg\n") {
		t.Errorf("file content = %q, want stripped \"ERROR: diskmsg\"", string(data))
	}
}

func TestLegacyShimsStillWork(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.SetNoColors(true)
	defer func() {
		logger.SetOutput(os.Stderr)
		logger.SetNoColors(false)
	}()
	logger.InfoMicroseconds("legacy")
	got := buf.String()
	if !tsLine.MatchString(got) || !strings.Contains(got, "INFO: legacy\n") {
		t.Errorf("legacy InfoMicroseconds output unexpected: %q", got)
	}
	// microsecond precision => 6 fractional digits in the timestamp.
	if !regexp.MustCompile(`\d{2}s\.\d{6}\]`).MatchString(got) {
		t.Errorf("expected microsecond precision timestamp: %q", got)
	}
}

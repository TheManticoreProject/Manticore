// Package logger is Manticore's logging facility. It is built on log/slog: a custom
// slog.Handler renders the human-readable "[timestamp] LEVEL: message" format, adds
// colour to level tags, can strip ANSI sequences, and can tee output to a file.
//
// Following the standard-library pattern (slog.Default / slog.New), the package exposes
// convenience functions backed by a default logger holding the global configuration
// (Info, Infof, SetLevel, SetOutput, SetNoColors, LogToFile, ...), and a Logger type that
// can be constructed with New and passed around for isolated configuration.
//
// The same method set is available with and without timestamps: the package-level
// functions timestamp their output, and Plain (which shares the same configuration) does
// not — logger.Info("x") vs logger.Plain.Info("x").
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// ColorMode controls when colour (and the ANSI it implies) is emitted.
type ColorMode int

const (
	// ColorAuto emits colour only when the output is a terminal (the default).
	ColorAuto ColorMode = iota
	// ColorAlways always emits colour.
	ColorAlways
	// ColorNever never emits colour and strips ANSI sequences from the output.
	ColorNever
)

// isTerminal reports whether w is a terminal (an *os.File backed by a tty).
func isTerminal(w io.Writer) bool {
	if f, ok := w.(interface{ Fd() uintptr }); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// Level is a logging level (an alias of slog.Level so the slog ecosystem interoperates).
type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug // -4
	LevelInfo  = slog.LevelInfo  // 0
	LevelWarn  = slog.LevelWarn  // 4
	LevelError = slog.LevelError // 8
	// LevelPrint is above all others so Print output is never filtered and carries no
	// level tag.
	LevelPrint = slog.Level(12)
)

// TimePrecision selects the fractional precision of the timestamp.
type TimePrecision int

const (
	Seconds TimePrecision = iota
	Milliseconds
	Microseconds
	Nanoseconds

	// useConfigPrecision tells emit to read the precision from the configuration.
	useConfigPrecision TimePrecision = -1
)

const colorReset = "\x1b[0m"

// ansiPattern matches ANSI/VT control sequences (CSI: colour, bold, cursor moves, ...).
var ansiPattern = regexp.MustCompile("\x1b\\[[0-9;?]*[ -/]*[@-~]")

func stripANSI(s string) string {
	if !strings.Contains(s, "\x1b") {
		return s
	}
	return ansiPattern.ReplaceAllString(s, "")
}

func layoutFor(p TimePrecision) string {
	switch p {
	case Seconds:
		return "2006-01-02 15h04m05s"
	case Microseconds:
		return "2006-01-02 15h04m05s.000000"
	case Nanoseconds:
		return "2006-01-02 15h04m05s.000000000"
	default: // Milliseconds
		return "2006-01-02 15h04m05s.000"
	}
}

func tagFor(l slog.Level) string {
	switch {
	case l >= LevelPrint:
		return ""
	case l >= LevelError:
		return "ERROR"
	case l >= LevelWarn:
		return "WARN"
	case l >= LevelInfo:
		return "INFO"
	default:
		return "DEBUG"
	}
}

func colorFor(l slog.Level) string {
	switch {
	case l >= LevelPrint:
		return ""
	case l >= LevelError:
		return "\x1b[1;31m" // bold red
	case l >= LevelWarn:
		return "\x1b[33m" // yellow
	case l >= LevelInfo:
		return "\x1b[36m" // cyan
	default:
		return "\x1b[90m" // bright black / grey
	}
}

// config is the shared, mutable state of a logger: where it writes, its level, timestamp
// precision and colour policy. Multiple Logger handles (e.g. timestamped and Plain) point
// at one config so a single setter reconfigures them together.
type config struct {
	mu            sync.Mutex
	out           io.Writer
	outIsTerminal bool
	file          io.Writer
	fileClose     io.Closer
	levelVar      *slog.LevelVar
	precision     TimePrecision
	colorMode     ColorMode
	utc           bool
}

func newConfig() *config {
	c := &config{out: os.Stderr, outIsTerminal: isTerminal(os.Stderr), levelVar: &slog.LevelVar{}, precision: Microseconds, colorMode: ColorAuto}
	c.levelVar.Set(LevelDebug) // emit everything by default
	return c
}

// useColor reports whether colour should be emitted (caller holds c.mu).
func (c *config) useColor() bool {
	switch c.colorMode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default: // ColorAuto
		return c.outIsTerminal
	}
}

// emit formats and writes one log line. The level tag is coloured unless colours are
// disabled; the primary writer receives the line (ANSI-stripped when colour is off) and
// the optional file receives an always-stripped copy.
func (c *config) emit(withTimestamp bool, precOverride TimePrecision, level slog.Level, msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if level < c.levelVar.Level() {
		return
	}
	prec := c.precision
	if precOverride >= 0 {
		prec = precOverride
	}

	color := c.useColor()
	var b strings.Builder
	if withTimestamp {
		now := time.Now()
		if c.utc {
			now = now.UTC()
		}
		b.WriteByte('[')
		b.WriteString(now.Format(layoutFor(prec)))
		b.WriteString("] ")
	}
	if tag := tagFor(level); tag != "" {
		if color {
			b.WriteString(colorFor(level))
			b.WriteString(tag)
			b.WriteString(colorReset)
		} else {
			b.WriteString(tag)
		}
		b.WriteString(": ")
	}
	b.WriteString(msg)
	b.WriteByte('\n')
	line := b.String()

	if c.out != nil {
		if color {
			io.WriteString(c.out, line)
		} else {
			io.WriteString(c.out, stripANSI(line)) // strip ANSI embedded in the message too
		}
	}
	if c.file != nil {
		io.WriteString(c.file, stripANSI(line))
	}
}

// handler is the slog.Handler that renders Manticore's format via config.emit.
type handler struct {
	cfg       *config
	timestamp bool
	attrs     string
}

func (h *handler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.cfg.levelVar.Level()
}

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	msg := r.Message
	if h.attrs != "" || r.NumAttrs() > 0 {
		var sb strings.Builder
		sb.WriteString(msg)
		sb.WriteString(h.attrs)
		r.Attrs(func(a slog.Attr) bool {
			sb.WriteByte(' ')
			sb.WriteString(a.Key)
			sb.WriteByte('=')
			sb.WriteString(a.Value.String())
			return true
		})
		msg = sb.String()
	}
	h.cfg.emit(h.timestamp, useConfigPrecision, r.Level, msg)
	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	s := h.attrs
	for _, a := range attrs {
		s += " " + a.Key + "=" + a.Value.String()
	}
	return &handler{cfg: h.cfg, timestamp: h.timestamp, attrs: s}
}

func (h *handler) WithGroup(string) slog.Handler { return h }

// Logger is a logging handle. Use New to create one, or the package-level functions for
// the default logger.
type Logger struct {
	sl  *slog.Logger
	cfg *config
}

var bg = context.Background()

func (l *Logger) Debug(msg string)               { l.sl.Log(bg, LevelDebug, msg) }
func (l *Logger) Debugf(format string, a ...any) { l.sl.Log(bg, LevelDebug, fmt.Sprintf(format, a...)) }
func (l *Logger) Info(msg string)                { l.sl.Log(bg, LevelInfo, msg) }
func (l *Logger) Infof(format string, a ...any)  { l.sl.Log(bg, LevelInfo, fmt.Sprintf(format, a...)) }
func (l *Logger) Warn(msg string)                { l.sl.Log(bg, LevelWarn, msg) }
func (l *Logger) Warnf(format string, a ...any)  { l.sl.Log(bg, LevelWarn, fmt.Sprintf(format, a...)) }
func (l *Logger) Error(msg string)               { l.sl.Log(bg, LevelError, msg) }
func (l *Logger) Errorf(format string, a ...any) { l.sl.Log(bg, LevelError, fmt.Sprintf(format, a...)) }
func (l *Logger) Print(msg string)               { l.sl.Log(bg, LevelPrint, msg) }
func (l *Logger) Printf(format string, a ...any) { l.sl.Log(bg, LevelPrint, fmt.Sprintf(format, a...)) }

// WithoutTimestamp returns a logger that shares this logger's configuration but emits no
// timestamp. (Plain is the package default's no-timestamp sibling.)
func (l *Logger) WithoutTimestamp() *Logger {
	return &Logger{sl: slog.New(&handler{cfg: l.cfg, timestamp: false}), cfg: l.cfg}
}

// Slog returns the underlying *slog.Logger for interoperability with slog-based code.
func (l *Logger) Slog() *slog.Logger { return l.sl }

// Option configures a Logger created with New.
type Option func(*config, *handler)

func WithOutput(w io.Writer) Option {
	return func(c *config, _ *handler) { c.out = w; c.outIsTerminal = isTerminal(w) }
}
func WithLevel(l Level) Option    { return func(c *config, _ *handler) { c.levelVar.Set(l) } }
func WithTimestamp(b bool) Option { return func(_ *config, h *handler) { h.timestamp = b } }
func WithTimePrecision(p TimePrecision) Option {
	return func(c *config, _ *handler) { c.precision = p }
}
func WithColorMode(m ColorMode) Option { return func(c *config, _ *handler) { c.colorMode = m } }
func WithUTC(b bool) Option            { return func(c *config, _ *handler) { c.utc = b } }

// WithNoColors disables colour (ColorNever) when true, otherwise restores ColorAuto.
func WithNoColors(b bool) Option {
	return func(c *config, _ *handler) {
		if b {
			c.colorMode = ColorNever
		} else {
			c.colorMode = ColorAuto
		}
	}
}

// New creates an independent logger with its own configuration (default: stderr, all
// levels, microsecond timestamps, colours on). Pass it where isolated logging is needed.
func New(opts ...Option) *Logger {
	c := newConfig()
	h := &handler{cfg: c, timestamp: true}
	for _, o := range opts {
		o(c, h)
	}
	return &Logger{sl: slog.New(h), cfg: c}
}

// --- default logger and package-level facade ---

var sharedConfig = newConfig()

var (
	std = &Logger{sl: slog.New(&handler{cfg: sharedConfig, timestamp: true}), cfg: sharedConfig}
	// Plain logs through the same configuration as the default logger but without a
	// timestamp prefix (the "same functions without timestamps").
	Plain = &Logger{sl: slog.New(&handler{cfg: sharedConfig, timestamp: false}), cfg: sharedConfig}
)

// Default returns the package default logger.
func Default() *Logger { return std }

func Debug(msg string)               { std.Debug(msg) }
func Debugf(format string, a ...any) { std.Debugf(format, a...) }
func Info(msg string)                { std.Info(msg) }
func Infof(format string, a ...any)  { std.Infof(format, a...) }
func Warn(msg string)                { std.Warn(msg) }
func Warnf(format string, a ...any)  { std.Warnf(format, a...) }
func Error(msg string)               { std.Error(msg) }
func Errorf(format string, a ...any) { std.Errorf(format, a...) }
func Print(msg string)               { std.Print(msg) }
func Printf(format string, a ...any) { std.Printf(format, a...) }

// SetOutput sets the primary writer of the default logger (and Plain). Default: stderr.
// Whether colour is emitted in ColorAuto mode is re-evaluated against the new writer.
func SetOutput(w io.Writer) {
	sharedConfig.mu.Lock()
	sharedConfig.out = w
	sharedConfig.outIsTerminal = isTerminal(w)
	sharedConfig.mu.Unlock()
}

// SetLevel sets the minimum level emitted by the default logger.
func SetLevel(l Level) { sharedConfig.levelVar.Set(l) }

// SetTimePrecision sets the timestamp precision of the default logger.
func SetTimePrecision(p TimePrecision) {
	sharedConfig.mu.Lock()
	sharedConfig.precision = p
	sharedConfig.mu.Unlock()
}

// SetColorMode sets when colour is emitted: ColorAuto (terminal only, the default),
// ColorAlways, or ColorNever.
func SetColorMode(m ColorMode) {
	sharedConfig.mu.Lock()
	sharedConfig.colorMode = m
	sharedConfig.mu.Unlock()
}

// SetNoColors is a convenience over SetColorMode: true selects ColorNever (no colour, ANSI
// stripped), false restores ColorAuto (colour only when the output is a terminal).
func SetNoColors(b bool) {
	if b {
		SetColorMode(ColorNever)
	} else {
		SetColorMode(ColorAuto)
	}
}

// SetUTC selects UTC timestamps for the default logger.
func SetUTC(b bool) {
	sharedConfig.mu.Lock()
	sharedConfig.utc = b
	sharedConfig.mu.Unlock()
}

// LogToFile additionally writes every log line to path (ANSI sequences always stripped, so
// the file is plain text). A previously opened log file is closed first. The file is
// opened for append.
func LogToFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("logger: open log file %s: %w", path, err)
	}
	sharedConfig.mu.Lock()
	defer sharedConfig.mu.Unlock()
	if sharedConfig.fileClose != nil {
		sharedConfig.fileClose.Close()
	}
	sharedConfig.file = f
	sharedConfig.fileClose = f
	return nil
}

// CloseLogFile stops file logging and closes the current log file, if any.
func CloseLogFile() error {
	sharedConfig.mu.Lock()
	defer sharedConfig.mu.Unlock()
	if sharedConfig.fileClose == nil {
		return nil
	}
	err := sharedConfig.fileClose.Close()
	sharedConfig.file = nil
	sharedConfig.fileClose = nil
	return err
}

// --- legacy compatibility shims ---
//
// The functions below preserve the original API. The fixed-precision variants honour the
// precision in their name (independent of SetTimePrecision); prefer the level functions
// plus SetTimePrecision for new code.

func InfoMilliseconds(message string)  { sharedConfig.emit(true, Milliseconds, LevelInfo, message) }
func InfoMicroseconds(message string)  { sharedConfig.emit(true, Microseconds, LevelInfo, message) }
func InfoNanoseconds(message string)   { sharedConfig.emit(true, Nanoseconds, LevelInfo, message) }
func WarnMilliseconds(message string)  { sharedConfig.emit(true, Milliseconds, LevelWarn, message) }
func WarnMicroseconds(message string)  { sharedConfig.emit(true, Microseconds, LevelWarn, message) }
func WarnNanoseconds(message string)   { sharedConfig.emit(true, Nanoseconds, LevelWarn, message) }
func ErrorMilliseconds(message string) { sharedConfig.emit(true, Milliseconds, LevelError, message) }
func ErrorMicroseconds(message string) { sharedConfig.emit(true, Microseconds, LevelError, message) }
func ErrorNanoseconds(message string)  { sharedConfig.emit(true, Nanoseconds, LevelError, message) }
func DebugMilliseconds(message string) { sharedConfig.emit(true, Milliseconds, LevelDebug, message) }
func DebugMicroseconds(message string) { sharedConfig.emit(true, Microseconds, LevelDebug, message) }
func DebugNanoseconds(message string)  { sharedConfig.emit(true, Nanoseconds, LevelDebug, message) }
func PrintMilliseconds(message string) { sharedConfig.emit(true, Milliseconds, LevelPrint, message) }
func PrintMicroseconds(message string) { sharedConfig.emit(true, Microseconds, LevelPrint, message) }
func PrintNanoseconds(message string)  { sharedConfig.emit(true, Nanoseconds, LevelPrint, message) }

// rawTimestamped writes a printf-style line with a leading timestamp and no level tag,
// honouring the output/file/colour configuration.
func (c *config) rawTimestamped(prec TimePrecision, format string, a ...any) {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.utc {
		now = now.UTC()
	}
	line := "[" + now.Format(layoutFor(prec)) + "] " + fmt.Sprintf(format, a...)
	if c.out != nil {
		if c.useColor() {
			io.WriteString(c.out, line)
		} else {
			io.WriteString(c.out, stripANSI(line))
		}
	}
	if c.file != nil {
		io.WriteString(c.file, stripANSI(line))
	}
}

func DatePrintf(format string, message ...any) {
	sharedConfig.rawTimestamped(Seconds, format, message...)
}
func DatePrintfMilliseconds(format string, message ...any) {
	sharedConfig.rawTimestamped(Milliseconds, format, message...)
}
func DatePrintfMicroseconds(format string, message ...any) {
	sharedConfig.rawTimestamped(Microseconds, format, message...)
}
func DatePrintfNanoseconds(format string, message ...any) {
	sharedConfig.rawTimestamped(Nanoseconds, format, message...)
}

// LoggerLock and Lock/Unlock are retained for compatibility. Logging is now internally
// synchronised, so they are no longer required to avoid interleaved output.
var LoggerLock sync.Mutex

func Lock()   { LoggerLock.Lock() }
func Unlock() { LoggerLock.Unlock() }

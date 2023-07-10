package logger

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

var std = NewLogger(DefaultOptions())

type Logger struct {
	zl zerolog.Logger
	dw diode.Writer

	output io.Writer
	pretty bool
}

type Options struct {
	Output io.Writer
	Pretty bool
	Level  zerolog.Level
}

func DefaultOptions() Options {
	return Options{
		Output: os.Stderr,
		Pretty: false,
		Level:  zerolog.InfoLevel,
	}
}

func NewLogger(opts Options) *Logger {
	l := Logger{
		pretty: opts.Pretty,
	}
	var w io.Writer = os.Stderr
	if opts.Output != nil {
		w = opts.Output
	}
	l.output = w

	l.dw = diode.NewWriter(w, 1000, 10*time.Millisecond, func(missed int) {
		_, _ = fmt.Fprintf(w, "Logger dropped %d messages", missed)
	})
	if opts.Pretty {
		cw := zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339Nano}
		l.zl = zerolog.New(cw).With().Timestamp().Logger().Level(opts.Level)
	} else {
		l.zl = zerolog.New(w).With().Timestamp().Logger().Level(opts.Level)
	}

	return &l
}

func (l *Logger) Close() error {
	return l.dw.Close()
}

func (l *Logger) Trace() *zerolog.Event {
	return l.zl.Trace()
}

func (l *Logger) Debug() *zerolog.Event {
	return l.zl.Debug()
}

func (l *Logger) Info() *zerolog.Event {
	return l.zl.Info()
}

func (l *Logger) Warn() *zerolog.Event {
	return l.zl.Warn()
}

func (l *Logger) Error() *zerolog.Event {
	return l.zl.Error()
}

func (l *Logger) IsPretty() bool {
	return l.pretty
}

func (l *Logger) Output() io.Writer {
	return l.output
}

func Close() error {
	if std != nil {
		return std.dw.Close()
	}

	return nil
}

func Trace() *zerolog.Event {
	return std.zl.Trace()
}

func Debug() *zerolog.Event {
	return std.zl.Debug()
}

func Info() *zerolog.Event {
	return std.zl.Info()
}

func Warn() *zerolog.Event {
	return std.zl.Warn()
}

func Error() *zerolog.Event {
	return std.zl.Error()
}

func IsPretty() bool {
	return std.pretty
}

func Output() io.Writer {
	return std.output
}

func SetupLogger(opts Options) {
	std = NewLogger(opts)
}

func DebugStacktrace(p interface{}) {
	debugStack := debug.Stack()
	if IsPretty() {
		formatter := newStackFormatter()
		out, err := formatter.format(p, debugStack)
		if err != nil {
			_, _ = fmt.Fprint(Output(), string(debugStack))
		} else {
			_, _ = fmt.Fprint(Output(), string(out))
		}
	} else {
		lines := strings.Split(string(debugStack), "\n")
		Debug().
			Strs("stacktrace", lines).
			Msg("Stacktrace")
	}
}

func ErrorStacktrace(p interface{}) {
	debugStack := debug.Stack()
	if IsPretty() {
		formatter := newStackFormatter()
		out, err := formatter.format(p, debugStack)
		if err != nil {
			_, _ = fmt.Fprint(Output(), string(debugStack))
		} else {
			_, _ = fmt.Fprint(Output(), string(out))
		}
	} else {
		lines := strings.Split(string(debugStack), "\n")
		Error().
			Strs("stacktrace", lines).
			Msg("Stacktrace")
	}
}

package log

import (
	"fmt"
	"io"

	glog "github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

// Logger is a wrapper around `zerolog.Logger` that provides an implementation of `echo.Logger` interface
type Logger struct {
	Log     zerolog.Logger
	out     io.Writer
	level   glog.Lvl
	prefix  string
	setters []Setter
}

// New returns a new Logger instance
func New(out io.Writer, setters ...Setter) *Logger {
	switch l := out.(type) {
	case zerolog.Logger:
		return newLogger(l, setters)
	default:
		return newLogger(zerolog.New(out), setters)
	}
}

// From returns a new Logger instance using existing zerolog log.
func From(l zerolog.Logger, setters ...Setter) *Logger {
	return newLogger(l, setters)
}

func newLogger(l zerolog.Logger, setters []Setter) *Logger {
	opts := newOptions(l, setters)
	return &Logger{
		Log:     opts.context.Logger(),
		out:     nil,
		level:   opts.level,
		prefix:  opts.prefix,
		setters: setters,
	}
}

func (l Logger) Debug(i ...any) {
	l.Log.Debug().Msg(fmt.Sprint(i...))
}

func (l Logger) Debugf(format string, i ...any) {
	l.Log.Debug().Msgf(format, i...)
}

func (l Logger) Debugj(j glog.JSON) {
	l.logJSON(l.Log.Debug(), j)
}

func (l Logger) Info(i ...any) {
	l.Log.Info().Msg(fmt.Sprint(i...))
}

func (l Logger) Infof(format string, i ...any) {
	l.Log.Info().Msgf(format, i...)
}

func (l Logger) Infoj(j glog.JSON) {
	l.logJSON(l.Log.Info(), j)
}

func (l Logger) Warn(i ...any) {
	l.Log.Warn().Msg(fmt.Sprint(i...))
}

func (l Logger) Warnf(format string, i ...any) {
	l.Log.Warn().Msgf(format, i...)
}

func (l Logger) Warnj(j glog.JSON) {
	l.logJSON(l.Log.Warn(), j)
}

func (l Logger) Error(i ...any) {
	l.Log.Error().Msg(fmt.Sprint(i...))
}

func (l Logger) Errorf(format string, i ...any) {
	l.Log.Error().Msgf(format, i...)
}

func (l Logger) Errorj(j glog.JSON) {
	l.logJSON(l.Log.Error(), j)
}

func (l Logger) Fatal(i ...any) {
	l.Log.Fatal().Msg(fmt.Sprint(i...))
}

func (l Logger) Fatalf(format string, i ...any) {
	l.Log.Fatal().Msgf(format, i...)
}

func (l Logger) Fatalj(j glog.JSON) {
	l.logJSON(l.Log.Fatal(), j)
}

func (l Logger) Panic(i ...any) {
	l.Log.Panic().Msg(fmt.Sprint(i...))
}

func (l Logger) Panicf(format string, i ...any) {
	l.Log.Panic().Msgf(format, i...)
}

func (l Logger) Panicj(j glog.JSON) {
	l.logJSON(l.Log.Panic(), j)
}

func (l Logger) Print(i ...any) {
	l.Log.WithLevel(zerolog.NoLevel).Str("level", "-").Msg(fmt.Sprint(i...))
}

func (l Logger) Printf(format string, i ...any) {
	l.Log.WithLevel(zerolog.NoLevel).Str("level", "-").Msgf(format, i...)
}

func (l Logger) Printj(j glog.JSON) {
	l.logJSON(l.Log.WithLevel(zerolog.NoLevel).Str("level", "-"), j)
}

func (l Logger) Output() io.Writer {
	return l.Log
}

func (l *Logger) SetOutput(newOut io.Writer) {
	l.out = newOut
	l.Log = l.Log.Output(newOut)
}

func (l Logger) Level() glog.Lvl {
	return l.level
}

func (l *Logger) SetLevel(level glog.Lvl) {
	zlvl, elvl := MatchEchoLevel(level)

	l.setters = append(l.setters, WithLevel(elvl))
	l.level = elvl
	l.Log = l.Log.Level(zlvl)
}

func (l Logger) Prefix() string {
	return l.prefix
}

func (l Logger) SetHeader(h string) {
	// not implemented
}

func (l *Logger) SetPrefix(newPrefix string) {
	l.setters = append(l.setters, WithPrefix(newPrefix))

	opts := newOptions(l.Log, l.setters)

	l.prefix = newPrefix
	l.Log = opts.context.Logger()
}

func (l *Logger) Unwrap() zerolog.Logger {
	return l.Log
}

func (l *Logger) logJSON(event *zerolog.Event, j glog.JSON) {
	for k, v := range j {
		event = event.Interface(k, v)
	}

	event.Msg("")
}

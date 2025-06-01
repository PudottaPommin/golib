package log

import (
	glog "github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

type (
	Options struct {
		context zerolog.Context
		level   glog.Lvl
		prefix  string
	}

	Setter func(opts *Options)
)

func newOptions(l zerolog.Logger, setters []Setter) *Options {
	elvl, _ := MatchZeroLevel(l.GetLevel())
	opts := &Options{
		context: l.With(),
		level:   elvl,
	}

	for _, set := range setters {
		set(opts)
	}

	return opts
}

func WithLevel(level glog.Lvl) Setter {
	return func(opts *Options) {
		zlvl, elvl := MatchEchoLevel(level)
		opts.context = opts.context.Logger().Level(zlvl).With()
		opts.level = elvl
	}
}

func WithField(name string, value any) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Interface(name, value)
	}
}

func WithFields(fields map[string]any) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Fields(fields)
	}
}

func WithTimestamp() Setter {
	return func(opts *Options) {
		opts.context = opts.context.Timestamp()
	}
}

func WithCaller() Setter {
	return func(opts *Options) {
		opts.context = opts.context.Caller()
	}
}

func WithCallerWithSkipFrameCount(skipFrameCount int) Setter {
	return func(opts *Options) {
		opts.context = opts.context.CallerWithSkipFrameCount(skipFrameCount)
	}
}

func WithPrefix(prefix string) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Str("prefix", prefix)
	}
}

func WithHook(hook zerolog.Hook) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Logger().Hook(hook).With()
	}
}

func WithHookFunc(hook zerolog.HookFunc) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Logger().Hook(hook).With()
	}
}

package log

import (
	glog "github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

var (
	echoLevels = map[glog.Lvl]zerolog.Level{
		glog.DEBUG: zerolog.DebugLevel,
		glog.INFO:  zerolog.InfoLevel,
		glog.WARN:  zerolog.WarnLevel,
		glog.ERROR: zerolog.ErrorLevel,
		glog.OFF:   zerolog.NoLevel,
	}

	zeroLevels = map[zerolog.Level]glog.Lvl{
		zerolog.TraceLevel: glog.DEBUG,
		zerolog.DebugLevel: glog.DEBUG,
		zerolog.InfoLevel:  glog.INFO,
		zerolog.WarnLevel:  glog.WARN,
		zerolog.ErrorLevel: glog.ERROR,
		zerolog.NoLevel:    glog.OFF,
	}
)

// MatchEchoLevel returns a zerolog level and echo level for a given echo level
func MatchEchoLevel(level glog.Lvl) (zerolog.Level, glog.Lvl) {
	if zlvl, ok := echoLevels[level]; ok {
		return zlvl, level
	}
	return zerolog.NoLevel, glog.OFF
}

// MatchZeroLevel returns an echo level and zerolog level for a given zerolog level
func MatchZeroLevel(level zerolog.Level) (glog.Lvl, zerolog.Level) {
	if elvl, ok := zeroLevels[level]; ok {
		return elvl, level
	}
	return glog.OFF, zerolog.NoLevel
}

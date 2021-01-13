package log

import (
	"github.com/curltech/go-colla-core/config"
	"github.com/kataras/golog"
)

func init() {
	level, _ := config.GetString("log.level", "debug")
	golog.SetLevel(level)
	timeFormat := config.AppParams.TimeFormat
	golog.SetTimeFormat(timeFormat)
	// Levels contains a map of the log levels and their attributes.
	errorAttrs := golog.Levels[golog.ErrorLevel]

	// Change a log level's text.
	customColorCode := 156
	errorAttrs.SetText("custom text", customColorCode)

	// Get (rich) text per log level.
	enableColors := true
	errorAttrs.Text(enableColors)
	golog.Infof("log config completed!")
}

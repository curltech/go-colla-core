package debug

import (
	"github.com/kataras/golog"
	"time"
)

/**
defer trace("")()
*/
func Trace(msg string) func() {
	start := time.Now()
	golog.Debugf("start %s", msg)

	return func() {
		golog.Debugf("end %s, time:%s", msg, time.Since(start))
	}
}

package debug

import (
	"github.com/curltech/go-colla-core/logger"
	"time"
)

/**
defer trace("")()
*/
func Trace(msg string) func() {
	start := time.Now()
	logger.Debugf("start %s", msg)

	return func() {
		logger.Debugf("end %s, time:%s", msg, time.Since(start))
	}
}

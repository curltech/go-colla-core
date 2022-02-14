package debug

import (
	"github.com/curltech/go-colla-core/logger"
	"time"
)

/**
defer trace("")()
*/
func TraceDebug(msg string) func() {
	start := time.Now()

	return func() {
		logger.Sugar.Debugf("%v start time:%v,end time:%v,spend time:%v", msg, start, time.Now(), time.Since(start))
	}
}

func TraceInfo(msg string) func() {
	start := time.Now()

	return func() {
		logger.Sugar.Infof("%v start time:%v,end time:%v,spend time:%v", msg, start, time.Now(), time.Since(start))
	}
}

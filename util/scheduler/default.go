package scheduler

import (
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/reflect"
	"time"
)

func Scheduler(d time.Duration, fn interface{}, args []interface{}) {
	ticker := time.NewTicker(time.Second)
	for v := range ticker.C { // 循环channel
		logger.Sugar.Infof("start invoke in %v", v)
		go reflect.Invoke(fn, args)
	}
}

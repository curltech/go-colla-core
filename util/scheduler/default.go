package scheduler

import (
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/reflect"
	"github.com/robfig/cron/v3"
	"time"
)

func Scheduler(d time.Duration, fn interface{}, args []interface{}) {
	ticker := time.NewTicker(d * time.Second)
	for v := range ticker.C { // 循环channel
		logger.Sugar.Infof("start invoke in %v", v)
		go reflect.Invoke(fn, args)
	}
}

func RunCron(express string, fn func(), args []interface{}) *cron.Cron {
	c := cron.New()
	c.AddFunc(express, fn)

	go c.Start()

	return c
}

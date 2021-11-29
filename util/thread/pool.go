package thread

import (
	"github.com/panjf2000/ants/v2"
	"sync"
)

type RoutinePool struct {
	pool *ants.PoolWithFunc
	cf   func(interface{})
	wg   sync.WaitGroup
}

func CreateRoutinePool(size int, pf func(para interface{}), cf func(para interface{})) *RoutinePool {
	routinePool := &RoutinePool{cf: cf}
	if size <= 0 {
		size = 10
	}
	pool, _ := ants.NewPoolWithFunc(size, func(para interface{}) {
		pf(para)
		routinePool.wg.Done()
	})
	routinePool.pool = pool

	return routinePool
}

func (this *RoutinePool) Invoke(para interface{}) {
	this.wg.Add(1)
	_ = this.pool.Invoke(para)
}

func (this *RoutinePool) Release() {
	this.pool.Release()
	ants.Release()
}

func (this *RoutinePool) Wait(para interface{}) {
	this.wg.Wait()
	if this.cf != nil {
		this.cf(para)
	}
}

func (this *RoutinePool) Running() int {
	return this.pool.Running()
}

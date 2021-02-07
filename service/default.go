package service

import (
	"database/sql"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/repository"
	"github.com/curltech/go-colla-core/util/collection"
	"strconv"
)

type BaseService interface {
	GetSeq() uint64
	GetSeqs(count int) []uint64
	NewEntity(data []byte) (interface{}, error)
	//返回数组的指针
	NewEntities(data []byte) (interface{}, error)
	ParseJSON(data []byte) ([]interface{}, error)
	Get(dest interface{}, locked bool, orderby string, conds string, params ...interface{}) bool
	Find(rowsSlicePtr interface{}, md interface{}, orderby string, from int, limit int, conds string, params ...interface{}) error
	Insert(mds ...interface{}) int64
	Update(md interface{}, columns []string, conds string, params ...interface{}) int64
	Upsert(mds ...interface{}) int64
	Delete(md interface{}, conds string, params ...interface{}) int64
	Save(mds ...interface{}) int64
	Exec(clause string, params ...interface{}) sql.Result
	Query(clause string, params ...interface{}) []map[string][]byte
	Count(bean interface{}, conds string, params ...interface{}) int64
	Transaction(fc func(s repository.DbSession) (interface{}, error)) (interface{}, error)
}

type SeqCache struct {
	name      string
	increment int
	queue     *collection.Queue
}

var idCaches = make(map[string]*SeqCache)

func RegistSeq(name string, increment uint64) {
	if increment == 0 {
		increment = 500
	}
	if config.DatabaseParams.Sequence == "table" {
		GetSequenceService().CreateSeq(name, increment, 500)
	}
	_, ok := idCaches[name]
	if !ok {
		queue := collection.Queue{}
		queue.Create()
		idCaches[name] = &SeqCache{name: name, increment: int(increment), queue: &queue}
	}
}

func GetSeq(name string, count int) []uint64 {
	if count < 1 {
		panic("ErrorCount")
	}
	idCache, ok := idCaches[name]
	if !ok {
		logger.Sugar.Errorf("seqname:%v no regist", name)
		panic("SeqNotRegist")
	}
	ids, c := enough(name, count)
	if c < count {
		gap := count - c
		step := gap / idCache.increment
		if step*idCache.increment != gap {
			step++
		}

		increment, err := strconv.ParseUint(strconv.Itoa(idCache.increment), 10, 64)
		if err != nil {
			panic(err)
		}
		//var i uint64
		for i := 0; i < step; i++ {
			id := GetSeqValue(name)
			if id != 0 {
				base := id - increment + 1
				var j uint64
				for j = 0; j < increment; j++ {
					if c < count {
						var len = len(ids)
						if len > c {
							ids[c] = (j + base)
							c++
						} else {
							logger.Sugar.Warnf("")
						}
					} else {
						idCache.queue.Push(j + base)
					}
				}
			}
		}
	}

	return ids
}

func enough(name string, count int) ([]uint64, int) {
	ids := make([]uint64, count)
	idCache, ok := idCaches[name]
	if !ok {
		panic("SeqNotExist")
	}

	for i := 0; i < count; i++ {
		id := idCache.queue.Pop()
		if id != nil {
			ids[i] = id.(uint64)
		} else {
			return ids, i
		}
	}

	return ids, count
}

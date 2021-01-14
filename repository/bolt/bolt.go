package bolt

import (
	"database/sql"
	"fmt"
	"github.com/boltdb/bolt"
	baseentity "github.com/curltech/go-colla-core/entity"
	_ "github.com/curltech/go-colla-core/log"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/repository"
	"github.com/curltech/go-colla-core/util/message"
	"github.com/curltech/go-colla-core/util/reflect"
	_ "github.com/lib/pq"
	goreflect "reflect"
	"time"
)

//每次事务开始时创建新会话
type BoltSession struct {
	tx *bolt.Tx
}

var boltdb *bolt.DB

func init() {
	var err error
	boltdb, err = bolt.Open("mydb.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logger.Errorf("open bolt db:%v", err)
	}
}

func NewBoltSession() repository.DbSession {
	tx, err := boltdb.Begin(true)
	if err != nil {
		panic(err)
	}

	return &BoltSession{tx: tx}
}

func (this *BoltSession) Sync(bean ...interface{}) {

}

// Get retrieve one record from database, bean's non-empty fields
// will be as conditions
func (this *BoltSession) Get(dest interface{}, locked bool, orderby string, conds string, params ...interface{}) bool {
	id, _ := reflect.GetValue(dest, baseentity.FieldName_Id)
	key := fmt.Sprintf("%v", id)
	var bucketname string
	var found bool
	var err error
	if id != nil {
		b := this.tx.Bucket([]byte(bucketname))
		bs := b.Get([]byte(key))
		if bs == nil {
			found = false
		} else {
			message.Unmarshal(bs, dest)
			found = true
		}
	}
	if err != nil {
		panic(err)
	}

	return found
}

// Find retrieve records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct everyone := make([]Userinfo, 0)
// err := engine.Find(&everyone)
func (this *BoltSession) Find(rowsSlicePtr interface{}, md interface{}, orderby string, from int, limit int, conds string, params ...interface{}) error {
	var err error

	if err != nil {
		panic(err)
	}
	return nil
}

// insert model data to database
func (this *BoltSession) Insert(mds ...interface{}) int64 {
	var err error
	var affected int64
	for md := range mds {
		id, _ := reflect.GetValue(md, baseentity.FieldName_Id)
		key := fmt.Sprintf("%v", id)
		var bucketname string
		if id != nil {
			b := this.tx.Bucket([]byte(bucketname))
			buf, err := message.Marshal(md)
			if err != nil {
				panic(err)
			}
			err = b.Put([]byte(key), buf)
			if err != nil {
				panic(err)
			} else {
				affected++
			}
		}
	}
	if err != nil {
		panic(err)
	}

	return affected
}

// update model to database.
// cols set the columns those want to update.
func (this *BoltSession) Update(md interface{}, columns []string, conds string, params ...interface{}) int64 {
	var err error
	var affected int64
	var mds []interface{}
	var ok bool
	kind := reflect.GetIndirectType(md)
	if kind == goreflect.Slice || kind == goreflect.Array {
		mds, ok = md.([]interface{})
		if !ok {
			mds = reflect.ToArray(md)
		}
	} else {
		mds = make([]interface{}, 1)
		mds[0] = md
	}
	for md := range mds {
		id, _ := reflect.GetValue(md, baseentity.FieldName_Id)
		key := fmt.Sprintf("%v", id)
		var bucketname string
		if id != nil {
			b := this.tx.Bucket([]byte(bucketname))
			buf, err := message.Marshal(md)
			if err != nil {
				panic(err)
			}
			err = b.Put([]byte(key), buf)
			if err != nil {
				panic(err)
			} else {
				affected++
			}
		}
	}
	if err != nil {
		panic(err)
	}

	return affected
}

// delete model in database
// Delete records, bean's non-empty fields are conditions
func (this *BoltSession) Delete(md interface{}, conds string, params ...interface{}) int64 {
	var err error
	var affected int64
	var mds []interface{}
	var ok bool
	kind := reflect.GetIndirectType(md)
	if kind == goreflect.Slice || kind == goreflect.Array {
		mds, ok = md.([]interface{})
		if !ok {
			mds = reflect.ToArray(md)
		}
	} else {
		mds = make([]interface{}, 1)
		mds[0] = md
	}
	for md := range mds {
		id, _ := reflect.GetValue(md, baseentity.FieldName_Id)
		key := fmt.Sprintf("%v", id)
		var bucketname string
		var err error
		if id != nil {
			b := this.tx.Bucket([]byte(bucketname))
			err = b.Put([]byte(key), nil)
			if err != nil {

			} else {
				affected++
			}
		}
	}
	if err != nil {
		panic(err)
	}

	return affected
}

//execute sql and get result
func (this *BoltSession) Exec(clause string, params ...interface{}) sql.Result {

	return nil
}

//execute sql and get result
func (this *BoltSession) Query(clause string, params ...interface{}) []map[string][]byte {

	return nil
}

func (this *BoltSession) Count(bean interface{}, conds string, params ...interface{}) int64 {
	var count int64

	return count
}

/**
	Transaction 的 f 参数类型为 一个在事务内处理的函数
    因此可以将 f 函数作为参数传入 Transaction 函数中。
    return Transaction(func(s *XormSession) error {
        if _,error := session.Insert(User{ID:5,Version:"abc"}); error != nil{
            return error
        }
	})
*/
func (this *BoltSession) Transaction(fc func(s repository.DbSession) error) {
	defer this.Close()
	tx, err := boltdb.Begin(true)
	if err != nil {
		panic(err)
	}
	defer func() {
		if p := recover(); p != nil {
			logger.Error("recover rollback:%s\r\n", p)
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			logger.Error("error rollback:%s\r\n", err)
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
		}
	}()
	// 执行在事务内的处理
	err = fc(&BoltSession{tx: tx})
	if err != nil {
		panic(err)
	}
}

func (this *BoltSession) Begin() {
	tx, err := boltdb.Begin(true)
	if err != nil {
		panic(err)
	}
	this.tx = tx
}

func (this *BoltSession) Rollback() {
	err := this.tx.Rollback()
	if err != nil {
		panic(err)
	}
}

func (this *BoltSession) Commit() {
	err := this.tx.Commit()
	if err != nil {
		panic(err)
	}
}

func (this *BoltSession) Close() {

}

//scan result
func (this *BoltSession) Scan(dest interface{}) *BoltSession {

	return this
}

func (this *BoltSession) Complex(qb *repository.QueryBuilder, dest []interface{}) {

}

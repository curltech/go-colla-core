package service

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/curltech/go-colla-core/config"
	baseentity "github.com/curltech/go-colla-core/entity"
	"github.com/curltech/go-colla-core/excel"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/repository"
	"github.com/curltech/go-colla-core/repository/gorm"
	"github.com/curltech/go-colla-core/repository/xorm"
	"github.com/curltech/go-colla-core/util/debug"
	"github.com/curltech/go-colla-core/util/reflect"
	"github.com/curltech/go-colla-core/util/security"
	"strconv"
	"strings"
)

func PlaceQuestionMark(n int) string {
	var b strings.Builder
	for i := 0; i < n-1; i++ {
		b.WriteString("?,")
	}
	if n > 0 {
		b.WriteString("?")
	}
	return b.String()
}

type OrmBaseService struct {
	GetSeqName      func() string
	FactNewEntity   func(data []byte) (interface{}, error)
	FactNewEntities func(data []byte) (interface{}, error)
}

var ormBaseService BaseService = &OrmBaseService{}

func GetOrmBaseService() BaseService {
	return ormBaseService
}

func GetSeqValue(name string) (uint64, error) {
	if config.DatabaseParams.Sequence == "table" {
		return GetSequenceService().GetSeqValue(name)
	} else if config.DatabaseParams.Sequence == "seq" {
		sql := fmt.Sprintf("select nextval('%v')", name)
		result, err := ormBaseService.Query(sql)
		if err != nil {
			return 0, err
		}

		if len(result) > 0 {
			r := result[0]
			str := string(r["nextval"])
			i64, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				panic(err)
			}
			logger.Sugar.Infof("from sequence %v get id %v", name, i64)

			return i64, err
		} else {
			logger.Sugar.Errorf("no query result")
			return 0, err
		}
	}
	return 0, nil
}

func (this *OrmBaseService) NewEntity(data []byte) (interface{}, error) {
	return this.FactNewEntity(data)
}

func (this *OrmBaseService) NewEntities(data []byte) (interface{}, error) {
	return this.FactNewEntities(data)
}

func (this *OrmBaseService) ParseJSON(data []byte) ([]interface{}, error) {
	entities, err := this.NewEntities(data)
	if err != nil {
		entity, err := this.NewEntity(data)
		if err != nil {
			return nil, err
		}
		es := make([]interface{}, 0)
		es = append(es, entity)

		return es, nil
	}
	es := reflect.ToArray(entities)

	return es, nil
}

func (this *OrmBaseService) GetSeq() uint64 {
	ids := this.GetSeqs(1)
	if ids != nil && len(ids) > 0 {
		return ids[0]
	}

	return 0
}

func (this *OrmBaseService) GetSeqs(count int) []uint64 {
	seqname := this.GetSeqName()
	ids := GetSeq(seqname, count)

	return ids
}

// Get retrieve one record from database, bean's non-empty fields
// will be as conditions
func (this *OrmBaseService) Get(dest interface{}, locked bool, orderby string, conds string, params ...interface{}) (bool, error) {
	if !reflect.IsPtr(dest) {
		return false, errors.New("DestinationNeedPtr")
	}
	result, err := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		result, err := session.Get(dest, locked, orderby, conds, params...)
		// return nil will commit the whole transaction
		return result, err
	})
	if result == nil {
		return false, err
	}
	return result.(bool), err
}

// Find retrieve records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct everyone := make([]Userinfo, 0)
// err := engine.Find(&everyone)
func (this *OrmBaseService) Find(rowsSlicePtr interface{}, condiBean interface{}, orderby string, from int, limit int, conds string, params ...interface{}) error {
	var err error
	if !reflect.IsPtr(rowsSlicePtr) {
		err = errors.New("ResultNeedPtr")

		return err
	}
	if condiBean != nil && !reflect.IsPtr(condiBean) {
		err = errors.New("CondiBeanNeedPtr")

		return err
	}
	_, err = this.Transaction(func(session repository.DbSession) (interface{}, error) {
		err = session.Find(rowsSlicePtr, condiBean, orderby, from, limit, conds, params...)

		return nil, err
	})

	return err
}

func (this *OrmBaseService) setId(rowPtr interface{}) bool {
	var id uint64
	v, err := reflect.GetValue(rowPtr, baseentity.FieldName_Id)
	if err != nil {
		return false
	}
	id, _ = v.(uint64)
	if id == 0 {
		id = this.GetSeq()
		reflect.SetValue(rowPtr, baseentity.FieldName_Id, id)

		return true
	}

	return false
}

// insert model data to database
func (this *OrmBaseService) Insert(mds ...interface{}) (int64, error) {
	affected, err := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		var affected int64
		var err error
		for _, rowPtr := range mds {
			if !reflect.IsPtr(rowPtr) {
				panic(errors.New("DestinationNeedPtr"))
			}
			this.setId(rowPtr)

			_, err = session.Insert(rowPtr)
			if err == nil {
				affected++
			} else {
				return 0, err
			}
		}

		// return nil will commit the whole transaction
		return affected, err
	})
	if affected == nil || affected == 0 {
		return 0, err
	}

	return affected.(int64), err
}

func (this *OrmBaseService) BatchInsert(mds ...interface{}) (int64, error) {
	batch := 1000
	ms := make([]interface{}, 0)
	var count int64
	for i := 0; i < len(mds); i = i + batch {
		for j := 0; j < batch; j++ {
			if i+j < len(mds) {
				md := mds[i+j]
				ms = append(ms, md)
			}
		}
		c, err := this.Insert(ms...)
		if err != nil {
			logger.Sugar.Errorf("Insert database error:%v", err.Error())
			return count, err
		} else {
			logger.Sugar.Infof("Insert database record:%v", len(ms))
		}
		count = count + c
		ms = make([]interface{}, 0)
	}

	return count, nil
}

// update model to database.
// cols set the columns those want to update.
func (this *OrmBaseService) Update(md interface{}, columns []string, conds string, params ...interface{}) (int64, error) {
	affected, err := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		var affected int64
		affected, err := session.Update(md, columns, conds, params...)

		return affected, err
	})
	if affected == nil || affected == 0 {
		return 0, err
	}

	return affected.(int64), err
}

// Upsert model data to database by id field
func (this *OrmBaseService) Upsert(mds ...interface{}) (int64, error) {
	affected, err := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		var affected int64
		var err error
		for _, md := range mds {
			if !reflect.IsPtr(md) && !reflect.IsSlice(md) {
				panic(errors.New("DestinationNeedPtr"))
			}
			ok := this.setId(md)
			if ok {
				affected, err = session.Insert(md)
			} else {
				affected, err = session.Update(md, nil, "")
			}
			if err != nil {
				return 0, err
			} else {
				affected++
			}
		}
		// return nil will commit the whole transaction
		return affected, err
	})
	if affected == nil || affected == 0 {
		return 0, err
	}

	return affected.(int64), err
}

// delete model in database
// Delete records, bean's non-empty fields are conditions
func (this *OrmBaseService) Delete(md interface{}, conds string, params ...interface{}) (int64, error) {
	affected, err := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		var affected int64
		var err error
		affected, err = session.Delete(md, conds, params...)
		// return nil will commit the whole transaction
		return affected, err
	})
	if affected == nil || affected == 0 {
		return 0, err
	}

	return affected.(int64), err
}

// save model data to database by state field
func (this *OrmBaseService) Save(mds ...interface{}) (int64, error) {
	affected, err := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		var affected int64
		var err error
		for _, md := range mds {
			if !reflect.IsPtr(md) && !reflect.IsSlice(md) {
				panic(errors.New("DestinationNeedPtr"))
			}
			state, _ := reflect.GetValue(md, "State")
			if state != nil {
				switch state {
				case baseentity.EntityState_New:
					this.setId(md)
					affected, err = session.Insert(md)
				case baseentity.EntityState_Modified:
					affected, err = session.Update(md, nil, "")
				case baseentity.EntityState_Deleted:
					affected, err = session.Delete(md, "")
				}
			}
			if err != nil {
				return 0, err
			}
		}
		// return nil will commit the whole transaction
		return affected, err
	})
	if affected == nil || affected == 0 {
		return 0, err
	}

	return affected.(int64), err
}

// execute sql and get result
func (this *OrmBaseService) Exec(clause string, params ...interface{}) (sql.Result, error) {
	result, err := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		result, err := session.Exec(clause, params...)
		// return nil will commit the whole transactionerr
		return result, err
	})
	if result == nil {
		return nil, err
	}

	return result.(sql.Result), err
}

// execute sql and get result
func (this *OrmBaseService) Query(clause string, params ...interface{}) ([]map[string][]byte, error) {
	result, err := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		result, err := session.Query(clause, params...)

		// return nil will commit the whole transaction
		return result, err
	})
	if result == nil {
		return nil, err
	}

	return result.([]map[string][]byte), err
}

func (this *OrmBaseService) Count(bean interface{}, conds string, params ...interface{}) (int64, error) {
	result, err := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		if bean == nil {
			return 0, errors.New("condiBean can't be nil")
		}
		result, err := session.Count(bean, conds, params...)

		// return nil will commit the whole transaction
		return result, err
	})
	if result == nil {
		return 0, err
	}

	return result.(int64), err
}

/*
*

		Transaction 的 f 参数类型为 一个在事务内处理的函数
	    因此可以将 f 函数作为参数传入 Transaction 函数中。
	    return Transaction(func(s *BaseService) error {
	        if _,error := session.Insert(User{ID:5,Version:"abc"}); error != nil{
	            return error
	        }
		})
*/
func (this *OrmBaseService) Transaction(fc func(s repository.DbSession) (interface{}, error)) (interface{}, error) {
	id := security.UUID()
	msg := fmt.Sprintf("XORM Transaction %v :", id)
	fn := debug.TraceDebug(msg)
	defer fn()
	//先获取新会话
	var session = GetSession()
	var err error
	defer session.Close()
	err = session.Begin()
	defer func() {
		if p := recover(); p != nil {
			logger.Sugar.Errorf("recover rollback:%s\r\n", p)
			session.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			logger.Sugar.Errorf("error rollback:%s\r\n", err)
			session.Rollback() // err is non-nil; don't change it
		} else {
			err = session.Commit() // err is nil; if Commit returns error update err
		}
	}()
	// 执行在事务内的处理
	result, err := fc(session)
	if err != nil {
		logger.Sugar.Errorf("Exception:%v", err.Error())
	}

	return result, err
}

/*
*
导入excel格式的数据
*/
func (this *OrmBaseService) Import(filenames []string) error {
	for _, filename := range filenames {
		rowsSlicePtr, err := this.NewEntities(nil)
		if err != nil {

		}
		err = excel.Read(filename, rowsSlicePtr)
		if err == nil {
			rowsSlicePtrs := reflect.ToArray(rowsSlicePtr)
			this.Insert(rowsSlicePtrs...)
		}
	}
	return nil
}

func (this *OrmBaseService) Export(condiBean interface{}) ([]byte, error) {
	rowsSlicePtr, err := this.NewEntities(nil)
	if err != nil {

	}
	this.Find(rowsSlicePtr, condiBean, "", 0, 0, "")

	return excel.Write(rowsSlicePtr)
}

func GetSession() repository.DbSession {
	var session repository.DbSession
	if config.DatabaseParams.Orm == "xorm" {
		session = xorm.NewXormSession()
	} else if config.DatabaseParams.Orm == "gorm" {
		session = gorm.NewGormSession()
	}

	return session
}

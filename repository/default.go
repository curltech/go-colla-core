package repository

import (
	"database/sql"
	baseentity "github.com/curltech/go-colla-core/entity"
	"github.com/curltech/go-colla-core/util/reflect"
)

type QueryBuilder struct {
	Clause     string
	Select     string
	Distinct   []string
	From       interface{}
	Join       []string
	On         string
	Where      string
	OrderBy    string
	GroupBy    string
	Limit      int
	Offset     int
	Args       []interface{}
	Containers []interface{}
}

type DbSession interface {
	Sync(bean ...interface{}) error
	Get(dest interface{}, locked bool, orderby string, conds string, params ...interface{}) (bool, error)
	Find(rowsSlicePtr interface{}, md interface{}, orderby string, from int, limit int, conds string, params ...interface{}) error
	Insert(mds ...interface{}) (int64, error)
	Update(md interface{}, columns []string, conds string, params ...interface{}) (int64, error)
	Delete(md interface{}, conds string, params ...interface{}) (int64, error)
	Exec(clause string, params ...interface{}) (sql.Result, error)
	Query(clause string, params ...interface{}) ([]map[string][]byte, error)
	Count(bean interface{}, conds string, params ...interface{}) (int64, error)
	Transaction(fc func(s DbSession) error) error
	Begin() error
	Rollback() error
	Commit() error
	Close() error
}

func GetId(md interface{}) (interface{}, bool) {
	var id interface{}
	idnames, _ := reflect.Call(md, "IdName", nil)
	if idnames != nil && len(idnames) > 0 {
		v, err := reflect.GetValue(md, idnames[0].(string))
		if err == nil {
			id = v
		}
	} else {
		v, err := reflect.GetValue(md, baseentity.FieldName_Id)
		if err == nil {
			id = v
		}
	}
	if id != nil {
		i, ok := id.(uint64)
		if ok && i > 0 {
			return id, true
		}
		s, ok := id.(string)
		if ok && s != "" {
			return id, true
		}
	}

	return id, false
}

func SetId(md interface{}, id interface{}) bool {
	idnames, _ := reflect.Call(md, "IdName", nil)
	if idnames != nil && len(idnames) > 0 {
		err := reflect.SetValue(md, idnames[0].(string), id)
		if err == nil {
			return true
		}
	} else {
		err := reflect.SetValue(md, baseentity.FieldName_Id, id)
		if err == nil {
			return true
		}
	}

	return false
}

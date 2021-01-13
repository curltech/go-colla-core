package gorm

import (
	"database/sql"
	"fmt"
	"github.com/curltech/go-colla-core/config"
	_ "github.com/curltech/go-colla-core/log"
	"github.com/curltech/go-colla-core/repository"
	"github.com/curltech/go-colla-core/util/reflect"
	"github.com/kataras/golog"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	goreflect "reflect"
	"time"
)

type GormSession struct {
	Session *gorm.DB
}

var engine *gorm.DB

func init() {
	if config.DatabaseParams.Orm != "gorm" {
		return
	}
	//drivername := config.DatabaseParams.Drivername
	host := config.DatabaseParams.Host
	port := config.DatabaseParams.Port
	dbname := config.DatabaseParams.Dbname
	user := config.DatabaseParams.User
	password := config.DatabaseParams.Password
	sslmode := config.DatabaseParams.Sslmode
	//timeZone, _ := config.GetString("database.timeZone")
	maxIdleConns := config.DatabaseParams.MaxIdleConns
	maxOpenConns := config.DatabaseParams.MaxOpenConns
	connMaxLifetime := config.DatabaseParams.ConnMaxLifetime
	//showSQL := config.DatabaseParams.ShowSQL
	//logLevel := config.DatabaseParams.LogLevel

	//dsn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=%v timeZone=%v", host, port, dbname, user, password, sslmode, timeZone)
	dsn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=%v", host, port, dbname, user, password, sslmode)
	var err error
	engine, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}
	sqlDB, err := engine.DB()
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(maxIdleConns)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(maxOpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Hour)
}

func NewGormSession() repository.DbSession {
	s := engine

	return &GormSession{Session: s}
}

func (this *GormSession) Sync(bean ...interface{}) {
	err := engine.AutoMigrate(bean...)
	if err != nil {
		golog.Errorf("%v", err)
	}
}

// Get retrieve one record from database, bean's non-empty fields
// will be as conditions
func (this *GormSession) Get(dest interface{}, locked bool, orderby string, conds string, params ...interface{}) bool {
	var found bool
	var err error
	var session = this.Session
	if conds != "" {
		session = session.Where(conds, params...)
	}
	if locked == true {
		//session = session.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if orderby != "" {
		session = session.Order(orderby)
	}
	result := session.First(dest)
	if result.Error != nil {
		panic(err)
	}

	return found
}

// Find retrieve records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct everyone := make([]Userinfo, 0)
// err := engine.Find(&everyone)
func (this *GormSession) Find(rowsSlicePtr interface{}, md interface{}, orderby string, from int, limit int, conds string, params ...interface{}) error {
	var session = this.Session
	if limit != 0 || from != 0 {
		session = session.Limit(limit).Offset(from)
	}
	if conds != "" {
		session = session.Where(conds, params...)
	}
	if orderby != "" {
		session = session.Order(orderby)
	}
	if md == nil {
		session = session.Find(rowsSlicePtr)
	} else {
		session = session.Find(rowsSlicePtr, md)
	}

	return session.Error
}

// insert model data to database
func (this *GormSession) Insert(mds ...interface{}) int64 {
	var session = this.Session
	session = session.Create(&mds)
	if session.Error != nil {
		panic(session.Error)
	}

	return session.RowsAffected
}

//第一个参数是更新的数据数组，当传入的为结构体指针时，只有非空和0的field才会被作为更新的字段
//第二个参数指定要被更新的字段名称，即使非空和0的field也会被更新
//不支持指定this.Session.Table(new(User))来指定表名，而是通过结构数组来指定，因此不支持map更新
//在数据没有Id的时候，使用第三个参数条件bean作为条件
func (this *GormSession) Update(md interface{}, columns []string, conds string, params ...interface{}) int64 {
	var session = this.Session
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
	for _, md := range mds {
		if (columns != nil && len(columns) > 0) || conds != "" {
			session = session.Model(md)
			if columns != nil && len(columns) > 0 {
				session = session.Select(columns)
			}
			if conds != "" {
				session = session.Where(conds, params...)
			}
			session = session.Updates(md)
		} else {
			session = session.Save(md)
		}
	}
	if session.Error != nil {
		panic(session.Error)
	}

	return session.RowsAffected
}

//第一个参数是删除的数据数组，当传入的为结构体指针时，非空和0的field会被作为删除的条件
//不支持指定this.Session.Table(new(User))来指定表名，而是通过结构数组来指定，因此不支持map删除
//在数据没有Id的时候，使用第二个参数作为条件
func (this *GormSession) Delete(md interface{}, conds string, params ...interface{}) int64 {
	var session = this.Session
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
	for _, md := range mds {
		id, ok := repository.GetId(md)
		if !ok {
			if conds != "" && len(conds) > 0 {
				session = session.Where(conds, params...).Delete(md)
			} else {
				session = session.Delete(md)
			}
		} else {
			session = session.Delete(md, id)
		}
	}
	if session.Error != nil {
		panic(session.Error)
	}

	return session.RowsAffected
}

//execute sql and get result
func (this *GormSession) Exec(clause string, params ...interface{}) sql.Result {
	var session = this.Session
	session = session.Exec(clause, params...)
	if session.Error != nil {
		panic(session.Error)
	}

	return nil
}

//execute sql and get result
func (this *GormSession) Query(clause string, params ...interface{}) []map[string][]byte {
	var session = this.Session
	session = session.Raw(clause, params...)
	if session.Error != nil {
		panic(session.Error)
	}

	return nil
}

func (this *GormSession) Count(bean interface{}, conds string, params ...interface{}) int64 {
	var session = this.Session
	var count int64
	session = session.Model(bean)
	if conds != "" && len(conds) > 0 {
		session = session.Where(conds, params...).Count(&count)
	} else {
		session = session.Count(&count)
	}
	if session.Error != nil {
		panic(session.Error)
	}

	return count
}

/**
	Transaction 的 f 参数类型为 一个在事务内处理的函数
    因此可以将 f 函数作为参数传入 Transaction 函数中。
    return Transaction(func(s *GormSession) error {
        if _,error := session.Insert(User{ID:5,Version:"abc"}); error != nil{
            return error
        }
	})
*/
func (this *GormSession) Transaction(fc func(s repository.DbSession) error) {
	defer this.Close()
	var session = this.Session
	session = session.Begin()
	if session.Error != nil {
		panic(session.Error)
	}
	defer func() {
		if p := recover(); p != nil {
			golog.Error("recover rollback:%s\r\n", p)
			session.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if session.Error != nil {
			golog.Error("error rollback:%s\r\n", session.Error)
			session.Rollback() // err is non-nil; don't change it
		} else {
			session = session.Commit() // err is nil; if Commit returns error update err
		}
	}()
	// 执行在事务内的处理
	err := fc(this)
	if err != nil {
		panic(err)
	}
}

func (this *GormSession) Begin() {
	var session = this.Session
	session = session.Begin()
	if session.Error != nil {
		panic(session.Error)
	}
}

func (this *GormSession) Rollback() {
	var session = this.Session
	session = session.Rollback()
	if session.Error != nil {
		panic(session.Error)
	}
}

func (this *GormSession) Commit() {
	var session = this.Session
	session = session.Commit()
	if session.Error != nil {
		panic(session.Error)
	}
}

func (this *GormSession) Close() {
	//err := this.Session.Close()
	//if err != nil {
	//	panic(err)
	//}
}

//scan result
func (this *GormSession) Scan(dest interface{}) *GormSession {

	return this
}

func (this *GormSession) Complex(qb *repository.QueryBuilder, dest []interface{}) {

}

package xorm

import (
	"database/sql"
	"fmt"
	"github.com/curltech/go-colla-core/config"
	_ "github.com/curltech/go-colla-core/log"
	"github.com/curltech/go-colla-core/repository"
	"github.com/curltech/go-colla-core/util/reflect"
	"github.com/kataras/golog"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	goreflect "reflect"
	"strings"
	"time"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

type XormSession struct {
	Session *xorm.Session
}

var engine *xorm.Engine

/**
如果结构体拥有 TableName() string 的成员方法，那么此方法的返回值即是该结构体对应的数据库表名
// TableName 会将 User 的表名重写为 `profiles`
func (User) TableName() string {
  return "profiles"
}
通过 engine.Table() 方法可以改变 struct 对应的数据库表的名称
db.Table("deleted_users")
通过 sturct 中 field 对应的 Tag 中使用 xorm:"'column_name'"
可以使该 field 对应的 Column 名称为指定名称
extends Person
type User struct {
	Id   int64
	Name string  `xorm:"varchar(25) notnull unique 'usr_name' comment('姓名')"`
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
	Version int `xorm:"version"`
	DeletedAt time.Time `xorm:"deleted"`
}
*/
func init() {
	if config.DatabaseParams.Orm != "xorm" {
		return
	}
	drivername := config.DatabaseParams.Drivername
	dsn := config.DatabaseParams.Dsn
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
	showSQL := config.DatabaseParams.ShowSQL
	//logLevel := config.DatabaseParams.LogLevel

	//dsn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=%v timeZone=%v", host, port, dbname, user, password, sslmode, timeZone)
	if dsn == "" {
		dsn = fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=%v", host, port, dbname, user, password, sslmode)
	}
	/**
	如果用sqlite3，则xorm.NewEngine("sqlite3", "./test.db")
	*/
	engine, _ = xorm.NewEngine(drivername, dsn)
	engine.ShowSQL(showSQL)
	engine.Logger().SetLevel(log.LOG_ERR)
	//engine.Logger().SetLevel(core.LOG_DEBUG)
	//engine.SetMapper(names.GonicMapper{})
	//engine.SetTableMapper(LowerMapper{})
	engine.SetColumnMapper(LowerMapper{})

	//tbMapper := names.NewPrefixMapper(names.SnakeMapper{}, "prefix_")
	//engine.SetTableMapper(tbMapper)

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	engine.SetMaxIdleConns(maxIdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	engine.SetMaxOpenConns(maxOpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	engine.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Hour)

	//engine.TZLocation, _ = time.LoadLocation(timeZone)
	//engine.DatabaseTZ, _ = time.LoadLocation(timeZone)
	engine.TZLocation = time.UTC
	engine.DatabaseTZ = time.UTC
}

// LowerMapper implements IMapper and provides lower name between struct and
// database table
type LowerMapper struct {
}

func (m LowerMapper) Obj2Table(o string) string {
	return strings.ToLower(o)
}

func (m LowerMapper) Table2Obj(t string) string {
	return t
}

func NewXormSession() repository.DbSession {
	s := engine.NewSession()

	return &XormSession{Session: s}
}

func (this *XormSession) Sync(bean ...interface{}) {
	err := engine.Sync2(bean...)
	if err != nil {
		golog.Errorf("%v", err)
	}
}

// Get retrieve one record from database, bean's non-empty fields
// will be as conditions
func (this *XormSession) Get(dest interface{}, locked bool, orderby string, conds string, params ...interface{}) bool {
	var found bool
	var err error
	var session = this.Session
	if conds != "" {
		session = session.Where(conds, params...)
	}
	if locked == true {
		session = session.ForUpdate()
	}
	if orderby != "" {
		session = session.OrderBy(orderby)
	}
	found, err = session.Get(dest)
	if err != nil {
		panic(err)
	}

	return found
}

// Find retrieve records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct everyone := make([]Userinfo, 0)
// err := engine.Find(&everyone)
func (this *XormSession) Find(rowsSlicePtr interface{}, md interface{}, orderby string, from int, limit int, conds string, params ...interface{}) error {
	var err error
	var session = this.Session
	if limit != 0 || from != 0 {
		session = session.Limit(limit, from)
	}
	if conds != "" {
		session = session.Where(conds, params...)
	}
	if orderby != "" {
		session = session.OrderBy(orderby)
	}
	if md == nil {
		err = session.Find(rowsSlicePtr)
	} else {
		err = session.Find(rowsSlicePtr, md)
	}

	return err
}

// insert model data to database
func (this *XormSession) Insert(mds ...interface{}) int64 {
	affected, err := this.Session.Insert(mds...)
	if err != nil {
		panic(err)
	}

	return affected
}

//第一个参数是更新的数据数组，当传入的为结构体指针时，只有非空和0的field才会被作为更新的字段
//第二个参数指定要被更新的字段名称，即使非空和0的field也会被更新
//不支持指定this.Session.Table(new(User))来指定表名，而是通过结构数组来指定，因此不支持map更新
//在数据没有Id的时候，使用第三个参数条件bean作为条件
func (this *XormSession) Update(md interface{}, columns []string, conds string, params ...interface{}) int64 {
	var affected int64
	var err error
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
		var session = this.Session
		if columns != nil && len(columns) > 0 {
			session = session.Cols(columns...)
		}
		if conds != "" {
			session = session.Where(conds, params...)
		}
		id, ok := repository.GetId(md)
		if !ok {
			if conds == "" && params != nil && len(params) > 0 {
				affected, err = this.Session.Update(md, params...)
			} else {
				affected, err = this.Session.Update(md)
			}
		} else {
			affected, err = this.Session.ID(id).Update(md)
		}
		if err == nil {
			affected++
		}
	}
	if err != nil {
		panic(err)
	}
	return affected
}

//第一个参数是删除的数据数组，当传入的为结构体指针时，非空和0的field会被作为删除的条件
//不支持指定this.Session.Table(new(User))来指定表名，而是通过结构数组来指定，因此不支持map删除
//在数据没有Id的时候，使用第二个参数作为条件
func (this *XormSession) Delete(md interface{}, conds string, params ...interface{}) int64 {
	var affected int64 = 0
	var err error
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
				affected, err = this.Session.Where(conds, params...).Delete(md)
			} else {
				affected, err = this.Session.Delete(md)
			}
		} else {
			blank := reflect.New(md)
			if blank != nil {
				affected, err = this.Session.ID(id).Delete(blank)
			}
		}
	}
	if err != nil {
		panic(err)
	}
	return affected
}

//execute sql and get result
func (this *XormSession) Exec(clause string, params ...interface{}) sql.Result {
	var sqlOrArgs = make([]interface{}, 0)
	sqlOrArgs = append(sqlOrArgs, clause)
	if params != nil && len(params) > 0 {
		for _, p := range params {
			sqlOrArgs = append(sqlOrArgs, p)
		}
	}
	result, err := this.Session.Exec(sqlOrArgs...)
	if err != nil {
		panic(err)
	}
	return result
}

//execute sql and get result
func (this *XormSession) Query(clause string, params ...interface{}) []map[string][]byte {
	var sqlOrArgs = make([]interface{}, 0)
	sqlOrArgs = append(sqlOrArgs, clause)
	if params != nil && len(params) > 0 {
		for _, p := range params {
			sqlOrArgs = append(sqlOrArgs, p)
		}
	}
	result, err := this.Session.Query(sqlOrArgs...)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *XormSession) Count(bean interface{}, conds string, params ...interface{}) int64 {
	var count int64
	var err error
	if conds != "" && len(conds) > 0 {
		count, err = this.Session.Where(conds, params...).Count(bean)
	} else {
		count, err = this.Session.Count(bean)
	}
	if err != nil {
		panic(err)
	}

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
func (this *XormSession) Transaction(fc func(s repository.DbSession) error) {
	defer this.Close()
	err := this.Session.Begin()
	if err != nil {
		panic(err)
	}
	defer func() {
		if p := recover(); p != nil {
			golog.Error("recover rollback:%s\r\n", p)
			this.Session.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			golog.Error("error rollback:%s\r\n", err)
			this.Session.Rollback() // err is non-nil; don't change it
		} else {
			err = this.Session.Commit() // err is nil; if Commit returns error update err
		}
	}()
	// 执行在事务内的处理
	err = fc(this)
	if err != nil {
		panic(err)
	}
}

func (this *XormSession) Begin() {
	err := this.Session.Begin()
	if err != nil {
		panic(err)
	}
}

func (this *XormSession) Rollback() {
	err := this.Session.Rollback()
	if err != nil {
		panic(err)
	}
}

func (this *XormSession) Commit() {
	err := this.Session.Commit()
	if err != nil {
		panic(err)
	}
}

func (this *XormSession) Close() {
	err := this.Session.Close()
	if err != nil {
		panic(err)
	}
}

//scan result
func (this *XormSession) Scan(dest interface{}) *XormSession {
	rows, err := this.Session.Rows(dest)
	if err != nil {
		panic(err)
	} else {
		defer rows.Close()
		if rows.Next() {
			err = rows.Scan(dest)
			if err != nil {
				panic(err)
			}
		}
	}

	return this
}

func (this *XormSession) Complex(qb *repository.QueryBuilder, dest []interface{}) {
	// 构建查询对象
	rows, err := this.Session.Select(qb.Select).
		Distinct(qb.Distinct...).
		Table(qb.From).
		Join(qb.Join[0], qb.Join[1], qb.Join[2]).
		Where(qb.Where, qb.Args...).
		OrderBy(qb.OrderBy).
		GroupBy(qb.GroupBy).
		Limit(qb.Limit, qb.Offset).Rows(nil)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(dest)
		if err != nil {
			panic(err)
		}
	}
}

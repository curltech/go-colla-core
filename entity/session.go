package entity

import (
	"time"
)

/**
活动会话的数据
*/
type Session struct {
	BaseEntity `xorm:"extends"`
	SessionId  string     `xorm:"varchar(255) notnull" json:",omitempty"`
	Status     string     `xorm:"varchar(16)" json:",omitempty"`
	StatusDate *time.Time `json:",omitempty"`
	LifeTime   int64      `json:",omitempty"`
}

func (Session) TableName() string {
	return "bas_session"
}

func (Session) IdName() string {
	return FieldName_Id
}

type SessionData struct {
	BaseEntity `xorm:"extends"`
	SessionId  string     `xorm:"varchar(255) notnull" json:",omitempty"`
	Key        string     `xorm:"varchar(512)" json:",omitempty"`
	Value      string     `xorm:"varchar(32000)" json:",omitempty"`
	ValueType  string     `xorm:"varchar(32000)" json:",omitempty"`
	Status     string     `xorm:"varchar(16)" json:",omitempty"`
	StatusDate *time.Time `json:",omitempty"`
	LifeTime   int64      `json:",omitempty"`
}

func (SessionData) TableName() string {
	return "bas_sessiondata"
}

func (SessionData) IdName() string {
	return FieldName_Id
}

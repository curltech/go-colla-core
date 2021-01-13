package entity

import (
	"github.com/curltech/go-colla-core/entity"
	"time"
)

/**
定义了两个实体和他们对应的表名
*/
/**
会话历史记录
*/
type SessionInstance struct {
	entity.BaseEntity   `xorm:"extends"`
	SessionId           string     `xorm:"varchar(255) notnull" json:",omitempty"`
	Status              string     `xorm:"varchar(32)" json:",omitempty"`
	StatusReason        string     `xorm:"varchar(255)" json:",omitempty"`
	StatusDate          *time.Time `json:",omitempty"`
	Host                string     `xorm:"varchar(255) notnull" json:",omitempty"`
	Locale              string     `xorm:"varchar(255) notnull" json:",omitempty"`
	Url                 string     `xorm:"varchar(255) notnull" json:",omitempty"`
	IsMobile            bool       `xorm:"notnull" json:",omitempty"`
	IsSSL               bool       `xorm:"notnull" json:",omitempty"`
	MaxInactiveInterval int32
	LastAccessTime      *time.Time `json:",omitempty"`
}

func (SessionInstance) TableName() string {
	return "bas_sessioninstance"
}

func (SessionInstance) IdName() string {
	return entity.FieldName_Id
}

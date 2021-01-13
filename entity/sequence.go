package entity

/**
活动会话的数据
*/
type Sequence struct {
	Name       string `xorm:"pk" json:"name,omitempty"`
	MinValue   uint64 `xorm:"notnull" json:"minvalue,omitempty"`
	Increment  uint64 `xorm:"notnull" json:"increment,omitempty"`
	CurrentVal uint64 `xorm:"notnull" json:"currentval,omitempty"`
}

func (Sequence) TableName() string {
	return "bas_sequence"
}

func (Sequence) IdName() string {
	return "Name"
}

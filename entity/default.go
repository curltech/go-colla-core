package entity

import (
	"encoding/json"
	"github.com/curltech/go-colla-core/util/collection"
	"github.com/curltech/go-colla-core/util/reflect"
	reflect2 "reflect"
	"time"
)

type BaseEntity struct {
	Id         uint64     `xorm:"pk" json:"id,omitempty"`
	CreateDate *time.Time `xorm:"created" json:"createDate,omitempty"`
	UpdateDate *time.Time `xorm:"updated" json:"updateDate,omitempty"`
	EntityId   string     `xorm:"-" json:"entityId,omitempty"`
	State      string     `xorm:"-" json:"state,omitempty"`
}

type IBaseEntity interface {
	SetId(id uint64)
	UpdateState(state string)
	Json() ([]byte, error)
	Map() map[string]interface{}
}

func (this *BaseEntity) SetId(id uint64) {
	this.Id = id
}

func (this *BaseEntity) UpdateState(state string) {
	// 现在是新的
	if EntityState_New == this.State && EntityState_Modified == state {
		return
	}

	this.State = state
}

func (this *BaseEntity) Json() ([]byte, error) {
	return json.Marshal(this)
}

func (this *BaseEntity) Map() map[string]interface{} {
	options := collection.MapOptions{}
	return collection.StructToMap(this, &options)
}

type UserEntity struct {
	BaseEntity   `xorm:"extends"`
	CreateUserId string `xorm:"varchar(32)" json:"createUserId,omitempty"`
	UpdateUserId string `xorm:"varchar(32)" json:"updateUserId,omitempty"`
}

type VersionEntity struct {
	BaseEntity `xorm:"extends"`
	Version    int `xorm:"version" json:"version,omitempty"`
}

type StatusEntity struct {
	BaseEntity   `xorm:"extends"`
	Status       string     `xorm:"varchar(16)" json:"status,omitempty"`
	StatusReason string     `xorm:"varchar(255)" json:"statusReason,omitempty"`
	StatusDate   *time.Time `json:"statusDate,omitempty"`
}

type DeletedEntity struct {
	BaseEntity `xorm:"extends"` //`gorm:"embedded"`
	DeleteDate *time.Time       `xorm:"deleted" json:"deleteDate,omitempty"`
}

const (
	EntityState_New      string = "New"
	EntityState_Modified string = "Modified"
	EntityState_Deleted  string = "Deleted"
	EntityState_None     string = "None"
)

const (
	EntityStatus_Draft     string = "Draft"
	EntityStatus_Effective string = "Effective"
	EntityStatus_Expired   string = "Expired"
	EntityStatus_Deleted   string = "Deleted"
	EntityStatus_Canceled  string = "Canceled"
	EntityStatus_Checking  string = "Checking"
	EntityStatus_Undefined string = "Undefined"
	EntityStatus_Locked    string = "Locked"
	EntityStatus_Checked   string = "Checked"
	EntityStatus_Unchecked string = "Unchecked"
	EntityStatus_Disable   string = "Disable"
	EntityStatus_Discarded string = "Discarded"
	EntityStatus_Merged    string = "Merged"
	EntityStatus_Reversed  string = "Reversed"
)

const (
	FieldName_Id         string = "Id"
	FieldName_TopId      string = "TopId"
	FieldName_Kind       string = "Kind"
	FieldName_SpecId     string = "SpecId"
	FieldName_ParentId   string = "ParentId"
	FieldName_SchemaName string = "SchemaName"
	FieldName_State      string = "State"
	FieldName_DirtyFlag  string = "DirtyFlag"
)

const (
	JsonFieldName_Id         string = "id"
	JsonFieldName_TopId      string = "topId"
	JsonFieldName_Kind       string = "kind"
	JsonFieldName_SpecId     string = "specId"
	JsonFieldName_ParentId   string = "parentId"
	JsonFieldName_SchemaName string = "schemaName"
	JsonFieldName_State      string = "state"
	JsonFieldName_DirtyFlag  string = "dirtyFlag"
	JsonFieldName_Path       string = "path"
)

/**
以下是对数据进行比较的代码
*/
type EntitySnapshot struct {
	entity interface{}
	data   map[string]interface{}
	state  string
}

func NewSnapshot(entity interface{}) *EntitySnapshot {
	options := collection.MapOptions{
		OmitEmpty:  true,
		OmitBool:   false,
		OmitNumber: false,
	}
	m := collection.StructToMap(entity, &options)
	snapshot := EntitySnapshot{entity: entity, data: m}

	return &snapshot
}

type EntityDiff struct {
	data  map[string]interface{}
	state string
}

type EntityContext struct {
	diffs     map[string]*EntityDiff
	snapshots map[string]*EntitySnapshot
}

func (this *EntityContext) GetEntityDiff(entity interface{}) (*EntityDiff, error) {
	entityId, err := reflect.GetValue(entity, "entityId")
	if err != nil {
		return nil, err
	}
	diff := this.diffs[entityId.(string)]
	if diff == nil {
		diff = &EntityDiff{}
		this.diffs[entityId.(string)] = diff
	}

	return diff, nil
}

func (this *EntityContext) GetSyncInfo(snapshots map[string]*EntitySnapshot) (map[string]*EntityDiff, error) {
	for _, snapshot := range snapshots {
		snapshotState := snapshot.state
		entity := snapshot.entity
		entityState, err := reflect.GetValue(entity, "state")
		if err != nil {
			return nil, err
		}
		if entityState != snapshotState {
			diff, _ := this.GetEntityDiff(entity)
			diff.state = entityState.(string)
		}
		if "NONE" != entityState.(string) {
			dataDiff := make(map[string]interface{})
			diff, _ := this.GetEntityDiff(entity)
			diff.data = dataDiff
			for name, _ := range snapshot.data {
				value, _ := reflect.GetValue(entity, name)
				dataDiff[name] = value
			}
		}
	}

	return this.diffs, nil
}

func (this *EntityContext) RegisterEntity(entity interface{}) {
	entityId, _ := reflect.GetValue(entity, "entityId")
	if entityId.(string) == "" {
		id, _ := reflect.GetValue(entity, FieldName_Id)
		if id != nil {
			entityId = id.(string)
		}
	}
	if entityId.(string) == "" {
		entityId = "101"
	}
	this.snapshots[entityId.(string)] = &EntitySnapshot{entity: entity}
	this.diffs[entityId.(string)] = nil

	typ := reflect2.TypeOf(entity)
	value := reflect2.ValueOf(entity)
	for i := 0; i < typ.NumField(); i++ {
		filedValue := value.Field(i).Interface()
		filedValueTyp := reflect.GetType(filedValue)
		if filedValueTyp == reflect2.Slice {
			for _, v := range filedValue.([]interface{}) {
				if v != nil {
					this.RegisterEntity(v)
				}
			}
		} else if filedValueTyp == reflect2.Struct {
			this.RegisterEntity(filedValue)
		}
	}
}

func BuildEntityContext(entity interface{}) *EntityContext {
	context := &EntityContext{}
	context.RegisterEntity(entity)
	return context
}

package service

import (
	"github.com/curltech/go-colla-core/base/entity"
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/service"
	"github.com/curltech/go-colla-core/util/message"
)

/**
同步表结构，服务继承基本服务的方法
*/
type SessionInstanceService struct {
	service.OrmBaseService
}

var sessionInstanceService = &SessionInstanceService{}

func GetSessionInstanceService() *SessionInstanceService {
	return sessionInstanceService
}

var seqname = "seq_base"

func (this *SessionInstanceService) GetSeqName() string {
	return seqname
}

func (this *SessionInstanceService) NewEntity(data []byte) (interface{}, error) {
	entity := &entity.SessionInstance{}
	if data == nil {
		return entity, nil
	}
	err := message.Unmarshal(data, entity)
	if err != nil {
		return nil, err
	}

	return entity, err
}

func (this *SessionInstanceService) NewEntities(data []byte) (interface{}, error) {
	entities := make([]*entity.SessionInstance, 0)
	if data == nil {
		return &entities, nil
	}
	err := message.Unmarshal(data, &entities)
	if err != nil {
		return nil, err
	}

	return &entities, err
}

func init() {
	service.GetSession().Sync(new(entity.SessionInstance))
	sessionInstanceService.OrmBaseService.GetSeqName = sessionInstanceService.GetSeqName
	sessionInstanceService.OrmBaseService.FactNewEntity = sessionInstanceService.NewEntity
	sessionInstanceService.OrmBaseService.FactNewEntities = sessionInstanceService.NewEntities
	service.RegistSeq(seqname, 0)
	container.RegistService("sessionInstance", sessionInstanceService)
}

func test() {
	filenames := make([]string, 0)
	filenames = append(filenames, "/Users/hujingsong/Downloads/bas_sessioninstance.xlsx")
	sessionInstanceService.Import(filenames)
}

package service

import (
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/entity"
	"github.com/curltech/go-colla-core/util/message"
)

type SessionService struct {
	OrmBaseService
}

var sessionService = &SessionService{}

func GetSessionService() *SessionService {
	return sessionService
}

var seqname = "seq_base"

func (this *SessionService) GetSeqName() string {
	return seqname
}

func (this *SessionService) NewEntity(data []byte) (interface{}, error) {
	entity := &entity.Session{}
	if data == nil {
		return entity, nil
	}
	err := message.Unmarshal(data, entity)
	if err != nil {
		return nil, err
	}

	return entity, err
}

func (this *SessionService) NewEntities(data []byte) (interface{}, error) {
	entities := make([]*entity.Session, 0)
	if data == nil {
		return &entities, nil
	}
	err := message.Unmarshal(data, entities)
	if err != nil {
		return nil, err
	}

	return &entities, err

	return &entities, nil
}

func init() {
	GetSession().Sync(new(entity.Session))
	RegistSeq(seqname, 0)
	sessionService.OrmBaseService.GetSeqName = sessionService.GetSeqName
	sessionService.OrmBaseService.FactNewEntity = sessionService.NewEntity
	sessionService.OrmBaseService.FactNewEntities = sessionService.NewEntities
	container.RegistService("session", sessionService)
}

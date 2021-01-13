package service

import (
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/entity"
)

type SessionDataService struct {
	OrmBaseService
}

var sessionDataService = &SessionDataService{}

func GetSessionDataService() *SessionDataService {
	return sessionDataService
}

func (this *SessionDataService) GetSeqName() string {
	return seqname
}

func init() {
	GetSession().Sync(new(entity.SessionData))
	sessionDataService.OrmBaseService.GetSeqName = sessionDataService.GetSeqName
	sessionDataService.OrmBaseService.FactNewEntity = sessionDataService.NewEntity
	sessionDataService.OrmBaseService.FactNewEntities = sessionDataService.NewEntities
	container.RegistService("sessionData", sessionDataService)
}

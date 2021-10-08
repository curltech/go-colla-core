package service

import (
	"github.com/curltech/go-colla-core/container"
	"github.com/curltech/go-colla-core/entity"
	"github.com/curltech/go-colla-core/repository"
	"github.com/curltech/go-colla-core/util/message"
)

type SequenceService struct {
	OrmBaseService
}

var sequenceService = &SequenceService{}

func GetSequenceService() *SequenceService {
	return sequenceService
}

func (this *SequenceService) GetSeqName() string {
	return ""
}

func (this *SequenceService) NewEntity(data []byte) (interface{}, error) {
	entity := &entity.Sequence{}
	if data == nil {
		return entity, nil
	}
	err := message.Unmarshal(data, entity)
	if err != nil {
		return nil, err
	}

	return entity, err
}

func (this *SequenceService) NewEntities(data []byte) (interface{}, error) {
	entities := make([]*entity.Sequence, 0)
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

func (this *SequenceService) GetSeqValue(name string) uint64 {
	affected, _ := this.Transaction(func(session repository.DbSession) (interface{}, error) {
		seq := &entity.Sequence{Name: name}
		ok,_ := session.Get(seq, true, "", "")
		var nextVal uint64
		if ok {
			if seq.Increment == 0 {
				seq.Increment = 1
			}
			nextVal = seq.CurrentVal + seq.Increment
			if nextVal < seq.MinValue {
				nextVal = seq.MinValue
			}
			seq.CurrentVal = nextVal
			session.Update(seq, nil, "")
		}
		return nextVal, nil
	})

	if affected == nil || affected == 0 {
		return 0
	}

	return affected.(uint64)
}

func (this *SequenceService) CreateSeq(name string, increment uint64, minValue uint64) int64 {
	seq := &entity.Sequence{Name: name}
	ok,_ := this.Get(seq, false, "", "")
	if !ok {
		seq := &entity.Sequence{Name: name, Increment: increment, MinValue: minValue}
		affected,_ := this.Insert(seq)

		return affected
	}

	return 0
}

func init() {
	GetSession().Sync(new(entity.Sequence))
	sequenceService.OrmBaseService.GetSeqName = sequenceService.GetSeqName
	sequenceService.OrmBaseService.FactNewEntity = sequenceService.NewEntity
	sequenceService.OrmBaseService.FactNewEntities = sequenceService.NewEntities
	container.RegistService("sequence", sequenceService)
}

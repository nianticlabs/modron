package nagatha

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/model"
)

func exceptionModelFromNagathaProto(ex *Exception) model.Exception {
	return model.Exception{
		Uuid:             ex.Uuid,
		SourceSystem:     ex.SourceSystem,
		UserEmail:        ex.UserEmail,
		NotificationName: ex.NotificationName,
		Justification:    ex.Justification,
		CreatedOn:        ex.CreatedOnTime.AsTime(),
		ValidUntil:       ex.ValidUntilTime.AsTime(),
	}
}

func exceptionNagathaProtoFromModel(ex model.Exception) *Exception {
	return &Exception{
		Uuid:             ex.Uuid,
		SourceSystem:     ex.SourceSystem,
		UserEmail:        ex.UserEmail,
		NotificationName: ex.NotificationName,
		Justification:    ex.Justification,
		CreatedOnTime:    timestamppb.New(ex.CreatedOn),
		ValidUntilTime:   timestamppb.New(ex.ValidUntil),
	}
}

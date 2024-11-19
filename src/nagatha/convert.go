package nagatha

import (
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/proto/generated/nagatha"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func exceptionModelFromNagathaProto(ex *nagatha.Exception) model.Exception {
	return model.Exception{
		UUID:             ex.Uuid,
		SourceSystem:     ex.SourceSystem,
		UserEmail:        ex.UserEmail,
		NotificationName: ex.NotificationName,
		Justification:    ex.Justification,
		CreatedOn:        ex.CreatedOnTime.AsTime(),
		ValidUntil:       ex.ValidUntilTime.AsTime(),
	}
}

func exceptionNagathaProtoFromModel(ex model.Exception) *nagatha.Exception {
	return &nagatha.Exception{
		Uuid:             ex.UUID,
		SourceSystem:     ex.SourceSystem,
		UserEmail:        ex.UserEmail,
		NotificationName: ex.NotificationName,
		Justification:    ex.Justification,
		CreatedOnTime:    timestamppb.New(ex.CreatedOn),
		ValidUntilTime:   timestamppb.New(ex.ValidUntil),
	}
}

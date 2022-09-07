package sideeffect

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)


func GenerateUUID(ctx workflow.Context) (uuid.UUID, error) {
	// UUIDErr allows us to extract both the id and the err from the sideeffect
	// not sure if there is a better way to do this
	type UUIDErr struct {
		id  uuid.UUID
		err error
	}

	var result UUIDErr
	encodedResult := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		uuid, err := uuid.NewUUID()

		return UUIDErr{
			id:  uuid,
			err: err,
		}
	})

	err := encodedResult.Get(&result)

	if err != nil {
		return uuid.UUID{}, errors.Wrap(err, "getting uuid from side effect")
	}

	return result.id, result.err
}
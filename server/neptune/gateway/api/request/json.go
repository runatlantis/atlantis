package request

import (
	"context"
	"encoding/json"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
)

// JSONRequestValidationProxy unmarshals a request's body and validates the result
// Consumers define how this struct is validated.
type JSONRequestValidationProxy[V validation.Validatable, R any] struct {
	Delegate delegate[V, R]
}

type delegate[T validation.Validatable, R any] interface {
	Convert(context context.Context, request T) (R, error)
}

func (c *JSONRequestValidationProxy[V, R]) Convert(from *http.Request) (R, error) {
	var v V
	var r R
	decoder := json.NewDecoder(from.Body)
	err := decoder.Decode(&v)

	if err != nil {
		return r, errors.Wrap(err, "decoding json")
	}

	if err := v.Validate(); err != nil {
		return r, errors.Wrap(err, "validating request")
	}

	return c.Delegate.Convert(from.Context(), v)
}

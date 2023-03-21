package request_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/gateway/api/request"
	"github.com/stretchr/testify/assert"
)

type validateable struct {
	Field string
}

func (v *validateable) GetField() string {
	return v.Field
}

func (v *validateable) Validate() error {
	return nil
}

type badValidateable struct {
	Field string
}

func (v *badValidateable) Validate() error {
	return assert.AnError
}

func (v *badValidateable) GetField() string {
	return v.Field
}

type fieldGetter interface {
	GetField() string
}

type converteable string
type converter[T fieldGetter] struct {
	result converteable
}

func (c converter[T]) Convert(ctx context.Context, v T) (converteable, error) {
	return converteable(v.GetField()), nil
}

func TestJsonRequestConverter_Success(t *testing.T) {
	expectedResult := converteable("hi")

	subject := &request.JSONRequestValidationProxy[*validateable, converteable]{
		Delegate: converter[*validateable]{
			result: expectedResult,
		},
	}

	request, err := http.NewRequestWithContext(context.Background(), "method", "www.url.com", strings.NewReader("{ \"Field\": \"hi\"}"))
	assert.NoError(t, err)

	result, err := subject.Convert(request)
	assert.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestJsonRequestConverter_FailedValidation(t *testing.T) {
	subject := &request.JSONRequestValidationProxy[*badValidateable, converteable]{
		Delegate: converter[*badValidateable]{
			result: converteable(""),
		},
	}

	request, err := http.NewRequestWithContext(context.Background(), "method", "www.url.com", strings.NewReader("{ \"Field\": \"hi\"}"))
	assert.NoError(t, err)

	_, err = subject.Convert(request)
	assert.Error(t, err)
}

func TestJsonRequestConverter_FailedDecoding(t *testing.T) {
	subject := &request.JSONRequestValidationProxy[*validateable, converteable]{
		Delegate: converter[*validateable]{
			result: converteable("hi"),
		},
	}

	request, err := http.NewRequestWithContext(context.Background(), "method", "www.url.com", strings.NewReader("body"))
	assert.NoError(t, err)

	_, err = subject.Convert(request)
	assert.Error(t, err)
}

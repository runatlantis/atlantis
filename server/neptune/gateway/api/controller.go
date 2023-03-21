package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Controller is a simple generic controller that converts a request and hands it off
// providing a level of consistency across API handling.
type Controller[Request any] struct {
	RequestConverter RequestConverter[Request]
	Handler          Handler[Request]
}

type Handler[Request any] interface {
	Handle(ctx context.Context, request Request) error
}

type RequestConverter[T any] interface {
	Convert(from *http.Request) (T, error)
}

func (c *Controller[Request]) Handle(w http.ResponseWriter, request *http.Request) {
	internalRequest, err := c.RequestConverter.Convert(request)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, errors.Wrap(err, "converting request"))
		return
	}

	err = c.Handler.Handle(request.Context(), internalRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, errors.Wrap(err, "handling request"))
		return
	}

	w.WriteHeader(http.StatusOK)

	//TODO: return something a bit more structured, like json
	fmt.Fprintln(w, "Request Submitted!")
}

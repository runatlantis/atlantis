package request

import (
	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/http"
)

type webhookParser interface {
	Parse(r *http.BufferedRequest, payload []byte) (interface{}, error)
}

type parser struct{}

func (e *parser) Parse(r *http.BufferedRequest, payload []byte) (interface{}, error) {
	return github.ParseWebHook(github.WebHookType(r.GetRequest()), payload)
}

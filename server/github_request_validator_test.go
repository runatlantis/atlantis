package server_test

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/atlantisnorth/atlantis/server"
	. "github.com/atlantisnorth/atlantis/testing"
	. "github.com/petergtz/pegomock"
)

func TestValidate_WithSecretErr(t *testing.T) {
	t.Log("if the request is not valid against the secret there is an error")
	RegisterMockTestingT(t)
	g := server.DefaultGithubRequestValidator{}
	buf := bytes.NewBufferString("")
	req, err := http.NewRequest("POST", "http://localhost/event", buf)
	Ok(t, err)
	req.Header.Set("X-Hub-Signature", "sha1=126f2c800419c60137ce748d7672e77b65cf16d6")
	req.Header.Set("Content-Type", "application/json")

	_, err = g.Validate(req, []byte("secret"))
	Assert(t, err != nil, "error should not be nil")
	Equals(t, "payload signature check failed", err.Error())
}

func TestValidate_WithSecret(t *testing.T) {
	t.Log("if the request is valid against the secret the payload is returned")
	RegisterMockTestingT(t)
	g := server.DefaultGithubRequestValidator{}
	buf := bytes.NewBufferString(`{"yo":true}`)
	req, err := http.NewRequest("POST", "http://localhost/event", buf)
	Ok(t, err)
	req.Header.Set("X-Hub-Signature", "sha1=126f2c800419c60137ce748d7672e77b65cf16d6")
	req.Header.Set("Content-Type", "application/json")

	bs, err := g.Validate(req, []byte("0123456789abcdef"))
	Ok(t, err)
	Equals(t, `{"yo":true}`, string(bs))
}

func TestValidate_WithoutSecretInvalidContentType(t *testing.T) {
	t.Log("if the request has an invalid content type an error is returned")
	RegisterMockTestingT(t)
	g := server.DefaultGithubRequestValidator{}
	buf := bytes.NewBufferString("")
	req, err := http.NewRequest("POST", "http://localhost/event", buf)
	Ok(t, err)
	req.Header.Set("Content-Type", "invalid")

	_, err = g.Validate(req, nil)
	Assert(t, err != nil, "error should not be nil")
	Equals(t, "webhook request has unsupported Content-Type \"invalid\"", err.Error())
}

func TestValidate_WithoutSecretJSON(t *testing.T) {
	t.Log("if the request is JSON the body is returned")
	RegisterMockTestingT(t)
	g := server.DefaultGithubRequestValidator{}
	buf := bytes.NewBufferString(`{"yo":true}`)
	req, err := http.NewRequest("POST", "http://localhost/event", buf)
	Ok(t, err)
	req.Header.Set("Content-Type", "application/json")

	bs, err := g.Validate(req, nil)
	Ok(t, err)
	Equals(t, `{"yo":true}`, string(bs))
}

func TestValidate_WithoutSecretFormNoPayload(t *testing.T) {
	t.Log("if the request is form encoded and does not contain a payload param an error is returned")
	RegisterMockTestingT(t)
	g := server.DefaultGithubRequestValidator{}
	buf := bytes.NewBufferString("")
	req, err := http.NewRequest("POST", "http://localhost/event", buf)
	Ok(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = g.Validate(req, nil)
	Assert(t, err != nil, "error should not be nil")
	Equals(t, "webhook request did not contain expected 'payload' form value", err.Error())
}

func TestValidate_WithoutSecretForm(t *testing.T) {
	t.Log("if the request is form encoded and does not contain a payload param an error is returned")
	RegisterMockTestingT(t)
	g := server.DefaultGithubRequestValidator{}
	form := url.Values{}
	form.Set("payload", `{"yo":true}`)
	buf := bytes.NewBufferString(form.Encode())
	req, err := http.NewRequest("POST", "http://localhost/event", buf)
	Ok(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	bs, err := g.Validate(req, nil)
	Ok(t, err)
	Equals(t, `{"yo":true}`, string(bs))
}

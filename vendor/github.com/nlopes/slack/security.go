package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"net/http"
)

// SecretsVerifier contains the information needed to verify that the request comes from Slack
type SecretsVerifier struct {
	slackSig  string
	timeStamp string
	hmac      hash.Hash
}

// NewSecretsVerifier returns a SecretsVerifier object in exchange for an http.Header object and signing secret
func NewSecretsVerifier(header http.Header, signingSecret string) (SecretsVerifier, error) {
	if header["X-Slack-Signature"][0] == "" || header["X-Slack-Request-Timestamp"][0] == "" {
		return SecretsVerifier{}, errors.New("Headers are empty, cannot create SecretsVerifier")
	}

	hash := hmac.New(sha256.New, []byte(signingSecret))
	hash.Write([]byte(fmt.Sprintf("v0:%s:", header["X-Slack-Request-Timestamp"][0])))
	return SecretsVerifier{
		slackSig:  header["X-Slack-Signature"][0],
		timeStamp: header["X-Slack-Request-Timestamp"][0],
		hmac:      hash,
	}, nil
}

func (v *SecretsVerifier) Write(body []byte) (n int, err error) {
	return v.hmac.Write(body)
}

// Ensure compares the signature sent from Slack with the actual computed hash to judge validity
func (v SecretsVerifier) Ensure() error {
	computed := "v0=" + string(hex.EncodeToString(v.hmac.Sum(nil)))
	if computed == v.slackSig {
		return nil
	}

	return fmt.Errorf("Expected signing signature: %s, but computed: %s", v.slackSig, computed)
}

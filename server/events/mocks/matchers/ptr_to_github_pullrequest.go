package matchers

import (
	"reflect"

	github "github.com/google/go-github/github"
	"github.com/petergtz/pegomock"
)

func AnyPtrToGithubPullRequest() *github.PullRequest {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(*github.PullRequest))(nil)).Elem()))
	var nullValue *github.PullRequest
	return nullValue
}

func EqPtrToGithubPullRequest(value *github.PullRequest) *github.PullRequest {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue *github.PullRequest
	return nullValue
}

package matchers

import (
	"reflect"

	go_gitlab "github.com/lkysow/go-gitlab"
	"github.com/petergtz/pegomock"
)

func AnyPtrToGoGitlabMergeRequest() *go_gitlab.MergeRequest {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(*go_gitlab.MergeRequest))(nil)).Elem()))
	var nullValue *go_gitlab.MergeRequest
	return nullValue
}

func EqPtrToGoGitlabMergeRequest(value *go_gitlab.MergeRequest) *go_gitlab.MergeRequest {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue *go_gitlab.MergeRequest
	return nullValue
}

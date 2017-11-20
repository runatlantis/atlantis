package matchers

import (
	"reflect"

	go_gitlab "github.com/lkysow/go-gitlab"
	"github.com/petergtz/pegomock"
)

func AnyGoGitlabMergeCommentEvent() go_gitlab.MergeCommentEvent {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(go_gitlab.MergeCommentEvent))(nil)).Elem()))
	var nullValue go_gitlab.MergeCommentEvent
	return nullValue
}

func EqGoGitlabMergeCommentEvent(value go_gitlab.MergeCommentEvent) go_gitlab.MergeCommentEvent {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue go_gitlab.MergeCommentEvent
	return nullValue
}

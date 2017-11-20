package matchers

import (
	"reflect"

	go_gitlab "github.com/lkysow/go-gitlab"
	"github.com/petergtz/pegomock"
)

func AnyGoGitlabMergeEvent() go_gitlab.MergeEvent {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(go_gitlab.MergeEvent))(nil)).Elem()))
	var nullValue go_gitlab.MergeEvent
	return nullValue
}

func EqGoGitlabMergeEvent(value go_gitlab.MergeEvent) go_gitlab.MergeEvent {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue go_gitlab.MergeEvent
	return nullValue
}

package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	vcs "github.com/runatlantis/atlantis/server/events/vcs"
)

func AnyVcsCommitStatus() vcs.CommitStatus {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(vcs.CommitStatus))(nil)).Elem()))
	var nullValue vcs.CommitStatus
	return nullValue
}

func EqVcsCommitStatus(value vcs.CommitStatus) vcs.CommitStatus {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue vcs.CommitStatus
	return nullValue
}

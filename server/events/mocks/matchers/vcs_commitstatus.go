package matchers

import (
	"reflect"

	vcs "github.com/atlantisnorth/atlantis/server/events/vcs"
	"github.com/petergtz/pegomock"
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

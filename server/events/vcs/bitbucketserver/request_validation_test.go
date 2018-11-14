package bitbucketserver_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	. "github.com/runatlantis/atlantis/testing"
)

func TestValidateSignature(t *testing.T) {
	body := `{"eventKey":"pr:comment:added","date":"2018-07-24T15:10:05+0200","actor":{"name":"lkysow","emailAddress":"lkysow@gmail.com","id":1,"displayName":"Luke Kysow","active":true,"slug":"lkysow","type":"NORMAL"},"pullRequest":{"id":5,"version":0,"title":"another.tf edited online with Bitbucket","state":"OPEN","open":true,"closed":false,"createdDate":1532365427513,"updatedDate":1532365427513,"fromRef":{"id":"refs/heads/lkysow/anothertf-1532365422773","displayId":"lkysow/anothertf-1532365422773","latestCommit":"b52b8e254e956654dcdb394d0ccba9199f420427","repository":{"slug":"atlantis-example","id":1,"name":"atlantis-example","scmId":"git","state":"AVAILABLE","statusMessage":"Available","forkable":true,"project":{"key":"AT","id":1,"name":"atlantis","public":false,"type":"NORMAL"},"public":false}},"toRef":{"id":"refs/heads/master","displayId":"master","latestCommit":"0a338874369017deba7c22e99e6000932724282f","repository":{"slug":"atlantis-example","id":1,"name":"atlantis-example","scmId":"git","state":"AVAILABLE","statusMessage":"Available","forkable":true,"project":{"key":"AT","id":1,"name":"atlantis","public":false,"type":"NORMAL"},"public":false}},"locked":false,"author":{"user":{"name":"lkysow","emailAddress":"lkysow@gmail.com","id":1,"displayName":"Luke Kysow","active":true,"slug":"lkysow","type":"NORMAL"},"role":"AUTHOR","approved":false,"status":"UNAPPROVED"},"reviewers":[],"participants":[]},"comment":{"properties":{"repositoryId":1},"id":65,"version":0,"text":"comment","author":{"name":"lkysow","emailAddress":"lkysow@gmail.com","id":1,"displayName":"Luke Kysow","active":true,"slug":"lkysow","type":"NORMAL"},"createdDate":1532437805137,"updatedDate":1532437805137,"comments":[],"tasks":[]}}`
	secret := "mysecret"
	sig := `sha256=ed11f92d1565a4de586727fe8260558277d58009e8957a79eb4749a7009ce083`
	err := bitbucketserver.ValidateSignature([]byte(body), sig, []byte(secret))
	Ok(t, err)
}

func TestValidateSignature_Invalid(t *testing.T) {
	body := `"eventKey":"pr:comment:added","date":"2018-07-24T15:10:05+0200","actor":{"name":"lkysow","emailAddress":"lkysow@gmail.com","id":1,"displayName":"Luke Kysow","active":true,"slug":"lkysow","type":"NORMAL"},"pullRequest":{"id":5,"version":0,"title":"another.tf edited online with Bitbucket","state":"OPEN","open":true,"closed":false,"createdDate":1532365427513,"updatedDate":1532365427513,"fromRef":{"id":"refs/heads/lkysow/anothertf-1532365422773","displayId":"lkysow/anothertf-1532365422773","latestCommit":"b52b8e254e956654dcdb394d0ccba9199f420427","repository":{"slug":"atlantis-example","id":1,"name":"atlantis-example","scmId":"git","state":"AVAILABLE","statusMessage":"Available","forkable":true,"project":{"key":"AT","id":1,"name":"atlantis","public":false,"type":"NORMAL"},"public":false}},"toRef":{"id":"refs/heads/master","displayId":"master","latestCommit":"0a338874369017deba7c22e99e6000932724282f","repository":{"slug":"atlantis-example","id":1,"name":"atlantis-example","scmId":"git","state":"AVAILABLE","statusMessage":"Available","forkable":true,"project":{"key":"AT","id":1,"name":"atlantis","public":false,"type":"NORMAL"},"public":false}},"locked":false,"author":{"user":{"name":"lkysow","emailAddress":"lkysow@gmail.com","id":1,"displayName":"Luke Kysow","active":true,"slug":"lkysow","type":"NORMAL"},"role":"AUTHOR","approved":false,"status":"UNAPPROVED"},"reviewers":[],"participants":[]},"comment":{"properties":{"repositoryId":1},"id":65,"version":0,"text":"comment","author":{"name":"lkysow","emailAddress":"lkysow@gmail.com","id":1,"displayName":"Luke Kysow","active":true,"slug":"lkysow","type":"NORMAL"},"createdDate":1532437805137,"updatedDate":1532437805137,"comments":[],"tasks":[]}}`
	secret := "mysecret"
	sig := `sha256=ed11f92d1565a4de586727fe8260558277d58009e8957a79eb4749a7009ce083`
	err := bitbucketserver.ValidateSignature([]byte(body), sig, []byte(secret))
	ErrEquals(t, "payload signature check failed", err)
}

package aws

import "github.com/aws/aws-sdk-go/aws/session"

func NewSession() (*session.Session, error) {
	awsSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return awsSession, nil
}

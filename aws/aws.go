package aws

import (
	"fmt"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

const assumeRoleSessionName = "atlantis"

type Config struct {
	Region      string
	RoleArn     string
	SessionName string
}

// CreateSession creates a new valid AWS session to be used by AWS clients
func (c *Config) CreateSession() (*session.Session, error) {
	session, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(c.Region)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new aws session: %v", err)
	}

	_, err = session.Config.Credentials.Get()
	if err != nil {
		return nil, fmt.Errorf("didn't find valid aws credentials. Please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables: %v", err)
	}

	// generate a new session if aws role is provided
	if c.RoleArn != "" {
		return c.assumeRole(session), nil
	}

	return session, nil
}

// assumeRole calls Amazon's Security Token Service and attempts to assume roleArn and provide credentials for that role
func (c *Config) assumeRole(s *session.Session) *session.Session {
	if c.SessionName == "" {
		c.SessionName = assumeRoleSessionName
	}
	stsClient := sts.New(s, s.Config)
	creds := stscreds.NewCredentialsWithClient(stsClient, c.RoleArn, func(p *stscreds.AssumeRoleProvider) {
		p.RoleSessionName = c.SessionName
		// override default 15 minute time
		p.Duration = time.Duration(30) * time.Minute
	})

	// now assume role
	return session.New(&aws.Config{Credentials: creds, Region: aws.String(c.Region)})
}

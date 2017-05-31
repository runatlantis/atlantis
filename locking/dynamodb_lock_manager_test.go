package locking_test

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/hootsuite/atlantis/locking"
	"testing"
	. "github.com/hootsuite/atlantis/testing_util"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type mockDynamoDB struct {
	dynamodbiface.DynamoDBAPI
}
func (m mockDynamoDB) Scan(s *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	return &dynamodb.ScanOutput{
		Items: []map[string]*dynamodb.AttributeValue{},
	}, nil
}

func TestListLocksEmpty(t *testing.T) {
	lock := &locking.DynamoDBLockManager{
		&mockDynamoDB{},
		"lockTable",
	}
	m, err := lock.ListLocks()
	Ok(t, err)
	Equals(t, 0, len(m))
}

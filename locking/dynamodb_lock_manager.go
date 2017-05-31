package locking

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"encoding/json"
	"fmt"
	"encoding/hex"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type DynamoDBLockManager struct {
	DB        dynamodbiface.DynamoDBAPI
	LockTable string
}

func NewDynamoDBLockManager(lockTable string, p client.ConfigProvider) *DynamoDBLockManager {
	return &DynamoDBLockManager{
		DB:        dynamodb.New(p),
		LockTable: lockTable,
	}
}

func (d *DynamoDBLockManager) TryLock(run Run) (TryLockResponse, error) {
	var r TryLockResponse
	newRunSerialized, err := d.serialize(run)
	if err != nil {
		return r, errors.Wrap(err, "serializing run data")
	}

	// check if there is an existing lock
	getItemParams := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockID": {
				B: run.StateKey(),
			},
		},
		TableName: aws.String(d.LockTable),
		ConsistentRead: aws.Bool(true),
	}
	item, err := d.DB.GetItem(getItemParams)
	if err != nil {
		return r, errors.Wrap(err, "checking if lock exists")
	}

	// if there is already a lock then we can't acquire a lock. Return the existing lock
	if len(item.Item) != 0 {
		runAttr, ok := item.Item["Run"]
		if !ok || runAttr == nil || len(runAttr.B) == 0 {
			return r, fmt.Errorf("found an existing lock at that id but it did not contain expected 'Run' key. We suggest manually deleting this key from DynamoDB")
		}
		var lockingRun Run
		if err := d.deserialize(runAttr.B, &lockingRun); err != nil {
			return r, errors.Wrap(err, "deserializing existing lock")
		}
		return TryLockResponse{
			LockAcquired: false,
			LockingRun: lockingRun,
			LockID: string(hex.EncodeToString(run.StateKey())),
		}, nil
	}

	// else we should be able to lock
	putItem := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"LockID": {B: run.StateKey()},
			"Run":   {B: newRunSerialized},
		},
		TableName:           aws.String(d.LockTable),
		// this will ensure that we don't insert the new item in a race situation
		// where someone has written this key just after our read
		ConditionExpression: aws.String("attribute_not_exists(LockID)"),
	}
	if _, err := d.DB.PutItem(putItem); err != nil {
		return r, errors.Wrap(err, "writing lock")
	}
	return TryLockResponse{
		LockAcquired: true,
		LockingRun: run,
		LockID: string(hex.EncodeToString(run.StateKey())),
	}, nil
}

func (d *DynamoDBLockManager) Unlock(lockID string) error {
	idAsBytes, err := hex.DecodeString(lockID)
	if err != nil {
		return errors.Wrap(err, "id was not in correct format")
	}

	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockID": {B: idAsBytes},
		},
		TableName: aws.String(d.LockTable),
	}
	_, err = d.DB.DeleteItem(params)
	return errors.Wrap(err, "deleting lock")
}

func (d *DynamoDBLockManager) ListLocks() (map[string]Run, error) {
	m := make(map[string]Run)
	params := &dynamodb.ScanInput{
		ProjectionExpression: aws.String("LockID,Run"),
		TableName: aws.String(d.LockTable),
	}

	// loop to get all locks since if datasize is over 1MB the client will page.
	// we're setting a counter here just in case something goes horribly wrong and we loop forever
	var i int
	var startKey map[string]*dynamodb.AttributeValue
	for ; i < 1000; i++ {
		params.SetExclusiveStartKey(startKey)
		scanOut, err := d.DB.Scan(params)
		if err != nil {
			return m, errors.Wrap(err, "reading dynamodb")
		}
		for _, item := range scanOut.Items {
			lockIDItem, ok := item["LockID"]
			if !ok || lockIDItem == nil {
				return m, fmt.Errorf("lock did not have expected key 'LockID'")
			}
			lockID := string(hex.EncodeToString(lockIDItem.B))
			runItem, ok := item["Run"]
			if !ok || runItem == nil {
				return m, fmt.Errorf("lock did not have expected key 'Run'")
			}

			var run Run
			if err := d.deserialize(runItem.B, &run); err != nil {
				return m, fmt.Errorf("deserializing run at key %q: %s", lockID, err)
			}
			m[lockID] = run
		}
		startKey = scanOut.LastEvaluatedKey

		// if there are no more pages then we're done
		if len(startKey) == 0 {
			return m, nil
		}
	}
	return m, errors.New("maxed out at 1000 scan iterations on the DynamoDB table. Something must be wrong")
}

func (d *DynamoDBLockManager) deserialize(bs []byte, run *Run) error {
	return json.Unmarshal(bs, run)
}

func (d *DynamoDBLockManager) serialize(run Run) ([]byte, error) {
	return json.Marshal(run)
}

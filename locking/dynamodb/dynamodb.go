package dynamodb

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/hootsuite/atlantis/models"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type Backend struct {
	DB        dynamodbiface.DynamoDBAPI
	LockTable string
}

// dynamoLock duplicates the fields of some models and adds LocksKey.
// This is so everything is a top-level field for serialization and querying
// and also so any changes to models won't affect
// how we're storing our data (or will at least cause a compile error)
type dynamoLock struct {
	LockKey      string
	RepoFullName string
	Path         string
	PullNum        int
	PullHeadCommit string
	PullBaseCommit string
	PullURL        string
	PullBranch     string
	PullAuthor     string
	UserUsername string
	Env          string
	Time         time.Time
}

func New(lockTable string, p client.ConfigProvider) Backend {
	return Backend{
		DB:        dynamodb.New(p),
		LockTable: lockTable,
	}
}

func (b Backend) key(project models.Project, env string) string {
	return fmt.Sprintf("%s/%s/%s", project.RepoFullName, project.Path, env)
}

func (b Backend) TryLock(newLock models.ProjectLock) (bool, models.ProjectLock, error) {
	var currLock models.ProjectLock
	key := b.key(newLock.Project, newLock.Env)
	newDynamoLock := b.toDynamo(key, newLock)
	newLockSerialized, err := dynamodbattribute.MarshalMap(newDynamoLock)
	if err != nil {
		return false, currLock, errors.Wrap(err, "serializing")
	}

	// check if there is an existing lock
	getItemParams := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockKey": {
				S: aws.String(key),
			},
		},
		TableName:      aws.String(b.LockTable),
		ConsistentRead: aws.Bool(true),
	}
	item, err := b.DB.GetItem(getItemParams)
	if err != nil {
		return false, currLock, errors.Wrap(err, "checking if lock exists")
	}

	// if there is already a lock then we can't acquire a lock. Return the existing lock
	var currDynamoLock dynamoLock
	if len(item.Item) != 0 {
		if err := dynamodbattribute.UnmarshalMap(item.Item, &currDynamoLock); err != nil {
			return false, currLock, errors.Wrap(err, "found an existing lock at that key but it could not be deserialized. We suggest manually deleting this key from DynamoDB")
		}
		return false, b.fromDynamo(currDynamoLock), nil
	}

	// else we should be able to lock
	putItem := &dynamodb.PutItemInput{
		Item:      newLockSerialized,
		TableName: aws.String(b.LockTable),
		// this will ensure that we don't insert the new item in a race situation
		// where someone has written this key just after our read
		ConditionExpression: aws.String("attribute_not_exists(LockKey)"),
	}
	if _, err := b.DB.PutItem(putItem); err != nil {
		return false, currLock, errors.Wrap(err, "writing lock")
	}
	return true, newLock, nil
}

func (b Backend) Unlock(project models.Project, env string) (*models.ProjectLock, error) {
	key := b.key(project, env)
	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockKey": {S: aws.String(key)},
		},
		TableName: aws.String(b.LockTable),
		ReturnValues: aws.String("ALL_OLD"),
	}
	output, err := b.DB.DeleteItem(params)
	if err != nil {
		return nil, errors.Wrap(err, "deleting lock")
	}

	// deserialize the lock so we can return it
	var dLock dynamoLock
	if err := dynamodbattribute.UnmarshalMap(output.Attributes, &dLock); err != nil {
		return nil, errors.Wrap(err, "found an existing lock at that key but it could not be deserialized. We suggest manually deleting this key from DynamoDB")
	}
	lock := b.fromDynamo(dLock)
	return &lock, nil
}

func (b Backend) List() ([]models.ProjectLock, error) {
	var locks []models.ProjectLock
	var err, internalErr error
	params := &dynamodb.ScanInput{
		TableName: aws.String(b.LockTable),
	}
	err = b.DB.ScanPages(params, func(out *dynamodb.ScanOutput, lastPage bool) bool {
		var dynamoLocks []dynamoLock
		if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &dynamoLocks); err != nil {
			internalErr = errors.Wrap(err, "deserializing locks")
			return false
		}
		for _, lock := range dynamoLocks {
			locks = append(locks, b.fromDynamo(lock))
		}
		return lastPage
	})

	if err == nil && internalErr != nil {
		err = internalErr
	}
	return locks, errors.Wrap(err, "scanning dynamodb")
}

func (b Backend) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	params := &dynamodb.ScanInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pullNum": {
				N: aws.String(strconv.Itoa(pullNum)),
			},
			":repoFullName": {
				S: aws.String(repoFullName),
			},
		},
		FilterExpression: aws.String("RepoFullName = :repoFullName and PullNum = :pullNum"),
		TableName:        aws.String(b.LockTable),
	}

	// scan DynamoDB for locks that match the pull request
	var dLocks []dynamoLock
	var locks []models.ProjectLock
	var err, internalErr error
	err = b.DB.ScanPages(params, func(out *dynamodb.ScanOutput, lastPage bool) bool {
		if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &dLocks); err != nil {
			internalErr = errors.Wrap(err, "deserializing locks")
			return false
		}
		return lastPage
	})
	if err == nil {
		err = internalErr
	}
	if err != nil {
		return locks, errors.Wrap(err, "scanning dynamodb")
	}

	// now we can unlock all of them
	for _, lock := range dLocks {
		if _, err := b.Unlock(models.NewProject(lock.RepoFullName, lock.Path), lock.Env); err != nil {
			return locks, errors.Wrapf(err, "unlocking repo %s, path %s, env %s", lock.RepoFullName, lock.Path, lock.Env)
		}
		locks = append(locks, b.fromDynamo(lock))
	}
	return locks, nil
}

func (b Backend) toDynamo(key string, l models.ProjectLock) dynamoLock {
	return dynamoLock{
		LockKey:      key,
		RepoFullName: l.Project.RepoFullName,
		Path:         l.Project.Path,
		PullNum:      l.Pull.Num,
		PullHeadCommit: l.Pull.HeadCommit,
		PullBaseCommit: l.Pull.BaseCommit,
		PullURL: l.Pull.URL,
		PullBranch: l.Pull.Branch,
		PullAuthor: l.Pull.Author,
		UserUsername: l.User.Username,
		Env:          l.Env,
		Time:         time.Now(),
	}
}

func (b Backend) fromDynamo(d dynamoLock) models.ProjectLock {
	return models.ProjectLock{
		Pull: models.PullRequest{
			Author: d.PullAuthor,
			Branch: d.PullBranch,
			URL: d.PullURL,
			BaseCommit: d.PullBaseCommit,
			HeadCommit: d.PullHeadCommit,
			Num: d.PullNum,
		},
		User: models.User{
			Username: d.UserUsername,
		},
		Project: models.Project{
			RepoFullName: d.RepoFullName,
			Path: d.Path,
		},
		Time: d.Time,
		Env: d.Env,
	}
}

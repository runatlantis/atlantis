// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements namespace initialization and validation (design §"Data
// model", §708 step 4). A fresh empty namespace is initialized atomically with a
// schema marker and a deployment record carrying the coordination epoch. An
// existing namespace is validated: an incompatible schema, mismatched deployment
// ID, or data without the schema marker refuses startup. Absence is never taken
// as successful completion of a partially started operation.
package etcd

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	schemaKind = "schema"
	schemaV1   = 1

	deploymentKind = "deployment"
	deploymentV1   = 1

	// product identifies these records as Atlantis so a foreign namespace is
	// rejected rather than misread.
	product = "atlantis"
)

// schemaRecord marks the namespace as Atlantis-owned at a known schema version.
type schemaRecord struct {
	Product string `json:"product"`
	Version int    `json:"version"`
}

// deploymentRecord binds the configured deployment ID to an opaque coordination
// epoch UUID. Every claim, command admission, and execution barrier includes the
// epoch (design §632). The epoch rotates on every migration and restore.
type deploymentRecord struct {
	DeploymentID string `json:"deployment_id"`
	Epoch        string `json:"epoch"`
}

// InitOrValidateNamespace atomically initializes a genuinely empty namespace, or
// validates an existing one, returning the coordination epoch to use for this
// process. It fails closed on any incompatibility (design §708 steps 4-6).
func InitOrValidateNamespace(ctx context.Context, kv clientv3.KV, keys Keyspace, deploymentID string) (string, error) {
	schemaKey := keys.SchemaKey()
	deployKey := keys.DeploymentKey()

	resp, err := kv.Get(ctx, schemaKey)
	if err != nil {
		return "", fmt.Errorf("reading schema marker: %w", err)
	}
	if len(resp.Kvs) > 0 {
		return validateExisting(ctx, kv, keys, deploymentID, resp.Kvs[0].Value)
	}

	// No schema marker. The namespace must be genuinely empty before we may
	// initialize it; any data key without the schema marker is refused.
	empty, err := namespaceEmpty(ctx, kv, keys)
	if err != nil {
		return "", err
	}
	if !empty {
		return "", errors.New("etcd namespace contains data but no Atlantis schema marker; refusing startup (an interrupted migration/restore or a foreign namespace)")
	}

	epoch := uuid.NewString()
	schemaVal, err := encodeValue(schemaKind, schemaV1, schemaRecord{Product: product, Version: SchemaVersion1})
	if err != nil {
		return "", err
	}
	deployVal, err := encodeValue(deploymentKind, deploymentV1, deploymentRecord{DeploymentID: deploymentID, Epoch: epoch})
	if err != nil {
		return "", err
	}

	// Initialize atomically: commit only if neither key was created by a racing
	// process between our empty check and now.
	txn, err := kv.Txn(ctx).
		If(
			clientv3.Compare(clientv3.CreateRevision(schemaKey), "=", 0),
			clientv3.Compare(clientv3.CreateRevision(deployKey), "=", 0),
		).
		Then(
			clientv3.OpPut(schemaKey, string(schemaVal)),
			clientv3.OpPut(deployKey, string(deployVal)),
		).
		Commit()
	if err != nil {
		return "", fmt.Errorf("initializing namespace: %w", err)
	}
	if !txn.Succeeded {
		// A concurrent initializer won; validate against what it wrote.
		got, gerr := kv.Get(ctx, schemaKey)
		if gerr != nil || len(got.Kvs) == 0 {
			return "", errors.New("namespace initialization raced and could not be validated")
		}
		return validateExisting(ctx, kv, keys, deploymentID, got.Kvs[0].Value)
	}
	return epoch, nil
}

// SchemaVersion1 is the integer schema version stored in the marker. It is kept
// separate from the string key-prefix version (SchemaVersion) so the two can
// evolve independently.
const SchemaVersion1 = 1

func validateExisting(ctx context.Context, kv clientv3.KV, keys Keyspace, deploymentID string, schemaData []byte) (string, error) {
	var sr schemaRecord
	if err := decodeValue(schemaData, schemaKind, schemaV1, schemaV1, &sr); err != nil {
		return "", fmt.Errorf("validating schema marker: %w", err)
	}
	if sr.Product != product {
		return "", fmt.Errorf("etcd namespace is not an Atlantis namespace (product %q)", sr.Product)
	}
	if sr.Version != SchemaVersion1 {
		return "", fmt.Errorf("etcd schema version %d is not readable by this binary (expected %d)", sr.Version, SchemaVersion1)
	}

	dresp, err := kv.Get(ctx, keys.DeploymentKey())
	if err != nil {
		return "", fmt.Errorf("reading deployment record: %w", err)
	}
	if len(dresp.Kvs) == 0 {
		return "", errors.New("etcd namespace has a schema marker but no deployment record; refusing startup")
	}
	var dr deploymentRecord
	if err := decodeValue(dresp.Kvs[0].Value, deploymentKind, deploymentV1, deploymentV1, &dr); err != nil {
		return "", fmt.Errorf("validating deployment record: %w", err)
	}
	if dr.DeploymentID != deploymentID {
		return "", fmt.Errorf("etcd namespace belongs to deployment %q but this process is configured for %q", dr.DeploymentID, deploymentID)
	}
	if dr.Epoch == "" {
		return "", errors.New("deployment record has an empty coordination epoch")
	}
	return dr.Epoch, nil
}

// namespaceEmpty reports whether no keys exist below the normalized prefix
// (design §635: genuinely empty).
func namespaceEmpty(ctx context.Context, kv clientv3.KV, keys Keyspace) (bool, error) {
	prefix := keys.Root() + "/"
	resp, err := kv.Get(ctx, prefix,
		clientv3.WithRange(clientv3.GetPrefixRangeEnd(prefix)),
		clientv3.WithLimit(1),
		clientv3.WithKeysOnly(),
	)
	if err != nil {
		return false, fmt.Errorf("checking namespace emptiness: %w", err)
	}
	return len(resp.Kvs) == 0, nil
}

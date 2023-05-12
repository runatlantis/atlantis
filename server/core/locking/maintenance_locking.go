package locking

import (
	"errors"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
)

//go:generate pegomock generate -m --package mocks -o mocks/mock_maintenance_lock_checker.go MaintenanceLockChecker

// MaintenanceLockChecker is an implementation of the global maintenance lock retrieval.
// It returns an object that contains information about maintenance locks status.
type MaintenanceLockChecker interface {
	CheckMaintenanceLock() (MaintenanceCommandLock, error)
}

//go:generate pegomock generate -m --package mocks -o mocks/mock_maintenance_locker.go MaintenanceLocker

// MaintenanceLocker interface that manages locks for maintenance command runner
type MaintenanceLocker interface {
	// LockMaintenance creates a lock for MaintenanceCommand if lock already exists it will
	// return existing lock without any changes
	LockMaintenance() (MaintenanceCommandLock, error)
	// UnlockMaintenance deletes maintenance lock created by LockMaintenance if present, otherwise
	// it is a no-op
	UnlockMaintenance() error
	MaintenanceLockChecker
}

// MaintenanceCommandLock contains information about maintenance command lock status.
type MaintenanceCommandLock struct {
	// Locked is true is when maintenance commands are locked
	// Either by using EnableMaintenance flag or creating a global MaintenanceCommandLock
	// EnableMaintenance lock take precedence when set
	Locked  bool
	Time    time.Time
	Failure string
}

type MaintenanceClient struct {
	backend               Backend
	enableMaintenanceFlag bool
}

func NewMaintenanceClient(backend Backend, enableMaintenanceFlag bool) MaintenanceLocker {
	return &MaintenanceClient{
		backend:               backend,
		enableMaintenanceFlag: enableMaintenanceFlag,
	}
}

// LockMaintenance acquires global maintenance lock.
// EnableMaintenanceFlag takes presedence to any existing locks, if it is set to true
// this function returns an error
func (c *MaintenanceClient) LockMaintenance() (MaintenanceCommandLock, error) {
	response := MaintenanceCommandLock{}

	if c.enableMaintenanceFlag {
		return response, errors.New("EnableMaintenanceFlag is set; Maintenance commands are locked globally until flag is unset")
	}

	maintenanceCmdLock, err := c.backend.LockCommand(command.Maintenance, time.Now())
	if err != nil {
		return response, err
	}

	if maintenanceCmdLock != nil {
		response.Locked = true
		response.Time = maintenanceCmdLock.LockTime()
	}
	return response, nil
}

// UnlockMaintenance releases a global maintenance lock.
// EnableMaintenanceFlag takes presedence to any existing locks, if it is set to true
// this function returns an error
func (c *MaintenanceClient) UnlockMaintenance() error {
	if c.enableMaintenanceFlag {
		return errors.New("maintenance commands are disabled until EnableMaintenance flag is unset")
	}

	err := c.backend.UnlockCommand(command.Maintenance)
	if err != nil {
		return err
	}

	return nil
}

// CheckMaintenanceLock retrieves an maintenance command lock if present.
// If EnableMaintenanceFlag is set it will always return a lock.
func (c *MaintenanceClient) CheckMaintenanceLock() (MaintenanceCommandLock, error) {
	response := MaintenanceCommandLock{}

	if c.enableMaintenanceFlag {
		return MaintenanceCommandLock{
			Locked: true,
		}, nil
	}

	maintenanceCmdLock, err := c.backend.CheckCommandLock(command.Maintenance)
	if err != nil {
		return response, err
	}

	if maintenanceCmdLock != nil {
		response.Locked = true
		response.Time = maintenanceCmdLock.LockTime()
	}

	return response, nil
}

// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package tfclient

import (
	"errors"
	"os"
	"syscall"
)

// tryLockFile attempts a non-blocking flock on file. exclusive maps to
// LOCK_EX, otherwise LOCK_SH is used so multiple readers can hold the lock
// at once. EWOULDBLOCK/EAGAIN simply means someone else holds it.
func tryLockFile(file *os.File, exclusive bool) (bool, error) {
	operation := syscall.LOCK_SH | syscall.LOCK_NB
	if exclusive {
		operation = syscall.LOCK_EX | syscall.LOCK_NB
	}
	if err := syscall.Flock(int(file.Fd()), operation); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func unlockFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

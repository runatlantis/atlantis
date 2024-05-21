package utils

import "os"

// RemoveIgnoreNonExistent removes a file, ignoring if it doesn't exist.
func RemoveIgnoreNonExistent(file string) error {
	err := os.Remove(file)
	if err == nil || os.IsNotExist(err) {
		return nil
	}

	return err
}

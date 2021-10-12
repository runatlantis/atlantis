package testing

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory and returns its path along
// with a cleanup function to be called via defer, ex:
//   dir, cleanup := TempDir()
//   defer cleanup()
func TempDir(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "")
	Ok(t, err)
	return tmpDir, func() {
		os.RemoveAll(tmpDir) // nolint: errcheck
	}
}

// DirStructure creates a directory structure in a temporary directory.
// structure describes the dir structure. If the value is another map, then the
// key is the name of a directory. If the value is nil, then the key is the name
// of a file. If val is a string then key is a file name and val is the file's content.
// It returns the path to the temp directory containing the defined
// structure and a cleanup function to delete the directory.
// Example usage:
// 	versionConfig := `
//  terraform {
// 	  required_version = "= 0.12.8"
//  }
//  `
//	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
//		"pulldir": map[string]interface{}{
//			"project1": map[string]interface{}{
//				"main.tf": nil,
//			},
//			"project2": map[string]interface{}{,
//				"main.tf": versionConfig,
//			},
//		},
//	})
//  defer cleanup()
//
func DirStructure(t *testing.T, structure map[string]interface{}) (string, func()) {
	tmpDir, cleanup := TempDir(t)
	dirStructureGo(t, tmpDir, structure)
	return tmpDir, cleanup
}

func dirStructureGo(t *testing.T, parentDir string, structure map[string]interface{}) {
	for key, val := range structure {
		// If val is nil then key is a filename and we just create it
		if val == nil {
			_, err := os.Create(filepath.Join(parentDir, key))
			Ok(t, err)
			continue
		}
		// If val is another map then key is a dir
		if dirContents, ok := val.(map[string]interface{}); ok {
			subDir := filepath.Join(parentDir, key)
			Ok(t, os.Mkdir(subDir, 0700))
			// Recurse and create contents.
			dirStructureGo(t, subDir, dirContents)
		} else if fileContent, ok := val.(string); ok {
			// If val is a string then key is a file name and val is the file's content
			err := os.WriteFile(filepath.Join(parentDir, key), []byte(fileContent), 0600)
			Ok(t, err)
		}
	}
}

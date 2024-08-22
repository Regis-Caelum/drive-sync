package common

import (
	"os"
	"path/filepath"
	"strings"
)

func PathExist(absPath string) bool {
	_, err := os.Stat(absPath)
	return !os.IsNotExist(err)
}

func IsHiddenPath(path string) bool {
	// Check if the path or any segment of the path starts with a dot
	segments := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, segment := range segments {
		if strings.HasPrefix(segment, ".") {
			return true
		}
	}
	return false
}

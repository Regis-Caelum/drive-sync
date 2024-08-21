package common

import (
	"os"
)

func PathExist(absPath string) bool {
	_, err := os.Stat(absPath)
	return !os.IsNotExist(err)
}

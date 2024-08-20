package common

import (
	"os"
)

type FileStatus int

type ActionType int

const (
	UNMODIFIED FileStatus = iota
	MODIFIED
	UPLOADED
	NOT_UPLOADED
)

const (
	UPDATE_NODES ActionType = iota
	UPDATE_WATCHLIST
)

func PathExist(absPath string) bool {
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return false
	}
	return true
}

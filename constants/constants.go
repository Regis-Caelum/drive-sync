package constants

type FileStatus int

type ActionType int

const (
	UNMODIFIED FileStatus = iota
	MODIFIED
	UPLOADED
	NOT_UPLOADED
)

const (
	AddNodes ActionType = iota
	DeleteNodes
	AddWatchlist
	DeleteWatchlist
)

package model

import (
	"github.com/Regis-Caelum/drive-sync/cmd/dsync/constant"
)

type Node struct {
	ID           int                 `gorm:"primaryKey"`
	Name         string              `gorm:"not null"`
	IsDir        bool                `gorm:"not null"`
	FileStatus   constant.FileStatus `gorm:"not null"`
	UploadStatus constant.FileStatus `gorm:"not null"`
	AbsolutePath string              `gorm:"not null;unique"`
}

type WatchList struct {
	ID           int    `gorm:"primaryKey"`
	Name         string `gorm:"not null"`
	AbsolutePath string `gorm:"not null;unique"`
}

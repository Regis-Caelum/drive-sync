package models

import (
	"github.com/Regis-Caelum/drive-sync/constants"
)

type Node struct {
	ID           int                  `gorm:"primaryKey"`
	Name         string               `gorm:"not null"`
	IsDir        bool                 `gorm:"not null"`
	FileStatus   constants.FileStatus `gorm:"not null"`
	UploadStatus constants.FileStatus `gorm:"not null"`
	AbsolutePath string               `gorm:"not null;unique"`
}

type WatchList struct {
	ID           int    `gorm:"primaryKey"`
	Name         string `gorm:"not null"`
	AbsolutePath string `gorm:"not null;unique"`
}

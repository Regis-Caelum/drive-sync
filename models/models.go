package models

import "github.com/Regis-Caelum/drive-sync/common"

type Node struct {
	ID           int               `gorm:"primaryKey"`
	Name         string            `gorm:"not null"`
	IsDir        bool              `gorm:"not null"`
	FileStatus   common.FileStatus `gorm:"not null"`
	UploadStatus common.FileStatus `gorm:"not null"`
	AbsolutePath string            `gorm:"not null;unique"`
}

type WatchList struct {
	ID           int    `gorm:"primaryKey"`
	Name         string `gorm:"not null"`
	AbsolutePath string `gorm:"not null;unique"`
}

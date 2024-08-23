package database

import (
	"fmt"
	pb "github.com/Regis-Caelum/drive-sync/proto/generated"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var (
	DB *gorm.DB
)

func init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("./daemon/database/database.sqlite"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("failed to get SQL DB from GORM DB:", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetConnMaxIdleTime(0)
	err = DB.AutoMigrate(
		&pb.Node{},
		&pb.WatchList{},
		&pb.OAuth2Token{})
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func ClearDatabase(db *gorm.DB) {
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	for _, table := range tables {
		if table != "sqlite_sequence" {
			db.Exec("DROP TABLE IF EXISTS " + table)
			log.Printf("Dropped table: %s", table)
		}
	}
	resetSequences(db)
}

func resetSequences(db *gorm.DB) {
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	for _, table := range tables {
		if table != "sqlite_sequence" {
			db.Exec("DELETE FROM sqlite_sequence WHERE name = ?", table)
			log.Printf("Reset sequence for table: %s", table)
		}
	}
}

func GetTx() (*gorm.DB, error) {
	tx := DB.Begin()
	if tx.Error != nil {
		//dbMutex.Unlock()
		return nil, tx.Error
	}
	return tx, nil
}

func CommitTx(tx *gorm.DB) {
	err := tx.Commit().Error
	if err != nil {
		log.Println("Error committing transaction:", err)
	}
}

func RollbackTx(tx *gorm.DB) {
	tx.Rollback()
}

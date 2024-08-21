package database

import (
	"fmt"
	"github.com/Regis-Caelum/drive-sync/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"sync"
)

var (
	DB      *gorm.DB
	dbMutex *sync.Mutex
)

func init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("./database/database.sqlite"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	dbMutex = new(sync.Mutex)

	// Set up the connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("failed to get SQL DB from GORM DB:", err)
	}

	sqlDB.SetMaxOpenConns(10)   // SQLite should have only one open connection at a time
	sqlDB.SetMaxIdleConns(10)   // One idle connection (same as max open conns)
	sqlDB.SetConnMaxLifetime(0) // Connection lifetime - 0 means connections are reused forever
	sqlDB.SetConnMaxIdleTime(0) // Idle time - 0 means no limit on how long a connection can be idle

	// Migrate the schema
	err = DB.AutoMigrate(
		&models.Node{},
		&models.WatchList{})
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func ClearDatabase(db *gorm.DB) {
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	// Drop each table
	for _, table := range tables {
		if table != "sqlite_sequence" { // Ignore the internal SQLite sequence table
			db.Exec("DROP TABLE IF EXISTS " + table)
			log.Printf("Dropped table: %s", table)
		}
	}
	resetSequences(db)
}

func resetSequences(db *gorm.DB) {
	// Get the list of all tables
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	// Iterate over each table and reset the sequence
	for _, table := range tables {
		if table != "sqlite_sequence" { // Ignore the internal SQLite sequence table
			// Reset the sequence
			db.Exec("DELETE FROM sqlite_sequence WHERE name = ?", table)
			log.Printf("Reset sequence for table: %s", table)
		}
	}
}

func GetTx() (*gorm.DB, error) {
	//dbMutex.Lock()
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
	//dbMutex.Unlock() // Only unlock if transaction was successfully committed
}

func RollbackTx(tx *gorm.DB) {
	tx.Rollback()
}

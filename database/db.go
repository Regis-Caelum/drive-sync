package database

import (
	"fmt"
	"github.com/Regis-Caelum/drive-sync/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var (
	DB *gorm.DB
)

func init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("./database/database.sqlite"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

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

func GetTx() *gorm.DB {
	return DB.Begin()
}

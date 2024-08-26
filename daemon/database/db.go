package database

import (
	"errors"
	"fmt"
	pb "github.com/Regis-Caelum/drive-sync/proto/generated"
	"google.golang.org/protobuf/proto"
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
		&pb.OAuth2Token{},
		&pb.DriveRecord{})
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func ClearDatabase() {
	var tables []string
	DB.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	for _, table := range tables {
		if table != "sqlite_sequence" {
			if table == "o_auth2_tokens" {
				DB.Model(&pb.OAuth2Token{}).Where("table_name = ?", "o_auth2_tokens").Updates(map[string]interface{}{
					"root": "",
					"host": "",
				})
				continue
			}
			DB.Exec("DROP TABLE IF EXISTS " + table)

			log.Printf("Dropped table: %s", table)
		}
	}
	resetSequences()
}

func resetSequences() {
	var tables []string
	DB.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	for _, table := range tables {
		if table != "sqlite_sequence" {
			DB.Exec("DELETE FROM sqlite_sequence WHERE name = ?", table)
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

// CreateNode creates a new Node record in a transaction.
func CreateNode(node *pb.Node) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(node).Error
	})
}

// GetNode retrieves a Node record by ID in a transaction.
func GetNode(id int32) (*pb.Node, error) {
	var node *pb.Node
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.First(&node, id).Error
	})
	return node, err
}

// UpdateNode updates an existing Node record in a transaction.
func UpdateNode(node *pb.Node) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var existingNode pb.Node

		if err := tx.First(&existingNode, "id = ?", node.Id).Error; err != nil {
			return err
		}

		if proto.Equal(&existingNode, node) {
			return nil
		}

		return tx.Save(node).Error
	})
}

// DeleteNode deletes a Node record by ID in a transaction.
func DeleteNode(id int32) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&pb.Node{}, id).Error
	})
}

// ListAllNodes retrieves all Node records in a transaction.
func ListAllNodes() ([]*pb.Node, error) {
	var nodes []*pb.Node
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.Find(&nodes).Error
	})
	return nodes, err
}

// CRUD for WatchList

// CreateWatchList creates a new WatchList record in a transaction.
func CreateWatchList(watchList *pb.WatchList) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(watchList).Error
	})
}

// GetWatchList retrieves a WatchList record by ID in a transaction.
func GetWatchList(id int32) (*pb.WatchList, error) {
	var watchList *pb.WatchList
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.First(&watchList, id).Error
	})
	return watchList, err
}

// UpdateWatchList updates an existing WatchList record in a transaction.
func UpdateWatchList(watchList *pb.WatchList) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var existingWatchList pb.WatchList

		if err := tx.First(&existingWatchList, "id = ?", watchList.Id).Error; err != nil {
			return err
		}

		if proto.Equal(&existingWatchList, watchList) {
			return nil
		}

		return tx.Save(watchList).Error
	})
}

// DeleteWatchList deletes a WatchList record by ID in a transaction.
func DeleteWatchList(id int32) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&pb.WatchList{}, id).Error
	})
}

// ListAllWatchLists retrieves all WatchList records in a transaction.
func ListAllWatchLists() ([]*pb.WatchList, error) {
	var watchLists []*pb.WatchList
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.Find(&watchLists).Error
	})
	return watchLists, err
}

// CRUD for OAuth2Token

// CreateOAuth2Token creates a new OAuth2Token record in a transaction.
func CreateOAuth2Token(token *pb.OAuth2Token) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(token).Error
	})
}

// GetOAuth2Token retrieves an OAuth2Token record by ID in a transaction.
func GetOAuth2Token(id int32) (*pb.OAuth2Token, error) {
	var token pb.OAuth2Token
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.First(&token, id).Error
	})
	return &token, err
}

// UpdateOAuth2Token updates an existing OAuth2Token record in a transaction.
func UpdateOAuth2Token(token *pb.OAuth2Token) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Save(token).Error
	})
}

// DeleteOAuth2Token deletes an OAuth2Token record by ID in a transaction.
func DeleteOAuth2Token(id int32) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&pb.OAuth2Token{}, id).Error
	})
}

// ListAllOAuth2Tokens retrieves all OAuth2Token records in a transaction.
func ListAllOAuth2Tokens() ([]*pb.OAuth2Token, error) {
	var tokens []*pb.OAuth2Token
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.Find(&tokens).Error
	})
	return tokens, err
}

// CRUD for DriveRecord

// CreateDriveRecord creates a new DriveRecord record in a transaction.
func CreateDriveRecord(record *pb.DriveRecord) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var existingRecord pb.DriveRecord

		if err := tx.Where("local_path = ?", record.LocalPath).First(&existingRecord).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if existingRecord.Id != 0 {
			fmt.Println("Record already exists with local_path:", record.LocalPath)
			return fmt.Errorf("record already exists")
		}

		return tx.Create(record).Error
	})
}

// GetDriveRecordByLocalPath retrieves an DriveRecord record by ID in a transaction.
func GetDriveRecordByLocalPath(path string) (*pb.DriveRecord, error) {
	var record pb.DriveRecord
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("local_path = ?", path).First(&record).Error
	})
	return &record, err
}

// UpdateDriveRecord updates an existing DriveRecord record in a transaction.
func UpdateDriveRecord(record *pb.DriveRecord) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Save(record).Error
	})
}

// DeleteDriveRecord deletes an DriveRecord record by ID in a transaction.
func DeleteDriveRecord(id int32) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&pb.DriveRecord{}, id).Error
	})
}

// ListAllDriveRecord retrieves all DriveRecord records in a transaction.
func ListAllDriveRecord() ([]*pb.DriveRecord, error) {
	var records []*pb.DriveRecord
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.Find(&records).Error
	})
	return records, err
}

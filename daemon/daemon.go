package main

import (
	"errors"
	"fmt"
	"github.com/Regis-Caelum/drive-sync/daemon/common"
	"github.com/Regis-Caelum/drive-sync/daemon/database"
	pb "github.com/Regis-Caelum/drive-sync/proto/generated"
	"github.com/fsnotify/fsnotify"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
)

var daemonChannel chan bool
var watcher *fsnotify.Watcher
var token *pb.OAuth2Token

func init() {
	token = new(pb.OAuth2Token)
	daemonChannel = make(chan bool)

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		tx.First(token)
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if token.GetValue() != "" {
		gDriveSync()
	} else {
		fmt.Println("cannot sync files, no drive connected")
	}

	err = initializeNodes()
	if err != nil {
		log.Fatal(err)
	}

	err = initializeDeleteFloatingNodes()
	if err != nil {
		log.Fatal(err)
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = initializeWatchList()
	if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
		log.Fatal(err)
	}

}

func initializeWatchList() error {
	watchList, err := database.ListAllWatchLists()
	if err != nil {
		return err
	}

	for i := len(watchList) - 1; i >= 0; i-- {
		w := watchList[i]
		if !common.PathExist(w.GetAbsolutePath()) {
			err := database.DeleteWatchList(w.GetId())
			if err != nil {
				return err
			}
			// Remove the element by slicing
			watchList = append(watchList[:i], watchList[i+1:]...)
		} else {
			err := traverseDirHelper(w.GetAbsolutePath())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func initializeNodes() error {
	nodes, err := database.ListAllNodes()
	if err != nil {
		return err
	}

	for i := len(nodes) - 1; i >= 0; i-- {
		n := nodes[i]
		if !common.PathExist(n.GetAbsolutePath()) {
			err := database.DeleteNode(n.GetId())
			if err != nil {
				return err
			}
			// Remove the node by slicing
			nodes = append(nodes[:i], nodes[i+1:]...)
		}
	}

	return nil
}

func initializeDeleteFloatingNodes() error {
	nodes, err := database.ListAllNodes()
	if err != nil {
		return err
	}

	watchList, err := database.ListAllWatchLists()
	if err != nil {
		return err
	}

	driveRecords, err := database.ListAllDriveRecord()
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if !common.PathExist(n.GetAbsolutePath()) {
			err = database.DeleteNode(n.GetId())
			if err != nil {
				return err
			}
		}
	}

	for _, w := range watchList {
		if !common.PathExist(w.GetAbsolutePath()) {
			err = database.DeleteWatchList(w.GetId())
			if err != nil {
				return err
			}
		}
	}

	for _, d := range driveRecords {
		if !common.PathExist(d.GetLocalPath()) {
			gDriveDeleteFromDriveRecord(d)
			err = database.DeleteDriveRecord(d.GetId())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func traverseDirHelper(dirPath string) error {
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	isDir := fileInfo.IsDir()
	if isDir && common.IsHiddenPath(dirPath) {
		return fmt.Errorf("%s is a hidden path", dirPath)
	}
	if !isDir {
		// Node doesn't exist, create a new one
		node := &pb.Node{
			Name:         fileInfo.Name(),
			IsDir:        isDir,
			FileStatus:   pb.FILE_STATUS_MODIFIED,
			UploadStatus: pb.FILE_STATUS_NOT_UPLOADED,
			AbsolutePath: dirPath,
		}
		err = database.CreateNode(node)
		if err != nil {
			return err
		}
		gDriveSyncFile(node)
	} else {
		files, err := os.ReadDir(dirPath)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		if len(files) >= 0 {
			err = watcher.Add(dirPath)
			if err != nil {
				log.Println("Error:", err)
				return err
			}
			watchList := &pb.WatchList{
				Name:         fileInfo.Name(),
				AbsolutePath: dirPath,
			}
			err = database.CreateWatchList(watchList)
			if err != nil {
				log.Println("Error:", err)
			} else {
				gDriveSyncFolder(watchList)
			}

			for _, file := range files {
				err = traverseDirHelper(filepath.Join(dirPath, file.Name()))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func handleCreate(path string) {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			err = traverseDirHelper(path)
			if err != nil {
				log.Println("Error:", err)
			}
		} else {
			node := &pb.Node{
				Name:         info.Name(),
				IsDir:        info.IsDir(),
				FileStatus:   pb.FILE_STATUS_MODIFIED,
				UploadStatus: pb.FILE_STATUS_NOT_UPLOADED,
				AbsolutePath: path,
			}
			err = database.CreateNode(node)
			if err != nil {
				log.Println("Error:", err)
			}

			gDriveSyncFile(node)
		}
	}
}

func handleRemove(path string) {
	if _, err := os.Stat(path); err != nil {
		handleRename(path)
	}
}

func handleRename(oldPath string) {
	nodes, err := database.GetNodesWithPrefix("absolute_path", oldPath)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	err = database.DeleteNodeWithPrefix("absolute_path", oldPath)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	for _, n := range nodes {
		gDriveDeleteFiles(n)
	}

	err = database.DeleteDriveRecordsWithPrefix("local_path", oldPath)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	watchList, err := database.GetWatchListWithPrefix("absolute_path", oldPath)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	err = database.DeleteWatchListWithPrefix("absolute_path", oldPath)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	for _, w := range watchList {
		gDriveDeleteFolders(w)
	}

	err = database.DeleteDriveRecordsWithPrefix("local_path", oldPath)
	if err != nil {
		log.Println("Error:", err)
		return
	}

}

func handleEventDaemon(event fsnotify.Event) {
	if event.Op&fsnotify.Create == fsnotify.Create {
		fmt.Println("Directory/File created:", event.Name)
		handleCreate(event.Name)

	} else if event.Op&fsnotify.Remove == fsnotify.Remove {
		fmt.Println("Directory/File removed:", event.Name)
		handleRemove(event.Name)

	} else if event.Op&fsnotify.Rename == fsnotify.Rename {
		fmt.Println("Directory/File renamed or moved:", event.Name)
		handleRename(event.Name)

	} else if event.Op&fsnotify.Write == fsnotify.Write {
		fmt.Println("Directory/File modified:", event.Name)

	}
}

func daemon() {
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}(watcher)

	daemonChannel <- true

	for {
		select {
		case event := <-watcher.Events:
			go handleEventDaemon(event)
		case err := <-watcher.Errors:
			fmt.Println("Error:", err)
		}
	}
}

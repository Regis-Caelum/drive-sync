package main

import (
	"fmt"
	"github.com/Regis-Caelum/drive-sync/daemon/common"
	"github.com/Regis-Caelum/drive-sync/daemon/database"
	pb "github.com/Regis-Caelum/drive-sync/proto/generated"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type WatchListMap map[string]*pb.WatchList
type NodesMap map[string]*pb.Node
type SharedResources struct {
	nodesMap     NodesMap
	watchListMap WatchListMap
	mutex        *sync.Mutex
}

var daemonChannel chan bool
var fileMasterChannel chan pb.FILE_ACTIONS

var sharedResources *SharedResources
var watcher *fsnotify.Watcher
var token *pb.OAuth2Token

func initializeWatchList() error {
	sharedResources.watchListMap = make(WatchListMap)

	var watchList []*pb.WatchList
	result := database.DB.Find(&watchList)
	if result.Error != nil {
		return result.Error
	}
	sharedResources.mutex.Lock()
	for _, item := range watchList {
		sharedResources.watchListMap[item.AbsolutePath] = item
	}
	sharedResources.mutex.Unlock()

	for path := range sharedResources.watchListMap {
		if !common.PathExist(path) {
			sharedResources.mutex.Lock()
			delete(sharedResources.watchListMap, path)
			sharedResources.mutex.Unlock()
		}
	}

	fileMasterChannel <- pb.FILE_ACTIONS_DELETE_WATCHLIST

	return nil
}

func initializeNodes() error {
	sharedResources.nodesMap = make(NodesMap)

	var nodes []*pb.Node
	result := database.DB.Find(&nodes)
	if result.Error != nil {
		return result.Error
	}

	sharedResources.mutex.Lock()
	for _, item := range nodes {
		sharedResources.nodesMap[item.AbsolutePath] = item
	}
	sharedResources.mutex.Unlock()

	for path := range sharedResources.nodesMap {
		if !common.PathExist(path) {
			sharedResources.mutex.Lock()
			delete(sharedResources.nodesMap, path)
			sharedResources.mutex.Unlock()
		}
	}

	fileMasterChannel <- pb.FILE_ACTIONS_DELETE_NODES

	return nil
}

func init() {
	sharedResources = new(SharedResources)
	sharedResources.mutex = new(sync.Mutex)
	token = new(pb.OAuth2Token)
	fileMasterChannel = make(chan pb.FILE_ACTIONS, 10)
	daemonChannel = make(chan bool)

	tx, err := database.GetTx()
	log.Println("Transaction started")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	tx.First(token)
	database.CommitTx(tx)
	log.Println("Transaction Ended")

	if token.GetValue() != "" {
		gDriveSync()
	} else {
		fmt.Println("cannot sync files, no drive connected")
	}

	go fileUpdateDaemon()

	err = initializeNodes()
	if err != nil {
		log.Fatal(err)
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = initializeWatchList()
	if err != nil {
		log.Fatal(err)
	}

}

func (n NodesMap) addNodesMap() {
	for _, node := range n {
		err := database.CreateNode(node)
		if err != nil {
			log.Println(err)
		}
	}
	fmt.Println("Uploading files to drive")
	gDriveSyncFiles()
}

func (n NodesMap) deleteNodesMap() {
	tx, err := database.GetTx()
	log.Println("Transaction started")
	if err != nil {
		log.Fatal(err)
	}
	defer database.RollbackTx(tx)
	absolutePaths := make([]string, 0, len(n))
	for _, node := range n {
		absolutePaths = append(absolutePaths, node.AbsolutePath)
	}
	tx.Where("absolute_path NOT IN (?)", absolutePaths).Delete(&pb.Node{})
	database.CommitTx(tx)
	log.Println("Transaction Ended")
}

func (w WatchListMap) addWatchListMap() {
	for _, watchList := range w {
		err := watcher.Add(watchList.AbsolutePath)
		if err != nil {
			log.Println("Error: ", err, watchList.AbsolutePath)
		}
	}

	for _, watchList := range w {
		err := database.CreateWatchList(watchList)
		if err != nil {
			log.Println(err)
		}
	}
	fmt.Println("Creating folders in drive")
	gDriveSyncFolders()
}

func (w WatchListMap) deleteWatchListMap() {
	tx, err := database.GetTx()
	log.Println("Transaction started")
	if err != nil {
		log.Fatal(err)
	}
	defer database.RollbackTx(tx)

	absolutePaths := make([]string, 0, len(w))
	for _, node := range w {
		absolutePaths = append(absolutePaths, node.AbsolutePath)
	}
	tx.Where("absolute_path NOT IN (?)", absolutePaths).Delete(&pb.WatchList{})
	database.CommitTx(tx)
	log.Println("Transaction Ended")
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
		if _, exists := sharedResources.nodesMap[dirPath]; !exists {
			// Node doesn't exist, create a new one
			sharedResources.mutex.Lock()
			sharedResources.nodesMap[dirPath] = &pb.Node{
				Name:         fileInfo.Name(),
				IsDir:        isDir,
				FileStatus:   pb.FILE_STATUS_MODIFIED,
				UploadStatus: pb.FILE_STATUS_NOT_UPLOADED,
				AbsolutePath: dirPath,
			}
			sharedResources.mutex.Unlock()
		}
	} else {
		files, err := os.ReadDir(dirPath)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		if len(files) > 0 {
			if _, ok := sharedResources.watchListMap[dirPath]; !ok {
				sharedResources.mutex.Lock()
				sharedResources.watchListMap[dirPath] = &pb.WatchList{
					Name:         fileInfo.Name(),
					AbsolutePath: dirPath,
				}
				sharedResources.mutex.Unlock()
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
		if _, exists := sharedResources.watchListMap[path]; !exists && info.IsDir() {
			sharedResources.mutex.Lock()
			sharedResources.watchListMap[path] = &pb.WatchList{
				Name:         info.Name(),
				AbsolutePath: path,
			}
			sharedResources.mutex.Unlock()
			err = traverseDirHelper(path)
			if err != nil {
				log.Println("Error:", err)
			}
			err = watcher.Add(path)
			if err != nil {
				log.Println("Error:", err)
			}
			fileMasterChannel <- pb.FILE_ACTIONS_ADD_WATCHLIST
			fileMasterChannel <- pb.FILE_ACTIONS_ADD_NODES
		} else if _, exists = sharedResources.nodesMap[path]; !exists {
			sharedResources.mutex.Lock()
			sharedResources.nodesMap[path] = &pb.Node{
				Name:         info.Name(),
				IsDir:        info.IsDir(),
				FileStatus:   pb.FILE_STATUS_MODIFIED,
				UploadStatus: pb.FILE_STATUS_NOT_UPLOADED,
				AbsolutePath: path,
			}
			sharedResources.mutex.Unlock()
			fileMasterChannel <- pb.FILE_ACTIONS_ADD_NODES
		}
	}
}

func handleRemove(path string) {
	if _, ok := sharedResources.watchListMap[path]; ok {
		sharedResources.mutex.Lock()
		delete(sharedResources.watchListMap, path)
		sharedResources.mutex.Unlock()
		fileMasterChannel <- pb.FILE_ACTIONS_DELETE_WATCHLIST
		log.Println("Path deleted from watchlist: ", path)
	} else if _, ok = sharedResources.nodesMap[path]; ok {
		sharedResources.mutex.Lock()
		delete(sharedResources.nodesMap, path)
		sharedResources.mutex.Unlock()
		fileMasterChannel <- pb.FILE_ACTIONS_DELETE_NODES
		log.Println("Path deleted from nodes: ", path)
	}
}

func handleRename(oldPath string) {
	if _, exists := sharedResources.watchListMap[oldPath]; exists {
		for path := range sharedResources.watchListMap {
			if strings.HasPrefix(path, oldPath) {
				sharedResources.mutex.Lock()
				delete(sharedResources.watchListMap, oldPath)
				sharedResources.mutex.Unlock()
			}
		}
		for path := range sharedResources.nodesMap {
			if strings.HasPrefix(path, oldPath+"/") {
				sharedResources.mutex.Lock()
				delete(sharedResources.nodesMap, path)
				sharedResources.mutex.Unlock()
			}
		}
		fileMasterChannel <- pb.FILE_ACTIONS_DELETE_WATCHLIST
		fileMasterChannel <- pb.FILE_ACTIONS_DELETE_NODES
	} else {
		sharedResources.mutex.Lock()
		delete(sharedResources.nodesMap, oldPath)
		sharedResources.mutex.Unlock()
		fileMasterChannel <- pb.FILE_ACTIONS_DELETE_NODES
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

func fileUpdateDaemon() {
	for {
		select {
		case updateChan := <-fileMasterChannel:
			switch updateChan {
			case pb.FILE_ACTIONS_ADD_NODES:
				go sharedResources.nodesMap.addNodesMap()
			case pb.FILE_ACTIONS_ADD_WATCHLIST:
				go sharedResources.watchListMap.addWatchListMap()
			case pb.FILE_ACTIONS_DELETE_NODES:
				go sharedResources.nodesMap.deleteNodesMap()
			case pb.FILE_ACTIONS_DELETE_WATCHLIST:
				go sharedResources.watchListMap.deleteWatchListMap()
			default:
				continue
			}
		}
	}
}

func daemon() {
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}(watcher)

	for path := range sharedResources.watchListMap {
		err := traverseDirHelper(path)
		if err != nil {
			log.Println("Error: ", err)
		}
	}
	fileMasterChannel <- pb.FILE_ACTIONS_ADD_WATCHLIST
	fileMasterChannel <- pb.FILE_ACTIONS_ADD_NODES
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

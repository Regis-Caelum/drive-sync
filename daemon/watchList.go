package daemon

import (
	"errors"
	"fmt"
	"github.com/Regis-Caelum/drive-sync/common"
	"github.com/Regis-Caelum/drive-sync/database"
	"github.com/Regis-Caelum/drive-sync/models"
	"github.com/fsnotify/fsnotify"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type WatchListMap map[string]*models.WatchList
type NodesMap map[string]*models.Node
type SharedResources struct {
	nodesMap     NodesMap
	watchListMap WatchListMap
	mutex        *sync.Mutex
}

var sharedResources *SharedResources
var watcher *fsnotify.Watcher
var updateChanel chan common.ActionType

func initializeWatchList() error {
	// Initialize the map
	sharedResources.watchListMap = make(WatchListMap)

	// Retrieve all entries from the database
	var watchList []*models.WatchList
	result := database.DB.Find(&watchList)
	if result.Error != nil {
		return result.Error
	}
	sharedResources.mutex.Lock()
	for _, item := range watchList {
		sharedResources.watchListMap[item.AbsolutePath] = item
	}
	sharedResources.mutex.Unlock()

	return nil
}

func initializeNodes() error {
	sharedResources.nodesMap = make(NodesMap)

	// Retrieve all entries from the database
	var nodes []*models.Node
	result := database.DB.Find(&nodes)
	if result.Error != nil {
		return result.Error
	}

	sharedResources.mutex.Lock()
	for _, item := range nodes {
		sharedResources.nodesMap[item.AbsolutePath] = item
	}
	sharedResources.mutex.Unlock()

	return nil
}

func init() {
	sharedResources = new(SharedResources)
	sharedResources.mutex = new(sync.Mutex)

	err := initializeNodes()
	if err != nil {
		log.Fatal(err)
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	updateChanel = make(chan common.ActionType, 10)

	err = initializeWatchList()
	if err != nil {
		log.Fatal(err)
	}

}

func (n NodesMap) updateNodesMap() {
	tx, err := database.GetTx()
	if err != nil {
		log.Fatal(err)
	}
	defer database.RollbackTx(tx)

	absolutePaths := make([]string, 0, len(n))
	for _, node := range n {
		absolutePaths = append(absolutePaths, node.AbsolutePath)
		var existingNode models.Node
		result := tx.Where("absolute_path = ?", node.AbsolutePath).First(&existingNode)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Println(result.Error)
		}
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			result = tx.Create(node)
			if result.Error != nil {
				log.Println(result.Error)
			}
		} else {
			if existingNode.Name != node.Name &&
				existingNode.IsDir != node.IsDir &&
				existingNode.AbsolutePath != node.AbsolutePath &&
				existingNode.UploadStatus != node.UploadStatus &&
				existingNode.FileStatus != node.FileStatus &&
				existingNode.ID == node.ID {
				result = tx.Save(node)
				if result.Error != nil {
					log.Println(result.Error)
				}
			}
		}
	}

	tx.Where("absolute_path NOT IN (?)", absolutePaths).
		Delete(&models.Node{})
	database.CommitTx(tx)
}

func (w WatchListMap) updateWatchListMap() {
	tx, err := database.GetTx()
	if err != nil {
		log.Fatal(err)
	}
	defer database.RollbackTx(tx)

	currentWatchList := watcher.WatchList()
	currentWatchListMap := make(map[string]bool)
	for _, watch := range currentWatchList {
		currentWatchListMap[watch] = true
	}
	for watch, _ := range currentWatchListMap {
		if _, ok := sharedResources.watchListMap[watch]; !ok {
			err = watcher.Remove(watch)
			if err != nil {
				log.Println("Error: ", err)
			}
		}
	}
	for _, watchList := range w {
		if _, ok := currentWatchListMap[watchList.AbsolutePath]; !ok {
			err = watcher.Add(watchList.AbsolutePath)
			if err != nil {
				log.Println("Error: ", err, watchList.AbsolutePath)
			}
			err = traverseDirHelper(watchList.AbsolutePath)
			if err != nil {
				fmt.Println("Error: ", err)
				return
			}
			updateChanel <- common.UPDATE_NODES
			currentWatchListMap[watchList.AbsolutePath] = true
		}
	}

	for _, watchList := range w {
		var existingWatchListItem models.WatchList
		result := tx.Where("absolute_path = ?", watchList.AbsolutePath).First(&existingWatchListItem)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Println(result.Error)
		}
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			result = tx.Create(watchList)
			if result.Error != nil {
				log.Println(result.Error)
			}
		}
	}

	absolutePaths := make([]string, 0, len(w))
	for _, node := range w {
		absolutePaths = append(absolutePaths, node.AbsolutePath)
	}
	tx.Where("absolute_path NOT IN (?)", absolutePaths).Delete(&models.Node{})
	database.CommitTx(tx)

}

func TraverseDir(dirPath string) {
	err := traverseDirHelper(dirPath)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	updateChanel <- common.UPDATE_NODES
	updateChanel <- common.UPDATE_WATCHLIST
}

func traverseDirHelper(dirPath string) error {
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	isDir := fileInfo.IsDir()
	if isDir && isHiddenPath(dirPath) {
		return fmt.Errorf("%s is a hidden path", dirPath)
	}
	if !isDir {
		if _, exists := sharedResources.nodesMap[dirPath]; !exists {
			// Node doesn't exist, create a new one
			sharedResources.mutex.Lock()
			sharedResources.nodesMap[dirPath] = &models.Node{
				Name:         fileInfo.Name(),
				IsDir:        isDir,
				FileStatus:   common.FileStatus(0),
				UploadStatus: common.FileStatus(4),
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
				sharedResources.watchListMap[dirPath] = &models.WatchList{
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

func StartDaemon(wg *sync.WaitGroup) {
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			fmt.Println("Error: ", err)
		}
		wg.Done()
	}(watcher)

	go func() {
		for {
			select {
			case updateChan := <-updateChanel:
				switch updateChan {
				case common.UPDATE_NODES:
					go sharedResources.nodesMap.updateNodesMap()
				case common.UPDATE_WATCHLIST:
					go sharedResources.watchListMap.updateWatchListMap()
				default:
					continue
				}
			}
		}
	}()

	for _, watchList := range sharedResources.watchListMap {
		TraverseDir(watchList.AbsolutePath)
	}

	for {
		select {
		case event := <-watcher.Events:
			go handleEvent(event)
		case err := <-watcher.Errors:
			fmt.Println("Error:", err)
		}
	}
}

func handleEvent(event fsnotify.Event) {
	if event.Op&fsnotify.Create == fsnotify.Create {
		log.Println("File or directory created:", event.Name)
		handleCreate(event.Name)

	} else if event.Op&fsnotify.Remove == fsnotify.Remove {
		log.Println("File or directory removed:", event.Name)
		handleRemove(event.Name)

	} else if event.Op&fsnotify.Rename == fsnotify.Rename {
		log.Println("File or directory renamed or moved:", event.Name)
		handleRename(event.Name)

	} else if event.Op&fsnotify.Write == fsnotify.Write {
		log.Println("File or directory modified:", event.Name)
		// Handle file modifications if needed
	}
}

func AddDirToWatch(path string) {
	if info, err := os.Stat(path); !os.IsNotExist(err) {
		sharedResources.mutex.Lock()
		sharedResources.watchListMap[path] = &models.WatchList{
			Name:         info.Name(),
			AbsolutePath: path,
		}
		sharedResources.mutex.Unlock()
		updateChanel <- common.UPDATE_WATCHLIST
	} else {
		fmt.Printf("Path '%s' doesn't exist ", path)
	}
}

func handleCreate(path string) {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			TraverseDir(path)
			return
		}
	}
	TraverseDir(path)
}

func handleRemove(path string) {
	for watchList, _ := range sharedResources.watchListMap {
		if watchList == path || strings.HasPrefix(watchList, path) {
			sharedResources.mutex.Lock()
			delete(sharedResources.watchListMap, path)
			sharedResources.mutex.Unlock()
			log.Println("Path deleted from watchlist: ", path)
		}
	}
	updateChanel <- common.UPDATE_WATCHLIST
	for nodePath, _ := range sharedResources.nodesMap {
		if nodePath == path || strings.HasPrefix(nodePath, path+"/") {
			sharedResources.mutex.Lock()
			delete(sharedResources.nodesMap, nodePath)
			sharedResources.mutex.Unlock()
		}
	}
	updateChanel <- common.UPDATE_NODES
}

func handleRename(oldPath string) {
	handleRemove(oldPath) // Remove the old entry
	TraverseDir(oldPath)
}

func isHiddenPath(path string) bool {
	// Check if the path or any segment of the path starts with a dot
	segments := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, segment := range segments {
		if strings.HasPrefix(segment, ".") {
			return true
		}
	}
	return false
}

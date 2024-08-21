package daemon

import (
	"fmt"
	"github.com/Regis-Caelum/drive-sync/common"
	"github.com/Regis-Caelum/drive-sync/constants"
	"github.com/Regis-Caelum/drive-sync/database"
	"github.com/Regis-Caelum/drive-sync/models"
	"github.com/fsnotify/fsnotify"
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

var Channel chan bool
var sharedResources *SharedResources
var watcher *fsnotify.Watcher
var updateChannel chan constants.ActionType

func initializeWatchList() error {
	sharedResources.watchListMap = make(WatchListMap)

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

	for _, watch := range watchList {
		if common.PathExist(watch.AbsolutePath) {
			sharedResources.mutex.Lock()
			delete(sharedResources.watchListMap, watch.AbsolutePath)
			sharedResources.mutex.Unlock()
		}
	}

	return nil
}

func initializeNodes() error {
	sharedResources.nodesMap = make(NodesMap)

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

	for _, node := range nodes {
		if common.PathExist(node.AbsolutePath) {
			sharedResources.mutex.Lock()
			delete(sharedResources.nodesMap, node.AbsolutePath)
			sharedResources.mutex.Unlock()
		}
	}

	return nil
}

func init() {
	sharedResources = new(SharedResources)
	sharedResources.mutex = new(sync.Mutex)
	Channel = make(chan bool)

	err := initializeNodes()
	if err != nil {
		log.Fatal(err)
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	updateChannel = make(chan constants.ActionType, 10)

	err = initializeWatchList()
	if err != nil {
		log.Fatal(err)
	}

}

func (n NodesMap) addNodesMap() {
	tx, err := database.GetTx()
	if err != nil {
		log.Fatal(err)
	}
	defer database.RollbackTx(tx)

	for _, node := range n {
		result := tx.Create(node)
		if result.Error != nil {
			log.Println(result.Error)
		}
	}
	tx.Commit()
}

func (n NodesMap) deleteNodesMap() {
	tx, err := database.GetTx()
	if err != nil {
		log.Fatal(err)
	}
	defer database.RollbackTx(tx)
	absolutePaths := make([]string, 0, len(n))
	for _, node := range n {
		absolutePaths = append(absolutePaths, node.AbsolutePath)
	}
	tx.Where("absolute_path NOT IN (?)", absolutePaths).Delete(&models.Node{})
	database.CommitTx(tx)
}

func (w WatchListMap) addWatchListMap() {
	tx, err := database.GetTx()
	if err != nil {
		log.Fatal(err)
	}
	defer database.RollbackTx(tx)

	for _, watchList := range w {
		err = watcher.Add(watchList.AbsolutePath)
		if err != nil {
			log.Println("Error: ", err, watchList.AbsolutePath)
		}
	}

	for _, watchList := range w {
		result := tx.Create(watchList)
		if result.Error != nil {
			log.Println(result.Error)
		}
	}
	tx.Commit()

}

func (w WatchListMap) deleteWatchListMap() {
	tx, err := database.GetTx()
	if err != nil {
		log.Fatal(err)
	}
	defer database.RollbackTx(tx)

	absolutePaths := make([]string, 0, len(w))
	for _, node := range w {
		absolutePaths = append(absolutePaths, node.AbsolutePath)
	}
	tx.Where("absolute_path NOT IN (?)", absolutePaths).Delete(&models.WatchList{})
	database.CommitTx(tx)
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
				FileStatus:   constants.FileStatus(0),
				UploadStatus: constants.FileStatus(4),
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
			case updateChan := <-updateChannel:
				switch updateChan {
				case constants.AddNodes:
					go sharedResources.nodesMap.addNodesMap()
				case constants.AddWatchlist:
					go sharedResources.watchListMap.addWatchListMap()
				case constants.DeleteNodes:
					go sharedResources.nodesMap.deleteNodesMap()
				case constants.DeleteWatchlist:
					go sharedResources.watchListMap.deleteWatchListMap()
				default:
					continue
				}
			}
		}
	}()

	for _, watchList := range sharedResources.watchListMap {
		err := traverseDirHelper(watchList.AbsolutePath)
		if err != nil {
			log.Println("Error: ", err)
		}
	}
	Channel <- true

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
		err = traverseDirHelper(path)
		if err != nil {
			log.Println("Error:", err)
		}
		updateChannel <- constants.AddNodes
		updateChannel <- constants.AddWatchlist
	} else {
		fmt.Printf("Path '%s' doesn't exist ", path)
	}
}

func handleCreate(path string) {
	if info, err := os.Stat(path); err == nil {
		if _, exists := sharedResources.watchListMap[path]; !exists && info.IsDir() {
			sharedResources.mutex.Lock()
			sharedResources.watchListMap[path] = &models.WatchList{
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
			updateChannel <- constants.AddWatchlist
			updateChannel <- constants.AddNodes
		} else if _, exists = sharedResources.nodesMap[path]; !exists {
			sharedResources.mutex.Lock()
			sharedResources.nodesMap[path] = &models.Node{
				Name:         info.Name(),
				IsDir:        info.IsDir(),
				FileStatus:   constants.MODIFIED,
				UploadStatus: constants.NOT_UPLOADED,
				AbsolutePath: path,
			}
			sharedResources.mutex.Unlock()
			updateChannel <- constants.AddNodes
		}
	}
}

func handleRemove(path string) {
	if _, ok := sharedResources.watchListMap[path]; ok {
		sharedResources.mutex.Lock()
		delete(sharedResources.watchListMap, path)
		sharedResources.mutex.Unlock()
		updateChannel <- constants.DeleteWatchlist
		log.Println("Path deleted from watchlist: ", path)
	} else if _, ok = sharedResources.nodesMap[path]; ok {
		sharedResources.mutex.Lock()
		delete(sharedResources.nodesMap, path)
		sharedResources.mutex.Unlock()
		updateChannel <- constants.DeleteNodes
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
		updateChannel <- constants.DeleteWatchlist
		updateChannel <- constants.DeleteNodes
	} else {
		sharedResources.mutex.Lock()
		delete(sharedResources.nodesMap, oldPath)
		sharedResources.mutex.Unlock()
		updateChannel <- constants.DeleteNodes
	}
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

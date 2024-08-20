package main

import (
	"fmt"
	"github.com/Regis-Caelum/drive-sync/database"
	"github.com/Regis-Caelum/drive-sync/models"
	"github.com/Regis-Caelum/drive-sync/utils"
	"sync"
)

func main() {
	//database.ClearDatabase(database.DB)
	{
		dirPath := "/home/regis/Desktop/projects/models"
		var wg = new(sync.WaitGroup)
		wg.Add(1)
		go utils.WatchDirs(wg)
		utils.AddDirToWatch(dirPath)

		printNodesAndClosure()

		wg.Wait()
	}
}

func printNodesAndClosure() {
	var nodes []models.Node
	database.DB.Find(&nodes)
	fmt.Println("Nodes:")
	for _, node := range nodes {
		fmt.Printf("ID: %d, Name: %s, IsDir: %t, Path: %s\n", node.ID, node.Name, node.IsDir, node.AbsolutePath)
	}
}

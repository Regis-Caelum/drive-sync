package main

import (
	"bytes"
	"fmt"
	"github.com/Regis-Caelum/drive-sync/daemon"
	"github.com/Regis-Caelum/drive-sync/database"
	"github.com/Regis-Caelum/drive-sync/models"
	"os/exec"
	"sync"
)

func main() {
	//prepForTest()
	prepForProd()
}

func prepForTest() {
	database.ClearDatabase(database.DB)
	cmd := exec.Command("mkdir", "./models/text")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}
	fmt.Println(out.String())

	cmd = exec.Command("touch", "./models/text/text.txt")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}
	fmt.Println(out.String())

	cmd = exec.Command("mkdir", "./models/text/text")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}
	fmt.Println(out.String())

	cmd = exec.Command("touch", "./models/text/text/text.txt")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}
	fmt.Println(out.String())
}

func prepForProd() {
	//dirPath := "/home/regis/Desktop/projects/models"
	var wg = new(sync.WaitGroup)
	wg.Add(1)
	go daemon.StartDaemon(wg)
	<-daemon.Channel
	//daemon.AddDirToWatch(dirPath)

	printNodesAndClosure()

	wg.Wait()
}

func printNodesAndClosure() {
	var nodes []models.Node
	database.DB.Find(&nodes)
	fmt.Println("Nodes:")
	for _, node := range nodes {
		fmt.Printf("ID: %d, Name: %s, IsDir: %t, Path: %s\n", node.ID, node.Name, node.IsDir, node.AbsolutePath)
	}
}

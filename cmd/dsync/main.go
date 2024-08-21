package main

import (
	"bytes"
	"fmt"
	"github.com/Regis-Caelum/drive-sync/cmd/dsync/common"
	"github.com/Regis-Caelum/drive-sync/cmd/dsync/database"
	"github.com/Regis-Caelum/drive-sync/cmd/dsync/model"
	"github.com/Regis-Caelum/drive-sync/cmd/dsync/service"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

func main() {
	app := &cobra.Command{}
	app.Use = "dsync"
	app.Short = "Command line client for LXD"
	app.Long = common.FormatSection("Description",
		`Command line client for Drive Sync

Add folders to watch list and automatically upload changes to Google Drive. --help.`)
	app.SilenceUsage = true
	app.SilenceErrors = true
	app.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

	app.PersistentFlags().BoolP("help", "h", false, "Show this help message")

	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//prepForTest()
	//prepForProd()
}

func cmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "foo",
		Short: "Prints 'Foo' from package2",
		Run: func(cmd *cobra.Command, args []string) {
			service.StartDaemon()
		},
	}

	return cmd
}

func prepForTest() {
	database.ClearDatabase(database.DB)
	cmd := exec.Command("mkdir", "./model/text")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}
	fmt.Println(out.String())

	cmd = exec.Command("touch", "./model/text/text.txt")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}
	fmt.Println(out.String())

	cmd = exec.Command("mkdir", "./model/text/text")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}
	fmt.Println(out.String())

	cmd = exec.Command("touch", "./model/text/text/text.txt")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}
	fmt.Println(out.String())
}

func prepForProd() {
	//dirPath := "/home/regis/Desktop/projects/drive-sync/model"
	go service.StartDaemon()
	<-service.Channel
	//daemon.AddDirToWatch(dirPath)

	printNodesAndClosure()
}

func printNodesAndClosure() {
	var nodes []model.Node
	database.DB.Find(&nodes)
	fmt.Println("Nodes:")
	for _, node := range nodes {
		fmt.Printf("ID: %d, Name: %s, IsDir: %t, Path: %s\n", node.ID, node.Name, node.IsDir, node.AbsolutePath)
	}
}

package main

import (
	"fmt"
	"github.com/Regis-Caelum/drive-sync/utils"
)

func traverseDir(path string) error {
	tree, err := utils.BuildTree(path)
	if err != nil {
		return err
	}

	utils.PrintTree(tree, 0)
	return nil
}

func main() {
	dirPath := "./"

	err := traverseDir(dirPath)
	if err != nil {
		fmt.Printf("Error traversing directory: %v\n", err)
	}
}

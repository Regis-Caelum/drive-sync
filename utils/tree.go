package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type TreeNode struct {
	Name     string      `json:"name"`
	Inode    uint64      `json:"inode"`
	IsDir    bool        `json:"isDir"`
	Path     string      `json:"path"`
	Children []*TreeNode `json:"children,omitempty"`
}

func BuildTree(path string) (*TreeNode, error) {
	node := new(TreeNode)
	fileInfo, err := os.Stat(path)
	if err != nil {
		return node, err
	}

	inode, err := getFileInode(path)
	if err != nil {
		return node, err
	}

	node.Name = filepath.Base(path)
	node.IsDir = fileInfo.IsDir()
	node.Inode = inode
	node.Path = path

	if node.IsDir {
		entries, err := os.ReadDir(path)
		if err != nil {
			return node, err
		}

		if len(entries) == 0 {
			return node, nil
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			childPath := filepath.Join(path, entry.Name())
			childNode, err := BuildTree(childPath)
			if err != nil {
				return node, err
			}

			if childNode != nil && childNode.Name != "" {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node, nil
}

func getFileInode(filePath string) (uint64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to get inode information")
	}

	return stat.Ino, nil
}

func PrintTree(node *TreeNode, level int) {
	indent := strings.Repeat("  ", level)
	if node.IsDir {
		fmt.Printf("%sDirectory: %s, Inode: %d, Path: %s\n", indent, node.Name, node.Inode, node.Path)
	} else {
		fmt.Printf("%sFile: %s, Inode: %d, Path: %s\n", indent, node.Name, node.Inode, node.Path)
	}

	for _, child := range node.Children {
		PrintTree(child, level+1)
	}
}

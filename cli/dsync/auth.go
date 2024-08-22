package main

import (
	"fmt"
	"github.com/Regis-Caelum/drive-sync/cli/dsync/common"
	pb "github.com/Regis-Caelum/drive-sync/proto/generated"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type cmdAuth struct {
	global *cmdGlobal
}

func (c *cmdAuth) command() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "add"
	cmd.Short = "Connect to drive daemon"
	cmd.Long = common.FormatSection("Description", `Get watch lists from dsync daemon.`)

	getListCmd := cmdLogin{global: c.global, login: c}
	cmd.AddCommand(getListCmd.command())

	cmd.Args = cobra.NoArgs
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_ = cmd.Usage()
		return nil
	}
	return cmd
}

type cmdLogin struct {
	global *cmdGlobal
	login  *cmdAuth
}

func (c *cmdLogin) command() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = fmt.Sprint("dir <PATH> <PATH> ...")
	cmd.Short = "Get the directories that are being watched"

	cmd.RunE = c.run
	return cmd
}

func (c *cmdLogin) run(_ *cobra.Command, args []string) error {
	if len(args) <= 1 {
		fmt.Println("Insufficient arguments")
		return nil
	}

	path := args
	for idx, val := range args {
		if common.PathExist(val) && common.IsDir(val) {
			path = append(path[:idx], path[idx+1:]...)
		}
	}

	err := c.global.initGrpcClient()
	if err != nil {
		return err
	}
	defer c.global.closeGrpcClient()

	client := pb.NewWatchListServiceClient(c.global.conn)

	req := &pb.PathList{Values: path}
	resp, err := client.AddDirectoriesToWatchList(context.Background(), req)
	if err != nil {
		fmt.Println("Error: ", err)
		return fmt.Errorf("failed to connect to dsync daemon: %s", err)
	}
	directoryResponses := resp.GetValues()
	if len(directoryResponses) <= 0 {
		fmt.Println(common.FormatSection("No directories were added", `You can add directories to watchlist by using the dsync add -d <absolute_path_to_directory> command`))
		return nil
	}

	fmt.Println("Result:")
	var headers []string
	var rows [][]string

	headers = []string{
		"Path",
		"Status",
		"Error",
	}

	for _, dir := range directoryResponses {
		rows = append(rows, []string{dir.GetPath(), dir.GetStatus().String(), dir.GetError()})
	}

	common.PrintTable(headers, rows)

	return nil
}

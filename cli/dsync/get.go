package main

import (
	"fmt"
	"github.com/Regis-Caelum/drive-sync/cli/dsync/common"
	pb "github.com/Regis-Caelum/drive-sync/proto/generated"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type cmdGet struct {
	global *cmdGlobal
}

func (c *cmdGet) command() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "get"
	cmd.Short = "Connect to drive daemon"
	cmd.Long = common.FormatSection("Description", `Get watch lists from dsync daemon.`)

	getListCmd := cmdGetList{global: c.global, get: c}
	cmd.AddCommand(getListCmd.command())

	cmd.Args = cobra.NoArgs
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_ = cmd.Usage()
		return nil
	}
	return cmd
}

type cmdGetList struct {
	global *cmdGlobal
	get    *cmdGet

	flagDirOnly     bool
	flagFileOnly    bool
	flagUploaded    bool
	flagNotUploaded bool
	flagModified    bool
	flagUnmodified  bool
}

func (c *cmdGetList) command() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = fmt.Sprint("list")
	cmd.Short = "Get the directories that are being watched"

	cmd.RunE = c.run
	cmd.Flags().BoolVarP(&c.flagDirOnly, "dir-only", "d", false, "Filter directories")
	cmd.Flags().BoolVarP(&c.flagFileOnly, "file-only", "f", false, "Filter files")
	cmd.Flags().BoolVarP(&c.flagUploaded, "uploaded", "u", false, "Filter uploaded")
	cmd.Flags().BoolVarP(&c.flagNotUploaded, "not-uploaded", "n", false, "Filter not uploaded")
	cmd.Flags().BoolVarP(&c.flagModified, "modified", "m", false, "Filter modified")
	cmd.Flags().BoolVarP(&c.flagUnmodified, "unmodified", "s", false, "Filter unmodified")

	return cmd
}

func (c *cmdGetList) run(cmd *cobra.Command, args []string) error {
	if !c.flagDirOnly && !c.flagFileOnly &&
		!c.flagUploaded && !c.flagNotUploaded &&
		!c.flagModified && !c.flagUnmodified {
		_ = cmd.Usage()
		return nil
	}

	err := c.global.initGrpcClient()
	if err != nil {
		return err
	}
	defer c.global.closeGrpcClient()

	client := pb.NewWatchListServiceClient(c.global.conn)

	resp, err := client.GetWatchList(context.Background(), &pb.Empty{})
	if err != nil {
		fmt.Println("Error: ", err)
		return fmt.Errorf("failed to connect to dsync daemon: %s", err)
	}
	directoryList := resp.GetDirectoryList()
	fileList := resp.GetFileList()

	if len(directoryList) <= 0 {
		fmt.Println(common.FormatSection("No directories are being watched", `You can add directories to watchlist by using the dsync add -d <absolute_path_to_directory> command`))
		return nil
	}

	fmt.Println("WatchList:")
	var headers []string
	var rows [][]string

	headers = []string{
		"Name",
		"Directory",
		"Track",
		"Status",
		"Path",
	}

	if c.flagDirOnly {
		for _, dir := range directoryList {
			rows = append(rows, []string{dir.GetName(), "Yes", pb.FILE_STATUS_UNTRACKED.String(), pb.FILE_STATUS_UNTRACKED.String(), dir.GetAbsolutePath()})
		}
	}

	if c.flagFileOnly {
		for _, file := range fileList {
			isDir := "No"
			if file.GetIsDir() {
				isDir = "Yes"
			}
			rows = append(rows, []string{file.GetName(), isDir, file.GetFileStatus().String(), file.GetUploadStatus().String(), file.GetAbsolutePath()})
		}
	}

	var filteredRows [][]string
	for _, row := range rows {
		if ((c.flagModified && row[2] == pb.FILE_STATUS_MODIFIED.String()) ||
			(c.flagUnmodified && row[2] == pb.FILE_STATUS_UNMODIFIED.String()) ||
			(!c.flagModified && !c.flagUnmodified)) &&
			((c.flagUploaded && row[3] == pb.FILE_STATUS_UPLOADED.String()) ||
				(c.flagNotUploaded && row[3] == pb.FILE_STATUS_NOT_UPLOADED.String()) ||
				(!c.flagUploaded && !c.flagNotUploaded)) {
			filteredRows = append(filteredRows, row)
		}
	}
	rows = filteredRows

	//headers = []string{
	//	"Directory name",
	//	"Path",
	//}
	//for _, dir := range directoryList {
	//	rows = append(rows, []string{dir.GetName(), dir.GetAbsolutePath()})
	//}
	//
	//common.PrintTable(headers, rows)
	//
	//headers = []string{
	//	"File name",
	//	"Track",
	//	"Status",
	//	"Path",
	//}
	//for _, file := range fileList {
	//	rows = append(rows, []string{file.GetName(), file.GetFileStatus().String(), file.GetUploadStatus().String(), file.GetAbsolutePath()})
	//}

	common.PrintTable(headers, rows)

	return nil

}

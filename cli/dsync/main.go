package main

import (
	"fmt"
	"github.com/Regis-Caelum/drive-sync/cli/dsync/common"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type cmdGlobal struct {
	conn *grpc.ClientConn
}

func (g *cmdGlobal) initGrpcClient() error {
	var err error
	if g.conn == nil {

		g.conn, err = grpc.NewClient("unix:///tmp/dsync.sock", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Printf("Error: %s", err)
			return fmt.Errorf("failed to connect to dsync daemon: %s", err)
		}
	}

	return nil
}

func (g *cmdGlobal) closeGrpcClient() {
	var err error
	err = g.conn.Close()
	if err != nil {
		fmt.Printf("failed to close connection with dsync daemon: %s", err)
	}
	g.conn = nil
}

func main() {
	app := &cobra.Command{}
	app.Use = "dsync"
	app.Short = "Command line client for Drive Sync"
	app.Long = common.FormatSection("Description",
		`Command line client for automatically uploading directories to google drive

All of dsync's features can be driven through the various commands below.
For help with any of those, simply call them with --help.`)
	app.SilenceUsage = true
	app.SilenceErrors = true
	app.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

	globalCmd := &cmdGlobal{}
	app.PersistentFlags().BoolP("help", "h", false, "Print help")

	app.InitDefaultHelpCmd()

	getCmd := &cmdGet{global: globalCmd}
	app.AddCommand(getCmd.command())

	addCmd := &cmdAdd{global: globalCmd}
	app.AddCommand(addCmd.command())

	authCmd := &cmdAuth{global: globalCmd}
	app.AddCommand(authCmd.command())

	//authCmd := &cmdAuth{global: globalCmd}
	//app.AddCommand(authCmd.command())

	app.SetArgs([]string{"add", "dir", "/home/regis/Desktop/projects/test"})
	//app.SetArgs([]string{"get", "list", "-df"})
	//app.SetArgs([]string{"auth", "login"})
	_ = app.Execute()
}

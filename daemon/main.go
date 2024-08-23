package main

import (
	"context"
	"fmt"
	"github.com/Regis-Caelum/drive-sync/daemon/database"
	pb "github.com/Regis-Caelum/drive-sync/proto/generated"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

var srv *grpc.Server

type server struct {
	pb.UnimplementedWatchListServiceServer
	pb.UnimplementedAuthenticationServiceServer
}

func (s *server) SaveToken(ctx context.Context, in *pb.OAuth2Token) (*pb.Empty, error) {
	tx, err := database.GetTx()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("unabe to connect to database")
	}
	defer tx.Rollback()

	in.Id = 1
	tx.Save(in)
	database.CommitTx(tx)
	return &pb.Empty{}, nil
}

func (s *server) GetToken(ctx context.Context, in *pb.Empty) (*pb.OAuth2Token, error) {
	return token, nil
}

func (s *server) GetWatchList(ctx context.Context, in *pb.Empty) (*pb.FileList, error) {
	var watchList []*pb.WatchList
	var nodeList []*pb.Node
	var resp = &pb.FileList{}

	tx, err := database.GetTx()
	if err != nil {
		fmt.Println("Error:", err)
	}

	tx.Find(&watchList)
	tx.Find(&nodeList)

	resp.DirectoryList = watchList
	resp.FileList = nodeList
	return resp, nil
}

func (s *server) AddDirectoriesToWatchList(ctx context.Context, in *pb.PathList) (*pb.ResponseList, error) {
	resp := new(pb.ResponseList)
	for _, path := range in.GetValues() {
		if info, err := os.Stat(path); !os.IsNotExist(err) {
			fmt.Printf("Adding path %s to watchlist...", path)
			sharedResources.mutex.Lock()
			sharedResources.watchListMap[path] = &pb.WatchList{
				Name:         info.Name(),
				AbsolutePath: path,
			}
			sharedResources.mutex.Unlock()
			err = traverseDirHelper(path)
			if err != nil {
				fmt.Println("	✔❌")
				resp.Values = append(resp.Values, &pb.AddDirectoryResponse{
					Status: pb.ADD_DIRECTORY_STATUS_PARTIAL,
					Error:  err.Error(),
					Path:   path,
				})
			} else {
				resp.Values = append(resp.Values, &pb.AddDirectoryResponse{
					Status: pb.ADD_DIRECTORY_STATUS_COMPLETE,
					Error:  "nil",
					Path:   path,
				})
			}
			fmt.Println("	✔")
			fileMasterChannel <- pb.FILE_ACTIONS_ADD_NODES
			fileMasterChannel <- pb.FILE_ACTIONS_ADD_WATCHLIST
		} else {
			fmt.Println("	❌")
			resp.Values = append(resp.Values, &pb.AddDirectoryResponse{
				Status: pb.ADD_DIRECTORY_STATUS_FAILED,
				Error:  err.Error(),
				Path:   path,
			})
		}
	}
	return resp, nil
}

func init() {
	srv = grpc.NewServer()
	pb.RegisterWatchListServiceServer(srv, &server{})
	pb.RegisterAuthenticationServiceServer(srv, &server{})
}

func main() {
	fmt.Println("Starting watchlist daemon...")
	go daemon()
	<-daemonChannel
	fmt.Println("Watchlist daemon up and running.")

	socketPath := "/tmp/dsync.sock"
	if err := os.RemoveAll(socketPath); err != nil {
		panic(err)
	}

	listen, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Println("Starting gRPC server on Unix socket...")
	if err = srv.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

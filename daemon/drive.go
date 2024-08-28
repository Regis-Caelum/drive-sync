package main

import (
	"encoding/json"
	"fmt"
	"github.com/Regis-Caelum/drive-sync/daemon/database"
	pb "github.com/Regis-Caelum/drive-sync/proto/generated"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var gDriveClient *http.Client
var gDriveService *drive.Service

func gDriveSync() {
	ctx := context.Background()

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	gDriveClient, err = gDriveGetClient(config)
	if err != nil {
		fmt.Printf("Unable to get google drive client: %v", err)
		return
	}

	gDriveService, err = drive.NewService(ctx, option.WithHTTPClient(gDriveClient))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	if token.GetRoot() == "" {
		createdRootFolder, err := gDriveCreateFolder("Computers", []string{}, "")
		if err != nil {
			log.Fatalf("Unable to create root folder: %v", err)
		}

		tx, err := database.GetTx()
		log.Println("Transaction started")
		if err != nil {
			fmt.Printf("Unable to get transaction: %v", err)
			return
		}

		token.Root = createdRootFolder.Id
		tx.Save(token)
		database.CommitTx(tx)
		log.Println("Transaction Ended")
		database.RollbackTx(tx)
	}

	if token.GetHost() == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatalf("Unable to get hostname: %v", err)
		}

		createdHostFolder, err := gDriveCreateFolder(hostname, []string{token.GetRoot()}, "")
		if err != nil {
			log.Fatalf("Unable to create host folder: %v", err)
		}

		token.Host = createdHostFolder.Id

		tx, err := database.GetTx()
		log.Println("Transaction started")
		if err != nil {
			fmt.Printf("Unable to get transaction: %v", err)
			return
		}

		tx.Save(token)
		database.CommitTx(tx)
		log.Println("Transaction Ended")
		database.RollbackTx(tx)
	}

	gDriveSyncFolders()
	gDriveSyncFiles()

}

func gDriveSyncFolders() {
	fmt.Println("Syncing Folders:")
	watchList, _ := database.ListAllWatchLists()
	if len(watchList) != 0 {
		for _, w := range watchList {
			_, err := database.GetDriveRecordByLocalPath(w.GetAbsolutePath())
			if err != nil {
				gDriveSyncFolder(w)
			}
		}
	}
}

func gDriveSyncFolder(w *pb.WatchList) {
	descPath := ""
	pathParts := strings.Split(w.GetAbsolutePath(), "/")
	currentParentID := token.GetHost()
	for _, part := range pathParts {
		if part == "" {
			continue
		}
		descPath += "/" + part
		if rec, err := database.GetDriveRecordByLocalPath(descPath); err == nil {
			currentParentID = rec.DriveId
			continue
		}
		folderID, err := gDriveCreateFolder(part, []string{currentParentID}, descPath)
		if err != nil {
			fmt.Printf("Unable to create folder: %v", err)
			continue
		}
		log.Printf("Host folder created: %s (%s)\n", folderID.Name, folderID.Id)
		err = database.CreateDriveRecord(&pb.DriveRecord{
			Name:      part,
			LocalPath: descPath,
			DriveId:   folderID.Id,
			ParentId:  currentParentID,
		})
		if err != nil {
			fmt.Printf("Unable to update watch list: %v", err)
		}
		currentParentID = folderID.Id
	}
	w.DriveId = currentParentID
	err := database.UpdateWatchList(w)
	if err != nil {
		fmt.Printf("Unable to update watch list: %v", err)
	}

}

func gDriveSyncFiles() {
	fmt.Println("Syncing Files:")
	fileNodes, _ := database.ListAllNodes()
	if len(fileNodes) != 0 {
		for _, f := range fileNodes {
			if f.GetUploadStatus() == pb.FILE_STATUS_NOT_UPLOADED || f.GetFileStatus() == pb.FILE_STATUS_MODIFIED {
				gDriveSyncFile(f)
			}
		}
	}
}

func gDriveSyncFile(f *pb.Node) {
	descPath := ""
	pathParts := strings.Split(f.GetAbsolutePath(), "/")
	currentParentID := token.GetHost()
	for i, part := range pathParts {
		if part == "" {
			continue
		}
		descPath += "/" + part
		if rec, err := database.GetDriveRecordByLocalPath(descPath); err == nil {
			currentParentID = rec.DriveId
			continue
		}
		if i == len(pathParts)-1 {
			localFile, err := os.Open(f.GetAbsolutePath())
			if err != nil {
				log.Fatalf("Unable to open local file: %v", err)
			}
			fileID, err := gDriveCreateFile(part, []string{currentParentID}, descPath, localFile)
			_ = localFile.Close()
			if err != nil {
				fmt.Printf("Unable to create file: %v", err)
				continue
			}
			f.DriveId = fileID.Id
			f.FileStatus = pb.FILE_STATUS_UNMODIFIED
			f.UploadStatus = pb.FILE_STATUS_UPLOADED
			err = database.UpdateNode(f)
			if err != nil {
				fmt.Printf("Unable to update watch list: %v", err)
			}
			err = database.CreateDriveRecord(&pb.DriveRecord{
				Name:      part,
				LocalPath: descPath,
				DriveId:   fileID.Id,
				ParentId:  currentParentID,
			})
			if err != nil {
				fmt.Printf("Unable to update watch list: %v", err)
			}
			continue
		}
		folderID, err := gDriveCreateFolder(part, []string{currentParentID}, descPath)
		if err != nil {
			fmt.Printf("Unable to create folder: %v", err)
			continue
		}
		log.Printf("Host folder created: %s (%s)\n", folderID.Name, folderID.Id)
		err = database.CreateDriveRecord(&pb.DriveRecord{
			Name:      part,
			LocalPath: descPath,
			DriveId:   folderID.Id,
			ParentId:  currentParentID,
		})
		if err != nil {
			fmt.Printf("Unable to update watch list: %v", err)
		}
		currentParentID = folderID.Id
	}
}

func gDriveDeleteFolders(watchList *pb.WatchList) {
	fmt.Println("Deleting Folders:")
	err := gDriveService.Files.Delete(watchList.GetDriveId()).Context(context.Background()).Do()
	if err != nil {
		log.Printf("Failed to delete folder with ID %s, %s: %v", watchList.GetDriveId(), watchList.GetName(), err)
	} else {
		fmt.Printf("Successfully deleted folder with ID %s, %s\n", watchList.GetDriveId(), watchList.GetName())
	}
}

func gDriveDeleteFiles(node *pb.Node) {
	fmt.Println("Deleting Files:")

	err := gDriveService.Files.Delete(node.GetDriveId()).Context(context.Background()).Do()
	if err != nil {
		log.Printf("Failed to delete file with ID %s, %s: %v", node.GetDriveId(), node.GetName(), err)
	} else {
		fmt.Printf("Successfully deleted file with ID %s, %s\n", node.GetDriveId(), node.GetName())
	}
}

func gDriveDeleteFromDriveRecord(driveRecord *pb.DriveRecord) {
	fmt.Println("Deleting Files:")

	err := gDriveService.Files.Delete(driveRecord.GetDriveId()).Context(context.Background()).Do()
	if err != nil {
		log.Printf("Failed to delete file with ID %s, %s: %v", driveRecord.GetDriveId(), driveRecord.GetName(), err)
	} else {
		fmt.Printf("Successfully deleted file with ID %s, %s\n", driveRecord.GetDriveId(), driveRecord.GetName())
	}
}

func gDriveGetClient(config *oauth2.Config) (*http.Client, error) {
	tok := &oauth2.Token{}
	err := json.Unmarshal([]byte(token.GetValue()), tok)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	return config.Client(context.Background(), tok), nil
}

func gDriveCreateFolder(name string, parents []string, localPath string) (*drive.File, error) {
	var query string

	if len(parents) > 0 {
		query = fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false and mimeType = 'application/vnd.google-apps.folder'", name, parents[0])
	} else {
		query = fmt.Sprintf("name = '%s' and 'root' in parents and trashed = false and mimeType = 'application/vnd.google-apps.folder'", name)
	}
	r, err := gDriveService.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return nil, err
	}

	if len(r.Files) > 0 {
		return r.Files[0], nil
	}

	folder := &drive.File{
		Name:        name,
		MimeType:    "application/vnd.google-apps.folder",
		Parents:     parents,
		Description: localPath,
	}

	folder, err = gDriveService.Files.Create(folder).Do()
	if err != nil {
		return nil, err
	}
	return folder, nil
}

func gDriveCreateFile(name string, parents []string, localPath string, fileContent io.Reader) (*drive.File, error) {
	var query string

	ext := filepath.Ext(name)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream" // Default to a generic binary stream
	}

	if len(parents) > 0 {
		query = fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false and mimeType = '%s'", name, parents[0], mimeType)
	} else {
		query = fmt.Sprintf("name = '%s' and 'root' in parents and trashed = false and mimeType = '%s'", name, mimeType)
	}

	r, err := gDriveService.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return nil, err
	}

	if len(r.Files) > 0 {
		return r.Files[0], nil
	}

	file := &drive.File{
		Name:        name,
		MimeType:    mimeType,
		Parents:     parents,
		Description: localPath,
	}

	file, err = gDriveService.Files.Create(file).Media(fileContent).Do()
	if err != nil {
		return nil, err
	}

	fmt.Printf("File created: %s (%s)\n", file.Name, file.Id)
	return file, nil
}

func gDriveGetAllFolders() ([]*drive.File, error) {
	query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and '%s' in parents and trashed = false", token.GetHost())
	files, err := gDriveService.Files.List().
		Q(query).
		Fields("files(id, name)").
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve folders: %v", err)
	}

	return files.Files, nil
}

func gDriveGetAllFiles() ([]*drive.File, error) {
	query := fmt.Sprintf("'%s' in parents and trashed = false and mimeType != 'application/vnd.google-apps.folder'", token.GetHost())
	files, err := gDriveService.Files.List().
		Q(query).
		Fields("files(id, name)").
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve folders: %v", err)
	}

	return files.Files, nil
}

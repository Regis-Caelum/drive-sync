package main

import (
	"encoding/json"
	"fmt"
	"github.com/Regis-Caelum/drive-sync/daemon/database"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
)

var gDriveClient *http.Client
var gDriveService *drive.Service

func syncWithDrive() {
	ctx := context.Background()

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	gDriveClient, err = getGDriveClient(config)
	if err != nil {
		fmt.Printf("Unable to get google drive client: %v", err)
		return
	}

	gDriveService, err = drive.NewService(ctx, option.WithHTTPClient(gDriveClient))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	if token.GetRoot() == "" {
		createdRootFolder, err := gDriveCreateFolder("Computers", nil, "")
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

	r, err := gDriveGetChildren(token.GetHost())
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		watchList, _ := database.ListAllWatchLists()
		if len(watchList) != 0 {
			//for _, w := range watchList {
			//	//_, err = gDriveCreateFolder(w.Name)
			//}
		}
	} else {
		for _, i := range r.Files {
			fmt.Printf("%s (%s)\n", i.Name, i.Id)
		}
	}
}

func getGDriveClient(config *oauth2.Config) (*http.Client, error) {
	tok := &oauth2.Token{}
	err := json.Unmarshal([]byte(token.GetValue()), tok)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	return config.Client(context.Background(), tok), nil
}

func gDriveCreateFolder(name string, parents []string, localPath string) (*drive.File, error) {
	folder := &drive.File{
		Name:        name,
		MimeType:    "application/vnd.google-apps.folder",
		Parents:     parents,
		Description: localPath,
	}
	folder, err := gDriveService.Files.Create(folder).Do()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Host folder created: %s (%s)\n", folder.Name, folder.Id)
	return folder, nil
}

func gDriveGetChildren(parent string) (*drive.FileList, error) {
	query := fmt.Sprintf("'%s' in parents and trashed = false", parent)
	r, err := gDriveService.Files.List().Q(query).
		Fields("nextPageToken, files(id, name)").Do()
	return r, err
}

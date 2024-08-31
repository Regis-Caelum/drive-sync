package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	pb "github.com/Regis-Caelum/drive-sync/proto/generated"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"log"
)

type cmdLogin struct {
	global *cmdGlobal
}

func (c *cmdLogin) command() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = fmt.Sprint("login")
	cmd.Short = "Login to google and authorize the cli access to google drive"

	cmd.Args = cobra.NoArgs
	cmd.RunE = c.run
	return cmd
}

func (c *cmdLogin) run(_ *cobra.Command, args []string) error {
	ctx := context.Background()

	err := c.global.initGrpcClient()
	if err != nil {
		return err
	}
	defer c.global.closeGrpcClient()

	client := pb.NewAuthenticationServiceClient(c.global.conn)

	authToken, err := client.GetToken(ctx, &pb.Empty{})
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	if authToken.GetValue() != "" {
		fmt.Println("User already logged in.")
		return nil
	}

	fmt.Println("User not logged in.")
	//b, err := os.ReadFile("credentials.json")
	//if err != nil {
	//	log.Fatalf("Unable to read client secret file: %v", err)
	//}

	hashedData := "eyJpbnN0YWxsZWQiOnsiY2xpZW50X2lkIjoiNjU5OTE0NDgzNTUwLXBuNW1icTliN21ibmI2cDFzaWNzM3FwMzU3azRsY3FiLmFwcHMuZ29vZ2xldXNlcmNvbnRlbnQuY29tIiwicHJvamVjdF9pZCI6ImRzeW5jLTQzMzMyMSIsImF1dGhfdXJpIjoiaHR0cHM6Ly9hY2NvdW50cy5nb29nbGUuY29tL28vb2F1dGgyL2F1dGgiLCJ0b2tlbl91cmkiOiJodHRwczovL29hdXRoMi5nb29nbGVhcGlzLmNvbS90b2tlbiIsImF1dGhfcHJvdmlkZXJfeDUwOV9jZXJ0X3VybCI6Imh0dHBzOi8vd3d3Lmdvb2dsZWFwaXMuY29tL29hdXRoMi92MS9jZXJ0cyIsImNsaWVudF9zZWNyZXQiOiJHT0NTUFgtR1UzeTI2b3dvOUF5TE01bFVPTFIzbkFESjB2dCIsInJlZGlyZWN0X3VyaXMiOlsiaHR0cDovL2xvY2FsaG9zdCJdfX0="

	// Unhash (decode) the data
	b, err := base64.URLEncoding.DecodeString(hashedData)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	token := getTokenFromWeb(config)
	jsonToken, err := json.Marshal(&token)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	_, err = client.SaveToken(ctx, &pb.OAuth2Token{
		Value: string(jsonToken),
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	fmt.Println("Successfully logged in.")
	return nil
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

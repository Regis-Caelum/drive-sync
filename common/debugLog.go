package common

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

var debugMode bool

func init() {
	err := godotenv.Load("debug.env")
	if err != nil {
		DebugLog("No .env file found or error loading .env file")
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if os.Getenv("DEBUG") == "true" {
		debugMode = true
	}
}

func DebugLog(message string, args ...interface{}) {
	if debugMode {
		log.Printf("DEBUG: "+message, args...)
	}
}

func IsEnvDebug() bool {
	return debugMode
}

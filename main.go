package main

import (
	"fmt"
	"goVideoStreaming/fileserver"
	"goVideoStreaming/manifestscheduler"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Could not open .env file. Returning.")
	}

	// TODO: Filescanner always runs when it opens. Fix this.
	var lastRun time.Time
	var isServerRunning = false
	var enableFilescan bool
	var lastRunInterval int
	var interval int
	enableFilescan = manifestscheduler.FilescanEnabled()

	now := time.Now()

	if enableFilescan {
		interval, err = manifestscheduler.GetFileScanInterval()

		if err != nil {
			log.Fatal("Could not get fileScanInterval. Returning.")
		}

		fmt.Println("-- Running scanner each: ", interval, " minutes.")
		// Run this once upon initializing. Then, in the for loop.
		lastRunInterval = int((now.Sub(lastRun)).Minutes())
	} else {
		lastRunInterval = 0
		interval = 0
		fmt.Println("-- Filescanner is OFF.")
	}

	// Main loop.
	for {
		// TODO: Ideally, this should be a separate service. Good enough for now.
		if lastRunInterval >= interval && enableFilescan {
			lastRun = time.Now()
			lastRunInterval = manifestscheduler.UpdateLastRunInterval(lastRun) //It should be zero.
			go manifestscheduler.GenerateManifest()
		}

		if enableFilescan {
			lastRunInterval = manifestscheduler.UpdateLastRunInterval(lastRun)
		}

		// If running locally, must forward the port using ngrok.
		if !isServerRunning {
			go fileserver.FileServerStart()
			go fileserver.NgrokStart()
			isServerRunning = true
			// TODO: Ideally should wait for the ngrok to be deployed, not just waiting randomly 5 seconds.
			time.Sleep(5 * time.Second)
			mainURL := fileserver.NgrokGetURL()
			manifestscheduler.PrintAvailableFiles(mainURL)
		}

	}

}

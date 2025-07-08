package manifestscheduler

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func FilescanEnabled() bool {
	enableFilescan := os.Getenv("ENABLE_FILESCAN")
	if enableFilescan == "TRUE" {
		return true
	}

	return false
}

// Get interval of when to run the Manifest.
func GetFileScanInterval() (int, error) {
	// How many minutes should it wait to scan for manifests.
	intervalStr := os.Getenv("FILESCAN_INTERVAL")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Fatal("Conversion error. Returning.")
	}

	return interval, err
}

// Updates the time (in minutes) that have passed since the Filescanner ran.
func UpdateLastRunInterval(lastRun time.Time) int {
	now := time.Now()
	lastRunInterval := int((now.Sub(lastRun)).Minutes())
	return lastRunInterval
}

// Utility to check subdirectories under the root file folder.
func PrintAvailableFiles(prefix string) {
	fmt.Println("=============================\nAvailable Files\n=============================")

	rootDir := os.Getenv("FILE_ROOT_DIR")

	subdirs, err := os.ReadDir(rootDir)

	if err != nil {
		fmt.Println("Unable to read root directory for files, returning...")
		return
	}

	if len(subdirs) == 0 {
		fmt.Println("No directories found under the root folder, returning...")
		return
	}

	if rootDir[0] == '.' {
		rootDir = rootDir[1:]
	}

	for _, folder := range subdirs {
		fmt.Println(prefix + rootDir + folder.Name() + "/playlist.m3u8")
	}
}

// Iterates through items in a slice of memory addresses, returns if target is found.
func SearchManifest(fileList []fs.DirEntry, target string) bool {
	for _, value := range fileList {
		if value.Name() == target {
			return true
		}
	}

	return false
}

// TODO. Handle return and errors. Improve security. Executes FFMPEG command to generate the .m3u8 manifest and .ts files for a given input.mp4 file.
func ExecuteFFMPEG(fileFolder string) {
	fullFolderPath, err := filepath.Abs(fileFolder)

	if err != nil {
		log.Fatal(err)
	}

	inputFile := fullFolderPath + "\\input.mp4"
	tsFileFormat := fullFolderPath + "\\file_%03d.ts"
	playlistFile := fullFolderPath + "\\playlist.m3u8"

	cmd := exec.Command("ffmpeg", "-i", inputFile, "-filter:v", "scale=1920:1080", "-c:v", "libx264", "-preset",
		"veryfast", "-b:v", "5000k",
		"-maxrate", "5350k", "-bufsize", "7500k", "-c:a", "aac", "-b:a", "128k", "-ar", "48000",
		"-f", "hls", "-hls_time", "5", "-hls_playlist_type", "vod", "-hls_segment_filename", tsFileFormat, playlistFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("FFMPEG Error:\n%s\n%v", string(output), err)
	}

	fmt.Println("   - Manifest generated for ", fileFolder, "!")
}

// Scans subdirectories in the /files/ folder and generates manifest if not already.
func GenerateManifest() {
	fmt.Println("-- Running Manifest Generator...")

	rootDir := os.Getenv("FILE_ROOT_DIR")

	subdirs, err := os.ReadDir(rootDir)
	if err != nil {
		fmt.Println("Unable to read root directory for files, returning...")
		return
	}

	if len(subdirs) == 0 {
		fmt.Println("No directories found under the root folder, returning...")
		return
	}

	for _, folder := range subdirs {
		currentFolder := folder.Name()
		fmt.Println("-- ", currentFolder)
		fileFolder := rootDir + currentFolder + "/"
		files, err := os.ReadDir(fileFolder)

		if err != nil {
			fmt.Println("Unable to read files subdirectory, returning...")
			return
		}

		found := SearchManifest(files, "playlist.m3u8")

		if found {
			fmt.Println("   - Manifest found! Skipping folder!")
		} else {
			// TODO: This runs a hardcoded ffmpeg command in the input.mp4 file inside the subdir. Must handle other file types.
			fmt.Println("   - Manifest not found! Generating manifest...")
			//TODO: Handle security. Handle errors.
			ExecuteFFMPEG(fileFolder)
		}

	}
	fmt.Println("-- End of File Scan: ", time.Now())
}

package globgrinder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type GlobGrinder struct {
	glob          string
	processedDir  string
	processingDir string
}

func New(glob string, out string) (*GlobGrinder, error) {

	gg := new(GlobGrinder)
	if _, err := filepath.Glob(glob); err != nil {
		return nil, fmt.Errorf("The file pattern you give must be a valid glob. %v", err)
	}
	gg.glob = glob

	gg.processedDir = out
	gg.processingDir = filepath.Join(gg.processedDir, "processing")

	// Create directories as required.
	if err := os.MkdirAll(gg.processingDir, 0733); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(gg.processedDir, 0733); err != nil {
		return nil, err
	}

	return gg, nil
}

func (gg *GlobGrinder) Run(process chan<- string, done <-chan bool) error {

	fileQueue := make(chan string)
	go gg.fileWatcher(fileQueue)

	// Process files as they come in
	for path := range fileQueue {

		fileBasePath := filepath.Base(path)

		// get lock on file by copying file (if it has already been moved, continue gracefully)
		processingPath := filepath.Join(gg.processingDir, fileBasePath)
		if err := os.Rename(path, processingPath); err != nil {
			log.Printf("Somebody else got the file first. %v\n", err)
			continue
		}

		// Put file in queue to be processed.
		log.Print("Adding file to processing queue.")
		process <- processingPath
		<-done // wait for processing.

		// Move file from processing to processed
		// We panic if we're not able to move the file because we want to not
		// keep processing files if the setup is messed up. E.g. the permissions
		// are wrong or have changed.
		processedPath := filepath.Join(gg.processedDir, fileBasePath)
		if exists, _ := exists(processedPath); exists == true {
			if err := os.Remove(processedPath); err != nil {
				log.Fatalf("Couldn't move a processed file out of the way. %v\n", err)
			}
		}
		if err := os.Rename(processingPath, processedPath); err != nil {
			log.Fatalf("Processed the file, but couldn't copy to processed directory. %v\n", err)
		}

	}
	return nil
}

func (gg *GlobGrinder) fileWatcher(fileQueue chan string) {

	for {
		matches, _ := filepath.Glob(gg.glob)
		for _, match := range matches {
			fileQueue <- match
		}
		time.Sleep(10 * time.Second)
	}

}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

package dirprocessor

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DirProcessor struct {
	WatchingDir   string
	ProcessingDir string
	ProcessedDir  string
}

func New(watchDir string) (*DirProcessor, error) {
	dp := new(DirProcessor)
	if exists, err := exists(watchDir); exists == false || err != nil {
		return dp, errors.New("The directory you specified to watch does not exist (or is not readable).")
	}
	dp.WatchingDir = watchDir
	dp.ProcessingDir = filepath.Join(dp.WatchingDir, "processing")
	dp.ProcessedDir = filepath.Join(dp.WatchingDir, "processed")

	// Check if watch dir exists. Check if subfolder processing and processed exist.
	// TODO: check for errors here
	os.MkdirAll(dp.ProcessingDir, 0733)
	os.MkdirAll(dp.ProcessedDir, 0733)
	return dp, nil
}

func (dp *DirProcessor) Run(process chan<- string, done <-chan bool) error {

	fileQueue := make(chan string)
	go fileWatcher(dp.WatchingDir, fileQueue)

	// Process files as they come in
	for path := range fileQueue {

		fileBasePath := filepath.Base(path)

		// get lock on file by copying file (if it has already been moved, continue gracefully)
		processingPath := filepath.Join(dp.ProcessingDir, fileBasePath)
		if err := os.Rename(path, processingPath); err != nil {
			log.Printf("Somebody else got the file first. %v\n", err)
			continue
		}

		// Put file in queue to be processed.
		log.Print("Adding file to processing queue.")
		process <- processingPath
		<-done // wait for processing.

		// Move file from processing to processed
		processedPath := filepath.Join(dp.ProcessedDir, fileBasePath)
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

func fileWatcher(directory string, fileQueue chan string) {

	matchFiles := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && (info.Name() == "processing" || info.Name() == "processed") {
			return filepath.SkipDir
		}
		// filepattern = sfduPLU_STAGE####_YYYYMMDDHHMMSS.dat
		if filepath.Ext(path) != ".dat" {
			return nil
		}
		if strings.HasPrefix(info.Name(), "sfduPLU_STAGE") {
			fileQueue <- path
		}
		return nil
	}

	for {
		err := filepath.Walk(directory, matchFiles)
		if err != nil {
			log.Fatalf("Could not walk the directory you specified: %v\n", err)
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

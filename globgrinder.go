package globgrinder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type GlobGrinder struct {
	glob         string
	processedDir string
	keep         bool
}

func New(glob string, out string) (*GlobGrinder, error) {

	gg := new(GlobGrinder)
	if _, err := filepath.Glob(glob); err != nil {
		return nil, fmt.Errorf("The file pattern you give must be a valid glob. %v", err)
	}
	gg.glob = glob

	gg.processedDir = out
	if out == "" {
		gg.keep = false
	} else {
		gg.keep = true
	}

	// Create output directory if needed
	if gg.keep {
		if err := os.MkdirAll(gg.processedDir, 0733); err != nil {
			return nil, err
		}
	}

	return gg, nil
}

func grindingPath(path string) string {
	return (path + ".grinding")
}

func Path(path string) string {
	return strings.TrimSuffix(path, ".grinding")
}

func (gg *GlobGrinder) Run(process chan<- string, done <-chan bool) error {

	fileQueue := make(chan string)
	go gg.fileWatcher(fileQueue)

	// Process files as they come in
	for path := range fileQueue {

		// never process a file already being ground
		if filepath.Ext(path) == "grinding" {
			continue
		}

		// get lock on file by renaming file (if it has already been renamed, skip it)
		if err := os.Rename(path, grindingPath(path)); err != nil {
			log.Printf("Somebody else got the file first. %v\n", err)
			continue
		}

		// Put file in queue to be processed.
		log.Print("Adding file to processing queue.")
		process <- grindingPath(path)
		<-done // wait for processing.

		if gg.keep {

			// Move the file to the processedDir
			// We panic if we're not able to move the file because we want to not
			// keep processing files if the setup is messed up. E.g. the permissions
			// are wrong or have changed.
			processedPath := filepath.Join(gg.processedDir, filepath.Base(path))
			processedPath = uniqueIncrementedPath(processedPath)

			if err := os.Rename(grindingPath(path), processedPath); err != nil {
				log.Fatalf("Processed the file, but couldn't copy to processed directory. %v\n", err)
			}

		} else {

			// if we don't need to keep it, let's just delete it
			if err := os.Remove(grindingPath(path)); err != nil {
				log.Fatalf("Could not remove a file after it was processed. It still has the .grinding extension.")
			}

		}

	}
	return nil
}

// Figure out the path to rename a file into. If it already exists in the processDir, give
// it a numbered extention
func uniqueIncrementedPath(path string) string {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	// if the file already exists, give it an incrementing number extension

	ms := path + ".[0-9]*"
	matches, _ := filepath.Glob(ms)
	if len(matches) == 0 {
		return path + ".1"
	}

	n := 0
	for _, match := range matches {
		ext := strings.Trim(filepath.Ext(match), ".()")
		if i, err := strconv.Atoi(ext); err == nil && i > n {
			n = i
		}
	}
	return path + "." + strconv.Itoa(n+1)
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

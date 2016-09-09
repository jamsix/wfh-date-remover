package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var log = new(Log)

type FileHistoryFile struct {
	Filename         string
	DirectoryName    string
	OriginalFilename string
	DateInt          int
	Newest           bool
}

func main() {
	dir, _ := os.Getwd()
	log.Info("Starting in %s", dir)
	RecurseFiles(dir)
}

func RecurseFiles(dirname string) {
	entries, _ := ioutil.ReadDir(dirname)

	var files, dirs []os.FileInfo

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	RenameFiles(dirname, files)

	files = nil

	for _, dir := range dirs {
		RecurseFiles(dirname + string(os.PathSeparator) + dir.Name())
	}

}

func RenameFiles(dirname string, files []os.FileInfo) {
	fhFiles := make(map[string][]FileHistoryFile)
	for _, file := range files {
		rp, _ := regexp.Compile("(.*)\\s\\(([0-9]+_[0-9]+_[0-9]+\\s[0-9]+_[0-9]+_[0-9]+)\\sUTC\\)(\\.[a-zA-Z0-9]+)$")
		found := rp.MatchString(file.Name())
		if found {
			matches := rp.FindAllStringSubmatch(file.Name(), -1)
			//fmt.Printf("%q", matches)
			//fmt.Printf("%s%s%s\n", dirname, string(os.PathSeparator), file.Name())
			originalFilename := matches[0][1] + matches[0][3]
			dateInt, _ := strconv.Atoi(strings.Replace(strings.Replace(matches[0][2], "_", "", 5), " ", "", 5))
			fhFile := FileHistoryFile{Filename: file.Name(), DirectoryName: dirname, OriginalFilename: originalFilename, DateInt: dateInt}
			fhFiles[originalFilename] = append(fhFiles[originalFilename], fhFile)
		}
	}
	for key, _ := range fhFiles {
		var newestFhFile = 0
		for i, fhFile := range fhFiles[key] {
			if fhFile.DateInt > fhFiles[key][newestFhFile].DateInt {
				newestFhFile = i
			}
		}
		fhFiles[key][newestFhFile].Newest = true
	}

	for key, _ := range fhFiles {
		for _, fhFile := range fhFiles[key] {
			relativeName := fhFile.DirectoryName + string(os.PathSeparator) + fhFile.Filename
			relativeOriginalName := fhFile.DirectoryName + string(os.PathSeparator) + fhFile.OriginalFilename
			if fhFile.Newest {
				if _, err := os.Stat(relativeOriginalName); err == nil {
					// File with OriginalFilename already exists
					os.Remove(relativeName)
					log.Info("Removed %s (%s already exists)", relativeName, relativeOriginalName)
				} else {
					os.Rename(relativeName, relativeOriginalName)
					log.Info("Renamed %s to %s", relativeName, relativeOriginalName)
				}
			} else {
				os.Remove(relativeName)
				log.Info("Removed %s (newer version exists)", relativeName)
			}
		}
	}

}

// Logger
type Log struct{}

func (log *Log) Info(format string, a ...interface{}) {
	fmt.Printf(time.Now().Format("2006-01-02 15:04:05 ")+" INFO  "+format+"\n", a...)
}
func (log *Log) Error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, time.Now().Format("2006-01-02 15:04:05 ")+" ERROR  "+format+"\n", a...)
}

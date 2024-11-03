package common

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/chebuya/sastsweep/common/logger"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"

var Err404 = errors.New("404 status code")
var Err422 = errors.New("422 status code")
var OsRuntime = runtime.GOOS

func isSafePath(path string, dest string) bool {

	// Zillion inodes/infinite directories
	depth := len(filepath.SplitList(path)) - 1
	if depth > 100 {
		logger.Debug(path + " was determined to be unsafe: path more than 100 directories deep")
		return false
	}

	// https://www.stackhawk.com/blog/golang-path-traversal-guide-examples-and-prevention/ - Validate user input
	if strings.Contains(path, "..") {
		logger.Debug(path + " was determined to be malicious: presence of ..")
		return false
	}

	// https://www.stackhawk.com/blog/golang-path-traversal-guide-examples-and-prevention/ - Canonicalize Paths
	if fileExists(path) {
		logger.Debug(path + " was determined to be malicious: attempting to overwrite a file")
		return false
	}

	// https://www.stackhawk.com/blog/golang-path-traversal-guide-examples-and-prevention/ - Establish a Trusted Root
	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
		logger.Debug(path + " was determined to be malicious: writing outside of intended path")
		return false
	}

	return true
}

func UnzipBytes(zipBytes []byte, dest string) error {
	reader := bytes.NewReader(zipBytes)
	zipReader, err := zip.NewReader(reader, int64(len(zipBytes)))
	if err != nil {
		logger.Error("Could not create reader for zipBytes")
		return err
	}

	// https://github.com/twbgc/sunzip: Restrict number of extracted files, zillion inodes
	if len(zipReader.File) > 500000 {
		return errors.New("ZIP Contains more than 500k files: " + strconv.Itoa(len(zipReader.File)))
	}

	var totalSize uint64
	for _, file := range zipReader.File {
		// Check for UNC path
		if strings.Contains(file.Name, `\\`) {
			continue
		}

		// https://github.com/twbgc/sunzip: Check if it's a nested zip file. (i.e. 42.zip)
		if strings.ToLower(filepath.Ext(file.Name)) == ".zip" {
			continue
		}

		fileSize := file.UncompressedSize64
		totalSize += fileSize
		// https://github.com/twbgc/sunzip: Restrict output file size
		if totalSize > 1024*1024*1024*60 {
			return errors.New(dest + " ZIP contains more than 60gb of uncompressed data, skipping")
		}

		// https://www.stackhawk.com/blog/golang-path-traversal-guide-examples-and-prevention/ - Clean FilePaths
		path := filepath.Clean(filepath.Join(dest, file.Name))

		if !isSafePath(path, dest) {
			logger.Debug(path + " was determined to be unsafe, skipping file write")
			continue
		}

		if file.FileInfo().IsDir() {
			err = os.Mkdir(path, 0700)
			if err != nil {
				logger.Error("Could not create " + path + " as a directory: " + err.Error())
			}
			continue
		}

		zipFileReader, err := file.Open()
		if err != nil {
			logger.Error("Could not open " + path + " for reading: " + err.Error())
			continue
		}
		defer zipFileReader.Close()

		// https://www.stackhawk.com/blog/golang-path-traversal-guide-examples-and-prevention/ - Canonicalize Paths (symlinks)
		outFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|syscall.O_NOFOLLOW, 0600)
		if err != nil {
			logger.Error("Could not open " + path + " for writing: " + err.Error())
			continue
		}
		defer outFile.Close()

		// https://github.com/twbgc/sunzip - Restrict output file size
		_, err = io.CopyN(outFile, zipFileReader, int64(fileSize))
		if err != nil {
			logger.Error("There was an error calling io.Copy on " + path + ": " + err.Error())
			continue
		}
	}

	return nil
}

func HTTPGet(client *http.Client, targetURL string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		logger.Error("Could not create the HTTP request: " + err.Error())
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Could not send the HTTP request: " + err.Error())
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, Err404
	} else if resp.StatusCode == 422 {
		return nil, Err422
	} else if resp.StatusCode != 200 {
		logger.Error(targetURL + ": non-200 status code was returned: " + strconv.Itoa(resp.StatusCode))
		return nil, errors.New("Non-200 status code")
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Could not read the HTTP response: " + err.Error())
		return nil, err
	}

	return body, nil
}

func CountFiles(path string) (int, error) {
	matchCount := 0
	err := filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			matchCount++
		}

		return nil
	})

	if err != nil {
		return -1, fmt.Errorf("error walking through directory: %v", err)
	}

	return matchCount, nil
}

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		logger.Error("Could not open " + path + ": " + err.Error())
		return nil, nil
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		logger.Error("Could not get file stats for " + path + ": " + err.Error())
		return nil, nil
	}

	fileBytes := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(fileBytes)
	if err != nil && err != io.EOF {
		logger.Error("Could not read from " + path + ": " + err.Error())
		return nil, nil
	}

	return fileBytes, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

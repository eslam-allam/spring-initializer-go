package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var logger *log.Logger = log.Default()

func UnzipFile(zipFile, destDir string) error {
	// Open the zip file for reading
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()
	// Create the destination directory if it doesn't exist
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		os.MkdirAll(destDir, os.ModePerm)
	}
	// Extract each file from the zip archive
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		path := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
		} else {
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ExpandAndMakeDir(targetDirectory string) (string, error) {
	if strings.HasPrefix(targetDirectory, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		targetDirectory = strings.Replace(targetDirectory, "~", home, 1)
	}
	if !path.IsAbs(targetDirectory) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %v", err)
		}
		targetDirectory = path.Join(cwd, targetDirectory)
	}
	if _, err := os.Stat(targetDirectory); errors.Is(err, os.ErrNotExist) {
		logger.Printf("Target directory %s does not exist. Creating it now", targetDirectory)
		err = os.MkdirAll(targetDirectory, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create target directory: %v", err)
		}
	}
	return targetDirectory, nil
}

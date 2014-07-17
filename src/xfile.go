// Copyright 2013-2014, Friedrich Paetzke. All rights reserved.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Returns all filenames in the given directory and sub-directories that passed
// the fn function
func GetFilenamesRecFunc(dir string, fn func(os.FileInfo) bool) (filenames []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if fn != nil && !fn(file) {
			continue
		}

		fPath := filepath.Join(dir, file.Name())
		if file.IsDir() {
			filenames = append(filenames, GetFilenamesRecFunc(fPath, fn)...)
		} else {
			filenames = append(filenames, fPath)
		}
	}
	return
}

// Returns all filenames in the given directory and sub-directories
func GetFilenamesRec(dir string) []string {
	return GetFilenamesRecFunc(dir, nil)
}

// Returns all filenames in the given directory
func GetFilenames(dir string) (filenames []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			fPath := filepath.Join(dir, file.Name())
			filenames = append(filenames, fPath)
		}
	}
	return
}

func GetDirectoriesRec(dir string) (dirnames []string) {
	dirnames = append(dirnames, dir)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			fPath := filepath.Join(dir, file.Name())
			dirnames = append(dirnames, GetDirectoriesRec(fPath)...)
		}
	}
	return
}

func GetFileExt(filename string) string {
	xs := strings.Split(filename, ".")
	idx := len(xs) - 1
	if idx < 1 {
		return ""
	}
	return strings.ToLower(xs[idx])
}

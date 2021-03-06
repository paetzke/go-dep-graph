// Copyright 2013-2014, Friedrich Paetzke. All rights reserved.

package main

import (
	"github.com/paetzke/godot/godot"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
)

var (
	doPackagesOnly bool
	doIgnoreStdLib bool
	sourceDir      string
	deps           map[string][]string = make(map[string][]string)
	stdLib         []string
)

func init() {
	// populate stdLib by examining GOROOT
	root := runtime.GOROOT() + "/src/pkg/"
	stdLib = GetDirectoriesRec(root)
	for i := range stdLib {
		stdLib[i] = strings.TrimPrefix(stdLib[i], root)
	}
}

func fileFilter(f os.FileInfo) bool {
	if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") {
		return true
	}
	if f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
		return true
	}
	return false
}

func containsString(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}
	return false
}

func removeDuplicates(dups []string) (uniqs []string) {
	set := make(map[string]bool)
	for _, dup := range dups {
		set[dup] = true
	}
	for uniq := range set {
		uniqs = append(uniqs, uniq)
	}
	return
}

func handleImport(filename, imported string) {
	imported = strings.TrimSpace(imported)

	if strings.HasPrefix(imported, "//") {
		return
	}

	// extract package name
	split := strings.Split(imported, "\"")
	if len(split) < 2 {
		return
	}
	imported = split[1]

	if doIgnoreStdLib && containsString(stdLib, imported) {
		return
	}

	if doPackagesOnly {
		filename = path.Dir(filename)
	}

	deps[filename] = append(deps[filename], imported)
}

func arrangeFilename(filename string) string {
	if len(filename) == 0 {
		return filename
	}

	gopath := path.Join(os.Getenv("GOPATH"), "/src/")
	pwd, _ := os.Getwd()
	mayBeRelPath := path.Clean(path.Join(pwd, filename))
	if strings.HasPrefix(mayBeRelPath, gopath) {
		s := mayBeRelPath[len(gopath):]
		if s[0:1] == "/" {
			filename = s[1:]
		}
	}
	mayBeAbsPath := path.Clean(filename)
	if strings.HasPrefix(mayBeAbsPath, gopath) {
		s := mayBeAbsPath[len(gopath):]
		if s[0:1] == "/" {
			filename = s[1:]
		}
	}

	filename = strings.TrimPrefix(filename, sourceDir)
	return filename
}

func printDot() {
	dotter, _ := godot.NewDotterEx(godot.OUT_DOT, godot.PROG_DOT, godot.GRAPH_DIRECTED,
		false, false, "")

	if doPackagesOnly {
		for file, importeds := range deps {
			deps[file] = removeDuplicates(importeds)
		}
	}

	for file, importeds := range deps {
		arrangeFile := arrangeFilename(file)
		for _, imported := range importeds {
			dotter.SetLink(arrangeFile, imported)
			dotter.SetLabel(imported, arrangeFilename(imported))
		}
		dotter.SetLabel(arrangeFile, arrangeFile)
	}

	if err := dotter.Close(); err != nil {
		panic(err)
	}
}

func extractImports(filename string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	hasMultiImport := false
	for _, line := range strings.Split(string(content), "\n") {
		if line == "import (" {
			hasMultiImport = true
		} else if strings.HasPrefix(line, "import ") {
			handleImport(filename, line)
		} else if hasMultiImport && line == ")" {
			hasMultiImport = false
		} else if hasMultiImport {
			handleImport(filename, line)
		}
	}
}

func main() {
	doPackagesOnly = containsString(os.Args[1:], "-p")
	doIgnoreStdLib = containsString(os.Args[1:], "-s")

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		// set sourceDir
		sourceDir = arg
		if !strings.HasSuffix(sourceDir, "/") {
			sourceDir += "/"
		}
		if !strings.HasSuffix(sourceDir, "src/") {
			sourceDir += "src/"
		}
		for _, file := range GetFilenamesRecFunc(arg, fileFilter) {
			extractImports(file)
		}
	}

	printDot()
}

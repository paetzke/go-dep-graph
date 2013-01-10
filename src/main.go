// Copyright 2013, Friedrich Paetzke. All rights reserved.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"github.com/paetzke/go/xfile"
)

var (
	doPackagesOnly bool
	doIgnoreStdLib bool
	deps map[string] []string = make(map[string] []string)
	stdLib []string
)

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

func handleImport(filename, imported string) {
	if strings.HasPrefix(imported, "import ") {
		imported = imported[7:]
	}
	imported = strings.Replace(imported, "\"", "", -1)
	imported = strings.TrimSpace(imported)

	if imported[0:2]== "//" || doIgnoreStdLib && containsString(stdLib, imported) {
		return
	}

	if strings.HasPrefix(imported, "_ ") {
		imported = imported[2:]
	}
	if doPackagesOnly {
		filename = path.Dir(filename)
	}

	imports, ok := deps[filename]
	if !ok {
		imports = make([]string, 0)
	}
	deps[filename] = append(imports, imported)
}

func escapeChars(s string) string {
	s = strings.Replace(s, ".", "DOT", -1)
	s = strings.Replace(s, "/", "SLASH", -1)
	s = strings.Replace(s, "-", "HIVEN", -1)
	return s
}



func arrangeFilename(filename string) string {
	if len(filename) == 0 {
		return filename
	}

	gopath := path.Join(os.Getenv("GOPATH"), "/src/")
	pwd, _ := os.Getwd()
	mayBeRelPath := path.Clean(path.Join(pwd, filename))
	if strings.HasPrefix(mayBeRelPath, gopath) {
		s :=  mayBeRelPath[len(gopath):]
		if s[0:1] == "/" {
			filename = s[1:]
		}
	}
	mayBeAbsPath := path.Clean(filename)
	if strings.HasPrefix(mayBeAbsPath, gopath) {
		s :=  mayBeAbsPath[len(gopath):]
		if s[0:1] == "/" {
			filename = s[1:]
		}
	}

	return filename
}

func printDot() {
	m := make(map[string] int)
	for file, importeds := range deps {
		arrangeFile := arrangeFilename(file)
		escArrangeFile := escapeChars(arrangeFile)

		for _, imported := range importeds {
			escImported := escapeChars(imported)
			arrImported := arrangeFilename(imported)

			m[fmt.Sprintf("\t%s -> %s\n", escArrangeFile, escImported)] = 0
			m[fmt.Sprintf("\t%s[label=\"%s\"]\n", escImported, arrImported)] = 0
		}
		m[fmt.Sprintf("\t%s[label=\"%s\"]\n", escArrangeFile, arrangeFile)] = 0
	}

	fmt.Println("digraph deps {")
	for key, _ := range m {
		fmt.Println(key)
	}
	fmt.Println("}")
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
	stdLib = getStdLib()

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		for _, file := range xfile.GetFilenamesRecFunc(arg, fileFilter) {
			extractImports(file)
		}
	}

	printDot()
}

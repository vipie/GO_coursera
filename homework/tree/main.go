package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

func GetNames(path string, isDir bool) ([]string, error) {

	fs, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, f := range fs {

		if f.IsDir() == isDir {
			files = append(files, f.Name())
		}

	}
	return files, nil
}

func GetFileNames(path string) ([]string, error) {

	return GetNames(path, false)
}

func GetFolderNames(path string) ([]string, error) {

	return GetNames(path, true)
}

func FormatFolderNames(path string, prevPath string) string {

	if prevPath == "." {
		return path
	} else {
		return prevPath + "/" + path
	}

}

func FormatFileNames(paths []string, prevString string) string {

	var sb strings.Builder
	for index, f := range paths {
		//fmt.Println(f)
		//sb.WriteString(prevString + f + "\n")
		if index == len(paths)-1 {
			sb.WriteString(strings.Replace(prevString, "├───", "└───", 1) + f)

		} else {
			sb.WriteString(prevString + f + "\n")
		}
	}

	return sb.String()
}

func GetPreString(path string) string {
	var sb strings.Builder

	if path == "." {
		return "├───"
	}

	for i := 0; i < strings.Count(path, "/"); i++ {
		sb.WriteString("\t|")
	}
	sb.WriteString("\t")
	sb.WriteString("├───")

	return sb.String()

}
func GetFolderPreString(path string) string {
	var sb strings.Builder

	for i := 0; i < strings.Count(path, "/"); i++ {
		sb.WriteString("\t|")
	}
	sb.WriteString("───")

	return sb.String()

}

func dirTree(out *os.File, path string, printFiles bool) error {

	preString := GetPreString(path)

	if printFiles {
		names, _ := GetFileNames(path)
		sort.Strings(names)
		fmt.Fprintln(out, FormatFileNames(names, preString))
	}

	names, _ := GetFolderNames(path)

	sort.Strings(names)

	for _, f := range names {
		formattedFoldersNames := FormatFolderNames(f, path)
		fmt.Fprintln(out, GetFolderPreString(formattedFoldersNames)+f)
		dirTree(out, formattedFoldersNames, printFiles)
	}

	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

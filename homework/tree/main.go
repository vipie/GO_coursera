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

func ConcatFolderName(path string, prevPath string) string {

	if prevPath == "." {
		return path
	} else {
		return prevPath + "/" + path
	}

}

// разбить путь на массив всех путей
func GetAllPaths(path string) []string {
	var ret []string

	cur_ind := 0
	ind := 0
	for true {
		ind = strings.Index(path[cur_ind:], "/")
		if ind == -1 {
			ret = append(ret, path)
			break
		}

		ret = append(ret, path[:ind+cur_ind])
		cur_ind = cur_ind + ind + 1
	}

	return ret
}

/*func FormatFileNames(paths []string, path string, mapa map[string]bool) string {

	var ret string

	for _, pt := range paths {
		ret = ret + FormatItemName(pt, path, mapa) + "\n"
	}
	return ret[:len(ret)-1]
}*/

func DeleteRootPath(path string, subPath string) string {
	str := strings.Replace(path, subPath, "", 1)

	return strings.Replace(str, "//", "", 1)
}

func FormatItemName(name string, path string, mapa map[string]bool, rootPath string) string {

	//noRootPath := DeleteRootPath(path, rootPath)
	var sb strings.Builder
	if len(GetAllPaths(ConcatFolderName(name, path))) > 1 {
		for _, pt := range GetAllPaths(path) {

			if mapa[pt] {
				sb.WriteString("\t")
			} else {
				sb.WriteString("│\t")
			}
		}
	}

	if mapa[ConcatFolderName(name, path)] {
		sb.WriteString("└───")
	} else {
		sb.WriteString("├───")
	}

	sb.WriteString(name)
	return sb.String()

}

func FillMapa(mapa map[string]bool, names []os.FileInfo, path string) map[string]bool {
	for idx, n := range names {
		mapa[ConcatFolderName(n.Name(), path)] = idx == len(names)-1
	}
	return mapa
}

const testDirResult = `├───project
├───static
│	├───a_lorem
│	│	└───ipsum
│	├───css
│	├───html
│	├───js
│	└───z_lorem
│		└───ipsum
└───zline
	└───lorem
		└───ipsum
`

const testFullResult = `├───project
│	├───file.txt (19b)
│	└───gopher.png (70372b)
├───static
│	├───a_lorem
│	│	├───dolor.txt (empty)
│	│	├───gopher.png (70372b)
│	│	└───ipsum
│	│		└───gopher.png (70372b)
│	├───css
│	│	└───body.css (28b)
│	├───empty.txt (empty)
│	├───html
│	│	└───index.html (57b)
│	├───js
│	│	└───site.js (10b)
│	└───z_lorem
│		├───dolor.txt (empty)
│		├───gopher.png (70372b)
│		└───ipsum
│			└───gopher.png (70372b)
├───zline
│	├───empty.txt (empty)
│	└───lorem
│		├───dolor.txt (empty)
│		├───gopher.png (70372b)
│		└───ipsum
│			└───gopher.png (70372b)
└───zzfile.txt (empty)
`

func dirTree_(out *os.File, path string, printFiles bool,
	mapa map[string]bool, rootPath string) (map[string]bool, error) {

	fs, _ := ioutil.ReadDir(rootPath + "/" + path)

	if !printFiles {

		var filteredFs []os.FileInfo

		for _, s := range fs {
			if s.IsDir() {
				filteredFs = append(filteredFs, s)
			}
		}
		fs = filteredFs
	}

	sort.Slice(fs, func(i, j int) bool {
		return fs[i].Name() < (fs[j].Name())
	})

	mapa = FillMapa(mapa, fs, path)

	for _, f := range fs {
		fmt.Fprintln(out, FormatItemName(f.Name(), path, mapa, rootPath))

		if f.IsDir() {
			mapa, _ = dirTree_(out, ConcatFolderName(f.Name(), path),
				printFiles, mapa, rootPath)
		}

	}

	return mapa, nil
}
func dirTree(out *os.File, path string, printFiles bool) error {

	ends_map := make(map[string]bool)
	_, err := dirTree_(out, ".", printFiles, ends_map, path)
	return err

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

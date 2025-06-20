package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var screenshotDir = flag.String("dir", "~/Pictures/Screenshots", "screenshot directory")
var rename = flag.Bool("rename", true, "rename screenshot files for simplicity and usability to remove spaces")
var outputPath = flag.Bool("outputPath", true, "output the entire file paths of screenshot files")

var isImgFilenameRegExp = regexp.MustCompile(`\.(?i)(png|jpg)$`)
var imgFilenameDateRegExp = regexp.MustCompile(`\s\d{4}-\d{2}-\d{2}.*`)
var isImgFilenameRenamedRegExp = regexp.MustCompile(`_\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}`)

func main() {
	flag.Parse()

	screenshotDir, cut := strings.CutPrefix(*screenshotDir, "~/")
	if cut {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Screenshot directory starts with user's home directory \"~\", error finding it: %s", err)
			os.Exit(1)
		}
		screenshotDir = filepath.Join(homeDir, screenshotDir)
	}

	amount := 1
	if flag.NArg() > 0 {
		var err error
		if amount, err = strconv.Atoi(flag.Arg(0)); err != nil {
			fmt.Printf("Invalid argument for amount of latest screenshots to include: %s\n", flag.Arg(0))
			os.Exit(1)
		}
	}

	dir, err := os.ReadDir(screenshotDir)
	if err != nil {
		fmt.Printf("Error reading screenshot directory: %s\n", err)
		os.Exit(1)
	}

	var imgFileInfos []os.FileInfo
	for _, file := range dir {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if !isImgFilename(info.Name()) {
			continue
		}

		imgFileInfos = append(imgFileInfos, info)
	}

	slices.SortFunc(imgFileInfos, func(a, b os.FileInfo) int {
		return b.ModTime().Compare(a.ModTime())
	})

	imgFileInfosLen := len(imgFileInfos)
	if amount >= imgFileInfosLen {
		amount = imgFileInfosLen - 1
	}

	imgs := imgFileInfos[:amount]
	imgPaths := make([]string, len(imgs))
	for i, img := range imgs {
		name := img.Name()
		pathName := filepath.Join(screenshotDir, name)
		if isImgFilenameAlreadyRenamed(name) || !(*rename) {
			if *outputPath {
				imgPaths[i] = pathName
			} else {
				imgPaths[i] = name
			}
			continue
		}

		newName := strings.ReplaceAll(imgFilenameDateRegExp.FindString(name), " ", "_")
		newPathName := filepath.Join(screenshotDir, newName)
		err := os.Rename(pathName, newPathName)
		if err != nil {
			fmt.Printf("Error renaming file: %s to %s: %s", name, newName, err)
			os.Exit(1)
		}

		if *outputPath {
			imgPaths[i] = newPathName
		} else {
			imgPaths[i] = newName
		}
	}

	fmt.Printf("%s\n", strings.Join(imgPaths, " "))
}

func isImgFilename(filename string) bool {
	return isImgFilenameRegExp.MatchString(filename)
}

func isImgFilenameAlreadyRenamed(filename string) bool {
	return isImgFilenameRenamedRegExp.MatchString(filename)
}

package main

import (
	"os"
	"path/filepath"
	"strings"
)

func findSidecars(root string) ([]string, error) {
	files := make([]string, 0)
	markFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// record this as a too-deep path we never want to traverse again
		if info.IsDir() && strings.Count(path, "/") > 0 {
			return filepath.SkipDir
		}

		if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), "-sidecar.json") {
			files = append(files, root+"/"+path)
		}
		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.Chdir(pwd)
		if err != nil {
			Log.Printf("findBuildScripts cannot restore working directory to %s: %v\n", pwd, err)
		}
	}()

	// Change directory to root so we have no need to know how many "/" root itself contains.
	if err := os.Chdir(root); err != nil {
		return nil, err
	}

	// Walk relative to root.
	err = filepath.Walk(".", markFn)
	if err != nil {
		return nil, err
	}

	return files, nil
}

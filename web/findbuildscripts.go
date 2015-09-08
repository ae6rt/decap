package main

import (
	"os"
	"path/filepath"
	"strings"
)

func findBuildScripts(root string) ([]string, error) {
	files := make([]string, 0)
	markFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(path, ".git") {
			return filepath.SkipDir
		}

		// record this as a too-deep path we never want to traverse again
		if info.IsDir() && strings.Count(path, "/") > 2 {
			return filepath.SkipDir
		}

		// We compare again the depth to 2 because ./build.sh is an undesirable possibility.
		if strings.Count(path, "/") == 2 && info.Mode().IsRegular() && info.Name() == "build.sh" {
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

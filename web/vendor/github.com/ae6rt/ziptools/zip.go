package ziptools

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Unzip unzips a zip file to a temporary directory, and returns the temporary directory name.
func Unzip(zipFile string) (string, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return "", err
	}
	defer r.Close()

	name, err := ioutil.TempDir("", "unzipped-")
	if err != nil {
		return "", err
	}

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "/") {
			continue
		}

		destinationFileName := name + "/" + f.Name
		parentDir := filepath.Dir(destinationFileName)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return "", err
		}

		dst, err := os.Create(destinationFileName)
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		_, err = io.Copy(dst, rc)
		if err != nil {
			return "", err
		}
		rc.Close()
		dst.Close()
	}
	return name, nil
}

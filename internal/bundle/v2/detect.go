package v2

import (
	"archive/zip"
)

func Detect(filename string) bool {
	zr, err := zip.OpenReader(filename)
	if err != nil {
		return false
	}
	defer zr.Close()
	f, err := zr.Open(versionFile)
	if err != nil {
		return false
	}
	_ = f.Close()

	return true
}

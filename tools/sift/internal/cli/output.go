package cli

import (
	"errors"
	"os"
	"path/filepath"
)

func createOutput(path string, inputInfo os.FileInfo) (*os.File, error) {
	if inputInfo != nil {
		info, err := os.Stat(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, &FileError{Op: "stat output", Path: path, Err: err}
		}

		if info != nil && os.SameFile(inputInfo, info) {
			return nil, &FileError{
				Op:   "open output",
				Path: path,
				Err:  errors.New("input and output must be different files"),
			}
		}
	}

	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return nil, &FileError{Op: "open output", Path: path, Err: err}
	}

	return file, nil
}

package util

import (
	"errors"
	"io/fs"
	"os"
)

func FileExist(file string) bool {
	_, err := os.Stat(file)
	b := err != nil && errors.Is(err, fs.ErrNotExist)
	return !b
}

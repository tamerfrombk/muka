package files

import (
	"fmt"
	"os"
)

// Deleter delete function interface
type Deleter interface {
	Delete(path string) error
}

type nopDeleter struct{}

func (nop nopDeleter) Delete(path string) error {

	fmt.Printf("'%s' would be removed.\n", path)

	return nil
}

type fileDeleter struct{}

func (impl fileDeleter) Delete(path string) error {

	return os.Remove(path)
}

// MakeDeleter a Deleter factory function
func MakeDeleter(isDryRun bool) Deleter {
	if isDryRun {
		return nopDeleter{}
	}

	return fileDeleter{}
}

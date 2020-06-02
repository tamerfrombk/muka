package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func TestMakeDeleterOnDryRunShouldKeepFile(t *testing.T) {
	f, err := ioutil.TempFile("", "TestMuka")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	deleter := MakeDeleter(true)
	if err := deleter.Delete(f.Name()); err != nil {
		t.Fatal(err)
	}

	if !fileExists(f.Name()) {
		t.Error("Dry run should not delete files.")
	}
}

func TestMakeDeleterOnWithoutDryRunShouldRemoveFile(t *testing.T) {
	f, err := ioutil.TempFile("", "TestMuka")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	deleter := MakeDeleter(false)
	if err := deleter.Delete(f.Name()); err != nil {
		t.Fatal(err)
	}

	if fileExists(f.Name()) {
		t.Error("The file should have been deleted.")
	}

}

package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

type testingDirOptions struct {
	CreateChildrenDirs bool
	CreateFiles        bool
}

func createTestingDirectory(options testingDirOptions) (string, error) {
	dir, err := ioutil.TempDir("", "TestMuka")
	if err != nil {
		return "", err
	}

	if options.CreateChildrenDirs {
		for i := 0; i < 5; i++ {
			if _, err := ioutil.TempDir(dir, "TestMukaChild"); err != nil {
				return "", err
			}
		}
	}

	if options.CreateFiles {
		for i := 0; i < 5; i++ {
			if _, err := ioutil.TempFile(dir, "TestMukaFile"); err != nil {
				return "", err
			}
		}
	}

	return dir, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

func TestDuplicateFileWithZeroDuplicates(t *testing.T) {
	fileHashes := []FileHash{
		{
			AbsolutePath: "file1.txt",
			Hash:         "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(fileHashes)
	if len(duplicateFiles) != 1 {
		t.Errorf("Expected '%d' entry but found '%d' instead.\n", 1, len(duplicateFiles))
	}

	if len(duplicateFiles[0].Duplicates) != 0 {
		t.Errorf("Expected '%d' entry but found '%d' instead.\n", 0, len(duplicateFiles[0].Duplicates))
	}
}

func TestDuplicateFileWithOneDuplicate(t *testing.T) {
	fileHashes := []FileHash{
		{
			AbsolutePath: "file1.txt",
			Hash:         "abcdefg",
		},
		{
			AbsolutePath: "file2.txt",
			Hash:         "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(fileHashes)
	if len(duplicateFiles) != 1 {
		t.Errorf("Expected '%d' entry but found '%d' instead.\n", 1, len(duplicateFiles))
	}

	if len(duplicateFiles[0].Duplicates) != 1 {
		t.Errorf("Expected '%d' entry but found '%d' instead.\n", 1, len(duplicateFiles[0].Duplicates))
	}
}

func TestDuplicateFileWithMoreThanOneDuplicate(t *testing.T) {
	fileHashes := []FileHash{
		{
			AbsolutePath: "file1.txt",
			Hash:         "abcdefg",
		},
		{
			AbsolutePath: "file2.txt",
			Hash:         "abcdefg",
		},
		{
			AbsolutePath: "file3.txt",
			Hash:         "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(fileHashes)
	if len(duplicateFiles) != 1 {
		t.Errorf("Expected '%d' entry but found '%d' instead.\n", 1, len(duplicateFiles))
	}

	if len(duplicateFiles[0].Duplicates) != 2 {
		t.Errorf("Expected '%d' entry but found '%d' instead.\n", 2, len(duplicateFiles[0].Duplicates))
	}
}

func TestCollectFilesEmptyDirectoryReturnsNoFiles(t *testing.T) {
	dir, err := createTestingDirectory(testingDirOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	files, err := CollectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) > 0 {
		t.Error("In an empty directory, no files should be collected.")
	}
}

func TestCollectFilesHasOnlyFiles(t *testing.T) {
	dir, err := createTestingDirectory(testingDirOptions{
		CreateChildrenDirs: true,
		CreateFiles:        true,
	})

	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	files, err := CollectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Errorf("Non empty list of files expected.")
	}

	for _, f := range files {
		if isDir(f.AbsolutePath) {
			t.Errorf("No directories should be returned but '%s' is a directory.\n", f.AbsolutePath)
		}
	}
}

func TestPromptToDeleteOriginal(t *testing.T) {
	dir, err := createTestingDirectory(testingDirOptions{
		CreateChildrenDirs: true,
		CreateFiles:        true,
	})

	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileHashes, err := CollectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)
	reader := strings.NewReader("o\n")
	var writer strings.Builder
	promptToDelete(&writer, reader, deleter, duplicates[0])

	if !strings.Contains(writer.String(), duplicates[0].String()) {
		t.Error("Duplicate should be displayed.")
	}

	if !strings.Contains(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") {
		t.Error("Incorrect prompt.")
	}

	if fileExists(duplicates[0].Original.AbsolutePath) {
		t.Error("The file was supposed to be deleted.")
	}

	for _, d := range duplicates[0].Duplicates {
		if !fileExists(d.AbsolutePath) {
			t.Error("The file was not supposed to be deleted.")
		}
	}
}

func TestPromptToDeleteDuplicates(t *testing.T) {
	dir, err := createTestingDirectory(testingDirOptions{
		CreateChildrenDirs: true,
		CreateFiles:        true,
	})

	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileHashes, err := CollectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)
	reader := strings.NewReader("d\n")
	var writer strings.Builder
	promptToDelete(&writer, reader, deleter, duplicates[0])

	if !strings.Contains(writer.String(), duplicates[0].String()) {
		t.Error("Duplicate should be displayed.")
	}

	if !strings.Contains(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") {
		t.Error("Incorrect prompt.")
	}

	if !fileExists(duplicates[0].Original.AbsolutePath) {
		t.Error("The file was not supposed to be deleted.")
	}

	for _, d := range duplicates[0].Duplicates {
		if fileExists(d.AbsolutePath) {
			t.Error("The file was supposed to be deleted.")
		}
	}
}

func TestPromptSkip(t *testing.T) {
	dir, err := createTestingDirectory(testingDirOptions{
		CreateChildrenDirs: true,
		CreateFiles:        true,
	})

	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileHashes, err := CollectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)
	reader := strings.NewReader("s\n")
	var writer strings.Builder
	promptToDelete(&writer, reader, deleter, duplicates[0])

	if !strings.Contains(writer.String(), duplicates[0].String()) {
		t.Error("Duplicate should be displayed.")
	}

	if !strings.Contains(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") {
		t.Error("Incorrect prompt.")
	}

	if !fileExists(duplicates[0].Original.AbsolutePath) {
		t.Error("The file was not supposed to be deleted.")
	}

	for _, d := range duplicates[0].Duplicates {
		if !fileExists(d.AbsolutePath) {
			t.Error("The file was not supposed to be deleted.")
		}
	}
}

func TestPromptInvalidResponseContinues(t *testing.T) {
	dir, err := createTestingDirectory(testingDirOptions{
		CreateChildrenDirs: true,
		CreateFiles:        true,
	})

	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileHashes, err := CollectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)
	reader := strings.NewReader("x\no\n")
	var writer strings.Builder
	promptToDelete(&writer, reader, deleter, duplicates[0])

	if strings.Count(writer.String(), duplicates[0].String()) != 2 {
		t.Error("Duplicate should be displayed.")
	}

	if strings.Count(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") != 2 {
		t.Error("Incorrect prompt.")
	}
}

func TestPromptEmptyResponseContinues(t *testing.T) {
	dir, err := createTestingDirectory(testingDirOptions{
		CreateChildrenDirs: true,
		CreateFiles:        true,
	})

	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileHashes, err := CollectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)
	reader := strings.NewReader("\no\n")
	var writer strings.Builder
	promptToDelete(&writer, reader, deleter, duplicates[0])

	if strings.Count(writer.String(), duplicates[0].String()) != 2 {
		t.Error("Duplicate should be displayed.")
	}

	if strings.Count(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") != 2 {
		t.Error("Incorrect prompt.")
	}
}

func TestPromptNoDuplicatesDoesNotPrintAnything(t *testing.T) {
	duplicates := []DuplicateFile{
		DuplicateFile{
			Original: FileHash{
				AbsolutePath: "foo.txt",
				Hash:         "abcd",
			},
			Duplicates: make([]FileHash, 0),
		},
	}

	deleter := MakeDeleter(false)
	reader := strings.NewReader("\no\n")
	var writer strings.Builder
	promptToDelete(&writer, reader, deleter, duplicates[0])

	if strings.Count(writer.String(), duplicates[0].String()) != 0 {
		t.Error("Duplicate should be displayed.")
	}

	if strings.Count(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") != 0 {
		t.Error("Incorrect prompt.")
	}
}

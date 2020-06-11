package muka

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
	if len(duplicateFiles) != 0 {
		t.Errorf("Expected '%d' entry but found '%d' instead.\n", 0, len(duplicateFiles))
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

	muka, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: dir,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(muka) > 0 {
		t.Error("In an empty directory, no muka should be collected.")
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

	muka, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: dir,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(muka) == 0 {
		t.Errorf("Non empty list of muka expected.")
	}

	for _, f := range muka {
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

	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: dir,
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)
	reader := strings.NewReader("o\n")
	var writer strings.Builder
	PromptToDelete(&writer, reader, deleter, duplicates[0])

	if !strings.Contains(writer.String(), duplicates[0].String()) {
		t.Error("Duplicate should be displayed.")
	}

	if !strings.Contains(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") {
		t.Error("Incorrect prompt.")
	}

	if FileExists(duplicates[0].Original.AbsolutePath) {
		t.Error("The file was supposed to be deleted.")
	}

	for _, d := range duplicates[0].Duplicates {
		if !FileExists(d.AbsolutePath) {
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

	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: dir,
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)
	reader := strings.NewReader("d\n")
	var writer strings.Builder
	PromptToDelete(&writer, reader, deleter, duplicates[0])

	if !strings.Contains(writer.String(), duplicates[0].String()) {
		t.Error("Duplicate should be displayed.")
	}

	if !strings.Contains(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") {
		t.Error("Incorrect prompt.")
	}

	if !FileExists(duplicates[0].Original.AbsolutePath) {
		t.Error("The file was not supposed to be deleted.")
	}

	for _, d := range duplicates[0].Duplicates {
		if FileExists(d.AbsolutePath) {
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

	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: dir,
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)
	reader := strings.NewReader("s\n")
	var writer strings.Builder
	PromptToDelete(&writer, reader, deleter, duplicates[0])

	if !strings.Contains(writer.String(), duplicates[0].String()) {
		t.Error("Duplicate should be displayed.")
	}

	if !strings.Contains(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") {
		t.Error("Incorrect prompt.")
	}

	if !FileExists(duplicates[0].Original.AbsolutePath) {
		t.Error("The file was not supposed to be deleted.")
	}

	for _, d := range duplicates[0].Duplicates {
		if !FileExists(d.AbsolutePath) {
			t.Error("The file was not supposed to be deleted.")
		}
	}
}

func TestPromptUnexpectedResponsesContinuePrompting(t *testing.T) {

	dir, err := createTestingDirectory(testingDirOptions{
		CreateChildrenDirs: true,
		CreateFiles:        true,
	})

	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: dir,
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)

	readers := []*strings.Reader{
		strings.NewReader("x\no\n"), // invalid response
		strings.NewReader("\no\n"),  // empty response
	}

	for _, reader := range readers {
		var writer strings.Builder
		PromptToDelete(&writer, reader, deleter, duplicates[0])

		if strings.Count(writer.String(), duplicates[0].String()) != 2 {
			t.Error("Duplicate should be displayed.")
		}

		if strings.Count(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") != 2 {
			t.Error("Incorrect prompt.")
		}
	}
}

func TestForceDeleteOnlyRemovesDuplicates(t *testing.T) {
	dir, err := createTestingDirectory(testingDirOptions{
		CreateChildrenDirs: true,
		CreateFiles:        true,
	})

	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: dir,
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(false)
	ForceDelete(duplicates, deleter)

	if !FileExists(duplicates[0].Original.AbsolutePath) {
		t.Error("The file was not supposed to be deleted.")
	}

	for _, d := range duplicates[0].Duplicates {
		if FileExists(d.AbsolutePath) {
			t.Error("The file was supposed to be deleted.")
		}
	}
}

func TestCompileSpaceSeparatedPatterns(t *testing.T) {
	inputs := map[string]int{
		"a":   1,
		"a b": 2,
		"":    0,
	}

	for input, patternsLength := range inputs {
		patterns, err := CompileSpaceSeparatedPatterns(input)
		if err != nil {
			t.Fatal(err)
		}

		if len(patterns) != patternsLength {
			t.Errorf("mismatched lengths for %q - expected %d but got %d", input, patternsLength, len(patterns))
		}
	}
}

func TestExcludeDirectories(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestIgnore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	ignoreDir, err := ioutil.TempDir(dir, ".ignoreMe")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ioutil.TempFile(ignoreDir, "should-not-be-picked-up"); err != nil {
		t.Fatal(err)
	}

	excludeDirs, err := CompileSpaceSeparatedPatterns(".ignoreMe")
	if err != nil {
		t.Fatal(err)
	}

	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: dir,
		ExcludeDirs:       excludeDirs,
	})

	if err != nil {
		t.Fatal(err)
	}

	for _, f := range fileHashes {
		if strings.Contains(f.AbsolutePath, "should-not-be-picked-up") {
			t.Errorf("%q should be excluded", "should-not-be-picked-up")
		}
	}

}

func TestExcludeFiles(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestIgnore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if _, err := ioutil.TempFile(dir, "should-not-be-picked-up"); err != nil {
		t.Fatal(err)
	}

	excludeFiles, err := CompileSpaceSeparatedPatterns("should-not-be-picked-up")
	if err != nil {
		t.Fatal(err)
	}

	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: dir,
		ExcludeFiles:      excludeFiles,
	})

	if err != nil {
		t.Fatal(err)
	}

	for _, f := range fileHashes {
		if strings.Contains(f.AbsolutePath, "should-not-be-picked-up") {
			t.Errorf("%q should be excluded", "should-not-be-picked-up")
		}
	}

}

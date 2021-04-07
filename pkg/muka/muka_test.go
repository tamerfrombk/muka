package muka

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func getTestingDir(dir string) string {
	return filepath.Join("..", "testdata", dir)
}

func assertEqualsI(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("expected %d but got %d", expected, actual)
	}
}

func assertEqualsF(t *testing.T, expected, actual float64) {
	// avoid float direct comparisons by taking a delta
	if math.Abs(expected-actual) >= 0.01 {
		t.Errorf("expected %f but got %f", expected, actual)
	}
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
			FileData: FileData{
				AbsolutePath: "file1.txt",
			},
			Hash: "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(Directory{HashedFiles: fileHashes})
	assertEqualsI(t, 0, len(duplicateFiles))
}

func TestDuplicateFileWithOneDuplicate(t *testing.T) {
	fileHashes := []FileHash{
		{
			FileData: FileData{
				AbsolutePath: "file1.txt",
			},
			Hash: "abcdefg",
		},
		{
			FileData: FileData{
				AbsolutePath: "file2.txt",
			},
			Hash: "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(Directory{HashedFiles: fileHashes})
	assertEqualsI(t, 1, len(duplicateFiles))
	assertEqualsI(t, 1, len(duplicateFiles[0].Duplicates))
}

func TestDuplicateFileWithMoreThanOneDuplicate(t *testing.T) {
	fileHashes := []FileHash{
		{
			FileData: FileData{
				AbsolutePath: "file1.txt",
			},
			Hash: "abcdefg",
		},
		{
			FileData: FileData{
				AbsolutePath: "file2.txt",
			},
			Hash: "abcdefg",
		},
		{
			FileData: FileData{
				AbsolutePath: "file3.txt",
			},
			Hash: "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(Directory{HashedFiles: fileHashes})
	assertEqualsI(t, 1, len(duplicateFiles))
	assertEqualsI(t, 2, len(duplicateFiles[0].Duplicates))
}

func TestCollectFilesHasOnlyFiles(t *testing.T) {
	d, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: getTestingDir("small"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(d.EncounteredFiles) == 0 {
		t.Error("non empty list of files expected")
	}

	for _, f := range d.EncounteredFiles {
		if isDir(f.AbsolutePath) {
			t.Errorf("%q is a directory", f.AbsolutePath)
		}
	}
}

func TestPromptToDeleteOriginal(t *testing.T) {
	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: getTestingDir("small"),
	})
	if err != nil {
		t.Fatal(err)
	}

	deleter := MakeDeleter(true)
	reader := strings.NewReader("o\n")
	var writer strings.Builder

	duplicates := FindDuplicateFiles(fileHashes)
	deletedFiles, err := PromptToDelete(&writer, reader, deleter, duplicates[0])
	if err != nil {
		t.Fatal(err)
	}

	assertEqualsI(t, 1, len(deletedFiles))

	if !strings.Contains(writer.String(), duplicates[0].String()) {
		t.Error("duplicate should be displayed")
	}

	if !strings.Contains(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") {
		t.Error("incorrect prompt")
	}
}

func TestPromptToDeleteDuplicates(t *testing.T) {
	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: getTestingDir("small"),
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(true)
	reader := strings.NewReader("d\n")
	var writer strings.Builder
	deletedFiles, err := PromptToDelete(&writer, reader, deleter, duplicates[0])
	if err != nil {
		t.Fatal(err)
	}

	assertEqualsI(t, len(deletedFiles), len(duplicates[0].Duplicates))

	if !strings.Contains(writer.String(), duplicates[0].String()) {
		t.Error("duplicate should be displayed")
	}

	if !strings.Contains(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") {
		t.Error("incorrect prompt")
	}
}

func TestPromptSkip(t *testing.T) {
	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: getTestingDir("small"),
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(true)
	reader := strings.NewReader("s\n")
	var writer strings.Builder
	PromptToDelete(&writer, reader, deleter, duplicates[0])

	if !strings.Contains(writer.String(), duplicates[0].String()) {
		t.Error("duplicate should be displayed")
	}

	if !strings.Contains(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > ") {
		t.Error("incorrect prompt")
	}
}

func TestPromptUnexpectedResponsesContinuePrompting(t *testing.T) {
	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: getTestingDir("small"),
	})
	if err != nil {
		t.Fatal(err)
	}

	readers := []*strings.Reader{
		strings.NewReader("x\no\n"), // invalid response
		strings.NewReader("\no\n"),  // empty response
	}

	duplicates := FindDuplicateFiles(fileHashes)
	deleter := MakeDeleter(true)
	for _, reader := range readers {
		var writer strings.Builder
		PromptToDelete(&writer, reader, deleter, duplicates[0])

		assertEqualsI(t, 2, strings.Count(writer.String(), duplicates[0].String()))
		assertEqualsI(t, 2, strings.Count(writer.String(), "Which file(s) do you wish to remove? [o/d/s] > "))
	}
}

func TestForceDeleteOnlyRemovesDuplicates(t *testing.T) {
	fileHashes, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: getTestingDir("small"),
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(fileHashes)

	deleter := MakeDeleter(true)
	deletedFiles := ForceDelete(duplicates, deleter)

	count := 0
	for _, d := range duplicates {
		count += len(d.Duplicates)
	}

	assertEqualsI(t, count, len(deletedFiles))
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
	excludeDirs, err := CompileSpaceSeparatedPatterns("d1")
	if err != nil {
		t.Fatal(err)
	}

	d, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: getTestingDir("small"),
		ExcludeDirs:       excludeDirs,
	})

	if err != nil {
		t.Fatal(err)
	}

	for _, f := range d.HashedFiles {
		if strings.Contains(f.AbsolutePath, "d1") {
			t.Errorf("%q should be excluded", "d1")
		}
	}

}

func TestExcludeFiles(t *testing.T) {
	excludeFiles, err := CompileSpaceSeparatedPatterns("exclude-me")
	if err != nil {
		t.Fatal(err)
	}

	d, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: getTestingDir("small"),
		ExcludeFiles:      excludeFiles,
	})

	if err != nil {
		t.Fatal(err)
	}

	for _, f := range d.HashedFiles {
		if strings.Contains(f.AbsolutePath, "exclude-me") {
			t.Errorf("%q should be excluded", "exclude-me")
		}
	}
}

func TestCalculateReport(t *testing.T) {

	dir, err := CollectFiles(FileCollectionOptions{
		DirectoryToSearch: getTestingDir("small"),
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicates := FindDuplicateFiles(dir)

	report := CalculateReport(dir, duplicates, []FileHash{})

	sumFileDataKB := func(fds []FileData) float64 {
		sum := int64(0)
		for _, f := range fds {
			sum += f.SizeInBytes
		}
		return float64(sum) / 1000.0
	}

	sumHashesKB := func(hashes []FileHash) float64 {
		sum := int64(0)
		for _, f := range hashes {
			sum += f.SizeInBytes
		}
		return float64(sum) / 1000.0
	}

	assertEqualsI(t, len(dir.EncounteredFiles), report.CollectedFileCount)
	assertEqualsF(t, sumFileDataKB(dir.EncounteredFiles), report.CollectedFileSizeInKB)

	duplicateCount := 0
	for _, f := range duplicates {
		duplicateCount += len(f.Duplicates)
	}

	assertEqualsI(t, duplicateCount, report.DuplicateFileCount)

	sumDuplicates := float64(0.0)
	for _, d := range duplicates {
		sumDuplicates += sumHashesKB(d.Duplicates)
	}

	assertEqualsF(t, sumDuplicates, report.DuplicateFileSizeInKB)
	assertEqualsF(t, (float64(duplicateCount)/float64(len(dir.EncounteredFiles)))*100, report.DuplicatePercentage)
	assertEqualsI(t, 0, report.DeletedFileCount)
	assertEqualsF(t, 0.0, report.DeletedFileSizeInKB)
}

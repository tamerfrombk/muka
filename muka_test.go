package main

import (
	"testing"
)

func TestDuplicateFileWithZeroDuplicates(t *testing.T) {
	fileHashes := []FileHash{
		{
			AbsolutePath: "file1.txt",
			Hash:         "abcdefg",
		},
	}

	duplicateFiles := FindDuplicateFiles(fileHashes)
	if len(duplicateFiles) != 0 {
		t.Fatalf("Expected '%d' duplicates but found '%d' instead.\n", 0, len(duplicateFiles))
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
		t.Fatalf("Expected '%d' duplicates but found '%d' instead.\n", 1, len(duplicateFiles))
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
	if len(duplicateFiles) != 2 {
		t.Fatalf("Expected '%d' duplicates but found '%d' instead.\n", 2, len(duplicateFiles))
	}
}

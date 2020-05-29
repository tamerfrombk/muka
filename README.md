# Muka

`muka` is a program for detecting and removing duplicate files from a directory. The word `muka` comes from the word for "duplicate" in Arabic: `mukarar` (مكرر).

## Getting Started

`muka` uses the Go standard library and build tools. 

By default, `muka` will recursively search the current working directory and list all duplicate files.

To review the full help menu, use the `-h` or `--help` flags.

### Examples

List all duplicate files in the current working directory:
```
> cd /tmp
> muka 

'/tmp/file2.md' is a duplicate of '/home/tamer/projects/go/muka/tmp/file1.txt'.
'/tmp/file3.foo' is a duplicate of '/home/tamer/projects/go/muka/tmp/file1.txt'.
```

List all duplicate files in a specific directory:
```
> muka -d /tmp

'/tmp/file2.md' is a duplicate of '/home/tamer/projects/go/muka/tmp/file1.txt'.
'/tmp/file3.foo' is a duplicate of '/home/tamer/projects/go/muka/tmp/file1.txt'.
```

List and interactively remove duplicates:
```
> muka -i

The following are duplicates:
1) /tmp/file1.txt
2) /tmp/file2.md
Which file do you wish to remove? [1/2] >
```

Remove duplicates automatically without prompting:

__Note__: When a file has multiple duplicates, this option will always remove the multiple duplicates over the "original". In the case of a file having a single duplicate, one of them is arbitrarily chosen to be removed.
```
> muka -f

'/tmp/file2.md' was removed.
'/tmp/file3.foo' was removed.
```

### Building

`go build muka.go`

### Installing

This will install `muka` to your `GOBIN` directory.

`go install muka.go`

Alternatively, you can move the `muka` executable after building to a directory of your choice.

### Running the tests

`go test -v`

## Contributing

Simply open a PR for your changes and I'll review it.
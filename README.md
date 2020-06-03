# Muka

`muka` is a program for detecting and removing duplicate files from a directory. The word `muka` is derived from the word for "duplicate" in Arabic: `mukarar` (مكرر).

`muka` attempts to follow the `KISS` principle while keeping data safety in mind. This means `muka` only focuses on delivering its core functionality (detecting and removing duplicate files) without many bells and whistles; `muka` comes with only a handful of command line flags and implements sane and data safe behavior by default.

By default, `muka` will recursively search the current working directory and list all duplicate files _without deleting anything_. This helps keep your data safe against accidental deletion by requiring you to explicitly state you wish to delete files. Additionally, `muka` comes with a dry run mode to simulate deleting files. For data safety reasons, it's recommended to run `muka` in dry run mode when pruning a directory for the first time.

To delete duplicate files, `muka` supports interactive and automatic behavior; use the `-i` flag to interactively delete duplicates or `-f` to have `muka` delete them for you without intervention. Dry run mode can be combined with both of these methods for additional data safety.

__Note__: When a file has multiple duplicates, the `-f` option will always remove the duplicates over the original. This is done to maximize the amount of freed space. In the case of a file having a single duplicate, the duplicate is still removed.

Please exercise caution when deleting files -- especially using `-f`; once a file is deleted, there is no easy way of getting it back.

## Getting Started

`muka` uses the Go standard library and build tools.

To review the full help menu for `muka`, use the `-h` or `--help` flags.

## Examples

List all duplicate files in the current working directory:
```
> cd /tmp
> muka 

Original: /tmp/file1.txt
Duplicates: [ /tmp/file2.md, /tmp/file3.foo ]

```

List all duplicate files in a specific directory:
```
> muka -d /tmp

Original: /tmp/file1.txt
Duplicates: [ /tmp/file2.md, /tmp/file3.foo ]

```

Interactively remove duplicates (`o` is for `original`, `d` is for `duplicates`, and `s` is to skip the current selection):
```
> muka -i

Original: /tmp/file1.txt
Duplicates: [ /tmp/file2.md, /tmp/file3.foo ]

Which file(s) do you wish to remove? [o/d/s] >
```

Remove duplicates automatically without prompting:
```
> muka -f

'/tmp/file2.md' was removed.
'/tmp/file3.foo' was removed.
```

`muka` also has a dry run option that can be combined with interactively or automatically removing files:

Automatic dry run:
```
> muka -f --dryrun

'/tmp/file2.md' would be removed.
'/tmp/file3.foo' would be removed.
```

Interactive dry run:
```
> muka -i --dryrun

Original: /tmp/file1.txt
Duplicates: [ /tmp/file2.md, /tmp/file3.foo ]

Which file(s) do you wish to remove? [o/d/s] > d
'/tmp/file2.md' would be removed.
'/tmp/file3.foo' would be removed.
```

### Building

`go build muka.go`

### Installing

This will install `muka` to your `GOBIN` directory.

`go install muka.go`

Alternatively, you can move the `muka` executable after building to a directory of your choice.

### Running the tests

`go test muka/files`

## Limitations
The following are known limitations of `muka`. Some of these will be built into the program in the future and some may not:

1. Specifying a recursion depth
    - As of now, `muka` does a full recursive search of the specified directory.
2. File filtering
    - As of now, `muka` looks for duplicates in all files under the specified directory.

## Contributing

Simply open a PR with your changes and I'll review it.
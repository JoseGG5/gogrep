`gogrep` is a small grep-like CLI tool written in Go. It searches for text or regex patterns in one or more files, and it can also search directories recursively.

## Compile

From the project root, run:

```bash
go build -o gogrep .
```

On Windows this creates `gogrep.exe`.

## Usage

```bash
gogrep [options] <pattern> <path1> <path2> ...
```

- `<pattern>` is the text or regex you want to search for.
- `<path1> <path2> ...` are one or more files.
- When using `-r`, the paths can also be directories.

## Options

- `-n`: print line numbers.
- `-i`: case-insensitive search.
- `-r`: recursively search inside directories.

## Examples

Search in one file:

```bash
gogrep hello test.txt
```

Search in multiple files:

```bash
gogrep hello file1.txt file2.txt
```

Show line numbers:

```bash
gogrep -n hello test.txt
```

Case-insensitive search:

```bash
gogrep -i hello test.txt
```

Search recursively in the current folder:

```bash
gogrep -r hello .
```

Use a regex pattern:

```bash
gogrep "^func" main.go
```

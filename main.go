package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type resultRecord struct {
	nline   int
	content string
}

func processFile(filePath string, pattern string, nFlag *bool, iFlag *bool, wg *sync.WaitGroup) error {
	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Failed to open file %w:", err)
	}
	defer file.Close() // close the file at the end

	// Create a bufio scanner to iterate line by line
	scanner := bufio.NewScanner(file)

	// Read line by line
	var nline int = 1
	for scanner.Scan() {
		line := scanner.Text()

		var isContained bool
		if *iFlag {
			isContained = strings.Contains(strings.ToLower(line), strings.ToLower(pattern))
		} else {
			isContained = strings.Contains(line, pattern)
		}

		if isContained {
			record := resultRecord{nline, line}
			if *nFlag {
				fmt.Println(filePath, record.nline, record.content)
			} else {
				fmt.Println(filePath, record.content)
			}
		}
		nline += 1
	}

	// check if we got to the end of the file or there was an error
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Failed to iterate through the file %w:", err)
	}

	return nil
}

func processFolder(
	folderPath string,
	wg *sync.WaitGroup,
	files chan<- string) error {

	defer wg.Done()

	filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("Error traversing folder %w:", err)
		}

		if d.IsDir() == false {
			files <- path
		}

		return nil

	})

	return nil
}

func main() {
	var nFlag = flag.Bool("n", false, "Print line number") // returns a pointer to a bool value
	var iFlag = flag.Bool("i", false, "Case insensitive mode")
	var rFlag = flag.Bool("r", false, "Recursive search")
	flag.Parse()

	args := flag.Args()
	if len(args) <= 1 {
		fmt.Fprintln(os.Stderr, "usage: gogrep [-n] [-i] <pattern> <file1> ... <filen>")
		os.Exit(1)
	}

	pattern := args[0]
	items := args[1:] // Could be files or folders

	var wg sync.WaitGroup         // To avoid main to finish before the goroutines
	fileChan := make(chan string) // Used to recover channels in case we are traversing folders

	for _, item := range items {
		wg.Add(1)

		if !*rFlag {
			go processFile(item, pattern, nFlag, iFlag, &wg)
		} else {

			// The idea is that there is a goroutine per folder sent by user
			// and each goroutine traverses the directory recursively and
			// sends files found to a channel. Then N workers grab files from the channel
			// and process them

			go processFolder(item, &wg, fileChan)

		}

	}

	wg.Wait()
}

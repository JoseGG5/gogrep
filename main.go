package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

const NumWorkers = 3

type resultRecord struct {
	nline   int
	content string
}

func processFile(filePath string, re *regexp.Regexp, nFlag *bool) error {

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

		isContained := re.MatchString(line)

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

	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("Error traversing folder %w:", err)
		}

		if d.IsDir() == false {
			files <- path
		}

		return nil

	})

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func workerFile(re *regexp.Regexp, nFlag *bool, channel <-chan string, fileWg *sync.WaitGroup) {
	defer fileWg.Done()

	for file := range channel { // This keeps going until chan is closed
		err := processFile(file, re, nFlag)
		if err != nil {
			fmt.Println(err)
		}
	}
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

	// Compile the pattern
	if *iFlag {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var itemWg sync.WaitGroup // To avoid main to finish before the goroutines
	var fileWg sync.WaitGroup
	fileChan := make(chan string) // Used to recover channels in case we are traversing folders

	// Create workers prior to create goroutines that send tasks through the channel if r is setted
	if *rFlag {
		for i := 0; i < NumWorkers; i++ {
			fileWg.Add(1)
			go workerFile(re, nFlag, fileChan, &fileWg)
		}
	}

	for _, item := range items {
		itemWg.Add(1)

		if !*rFlag {
			go func(item string) {
				defer itemWg.Done()
				err := processFile(item, re, nFlag)
				if err != nil {
					fmt.Println("error: ", err)
				}
			}(item)
		} else {
			// The idea is that there is a goroutine per folder sent by user
			// and each goroutine traverses the directory recursively and
			// sends files found to a channel. Then N workers grab files from the channel
			// and process them
			go processFolder(item, &itemWg, fileChan)
		}
	}

	itemWg.Wait()
	close(fileChan) // We close the channel so that we can then wait for workers to finish

	fileWg.Wait()

}

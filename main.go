package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const filePath = "messages.txt"

func main() {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Err opening messages.txt: %v\n", err)
	}
	fmt.Printf("Reading file: %s\n", filePath)
	fmt.Println(strings.Repeat("=", len(filePath)+14))

	linesChan := getLinesChannel(f)

	for line := range linesChan {
		fmt.Println("read:", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
		defer close(lines)
		currLine := ""
		for {
			buff := make([]byte, 8, 8)
			n, err := f.Read(buff)
			if err != nil {
				if currLine != "" {
					lines <- currLine
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				break
			}
			curr_read := string(buff[:n])
			parts := strings.Split(curr_read, "\n")
			for _, part := range parts[:len(parts)-1] {
				lines <- fmt.Sprintf("%s%s", currLine, part)
				currLine = ""
			}
			currLine += parts[len(parts)-1]
		}

	}()
	return lines
}

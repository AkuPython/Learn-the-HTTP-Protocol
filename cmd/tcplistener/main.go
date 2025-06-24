package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

const host = "127.0.0.1:42069"

func main() {
	s, err := net.Listen("tcp4", host)
	if err != nil {
		fmt.Printf("Error creating socket: %v\n", err)
	}
	defer s.Close()
	fmt.Printf("Listening on: %s\n", host)
	fmt.Println(strings.Repeat("=", len(host)+14))

	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
		}
		fmt.Printf("Connection accepted from: %v\n", c.RemoteAddr().String())

		lineChan := getLinesChannel(c)
		for line := range lineChan {
			fmt.Println(line)
		}

		fmt.Printf("Connection closed from: %v\n", c.RemoteAddr().String())
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

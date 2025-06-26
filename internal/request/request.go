package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type State int

const (
	initialized = iota
	done
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == initialized {
		rl, b, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if b == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.state = done
		return b, nil
	}
	if r.state == done {
		return 0, errors.New("error: trying to read data in done state")
	}
	return 0, errors.New("error: unknown state")
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := Request{state: initialized}
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	for r.state != done {
		if readToIndex >= cap(buf) {
			buf = append(buf, make([]byte, len(buf)*2, len(buf)*2)...)
		}
		i, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				r.state = done
				break
			}
			return nil, errors.New(fmt.Sprintf("Could not read from reader: %v\n", err))
		}
		readToIndex += i

		i, err = r.parse(buf[:readToIndex])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Could not parse request line: %v\n", err))
		}
		if i > 0 {
			copy(buf, buf[i:])
			readToIndex -= i
		}
	}
	return &r, nil
}

func isUpper(s string) bool {
	for _, charNumber := range s {
		if charNumber > 90 || charNumber < 65 {
			return false
		}
	}
	return true
}

func parseRequestLine(request []byte) (*RequestLine, int, error) {
	splitReq := strings.Split(string(request), "\r\n")
	if len(splitReq) < 2 {
		return nil, 0, nil
	}
	firstLine := splitReq[0]
	parts := strings.Split(firstLine, " ")
	if len(parts) != 3 {
		return nil, 0, errors.New(fmt.Sprintf("Invalid request line: %s\n", firstLine))
	}
	method := parts[0]
	target := parts[1]
	version := parts[2]

	if version != "HTTP/1.1" {
		return nil, 0, errors.New(fmt.Sprintf("Invalid HTTP Version: %s\n", version))
	}
	if !isUpper(method) {
		return nil, 0, errors.New(fmt.Sprintf("Invalid Method: %s\n", method))
	}
	return &RequestLine{HttpVersion: strings.Split(version, "/")[1], RequestTarget: target, Method: method}, len([]byte(firstLine + "\r\n")), nil
}

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

type Request struct {
	RequestLine RequestLine
	State       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	if r.State == initialized {
		rl, b, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if b == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.State = done
		return b, nil
	}
	if r.State == done {
		return 0, errors.New("error: trying to read data in done state")
	}
	return 0, errors.New("error: unknown state")
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not read from reader: %v\n", err))
	}
	reqLine, consumed, err := parseRequestLine(r)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse request line: %v\n", err))
	}
	if consumed == 0 {

	}
	return &Request{
		RequestLine: *reqLine,
	}, nil
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

package request

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/AkuPython/Learn-the-HTTP-Protocol/internal/headers"
)

type State int

const (
	initialized = iota
	requestStateParsingHeaders
	done
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	bytesParsed := 0
	for r.state != done {
		n, err := r.parseSingle(data[bytesParsed:])
		if err != nil {
			return 0, err
		}
		bytesParsed += n
		if n == 0 {
			break
		}
	}
	return bytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case initialized:
		rl, b, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if b == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.state = requestStateParsingHeaders
		return b, nil
	case requestStateParsingHeaders:
		i, d, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if i == 0 {
			return 0, nil
		}
		if d {
			r.state = done
		}
		return i, nil
	case done:
		return 0, errors.New("error: trying to read data in done state")
	default:
		return 0, errors.New("error: unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := Request{
		state:   initialized,
		Headers: headers.NewHeaders(),
	}
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	for r.state != done {
		if readToIndex >= len(buf) {
			buf = append(buf, make([]byte, len(buf)*2)...)
		}
		i, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if r.state != done {
					return nil, errors.New(fmt.Sprintf("Error: Incomplete request, in state %v, read %v bytes on EOF\n", r.state, i))
				}
				break
			}
			return nil, errors.New(fmt.Sprintf("Error: Could not read from reader: %v\n", err))
		}
		readToIndex += i

		i, err = r.parse(buf[:readToIndex])
		if err != nil {
			switch r.state {
			case initialized:
				return nil, errors.New(fmt.Sprintf("Error: Could not parse request line: %v\n", err))
			case requestStateParsingHeaders:
				return nil, errors.New(fmt.Sprintf("Error: Could not parse headers: %v\n", err))
			}
		}
		copy(buf, buf[i:])
		readToIndex -= i
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

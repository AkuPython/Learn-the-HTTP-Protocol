package response

import (
	"fmt"
	"io"

	"github.com/AkuPython/Learn-the-HTTP-Protocol/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := "HTTP/1.1 " + fmt.Sprintf("%d", statusCode) + " "
	switch statusCode {
	case StatusOK:
		statusLine += "OK"
	case StatusBadRequest:
		statusLine += "Bad Request"
	case StatusInternalServerError:
		statusLine += "Internal Server Error"
	}
	statusLine += "\r\n"
	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	newHeaders := headers.NewHeaders()
	newHeaders.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	newHeaders.Set("Connnection", "close")
	newHeaders.Set("Content-Type", "text/plain")

	return newHeaders
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

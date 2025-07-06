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

type Writer struct {
	W           io.Writer
	writerState WriterState
}

type WriterState int

const (
	WriteStatusLineState = iota
	WriteHeadersState
	WriteBodyState
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != WriteStatusLineState {
		return fmt.Errorf("Incorrect writer state: %d - WriteStatusLine should be called first", w.writerState)
	}
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
	_, err := w.W.Write([]byte(statusLine))
	w.writerState = WriteHeadersState
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	newHeaders := headers.NewHeaders()
	newHeaders.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	newHeaders.Set("Connnection", "close")
	newHeaders.Set("Content-Type", "text/plain")

	return newHeaders
}

func internalWriteHeaders(w *Writer, headers headers.Headers) error {
	for k, v := range headers {
		p := []byte{}
		_, err := w.W.Write(fmt.Appendf(p, "%s: %s\r\n", k, v))
		if err != nil {
			return err
		}
	}

	_, err := w.W.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.writerState = WriteBodyState
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != WriteHeadersState {
		return fmt.Errorf("Incorrect writer state %d - WriteHeaders should be called second", w.writerState)
	}
	return internalWriteHeaders(w, headers)
}

func (w *Writer) WriteTrailers(headers headers.Headers) error {
	payload := []byte("0\r\n")
	_, err := w.WriteBody(payload)
	if err != nil {
		return fmt.Errorf("error writing Trailer: %v", err)
	}
	return internalWriteHeaders(w, headers)
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != WriteBodyState {
		return 0, fmt.Errorf("Incorrect writer state %d - WriteBody should be called last", w.writerState)
	}
	return w.W.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	payload := []byte{}
	payload = fmt.Appendf(payload, "%X\r\n", len(p))
	payload = append(payload, p...)
	payload = append(payload, []byte("\r\n")...)
	return w.WriteBody(payload)
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	p := make([]byte, 0, 0)
	return w.WriteChunkedBody(p)
}

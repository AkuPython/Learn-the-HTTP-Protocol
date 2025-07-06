package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/AkuPython/Learn-the-HTTP-Protocol/internal/request"
	"github.com/AkuPython/Learn-the-HTTP-Protocol/internal/response"
	"github.com/AkuPython/Learn-the-HTTP-Protocol/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func basicHtmlWriter(w *response.Writer, req *request.Request, sc response.StatusCode) {
	sl := ""
	reason := ""
	switch sc {
	case response.StatusOK:
		sl += "OK"
		reason = "Your request was an absolute banger."
	case response.StatusBadRequest:
		sl += "Bad Request"
		reason = "Your request honestly kinda sucked."
	case response.StatusInternalServerError:
		sl += "Internal Server Error"
		reason = "Okay, you know what? This one is on me."
	}

	payload := []byte("<html>" +
		"  <head>" +
		fmt.Sprintf("	<title>%d %s</title>", sc, sl) +
		"  </head>" +
		"  <body>" +
		fmt.Sprintf("	<h1>%s</h1>", sl) +
		fmt.Sprintf("	<p>%s</p>", reason) +
		"  </body>" +
		"</html>")
	err := w.WriteStatusLine(sc)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	h := response.GetDefaultHeaders(len(payload))
	h.Override("Content-Type", "text/html")
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
	i, err := w.WriteBody(payload)
	if err != nil {
		log.Printf("Error writing body: %v", err)
		return
	}
	log.Printf("bytes written: %d\n", i)
}

func httpbinWriter(w *response.Writer, req *request.Request) {
	destUrl := "https://httpbin.org" + strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	fmt.Println("Proxying to", destUrl)

	resp, err := http.Get(destUrl)
	if err != nil {
		log.Printf("Error: request to httpbin.com failed: %v", err)
		basicHtmlWriter(w, req, response.StatusInternalServerError)
		return

	}
	sc := response.StatusOK

	err = w.WriteStatusLine(sc)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	h := response.GetDefaultHeaders(0)
	h.Remove("content-length")
	h.Set("transfer-encoding", "chunked")

	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
	for {
		buf := make([]byte, 1024, 1024)
		int, err := resp.Body.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				_, err = w.WriteChunkedBodyDone()
				if err != nil {
					log.Printf("Error writing chunked body end: %v", err)
				}
				return
			}
			log.Printf("Error reading from httpbin to buffer: %v", err)
			basicHtmlWriter(w, req, response.StatusInternalServerError)
			break
		}
		log.Printf("Bytes read: %d", int)
		_, err = w.WriteChunkedBody(buf)
		if err != nil {
			log.Printf("Error writing chunked body: %v", err)
			basicHtmlWriter(w, req, response.StatusInternalServerError)
			return
		}
	}

}

func handler(w *response.Writer, req *request.Request) {
	var sc response.StatusCode
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		httpbinWriter(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/yourproblem" {
		sc = response.StatusBadRequest
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		sc = response.StatusInternalServerError
	} else {
		sc = response.StatusOK

	}
	basicHtmlWriter(w, req, sc)
	return

}

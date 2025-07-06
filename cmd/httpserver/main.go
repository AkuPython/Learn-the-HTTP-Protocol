package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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

func handler(w *response.Writer, req *request.Request) {
	var sc response.StatusCode
	sl := ""
	reason := ""
	if req.RequestLine.RequestTarget == "/yourproblem" {
		sc = response.StatusBadRequest
		reason = "Your request honestly kinda sucked."
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		sc = response.StatusInternalServerError
		reason = "Okay, you know what? This one is on me."
	} else {
		sc = response.StatusOK
		reason = "Your request was an absolute banger."

	}
	switch sc {
	case response.StatusOK:
		sl += "OK"
	case response.StatusBadRequest:
		sl += "Bad Request"
	case response.StatusInternalServerError:
		sl += "Internal Server Error"
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

package main

import (
	"HTTP_FROM_TCP/internal/headers"
	"HTTP_FROM_TCP/internal/request"
	"HTTP_FROM_TCP/internal/response"
	"HTTP_FROM_TCP/internal/server"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

var badRequestBody = `
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`

var internalServerErrorBody = `
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`

var allGoodBody = `
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`

var videoReadErrorBody = `
<html>
	<head>
		<title>500 Internal Server Error</title>
	</head>
	<body>
		<h1>Internal Server Error</h1>
		<p>Something bad happened when reading the video file</p>
	</body>
</html>
`

func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request, h *headers.Headers) error {
		var status response.StatusCode
		var responseBody []byte
		if req.RequestLine.RequestTarget == "/yourproblem" {
			status = response.StatusBadRequest
			h.Update("Content-type", "text/html")
			h.Update("Content-length", fmt.Sprintf("%d", len(badRequestBody)))
			responseBody = []byte(badRequestBody)
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			status = response.StatusInternalServerError
			h.Update("Content-type", "text/html")
			h.Update("Content-length", fmt.Sprintf("%d", len(internalServerErrorBody)))
			responseBody = []byte(internalServerErrorBody)
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			streamN := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
			// fmt.Printf("stream/x: %s\n", streamN)

			// TODO can implement own http.Get
			res, err := http.Get("https://httpbin.org/" + streamN)
			if err != nil {
				return err
			}

			h.Delete("Content-length")
			h.Set("Transfer-Encoding", "chunked")
			h.Set("Trailer", "X-Content-SHA256")
			h.Set("Trailer", "X-Content-Length")

			_ = w.WriteStatusLine(response.StatusOK)
			_ = w.WriteHeaders(h)

			fullBody := []byte{}
			for {
				buff := make([]byte, 32)
				n, err := res.Body.Read(buff)
				if err != nil {
					if err == io.EOF {
						fmt.Println("Brother EOF, I'm done")
					}
					break
				}
				fullBody = append(fullBody, buff[:n]...)
				_, _ = w.WriteChunkedBody(n, buff[:n])
			}
			_, _ = w.WriteChunkedBodyDone()

			trailers := headers.NewHeaders()
			sha := sha256.Sum256(fullBody)
			trailers.Set("X-Content-SHA256", toStr(sha[:]))
			trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

			_ = w.WriteTrailers(trailers)
			return nil
		} else if req.RequestLine.RequestTarget == "/video" {
			f, err := os.ReadFile("D:/httpfromtcp_go/assets/vim.mp4")
			if err != nil {
				status = response.StatusInternalServerError
				h.Update("Content-type", "text/html")
				h.Update("Content-length", fmt.Sprintf("%d", len(videoReadErrorBody)))
				responseBody = []byte(videoReadErrorBody)
			} else {
				status = response.StatusOK
				h.Update("Content-type", "video/mp4")
				h.Update("Content-length", fmt.Sprintf("%d", len(f)))
				responseBody = []byte(f)
			}
		} else {
			status = response.StatusOK
			h.Update("Content-type", "text/html")
			h.Update("Content-length", fmt.Sprintf("%d", len(allGoodBody)))
			responseBody = []byte(allGoodBody)
		}

		err := w.WriteStatusLine(status)
		if err != nil {
			return err
		}
		err = w.WriteHeaders(h)
		if err != nil {
			return err
		}
		_, err = w.WriteBody(responseBody)
		if err != nil {
			return err
		}
		return nil
	})
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

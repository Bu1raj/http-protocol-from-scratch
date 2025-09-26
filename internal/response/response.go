package response

import (
	"HTTP_FROM_TCP/internal/headers"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type StatusCode int
type writeState string

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

const (
	statusLine     writeState = "status-line"
	responseHeader writeState = "response-headers"
	responseBody   writeState = "response-body"
	responseTrailers writeState = "response-trailers"
	done           writeState = "done"
)

var ErrIncorrectResponseFormat = errors.New("incorrect order of writing to response")

type Writer struct {
	connection io.Writer
	writeState writeState
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		connection: conn,
		writeState: statusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writeState != statusLine {
		return ErrIncorrectResponseFormat
	}

	statusLine := []byte{}
	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	}

	_, err := w.connection.Write(statusLine)
	if err != nil {
		return err
	}
	w.writeState = responseHeader
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := headers.NewHeaders()

	defaultHeaders.Set("Content-Length", strconv.Itoa(contentLen))
	defaultHeaders.Set("Connection", "close")
	defaultHeaders.Set("Content-Type", "text/plain")

	return *defaultHeaders
}

func (w *Writer) WriteHeaders(headers *headers.Headers) error {
	if w.writeState != responseHeader {
		return ErrIncorrectResponseFormat
	}

	headersString := ""

	headers.ForEach(func(k, v string) {
		headersString += fmt.Sprintf("%s: %s\r\n", k, v)
	})
	headersString += "\r\n" // extra CRLF to separate the headers from the body
	_, err := w.connection.Write([]byte(headersString))
	if err != nil {
		return err
	}
	w.writeState = responseBody
	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.writeState != responseBody {
		return 0, ErrIncorrectResponseFormat
	}
	w.writeState = done
	return w.connection.Write([]byte(body))
}

func (w *Writer) WriteChunkedBody(n int, p []byte) (int, error) {
	if w.writeState != responseBody {
		return 0, ErrIncorrectResponseFormat
	}
	noBytes := 0
	n, err := fmt.Fprintf(w.connection, "%x\r\n", n)
	if err != nil {
		return 0, err
	}
	noBytes += n
	n, err = w.connection.Write(p)
	if err != nil {
		return n, err
	}
	noBytes += n
	n, err = w.connection.Write([]byte("\r\n"));
	if err != nil {
		return n, err
	}
	noBytes += n
	return noBytes, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	w.writeState = responseTrailers
	return w.connection.Write([]byte("0\r\n"))
}

func (w * Writer) WriteTrailers(h *headers.Headers) error {
	if w.writeState != responseTrailers {
		return ErrIncorrectResponseFormat
	}

	trailersString := ""

	h.ForEach(func(k, v string) {
		trailersString += fmt.Sprintf("%s: %s\r\n", k, v)
	})
	trailersString += "\r\n" // extra CRLF to separate to close the request
	_, err := w.connection.Write([]byte(trailersString))
	if err != nil {
		return err
	}
	w.writeState = done
	return nil
}

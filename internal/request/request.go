package request

import (
	"HTTP_FROM_TCP/internal/headers"
	"bytes"
	"errors"
	"io"
	"strconv"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
	state       parserState
}

func newRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		state:   StateInit,
	}
}

var SEPARATOR = []byte("\r\n")

var ErrInvalidHTTPVersion = errors.New("HTTP version not supported")
var ErrMalformedRequestLine = errors.New("malformed request line")
var ErrMalformedHTTPVersion = errors.New("malformed HTTP version")
var ErrRequestInErrorState = errors.New("something wrong with the request line")

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	read := bytes.Index(data, SEPARATOR)
	if read == -1 {
		return nil, 0, nil
	}
	line := data[:read]

	parts := bytes.Split(line, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, ErrMalformedRequestLine
	}

	http_parts := bytes.Split(parts[2], []byte("/"))
	if len(http_parts) != 2 || string(http_parts[0]) != "HTTP" || string(http_parts[1]) != "1.1" {
		return nil, 0, ErrInvalidHTTPVersion
	}

	req_line := &RequestLine{
		HttpVersion:   string(http_parts[1]),
		RequestTarget: string(parts[1]),
		Method:        string(parts[0]),
	}

	return req_line, read, nil
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}

	valInt, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}

	return valInt
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		switch r.state {
		case StateError:
			return 0, ErrRequestInErrorState
		case StateInit:
			req_line, num_read, err := parseRequestLine(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if num_read == 0 {
				break outer
			}
			r.RequestLine = *req_line
			read += (num_read + 2)
			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}

			read += n
	
			if done {
				read += len(SEPARATOR)
				r.state = StateBody
			} else if n == 0 {
				break outer
			}
			
			// if n == 0 {
			// 	break outer
			// }
		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			
			if length == 0 {
				r.state = StateDone
				break outer
			}

			remaining := min(length-len(r.Body), len(data[read:]))
			if remaining == 0 && len(data[read:]) == 0 {
				break outer
			}
			r.Body = append(r.Body, data[read:read+remaining]...)
			read += remaining

			if len(r.Body) == length {
				r.state = StateDone
			}
		case StateDone:
			break outer
		default:
			panic("have done something wrong")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func (r *Request) error() bool {
	return r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()
	// NOTE: buffer can get overrun, header that exceeds 1k would do that
	buf := make([]byte, 1024)
	buf_len := 0

	for !req.done() && !req.error() {
		n, err := reader.Read(buf[buf_len:])
		// TODO what to do here
		if err != nil {
			// if err == io.EOF {
			// 	// req.state = StateDone
			// 	break
			// }
			return nil, err
		}
		buf_len += n
		no_parsed, err := req.parse(buf[:buf_len])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[no_parsed:buf_len])
		buf_len -= no_parsed
	}

	return req, nil
}

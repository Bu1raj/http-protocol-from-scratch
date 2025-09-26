package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var rn = []byte("\r\n")
var specialChars = []byte("!#$%&'*+-.^_`|~")

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func validHeaderName(name []byte) bool {
	for _, b := range name {
		if !((b >= 'A' && b <= 'Z') ||
			(b >= 'a' && b <= 'z') ||
			(b >= '0' && b <= '9') ||
			bytes.ContainsRune(specialChars, rune(b))) {
			return false
		}
	}
	return true
}

func parseHeaderLine(data []byte) (string, string, error) {
	parts := bytes.SplitN(data, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed header line")
	}

	fieldName := parts[0]
	fieldValue := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(fieldName, []byte(" ")) ||
		len(fieldName) < 1 ||
		!validHeaderName(fieldName) {
		return "", "", fmt.Errorf("malformed header field name")
	}

	return string(fieldName), string(fieldValue), nil
}

type Headers struct {
	headers map[string]string
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)

	if val, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", val, value)
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) Update(name, value string) {
	name = strings.ToLower(name)
	h.headers[name] = value
}

func (h *Headers) Get(name string) (string, bool) {
	valueStr, ok := h.headers[strings.ToLower(name)]
	return valueStr, ok
}

func (h *Headers) ForEach(kv func(k, v string)) {
	for key, val := range h.headers {
		kv(key, val)
	}
}

func (h *Headers) Delete(name string) {
	name = strings.ToLower(name)
	delete(h.headers, name)
}


func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}

		//EMPTY HEADER
		if idx == 0 {
			done = true
			break
		}

		name, value, err := parseHeaderLine(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		read += (idx + len(rn))
		h.Set(name, value)
	}
	return read, done, nil
}
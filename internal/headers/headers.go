package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	i := bytes.Index(data, []byte(crlf))
	if i == -1 {
		return 0, false, nil
	}
	if i == 0 {
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:i], []byte(":"), 2)
	key := strings.ToLower(string(parts[0]))

	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)
	if !validTokens([]byte(key)) {
		return 0, false, fmt.Errorf("invalid header token found: %s", key)
	}
	h.Set(key, string(value))
	return i + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if v, ok := h[key]; ok {
		value = strings.Join([]string{v, value}, ", ")
	}
	h[key] = value
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	v, ok := h[key]
	return v, ok
}

var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

func validTokens(data []byte) bool {
	for _, c := range data {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}

func isTokenChar(c byte) bool {
	if c >= 'A' && c <= 'Z' ||
		c >= 'a' && c <= 'z' ||
		c >= '0' && c <= '9' {
		return true
	}

	return slices.Contains(tokenChars, c)
}

package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/d-darac/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine    RequestLine
	Headers        headers.Headers
	Body           []byte
	state          parserState
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const bufferSize = 8
const crlf = "\r\n"
const (
	stateInit           parserState = "init"
	stateDone           parserState = "done"
	stateParsingHeaders parserState = "parsing_headers"
	stateParsingBody    parserState = "parsing_body"
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state:   stateInit,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
	buf := make([]byte, bufferSize)
	readToIndex := 0

	for !req.done() {
		if readToIndex >= len(buf) {
			temp := make([]byte, len(buf)*2)
			copy(temp, buf)
			buf = temp
		}

		nRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != stateDone {
					return nil, fmt.Errorf("incomplete request, in state: %v, read n bytes on EOF: %d", req.state, nRead)
				}
				break
			}
			return nil, err
		}

		readToIndex += nRead

		nParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[nParsed:])
		readToIndex -= nParsed
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for !r.done() {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case stateInit:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.state = stateParsingHeaders
		return n, nil
	case stateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = stateParsingBody
		}
		return n, nil
	case stateParsingBody:
		clString, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.state = stateDone
			return len(data), nil
		}
		cl, err := strconv.Atoi(clString)
		if err != nil {
			return 0, fmt.Errorf("malformed Content-Length: %s", clString)
		}
		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)
		if r.bodyLengthRead > cl {
			return 0, errors.New("Content-Length too large")
		}
		if r.bodyLengthRead == cl {
			r.state = stateDone
		}
		return len(data), nil
	case stateDone:
		return 0, errors.New("error: trying to read data in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	i := bytes.Index(data, []byte(crlf))
	if i == -1 {
		return nil, 0, nil
	}
	str := string(data[:i])
	requestLine, err := requestLineFromString(str)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, i + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}

func (r *Request) done() bool {
	return r.state == stateDone
}

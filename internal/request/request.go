package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/GhostVox/httptcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	state       state
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type state int

const (
	requestIntialized state = iota
	requestStateParsingHeaders
	requestParsingBody
	requestDone
)
const buffSize int = 8

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, buffSize, buffSize)

	readToIndex := 0
	request := &Request{
		state:   requestIntialized,
		Headers: headers.NewHeaders(),
	}
	for request.state != requestDone {
		if readToIndex >= len(buf) {
			newBuff := make([]byte, len(buf)*2)
			copy(newBuff, buf)
			buf = newBuff
		}
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.state != requestDone {
					return nil, fmt.Errorf("unexpected EOF")
				}

				if request.RequestLine == (RequestLine{}) {
					return nil, fmt.Errorf("no request-line found")
				}
				break

			}
			return nil, fmt.Errorf("error reading from reader: %w", err)
		}
		readToIndex += n
		bytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("error parsing request: %w", err)
		}
		copy(buf, buf[bytesParsed:])
		readToIndex -= bytesParsed

	}
	return request, nil

}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + len(crlf), nil
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

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestDone {
		bytesParsed, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += bytesParsed
		if bytesParsed == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestIntialized:
		requestLine, bytesParsed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if bytesParsed == 0 {
			return bytesParsed, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return bytesParsed, nil
	case requestStateParsingHeaders:
		bytesParsed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestParsingBody
		}
		return bytesParsed, nil
	case requestParsingBody:
		contentLength := r.Headers.Get("Content-Length")
		if contentLength == "" {
			r.state = requestDone
			return 0, nil
		}
		cLength, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, fmt.Errorf("invalid content-length: %s", contentLength)
		}
		fmt.Println("Content-Length: ", cLength, "data: ", data)
		r.Body = append(r.Body, data...)
		if len(r.Body) > cLength {
			return 0, fmt.Errorf("body is larger than content-length")
		}
		if cLength == len(r.Body) {
			r.state = requestDone
		}
		fmt.Printf("Content-Length: %d, BodyLength: %d", cLength, len(r.Body))
		return len(data), nil

	case requestDone:
		return 0, fmt.Errorf("trying to read data in done state")

	default:
		return 0, fmt.Errorf("unknown state: %d", r.state)
	}
}

package main

import (
	"bufio"
	"errors"
	"net"
	"slices"
	"strings"
)

type Request struct {
	Method        string
	RequestTarget string
	HTTPVersion   string
	Headers       map[string]string
	Body          string
}

const (
	HeaderKeyHost = "host"
)

var (
	ErrInvalidRequestLine   = errors.New("invalid request line")
	ErrInvalidMethod        = errors.New("invalid request method")
	ErrInvalidHTTPVersion   = errors.New("invalid HTTP version")
	ErrMissingHostHeader    = errors.New("missing Host header")
	ErrMultipleHostHeader   = errors.New("multiple Host headers")
	ErrInvalidRequestTarget = errors.New("invalid request target")
	errEmptyHeaderFieldName = errors.New("empty header field name")
	methods                 = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
)

func (r *Request) parseRequestLine(rl string) error {
	rlc := strings.Split(rl, " ")
	if len(rlc) != 3 {
		return ErrInvalidRequestLine
	}

	if !slices.Contains(methods, rlc[0]) {
		return ErrInvalidMethod
	}

	if rlc[1][0] != '/' {
		return ErrInvalidRequestTarget
	}

	if rlc[2] != "HTTP/1.1" {
		return ErrInvalidHTTPVersion
	}

	r.Method = rlc[0]
	r.RequestTarget = rlc[1]
	r.HTTPVersion = rlc[2]

	return nil
}

func (r *Request) parseHeader(header string) error {
	if header == "" {
		return errEmptyHeaderFieldName
	}

	h := strings.SplitN(header, ":", 2)
	key := strings.ToLower(h[0])
	if key == HeaderKeyHost {
		if _, ok := r.Headers[HeaderKeyHost]; ok {
			return ErrMultipleHostHeader
		}
	}

	var headerValue string
	if len(h[1]) < 2 {
		headerValue = ""
	} else {
		headerValue = h[1][1:]
	}
	r.Headers[key] = headerValue

	return nil
}

func parseRequest(c net.Conn) (*Request, error) {
	// create scanner to read line by line
	sc := bufio.NewScanner(c)
	sc.Split(bufio.ScanLines)

	request := Request{}
	request.Headers = make(map[string]string)

	// parse request line
	rl := sc.Scan()
	if !rl {
		return nil, ErrInvalidRequestLine
	}
	err := request.parseRequestLine(sc.Text())
	if err != nil {
		return nil, err
	}

	// parse headers
	for sc.Scan() {
		header := sc.Text()

		err := request.parseHeader(header)
		if err != nil {
			if err == errEmptyHeaderFieldName {
				break
			}
			return nil, err
		}
	}
	if _, ok := request.Headers[HeaderKeyHost]; !ok {
		return nil, ErrMissingHostHeader
	}

	return &request, nil
}

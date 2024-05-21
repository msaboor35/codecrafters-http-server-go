package main

import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	Method        string
	RequestTarget string
	HTTPVersion   string
	Headers       map[string]string
	ContentLength int
	Body          []byte
}

const (
	HeaderKeyHost = "host"
)

var (
	ErrInvalidRequestLine          = errors.New("invalid request line")
	ErrInvalidMethod               = errors.New("invalid request method")
	ErrInvalidHTTPVersion          = errors.New("invalid HTTP version")
	ErrMissingHostHeader           = errors.New("missing Host header")
	ErrMultipleHostHeader          = errors.New("multiple Host headers")
	ErrInvalidRequestTarget        = errors.New("invalid request target")
	errEmptyHeaderFieldName        = errors.New("empty header field name")
	ErrUnsupportedTransferEncoding = errors.New("unsupported transfer encoding")
	ErrInvalidContentLength        = errors.New("invalid content length")
	ErrFailedToReadBody            = errors.New("failed to read body")
)

func (r *Request) parseRequestLine(rl string) error {
	rlc := strings.Split(rl, " ")
	if len(rlc) != 3 {
		return ErrInvalidRequestLine
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

func readLine(byteScanner *bufio.Scanner) (string, error) {
	line := ""
	prevByte := byte(0)
	first := true
	for byteScanner.Scan() {
		b := byteScanner.Bytes()
		if prevByte == '\r' && b[0] == '\n' {
			break
		}
		if first {
			first = false
			prevByte = b[0]
			continue
		}
		line += string(prevByte)
		prevByte = b[0]
	}

	if err := byteScanner.Err(); err != nil {
		return "", err
	}

	if len(line) == 0 {
		return "", nil
	}

	// remove the last byte if it' part of a '\r\n' sequence
	if line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	return line, nil
}

func parseRequest(c net.Conn) (*Request, error) {
	// create scanner to read line by line
	sc := bufio.NewScanner(c)
	sc.Split(bufio.ScanBytes)

	request := Request{}
	request.Headers = make(map[string]string)

	// parse request line
	rl, err := readLine(sc)
	if err != nil {
		return nil, ErrInvalidRequestLine
	}
	err = request.parseRequestLine(rl)
	if err != nil {
		return nil, err
	}

	// parse headers
	for {
		header, err := readLine(sc)
		if err != nil {
			return nil, err
		}

		err = request.parseHeader(header)
		if err != nil {
			if err == errEmptyHeaderFieldName {
				break
			}
			return nil, err
		}
	}

	// parse body
	if _, ok := request.Headers["transfer-encoding"]; ok {
		return nil, ErrUnsupportedTransferEncoding
	}

	if clv, ok := request.Headers["content-length"]; ok {
		// parse content length field value
		clv = strings.ReplaceAll(clv, " ", "")
		clvList := strings.Split(clv, ",")
		allSame := true
		prevCl := clvList[0]
		for _, clv := range clvList {
			if clv != prevCl {
				allSame = false
				break
			}
			prevCl = clv
		}

		if !allSame {
			return nil, ErrInvalidContentLength
		}

		contentLength, err := strconv.Atoi(clv)
		if err != nil {
			return nil, ErrInvalidContentLength
		}

		request.ContentLength = contentLength

		// read body
		body := make([]byte, contentLength)
		for i := 0; i < contentLength; i++ {
			if !sc.Scan() {
				return nil, ErrFailedToReadBody
			}
			body[i] = sc.Bytes()[0]
		}
		request.Body = body

	}

	return &request, nil
}

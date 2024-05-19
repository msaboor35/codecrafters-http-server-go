package main

import (
	"bufio"
	"errors"
	"fmt"
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

// parsing state machine
const (
	ParseRequestLine = iota + 1
	ParseHeaders
)

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
	methods                 = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
)

func parseRequest(c net.Conn) (*Request, error) {
	sc := bufio.NewScanner(c)
	sc.Split(bufio.ScanLines)

	state := ParseRequestLine
	request := Request{}
	request.Headers = make(map[string]string)
	fmt.Println("Parsing request")
scanLoop:
	for sc.Scan() {
		fmt.Printf("state in current loop: %d\n", state)

	stateSwitch:
		switch state {
		case ParseRequestLine:
			rl := sc.Text()

			rlc := strings.Split(rl, " ")
			if len(rlc) != 3 {
				return nil, ErrInvalidRequestLine
			}

			if !slices.Contains(methods, rlc[0]) {
				return nil, ErrInvalidMethod
			}

			if rlc[1][0] != '/' {
				return nil, ErrInvalidRequestTarget
			}

			if rlc[2] != "HTTP/1.1" {
				return nil, ErrInvalidHTTPVersion
			}

			request.Method = rlc[0]
			request.RequestTarget = rlc[1]
			request.HTTPVersion = rlc[2]

			state = ParseHeaders

			fmt.Printf("request after parsing request line: %+v\n", request)

			break stateSwitch
		case ParseHeaders:
			fmt.Printf("parsing headers\n")

			header := sc.Text()
			fmt.Printf("header#%d: %s\n", len(request.Headers)+1, header)
			if header == "" {
				// nothing else to parse in the headers
				break scanLoop
			}

			h := strings.SplitN(header, ":", 2)
			key := strings.ToLower(h[0])
			if key == HeaderKeyHost {
				if _, ok := request.Headers[HeaderKeyHost]; ok {
					return nil, ErrMultipleHostHeader
				}
			}
			request.Headers[key] = h[1]

			break stateSwitch
		default:
			panic("unknown state")
		}
		fmt.Printf("state after switch: %d\n", state)
		fmt.Printf("request after switch: %+v\n", request)
		fmt.Printf("scan error: %v\n", sc.Err())
		fmt.Println("===========================")
	}

	if _, ok := request.Headers[HeaderKeyHost]; !ok {
		return nil, ErrMissingHostHeader
	}

	return &request, nil
}

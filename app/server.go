package main

import (
	"errors"
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	c, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	req, err := parseRequest(c)
	if err != nil {
		if errors.Is(err, ErrInvalidRequestLine) {
			c.Write([]byte(ResposneBadRequest))
			return
		} else if errors.Is(err, ErrInvalidMethod) {
			c.Write([]byte(ResponseNotImplemented))
			return
		} else if errors.Is(err, ErrInvalidHTTPVersion) {
			c.Write([]byte(ResponseHTTPVersionNotSupported))
			return
		} else if errors.Is(err, ErrMissingHostHeader) {
			c.Write([]byte(ResposneBadRequest))
			return
		} else if errors.Is(err, ErrMultipleHostHeader) {
			c.Write([]byte(ResposneBadRequest))
			return
		} else if errors.Is(err, ErrInvalidRequestTarget) {
			c.Write([]byte(ResposneBadRequest))
			return
		}

		fmt.Println("Error parsing request: ", err.Error())
		os.Exit(1)
	}

	if req.Method == "GET" && req.RequestTarget == "/" {
		c.Write([]byte(ResponseOK))
		return
	} else {
		c.Write([]byte(ResponseNotFound))
		return
	}

}

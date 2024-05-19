package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	HTTPVersion = "HTTP/1.1"
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
	defer c.Close()

	req, err := parseRequest(c)
	if err != nil {
		if errors.Is(err, ErrInvalidRequestLine) {
			err = NewResponse(404).Write(c)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}
		} else if errors.Is(err, ErrInvalidMethod) {
			err = NewResponse(501).Write(c)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}
		} else if errors.Is(err, ErrInvalidHTTPVersion) {
			err = NewResponse(505).Write(c)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}
		} else if errors.Is(err, ErrMissingHostHeader) {
			err = NewResponse(404).Write(c)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}
		} else if errors.Is(err, ErrMultipleHostHeader) {
			err = NewResponse(404).Write(c)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}
		} else if errors.Is(err, ErrInvalidRequestTarget) {
			err = NewResponse(404).Write(c)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}
		}

		fmt.Println("Error parsing request: ", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Request: %+v\n", req)

	if req.Method == "GET" {
		if req.RequestTarget == "/" {
			err = NewResponse(200).Write(c)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}

			return
		}

		if strings.HasPrefix(req.RequestTarget, "/echo/") {
			str := strings.TrimPrefix(req.RequestTarget, "/echo/")
			err = NewResponse(200).
				SetHeader("Content-Type", "text/plain").
				SetHeader("Content-Length", fmt.Sprintf("%d", len(str))).
				SetBody([]byte(str)).
				Write(c)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}

			return
		}

		if req.RequestTarget == "/user-agent" {
			var userAgent string
			var ok bool
			if userAgent, ok = req.Headers["user-agent"]; !ok {
				userAgent = ""
			}

			err = NewResponse(200).
				SetHeader("Content-Type", "text/plain").
				SetHeader("Content-Length", fmt.Sprintf("%d", len(userAgent))).
				SetBody([]byte(userAgent)).
				Write(c)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}

			return
		}

		err = NewResponse(404).Write(c)
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			os.Exit(1)
		}
	} else {
		err = NewResponse(404).Write(c)
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			os.Exit(1)
		}
	}

}

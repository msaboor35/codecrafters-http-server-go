package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/samber/lo"
)

const (
	HTTPVersion = "HTTP/1.1"
)

func processRequest(c net.Conn) error {
	req, err := parseRequest(c)
	if err != nil {
		if errors.Is(err, ErrInvalidRequestLine) {
			return NewResponse(404).Send(c)
		} else if errors.Is(err, ErrInvalidHTTPVersion) {
			return NewResponse(505).Send(c)
		} else if errors.Is(err, ErrMultipleHostHeader) {
			return NewResponse(400).Send(c)
		} else if errors.Is(err, ErrInvalidRequestTarget) {
			return NewResponse(400).Send(c)
		} else if errors.Is(err, ErrUnsupportedTransferEncoding) {
			return NewResponse(501).Send(c)
		} else if errors.Is(err, ErrInvalidContentLength) {
			return NewResponse(400).Send(c)
		} else if errors.Is(err, ErrFailedToReadBody) {
			return NewResponse(500).Send(c)
		}

		fmt.Printf("Error parsing request: %v\n", err)
		return err
	}

	fmt.Printf("Request: %+v\n", req)

	if req.Method == "GET" {
		if req.RequestTarget == "/" {
			return NewResponse(200).Send(c)
		}

		if strings.HasPrefix(req.RequestTarget, "/echo/") {
			str := strings.TrimPrefix(req.RequestTarget, "/echo/")
			resp := NewResponse(200)

			if enc, ok := req.Headers["accept-encoding"]; ok {
				if enc == "gzip" {
					resp = resp.SetHeader("Content-Encoding", "gzip")
				}
			}

			return resp.SendString(c, str)
		}

		if req.RequestTarget == "/user-agent" {
			userAgent := lo.ValueOr(req.Headers, "user-agent", "")
			return NewResponse(200).SendString(c, userAgent)
		}

		if strings.HasPrefix(req.RequestTarget, "/files/") {
			filename := strings.TrimPrefix(req.RequestTarget, "/files/")
			fileData, err := os.ReadFile(fmt.Sprintf("%s/%s", servingDirectory, filename))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return NewResponse(404).Send(c)
				}

				return NewResponse(500).Send(c)
			}

			return NewResponse(200).SendBytes(c, fileData)
		}

		return NewResponse(404).Send(c)
	} else if req.Method == "POST" {
		if strings.HasPrefix(req.RequestTarget, "/files/") {
			filename := strings.TrimPrefix(req.RequestTarget, "/files/")
			fileData := req.Body

			err := os.WriteFile(fmt.Sprintf("%s/%s", servingDirectory, filename), fileData, 0644)
			if err != nil {
				fmt.Println("Error writing file: ", err)
				return NewResponse(500).Send(c)
			}

			return NewResponse(201).Send(c)
		}

		return NewResponse(404).Send(c)
	}

	return NewResponse(404).Send(c)
}

func handleConnection(c net.Conn) {
	defer c.Close()
	err := processRequest(c)
	if err != nil {
		fmt.Println("Error processing request: ", err.Error())
		return
	}
}

var servingDirectory string

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	flag.StringVar(&servingDirectory, "directory", "", "Directory to serve files from")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(c)
	}

}

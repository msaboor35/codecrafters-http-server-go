package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	HTTPVersion = "HTTP/1.1"
)

func processRequest(c net.Conn) error {
	req, err := parseRequest(c)
	if err != nil {
		if errors.Is(err, ErrInvalidRequestLine) {
			err := NewResponse(404).Write(c)
			if err != nil {
				fmt.Printf("Error writing response: %v\n", err)
				return err
			}
		} else if errors.Is(err, ErrInvalidHTTPVersion) {
			err := NewResponse(505).Write(c)
			if err != nil {
				fmt.Printf("Error writing response: %v\n", err)
				return err
			}
		} else if errors.Is(err, ErrMultipleHostHeader) {
			err := NewResponse(404).Write(c)
			if err != nil {
				fmt.Printf("Error writing response: %v\n", err)
				return err
			}
		} else if errors.Is(err, ErrInvalidRequestTarget) {
			err := NewResponse(404).Write(c)
			if err != nil {
				fmt.Printf("Error writing response: %v\n", err)
				return err
			}
		}

		fmt.Printf("Error parsing request: %v\n", err)
		return err
	}

	fmt.Printf("Request: %+v\n", req)

	if req.Method == "GET" {
		if req.RequestTarget == "/" {
			err = NewResponse(200).Write(c)
			if err != nil {
				fmt.Printf("Error writing response: %v\n", err)
				return err
			}

			return nil
		}

		if strings.HasPrefix(req.RequestTarget, "/echo/") {
			str := strings.TrimPrefix(req.RequestTarget, "/echo/")
			err = NewResponse(200).
				SetHeader("Content-Type", "text/plain").
				SetHeader("Content-Length", fmt.Sprintf("%d", len(str))).
				SetBody([]byte(str)).
				Write(c)
			if err != nil {
				fmt.Printf("Error writing response: %v\n", err)
				return err
			}

			return nil
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
				fmt.Printf("Error writing response: %v\n", err)
				return err
			}

			return nil
		}

		if strings.HasPrefix(req.RequestTarget, "/files/") {
			filename := strings.TrimPrefix(req.RequestTarget, "/files/")
			fileData, err := os.ReadFile(fmt.Sprintf("%s/%s", servingDirectory, filename))
			if err != nil {
				err = NewResponse(500).Write(c)
				if err != nil {
					fmt.Printf("Error writing response: %v\n", err)
					return err
				}
			}

			err = NewResponse(200).
				SetHeader("Content-Type", "application/octet-stream").
				SetHeader("Content-Length", fmt.Sprintf("%d", len(fileData))).
				SetBody(fileData).
				Write(c)
			if err != nil {
				fmt.Printf("Error writing response: %v\n", err)
				return err
			}

			return nil
		}

		err = NewResponse(404).Write(c)
		if err != nil {
			fmt.Printf("Error writing response: %v\n", err)
			return err
		}
	} else {
		err = NewResponse(404).Write(c)
		if err != nil {
			fmt.Printf("Error writing response: %v\n", err)
			return err
		}
	}

	return nil
}

func handleConnection(c net.Conn) {
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
	if servingDirectory == "" {
		fmt.Println("Directory is required")
		os.Exit(1)
	}

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

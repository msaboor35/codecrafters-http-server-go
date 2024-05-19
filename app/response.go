package main

import (
	"bufio"
	"fmt"
	"net"
)

var statusText = map[int]string{
	200: "OK",
	400: "Bad Request",
	404: "Not Found",
	501: "Not Implemented",
	505: "HTTP Version Not Supported",
}

type Response struct {
	StatusCode int
	Status     string
	Headers    map[string]string
	Body       []byte
}

func NewResponse(statusCode int) *Response {
	return &Response{
		StatusCode: statusCode,
		Status:     statusText[statusCode],
		Headers:    make(map[string]string),
		Body:       []byte{},
	}
}

func (resp *Response) SetHeader(key, value string) *Response {
	resp.Headers[key] = value
	return resp
}

func (resp *Response) SetBody(data []byte) *Response {
	resp.Body = data
	return resp
}

func (resp *Response) Write(c net.Conn) error {
	wr := bufio.NewWriter(c)
	defer func() error {
		err := wr.Flush()
		if err != nil {
			return err
		}
		return nil
	}()

	_, err := wr.Write(resp.generateStatusLine())
	if err != nil {
		return err
	}

	_, err = wr.Write(resp.generateHeaders())
	if err != nil {
		return err
	}

	if len(resp.Body) == 0 {
		return nil
	}

	_, err = wr.Write(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (resp *Response) generateStatusLine() []byte {
	return []byte(fmt.Sprintf("%s %d %s\r\n", HTTPVersion, resp.StatusCode, resp.Status))
}

func (resp *Response) generateHeaders() []byte {
	if resp.Headers == nil || len(resp.Headers) == 0 {
		return []byte("\r\n\r\n")
	}

	headers := ""
	for k, v := range resp.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	headers += "\r\n"
	return []byte(headers)
}

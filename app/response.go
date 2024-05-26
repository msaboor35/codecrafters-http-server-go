package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type Response struct {
	StatusCode int
	Status     string
	Headers    map[string]string
	Body       []byte
}

func NewResponse(statusCode int) *Response {
	return &Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
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

func (resp *Response) write(c net.Conn) error {
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

func (resp *Response) Send(c net.Conn) error {
	err := resp.write(c)
	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
		return err
	}

	return nil
}

func (resp *Response) SendString(c net.Conn, data string) error {
	return resp.SetHeader("Content-Type", "text/plain").
		SetHeader("Content-Length", fmt.Sprintf("%d", len(data))).
		SetBody([]byte(data)).
		Send(c)
}

func (resp *Response) SendBytes(c net.Conn, data []byte) error {
	return resp.SetHeader("Content-Type", "application/octet-stream").
		SetHeader("Content-Length", fmt.Sprintf("%d", len(data))).
		SetBody(data).
		Send(c)
}

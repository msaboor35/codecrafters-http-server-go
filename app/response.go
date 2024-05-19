package main

const (
	ResponseOK                      = "HTTP/1.1 200 OK\r\n\r\n"
	ResposneBadRequest              = "HTTP/1.1 400 Bad Request\r\n\r\n"
	ResponseNotFound                = "HTTP/1.1 404 Not Found\r\n\r\n"
	ResponseNotImplemented          = "HTTP/1.1 501 Not Implemented\r\n\r\n"
	ResponseHTTPVersionNotSupported = "HTTP/1.1 505 HTTP Version Not Supported\r\n\r\n"
)

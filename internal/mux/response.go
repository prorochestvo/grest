package mux

import (
	"bytes"
	"net/http"
	"strings"
	"sync"
)

func newResponse(w http.ResponseWriter) *Response {
	result := Response{}
	result.writer = w
	result.code = http.StatusNotImplemented
	result.head = make(http.Header, 0)
	result.body = bytes.NewBufferString("")
	return &result
}

type Response struct {
	lock   *sync.RWMutex
	code   int
	head   http.Header
	body   *bytes.Buffer
	writer http.ResponseWriter
}

func (this *Response) Header() http.Header {
	return this.head
}

func (this *Response) Write(b []byte) (int, error) {
	return this.body.Write(b)
}

func (this *Response) WriteHeader(statusCode int) {
	this.code = statusCode
}

func (this *Response) Send() (int, error) {
	for key, val := range this.head {
		this.writer.Header().Set(key, strings.Join(val, ", "))
	}
	this.writer.WriteHeader(this.code)
	return this.writer.Write(this.body.Bytes())
}

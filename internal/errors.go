package internal

import (
	"fmt"
	"net/http"
)

func NewError(status Status, format string, a ...interface{}) Error {
	return &httpError{status: status, description: fmt.Sprintf(format, a...)}
}

type Error interface {
	Status() int
	Text() string
	Error() string
}

type httpError struct {
	status      Status
	description string
}

func (this *httpError) Status() int {
	return int(this.status)
}

func (this *httpError) Text() string {
	return http.StatusText(int(this.status))
}

func (this *httpError) Error() string {
	return this.description
}

package mux

import (
	"net/http"
	"net/url"
)

func newRequest(r *http.Request) *Request {
	result := Request{Request: r}
	result.URL = urlEx{URL: r.URL}
	result.URL.ID.Name = ""
	result.URL.ID.Value = nil
	result.URL.parameters = make(map[string]string)
	result.URL.query = r.URL.Query()
	return &result
}

type Request struct {
	URL urlEx
	*http.Request
}

type urlEx struct {
	ID struct {
		Value []byte
		Name  string
	}
	parameters map[string]string
	query      url.Values
	*url.URL
}

func (this *urlEx) Query() url.Values {
	return this.query
}

func (this *urlEx) SetQuery(value url.Values) {
	this.query = value
}

func (this *urlEx) Param(name string) string {
	if p, ok := this.parameters[name]; ok {
		return p
	}
	return ""
}

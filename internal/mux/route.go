package mux

import (
	"github.com/prorochestvo/grest/internal/helper"
	"net/http"
	"regexp"
)

// Route
type route struct {
	handler func(w http.ResponseWriter, r *Request)
	methods methods
	path    *regexp.Regexp
	varID   string
}

func (this *route) match(path string) bool {
	return this.path.MatchString(path)
}

func (this *route) Methods() []string {
	return this.methods
}

// Methods
type methods []string

func (this methods) match(method string) bool {
	return helper.StringsIndexOf(this, method) >= 0
}

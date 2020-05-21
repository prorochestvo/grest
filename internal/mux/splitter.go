package mux

import (
	"bytes"
	"fmt"
	"grest/internal/helper"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type Splitter interface {
	NewRoute(methods []string, url string, handler func(w http.ResponseWriter, r *Request)) error
	http.Handler
}

func NewSplitter() Splitter {
	result := splitter{}
	result.routes = make([]*route, 0)
	return &result
}

type splitter struct {
	routes []*route
}

func (this *splitter) NewRoute(methods []string, url string, handler func(w http.ResponseWriter, r *Request)) error {
	url = helper.HttpPathTrim(url)
	route := &route{path: nil}
	route.handler = handler
	route.methods = make([]string, 0)
	// set methods
	for _, method := range methods {
		route.methods = append(route.methods, strings.ToUpper(method))
	}
	// set path
	pattern := bytes.NewBufferString("")
	pattern.WriteByte('^')
	pattern.WriteByte('/')
	idxs, err := braceIndices(url)
	if err != nil {
		return err
	}
	var end int
	last := strings.LastIndex(url, "/")
	for i := 0; i < len(idxs); i += 2 {
		raw := url[end:idxs[i]]
		end = idxs[i+1]
		parts := strings.SplitN(url[idxs[i]+1:end-1], ":", 2)
		n := regexp.MustCompile(`[^A-Za-z0-9_-]`).ReplaceAllString(parts[0], "")
		p := "[^/]+"
		if len(parts) == 2 {
			p = parts[1]
		}
		if n == "" || p == "" {
			return fmt.Errorf("missing name or pattern in %q", url[idxs[i]:end])
		}
		_, _ = fmt.Fprintf(pattern, "%s(?P<%s>%s)", regexp.QuoteMeta(raw), n, p)
		// if last var when is id
		if last < idxs[i]+1 {
			route.varID = n
		}
	}
	raw := url[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))
	pattern.WriteString("[/]{0,1}$")
	if route.path, err = regexp.Compile(pattern.String()); err != nil {
		return err
	}
	if route.path.NumSubexp() != (len(idxs) >> 1) {
		return fmt.Errorf("route '%s' contains capture groups in its regexp. Only non-capturing groups are accepted: e.g. (?:pattern) instead of (pattern)", url)
	}
	this.routes = append(this.routes, route)
	return nil
}

func (this *splitter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path //path = r.URL.EscapedPath()
	if p := cleanPath(path); p != path {
		w.Header().Set("Location", p)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	for _, route := range this.routes {
		if route == nil || !route.path.MatchString(path) || !route.methods.match(r.Method) {
			continue
		}
		req := newRequest(r)
		res := newResponse(w)
		m := route.path.FindStringSubmatch(path)
		if n := route.path.SubexpNames(); len(n) > 0 { //  m != nil && route.args != nil && len(route.args) == len(m)
			for i, name := range n {
				if name != "" {
					req.URL.parameters[name] = m[i]
				}
			}
			if len(route.varID) > 0 {
				req.URL.ID.Name = route.varID
				if id, ok := req.URL.parameters[route.varID]; ok {
					req.URL.ID.Value = []byte(id)
					delete(req.URL.parameters, route.varID)
				}
			}
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			route.handler(res, req)
			wg.Done()
		}()
		wg.Wait()
		_, _ = res.Send()
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

package grest

import (
	"fmt"
	"github.com/prorochestvo/grest/db"
	"github.com/prorochestvo/grest/internal/mux"
	"net/http"
	"strings"
)

func newRoute(router *Router, driver db.Driver, controller Controller, action Action) *route {
	result := route{}
	result.DB = driver
	result.Controller = controller
	result.Model = getControllerModel(controller)
	result.router = router
	result.action = action
	return &result
}

type route struct {
	DB         db.Driver
	Controller Controller
	Model      Model
	router     *Router
	action     Action
}

func (this *route) Path() string {
	return makeControllerActionPath(this.Controller, this.action)
}

func (this *route) Methods() []string {
	result := this.action.Methods()
	if result == nil {
		result = make([]string, 0)
	}
	return result
}

func (this *route) run(w http.ResponseWriter, r *mux.Request) {
	defer func() {
		_ = r.Body.Close()
	}()
	var err error
	var code = http.StatusInternalServerError
	var head map[string]string = nil
	var body interface{} = nil
	req := newRequest(r, this)
	if req.User, err = this.router.AccessControl.User(req); err != nil {
		code, head, body = this.router.AccessControl.Error(req, http.StatusUnauthorized, head, err)
	} else if roles := this.action.Roles(); roles != nil && len(roles) > 0 && roles.IndexOf(req.User.Role()) < 0 {
		code, head, body = this.router.AccessControl.Error(req, http.StatusForbidden, head, fmt.Errorf("don't have permission"))
	} else if code, head, body, err = this.action.Run(req); err != nil {
		code, head, body = this.router.AccessControl.Error(req, code, head, err)
	}
	_, _ = this.send(w, code, head, body)
}

func (this *route) cors(w http.ResponseWriter, r *mux.Request) {
	defer func() {
		_ = r.Body.Close()
	}()
	var err error
	var code = http.StatusInternalServerError
	var head map[string]string = nil
	var body interface{} = nil
	req := newRequest(r, this)
	if req.User, err = this.router.AccessControl.User(req); err != nil {
		code, head, body = this.router.AccessControl.Error(req, http.StatusUnauthorized, head, err)
	} else if code, head, body, err = this.router.AccessControl.Origin(req); err != nil {
		code, head, body = this.router.AccessControl.Error(req, http.StatusInternalServerError, head, err)
	}
	_, _ = this.send(w, code, head, body)
}

func (this *route) send(w http.ResponseWriter, code int, head map[string]string, body interface{}) (int, error) {
	var data []byte = nil
	if head == nil {
		head = make(map[string]string, 0)
	}
	if _, ok := head["Content-Type"]; !ok {
		if b, err := this.router.ContentType.Marshal(body); err != nil {
			code = http.StatusInternalServerError
		} else if b != nil {
			head["Content-Type"] = this.router.ContentType.Format
			data = b
		}
	} else if d, ok := body.([]string); ok && d != nil && len(d) > 0 {
		data = []byte(strings.Join(d, "\n"))
	} else if d, ok := body.(string); ok && len(d) > 0 {
		data = []byte(d)
	} else if d, ok := body.([]byte); ok && d != nil && len(d) > 0 {
		data = d
	}
	if data == nil {
		data = make([]byte, 0)
	}
	// write to client
	for key, val := range head {
		w.Header().Set(key, val)
	}
	// reset version
	if len(this.router.Version) > 0 {
		w.Header().Set("FREST-Version", this.router.Version)
	}
	w.WriteHeader(code)
	return w.Write(data)
}

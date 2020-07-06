package grest

import (
	"fmt"
	"github.com/prorochestvo/grest/db"
	"github.com/prorochestvo/grest/internal"
	"github.com/prorochestvo/grest/internal/helper"
	"github.com/prorochestvo/grest/internal/mux"
	"github.com/prorochestvo/grest/usr"
	"io"
	"net/http"
	"os"
)

func newRouter(db db.Driver) *Router {
	result := &Router{controllers: make([]Controller, 0), Stdout: os.Stdout, Stderr: os.Stderr}
	result.Version = ""
	result.Migration = NewMigration(db, result)
	result.AccessControl.User = func(_ *Request) (usr.User, error) {
		return usr.DefaultUser, nil
	}
	result.AccessControl.Origin = func(_ *Request) (int, map[string]string, interface{}, error) {
		head := map[string]string{
			"Connection":                    "Keep-Alive",
			"Cache-Control":                 "no-cache",
			"Pragma":                        "no-cache",
			"Access-Control-Allow-Origin":   "*",
			"Access-Control-Expose-Headers": "FREST-Version, *",
			"X-Accel-Buffering":             "no",
			"Transfer-Encoding":             "identity",
			"Content-Type":                  "application/json",
		}
		return 204, head, make([]byte, 0), nil
	}
	result.AccessControl.Error = func(_ *Request, code int, head map[string]string, body error) (int, map[string]string, interface{}) {
		b := struct {
			Error string `json:"error"`
		}{
			Error: body.Error(),
		}
		return code, head, b
	}
	return result
}

func NewJSONRouter(db db.Driver) *Router {
	result := newRouter(db)
	result.ContentType = internal.JSON
	return result
}

func NewXMLRouter(db db.Driver) *Router {
	result := newRouter(db)
	result.ContentType = internal.XML
	return result
}

type Router struct {
	Version       string
	Migration     *migration
	controllers   []Controller
	ContentType   internal.MimeType
	AccessControl accessControl
	Stderr        io.Writer
	Stdout        io.Writer
	http.Handler
}

func (this *Router) Listen(value ...Controller) error {
	items := this.controllers
	for _, v := range value {
		for _, item := range items {
			if item == nil || v == nil || item.Path() != v.Path() {
				continue
			}
			return fmt.Errorf("duplicate path controller (%s)", item.Path())
		}
		// change controllers
		items = append(items, v)
	}
	this.controllers = items
	this.Handler = this.init()
	return nil
}

func (this *Router) Shutdown(value ...Controller) error {
	for _, c := range value {
		removed := false
		for i, item := range this.controllers {
			if item == nil || c == nil || item.Path() != c.Path() {
				continue
			}
			this.controllers = append(this.controllers[:i], this.controllers[i+1:]...)
			if len(this.controllers) == 0 {
				this.controllers = this.controllers[:0]
			}
			removed = true
		}
		if !removed && c != nil {
			return fmt.Errorf("not found path controller (%s)", c.Path())
		}
	}
	this.Handler = this.init()
	return nil
}

func (this *Router) init() mux.Splitter {
	result := mux.NewSplitter()
	for _, controller := range this.controllers {
		if controller == nil {
			continue
		}
		// merge action
		for _, action := range controller.Actions() {
			if action == nil {
				continue
			}
			route := newRoute(this, this.Migration.driver, controller, action)
			path := route.Path()
			methods := route.Methods()
			_ = result.NewRoute(methods, path, route.run)
			if pos := helper.StringsIndexOf(methods, http.MethodOptions); pos < 0 {
				_ = result.NewRoute([]string{http.MethodOptions}, path, route.cors)
			}
		}
		// custom controller
		if c, ok := controller.(ControllerWithRouteCustom); ok == true && c != nil {
			c.CustomRoutes(result)
		}
		// migration.Connection
	}
	return result
}

// internal types
type accessControl struct {
	User   func(r *Request) (usr.User, error)
	Origin func(r *Request) (int, map[string]string, interface{}, error)
	Error  func(r *Request, status int, head map[string]string, body error) (int, map[string]string, interface{})
}

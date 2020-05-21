package grest

import (
	"encoding/json"
	"fmt"
	"grest/db"
	"grest/internal/helper"
	"grest/usr"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

/***********************************************************************************************************************
 * ModuleControllersShare
 * Раскрвает все контроллеры
 */
type ModuleControllersShare struct {
	path string
	Role []usr.Role
	CSS  []string
}

func (this *ModuleControllersShare) WithID() bool {
	return true
}

func (this *ModuleControllersShare) Methods() []string {
	return []string{
		http.MethodGet,
		http.MethodHead,
	}
}

func (this *ModuleControllersShare) Path() string {
	if len(this.path) > 0 {
		return this.path
	}
	return "controller"
}

func (this *ModuleControllersShare) SetPath(value string) {
	this.path = value
}

func (this *ModuleControllersShare) Id() (name, pattern string) {
	return "route", "[A-Za-z0-9-]*"
}

func (this *ModuleControllersShare) Roles() usr.Roles {
	if this.Role == nil {
		return make([]usr.Role, 0)
	}
	return this.Role
}

func (this *ModuleControllersShare) Run(r *Request) (int, map[string]string, interface{}, error) {
	// controllers
	controller, controllers := this.getHtmlControllers(r)
	// version
	version := r.router.Version
	if len(version) > 0 {
		version = "v " + version
	}
	//
	head := map[string]string{
		"Content-Type": "text/html; charset=utf-8",
	}
	html := fmt.Sprintf(`
    <h3 style="margin-bottom: 5px">
      API Controllers %s
    </h2>
    <div style="text-transform: capitalize">
      <ul class="hor small">
        <li><a href=".">LIST</a></li>
        <li>\</li>
        <li class="font-bold">%s</li>
      </ul>
    </div>
    <hr class="m-b-md m-t-xs">
    %s`, version, this.getControllerName(controller), controllers)
	return http.StatusOK, head, helper.HtmlDocument(html, this.CSS...), nil
}

func (this *ModuleControllersShare) getHtmlControllers(r *Request) (ctrl Controller, html string) {
	html = ""
	ctrl = nil
	// tmp
	result := make([]string, 0)
	rx := regexp.MustCompile(`[^A-Za-z0-9-]+`)
	tmp := make(map[string]Controller, 0)
	for _, controller := range r.router.controllers {
		if controller == nil {
			continue
		}
		path := controller.Path()
		path = rx.ReplaceAllString(path, "-")
		path = strings.ToLower(path)
		tmp[path] = controller
	}
	// controller
	if controller, ok := tmp[string(r.URL.ID.Value)]; ok && controller != nil {
		// name
		name := this.getControllerName(controller)
		// actions
		actions := make([]string, 0)
		for _, action := range controller.Actions() {
			m := action.Methods()
			p := makeControllerActionPath(controller, action)
			if id, name, pattern := helper.HttpPathID(p); len(id) > 0 && len(name) > 0 && len(pattern) > 0 {
				p = strings.ReplaceAll(p, id, "")
				p += fmt.Sprintf("{%s}", name)
				p += fmt.Sprintf("<i class=\"m-l-xs small\">%s</i>", pattern)
			}
			actions = append(actions, fmt.Sprintf("<li>%s %s</li>", strings.ToUpper(strings.Join(m, " ")), p))
		}
		// save
		result = append(result, fmt.Sprintf(`
       <div class="m-b-sm m-t-sm">
         <div><span class="font-bold" style="text-transform: capitalize">%s</span></div>
         <div class="font-normal text-monospaced"><ul style="padding-left: 10px">%s</ul></div>
       </div>`, name, strings.Join(actions, "\n")))
		ctrl = controller
	} else
	// controllers
	if len(tmp) > 0 {
		for hash, controller := range tmp {
			if controller == nil {
				continue
			}
			// name
			name := this.getControllerName(controller)
			// actions
			actions := make([]string, 0)
			for _, action := range controller.Actions() {
				p := makeControllerActionPath(controller, action)
				if id, name, pattern := helper.HttpPathID(p); len(id) > 0 && len(name) > 0 && len(pattern) > 0 {
					p = strings.ReplaceAll(p, id, "")
					p += fmt.Sprintf("{%s}", name)
					p += fmt.Sprintf("<i class=\"m-l-xs small\">%s</i>", pattern)
				}
				actions = append(actions, fmt.Sprintf("<li>%s</li>", p))
			}
			// save
			result = append(result, fmt.Sprintf(`
       <div class="m-b-sm m-t-sm">
         <div><a class="link font-bold" href="%s" style="text-transform: capitalize">%s</a></div>
         <div class="font-normal"><ul style="padding: 0 0 0 10px; margin: 0; list-style-type: none;">%s</ul></div>
       </div>`, hash, name, strings.Join(actions, "\n")))
		}
		sort.Strings(result)
	}
	html = strings.Join(result, "\n")
	return
}

func (this *ModuleControllersShare) getControllerName(controller Controller) string {
	if controller == nil {
		return ""
	}
	result := strings.ReplaceAll(reflect.TypeOf(controller).String(), "*", "")
	if pos := strings.Index(result, "."); pos > 0 {
		p := strings.ToUpper(result[0:pos])
		c := result[pos:]
		result = p + c
	}
	return result
}

/***********************************************************************************************************************
 * ModuleSqlEditor
 * Разбирает HTTP запрос и преобразует его в SQL запрос
 */
type ModuleSqlEditor struct {
	path string
	Role []usr.Role
}

func (this *ModuleSqlEditor) WithID() bool {
	return true
}

func (this *ModuleSqlEditor) Methods() []string {
	return []string{
		http.MethodGet,
		http.MethodHead,
	}
}

func (this *ModuleSqlEditor) Path() string {
	if len(this.path) > 0 {
		return this.path
	}
	return "sql/editor"
}

func (this *ModuleSqlEditor) SetPath(value string) {
	this.path = value
}

func (this *ModuleSqlEditor) Id() (name, pattern string) {
	return "table", "[A-Za-z0-9_-]+"
}

func (this *ModuleSqlEditor) Roles() usr.Roles {
	if this.Role == nil {
		return make([]usr.Role, 0)
	}
	return this.Role
}

func (this *ModuleSqlEditor) Run(r *Request) (int, map[string]string, interface{}, error) {
	where, groupBy, orderBy, having, limit, offset := db.SQLParser(r.Request.URL.Query())
	body := db.NewSQLLinker().Select(db.NewSQLTable(string(r.URL.ID.Value)), []db.SQLField{db.NewSQLField("*", nil)}, where, groupBy, having, orderBy, limit, offset)
	head := map[string]string{
		"Content-Type": "text/plain; charset=utf-8",
	}
	return http.StatusOK, head, body, nil
}

/***********************************************************************************************************************
 * ModuleHttpParser
 * Разбирает HTTP запрос и возвращает результат
 */
type ModuleHttpParser struct {
	path string
	Role []usr.Role
}

func (this *ModuleHttpParser) WithID() bool {
	return false
}

func (this *ModuleHttpParser) Methods() []string {
	return []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodTrace,
	}
}

func (this *ModuleHttpParser) Path() string {
	if len(this.path) > 0 {
		return this.path
	}
	return "request/parser"
}

func (this *ModuleHttpParser) SetPath(value string) {
	this.path = value
}

func (this *ModuleHttpParser) Id() (name, pattern string) {
	return "id", "[0-9]+"
}

func (this *ModuleHttpParser) Roles() usr.Roles {
	if this.Role == nil {
		return make([]usr.Role, 0)
	}
	return this.Role
}

func (this *ModuleHttpParser) Run(r *Request) (int, map[string]string, interface{}, error) {
	// response
	type inside struct {
		Header map[string]string      `json:"headers"`
		Url    map[string]interface{} `json:"url"`
		Body   interface{}            `json:"body"`
	}
	response := inside{
		Header: make(map[string]string, 0),
		Url:    map[string]interface{}{"id": r.URL.ID, "path": r.URL.Path, "query": r.URL.Query()},
		Body:   nil,
	}
	if b, err := ioutil.ReadAll(r.Body); r.Body != nil && err != nil {
		return http.StatusInternalServerError, nil, nil, err
	} else if b != nil {
		response.Body = string(b)
	}
	for key, val := range r.Header {
		response.Header[key] = strings.Join(val, ", ")
	}
	// result
	body := make([]byte, 0)
	if b, err := json.MarshalIndent(response, "", "  "); err != nil {
		return http.StatusInternalServerError, nil, nil, err
	} else {
		body = b
	}
	head := map[string]string{
		"Content-Type": "application/json",
	}
	return http.StatusOK, head, body, nil
}

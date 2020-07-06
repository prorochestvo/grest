package grest

import (
	"fmt"
	"github.com/prorochestvo/grest/db"
	"github.com/prorochestvo/grest/internal/helper"
	"github.com/prorochestvo/grest/usr"
	"math"
	"net/http"
	"regexp"
	"strconv"
)

type Action interface {
	Path() string
	Methods() []string
	WithID() bool
	Roles() usr.Roles
	Run(*Request) (int, map[string]string, interface{}, error)
}

type ActionWithID interface {
	Id() (name, pattern string)
	Action
}

func NewActionPagination(roles ...usr.Role) Action {
	return NewAction(MethodGet|MethodHead, "", actionPagination, roles...)
}

func NewActionList(roles ...usr.Role) Action {
	return NewAction(MethodGet|MethodHead, "", actionList, roles...)
}

func NewActionView(roles ...usr.Role) Action {
	return NewAction(MethodGet|MethodHead|WithID, "", actionView, roles...)
}

func NewActionCreate(roles ...usr.Role) Action {
	return NewAction(MethodPost, "", actionCreate, roles...)
}

func NewActionUpdate(roles ...usr.Role) Action {
	return NewAction(MethodPut|MethodPatch|WithID, "", actionUpdate, roles...)
}

func NewActionDelete(roles ...usr.Role) Action {
	return NewAction(MethodDelete|WithID, "", actionDelete, roles...)
}

func NewAction(options uint32, path string, handler func(*Request) (int, map[string]string, interface{}, error), role ...usr.Role) Action {
	result := action{}
	result.path = path
	result.options = options
	result.handler = handler
	result.roles = role
	return &result
}

type action struct {
	path    string
	options uint32
	roles   []usr.Role
	handler func(*Request) (int, map[string]string, interface{}, error)
}

func (this *action) Path() string {
	return this.path
}

func (this *action) Methods() []string {
	result := make([]string, 0)
	if this.options&MethodGet == MethodGet {
		result = append(result, http.MethodGet)
	}
	if this.options&MethodHead == MethodHead {
		result = append(result, http.MethodHead)
	}
	if this.options&MethodPost == MethodPost {
		result = append(result, http.MethodPost)
	}
	if this.options&MethodPut == MethodPut {
		result = append(result, http.MethodPut)
	}
	if this.options&MethodPatch == MethodPatch {
		result = append(result, http.MethodPatch)
	}
	if this.options&MethodDelete == MethodDelete {
		result = append(result, http.MethodDelete)
	}
	if this.options&MethodConnect == MethodConnect {
		result = append(result, http.MethodConnect)
	}
	if this.options&MethodOptions == MethodOptions {
		result = append(result, http.MethodOptions)
	}
	if this.options&MethodTrace == MethodTrace {
		result = append(result, http.MethodTrace)
	}
	return result
}

func (this *action) WithID() bool {
	return this.options&WithID == WithID
}

func (this *action) Roles() usr.Roles {
	if this.roles == nil {
		return make([]usr.Role, 0)
	}
	return this.roles
}

func (this *action) Run(r *Request) (int, map[string]string, interface{}, error) {
	if this.handler == nil {
		return http.StatusInternalServerError, make(map[string]string, 0), nil, fmt.Errorf("not found handler function")
	}
	return this.handler(r)
}

func actionPagination(r *Request) (int, map[string]string, interface{}, error) {
	if r.Model == nil {
		return http.StatusInternalServerError, nil, nil, fmt.Errorf("missing model in %s", helper.TypeName(r.Controller))
	}
	if f := getModelFields(r.Model, r.User.Role(), usr.ALEVEL_READ); f != nil && len(f) > 0 {
		// general options
		parsers := make(map[string]func(value string) (interface{}, error), 0)
		fields := make([]db.SQLField, 0)
		table := db.NewSQLTable(r.Model.Table())
		for _, field := range f {
			parsers[field.Name()] = field.Parser
			fields = append(fields, db.NewSQLField(field.Name(), nil))
		}
		where, groupBy, orderBy, having, _, _ := db.SQLParserEx(r.Request.Request, parsers)
		if r.URL.ID.Value != nil {
			where = append(where, db.NewSQLWhere(r.URL.ID.Name, string(r.URL.ID.Value)))
		}
		// page options
		var pageNumber int64 = 0
		var pageSize int64 = 10
		var totalRows int64 = -1
		if v, err := strconv.ParseInt(r.Header.Get("PaginationPerPage"), 10, 64); err == nil && v >= 0 {
			pageSize = v
		} else if v, err := strconv.ParseInt(r.URL.Query().Get("page[size]"), 10, 64); err == nil && v >= 0 {
			pageSize = v
		}
		if pageSize <= 0 {
			pageSize = 1
		}
		if v, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64); err == nil && v > 0 {
			pageNumber = v - 1
		}
		// get rows count
		if body, err := r.DB.Select(table, []db.SQLField{db.NewSQLField(`COUNT(*) as "cnt"`, nil)}, where, groupBy, having, nil, nil, nil); err != nil {
			return http.StatusInternalServerError, nil, nil, err
		} else if body == nil || len(body) != 1 {
			return http.StatusInternalServerError, nil, nil, fmt.Errorf("empty dataset")
		} else if d, ok := body[0]["cnt"]; !ok {
			return http.StatusInternalServerError, nil, nil, fmt.Errorf("wrong dataset")
		} else if rows, ok := d.(int64); !ok || rows < 0 {
			return http.StatusInternalServerError, nil, nil, fmt.Errorf("wrong dataset")
		} else {
			totalRows = rows
		}
		// get page rows
		if body, err := r.DB.Select(table, fields, where, groupBy, having, orderBy, db.NewSQLLimit(pageSize), db.NewSQLOffset(pageNumber*pageSize)); err != nil {
			return http.StatusInternalServerError, nil, nil, err
		} else if body != nil {
			d := struct {
				Data interface{}            `json:"data"`
				Meta map[string]interface{} `json:"meta"`
			}{
				Data: r.expand(r.Model, body),
				Meta: map[string]interface{}{
					"total_entries": totalRows,
					"current_page":  pageNumber + 1,
					"total_pages":   int64(math.Ceil(float64(totalRows) / float64(pageSize))),
					"per_page":      pageSize,
				},
			}
			return http.StatusOK, nil, d, nil
		}
	}
	return http.StatusInternalServerError, nil, nil, fmt.Errorf("unknown error")
}

func actionList(r *Request) (int, map[string]string, interface{}, error) {
	if r.Model == nil {
		return http.StatusInternalServerError, nil, nil, fmt.Errorf("missing model in %s", helper.TypeName(r.Controller))
	}
	if f := getModelFields(r.Model, r.User.Role(), usr.ALEVEL_READ); f != nil && len(f) > 0 {
		parsers := make(map[string]func(value string) (interface{}, error), 0)
		fields := make([]db.SQLField, 0)
		table := db.NewSQLTable(r.Model.Table())
		for _, field := range f {
			parsers[field.Name()] = field.Parser
			fields = append(fields, db.NewSQLField(field.Name(), nil))
		}
		where, groupBy, orderBy, having, limit, offset := db.SQLParserEx(r.Request.Request, parsers)
		if r.URL.ID.Value != nil {
			where = append(where, db.NewSQLWhere(r.URL.ID.Name, string(r.URL.ID.Value)))
		}
		if body, err := r.DB.Select(table, fields, where, groupBy, having, orderBy, limit, offset); err != nil {
			return http.StatusInternalServerError, nil, nil, err
		} else if body != nil {
			return http.StatusOK, nil, r.expand(r.Model, body), nil
		}
	}
	return http.StatusInternalServerError, nil, nil, fmt.Errorf("unknown error")
}

func actionView(r *Request) (int, map[string]string, interface{}, error) {
	if r.Model == nil {
		return http.StatusInternalServerError, nil, nil, fmt.Errorf("missing model in %s", helper.TypeName(r.Controller))
	} else if r.URL.ID.Value == nil {
		return http.StatusBadRequest, nil, nil, fmt.Errorf("missing identifier")
	}
	if f := getModelFields(r.Model, r.User.Role(), usr.ALEVEL_READ); f != nil && len(f) > 0 {
		fields := make([]db.SQLField, 0)
		table := db.NewSQLTable(r.Model.Table())
		where := []db.SQLWhere{db.NewSQLWhere(r.URL.ID.Name, string(r.URL.ID.Value))}
		limit := db.NewSQLLimit(1)
		for _, field := range f {
			fields = append(fields, db.NewSQLField(field.Name(), nil))
		}
		if body, err := r.DB.Select(table, fields, where, nil, nil, nil, limit, nil); err != nil {
			return http.StatusInternalServerError, nil, nil, err
		} else if body != nil && len(body) == 1 {
			return http.StatusOK, nil, r.expand(r.Model, body[0]), err
		}
	}
	return http.StatusInternalServerError, nil, nil, fmt.Errorf("unknown error")
}

func actionCreate(r *Request) (int, map[string]string, interface{}, error) {
	if r.Model == nil {
		return http.StatusInternalServerError, nil, nil, fmt.Errorf("missing model in %s", helper.TypeName(r.Controller))
	}
	if data, err := r.body(); err != nil {
		return err.Status(), nil, nil, err
	} else if ctrl, ok := r.Controller.(ControllerWithID); ok && ctrl != nil {
		table := db.NewSQLTable(r.Model.Table())
		fields := make([]db.SQLField, 0)
		for name, value := range data {
			fields = append(fields, db.NewSQLField(name, value))
		}
		if res, err := r.DB.Insert(table, fields); err != nil {
			return http.StatusInternalServerError, nil, nil, err
		} else if id := getModelField(r.Model, r.URL.ID.Name); id != nil {
			if m, ok := res.(map[string]interface{}); ok && m != nil {
				if id, ok := m[id.Name()]; ok && id != nil {
					res = id
				}
			}
			if f := getModelFields(r.Model, r.User.Role(), usr.ALEVEL_READ); id.Validate(res) && f != nil && len(f) > 0 && res != nil {
				table := db.NewSQLTable(r.Model.Table())
				fields := make([]db.SQLField, 0)
				where := []db.SQLWhere{db.NewSQLWhere(id.Name(), res)}
				for _, field := range f {
					fields = append(fields, db.NewSQLField(field.Name(), nil))
				}
				if body, err := r.DB.Select(table, fields, where, nil, nil, nil, nil, nil); err != nil {
					return http.StatusInternalServerError, nil, nil, err
				} else if body != nil && len(body) == 1 {
					return http.StatusCreated, nil, r.expand(r.Model, body[0]), err
				}
			}
		}
	}
	return http.StatusInternalServerError, nil, nil, fmt.Errorf("unknown error")
}

func actionUpdate(r *Request) (int, map[string]string, interface{}, error) {
	if r.Model == nil {
		return http.StatusInternalServerError, nil, nil, fmt.Errorf("missing model in %s", helper.TypeName(r.Controller))
	} else if r.URL.ID.Value == nil {
		return http.StatusBadRequest, nil, nil, fmt.Errorf("missing identifier")
	}
	if data, err := r.body(); err != nil {
		return err.Status(), nil, nil, err
	} else if ctrl, ok := r.Controller.(ControllerWithID); ok && ctrl != nil {
		table := db.NewSQLTable(r.Model.Table())
		fields := make([]db.SQLField, 0)
		where := []db.SQLWhere{db.NewSQLWhere(r.URL.ID.Name, string(r.URL.ID.Value))}
		for name, value := range data {
			fields = append(fields, db.NewSQLField(name, value))
		}
		if err := r.DB.Update(table, fields, where); err != nil {
			return http.StatusInternalServerError, nil, nil, err
		} else if f := getModelFields(r.Model, r.User.Role(), usr.ALEVEL_READ); f != nil && len(f) > 0 {
			fields = make([]db.SQLField, 0)
			limit := db.NewSQLLimit(1)
			for _, field := range f {
				fields = append(fields, db.NewSQLField(field.Name(), nil))
			}
			if body, err := r.DB.Select(table, fields, where, nil, nil, nil, limit, nil); err != nil {
				return http.StatusInternalServerError, nil, nil, err
			} else if body != nil && len(body) == 1 {
				return http.StatusAccepted, nil, r.expand(r.Model, body[0]), err
			}
		}
	}
	return http.StatusInternalServerError, nil, nil, fmt.Errorf("unknown error")
}

func actionDelete(r *Request) (int, map[string]string, interface{}, error) {
	if r.Model == nil {
		return http.StatusInternalServerError, nil, nil, fmt.Errorf("wrong model by %s", helper.TypeName(r.Controller))
	} else if r.URL.ID.Value == nil {
		return http.StatusBadRequest, nil, nil, fmt.Errorf("missing identifier")
	}
	if ctrl, ok := r.Controller.(ControllerWithID); ok && ctrl != nil {
		table := db.NewSQLTable(r.Model.Table())
		where := []db.SQLWhere{db.NewSQLWhere(r.URL.ID.Name, string(r.URL.ID.Value))}
		if f := getModelFields(r.Model, r.User.Role(), usr.ALEVEL_READ); f != nil && len(f) > 0 {
			fields := make([]db.SQLField, 0)
			limit := db.NewSQLLimit(1)
			for _, field := range f {
				fields = append(fields, db.NewSQLField(field.Name(), nil))
			}
			if body, err := r.DB.Select(table, fields, where, nil, nil, nil, limit, nil); err != nil {
				return http.StatusInternalServerError, nil, nil, err
			} else if err := r.DB.Delete(table, where); err != nil {
				return http.StatusInternalServerError, nil, nil, err
			} else if body != nil && len(body) == 1 {
				return http.StatusAccepted, nil, r.expand(r.Model, body[0]), err
			}
		}
	}
	return http.StatusInternalServerError, nil, nil, fmt.Errorf("unknown error")
}

/***********************************************************************************************************************
 * helper
 */
func getActionID(action Action) string {
	name := ""
	pattern := ""
	if c, ok := action.(ActionWithID); ok == true && c != nil {
		name, pattern = c.Id()
		if len(pattern) == 0 {
			pattern = "[0-9]+"
			name = helper.HttpPathTrim(name)
		}
		if len(name) == 0 {
			name = "id"
		}
		pattern = helper.HttpPathTrim(pattern)
		name = regexp.MustCompile(`[^A-Za-z_-]+`).ReplaceAllString(name, "")
		name = helper.HttpPathTrim(name)
		return fmt.Sprintf("{%s:%s}", name, pattern)
	}
	return ""
}

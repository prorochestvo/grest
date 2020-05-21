package grest

import (
	"fmt"
	"grest/db"
	"grest/internal"
	"grest/internal/helper"
	"grest/internal/logger"
	"grest/internal/mux"
	"grest/usr"
	"io/ioutil"
	"time"
)

func newRequest(r *mux.Request, route *route) *Request {
	result := Request{Request: r}
	result.route = route
	result.User = nil
	_, result.URL.ID.Name, _ = helper.HttpPathID(fmt.Sprintf("/%s", makeControllerActionID(route.Controller, route.action)))
	return &result
}

type Request struct {
	User usr.User
	*route
	*mux.Request
}

func (this *Request) Logger(format string, arg ...interface{}) {
	if err := logger.FileSavef(format, arg); err != nil {
		_, _ = this.router.Stderr.Write([]byte(fmt.Sprintf("logger: %s, %s", time.Now().Format("2006-01-02"), err.Error())))
	}
}

func (this *Request) body() (map[string]interface{}, internal.Error) {
	var result map[string]interface{} = nil
	var tmp interface{} = nil
	if b, err := ioutil.ReadAll(this.Body); this.Body != nil && err != nil {
		return nil, internal.NewError(internal.StatusBadRequest, err.Error())
	} else if err := this.router.ContentType.Unmarshal(b, &tmp); err != nil {
		return nil, internal.NewError(internal.StatusBadRequest, err.Error())
	} else if data, ok := tmp.(map[string]interface{}); !ok || data == nil || len(data) == 0 {
		return nil, internal.NewError(internal.StatusBadRequest, "wrong request body")
	} else if fields := getModelFields(this.Model, this.User.Role(), usr.ALEVEL_WRITE); fields == nil || len(fields) == 0 {
		return nil, internal.NewError(internal.StatusForbidden, "fields not found by %s", helper.TypeName(this.Model))
	} else {
		result = make(map[string]interface{}, 0)
		for name, value := range data {
			if field, ok := fields[name]; !ok || !field.Validate(value) {
				return nil, internal.NewError(internal.StatusUnprocessableEntity, "wrong field %s", name)
			}
			result[name] = value
		}
	}
	return result, nil
}

func (this *Request) expand(model Model, data interface{}) interface{} {
	const interimKeyName string = "tmp_key_a7271a8b5f3b9ca7d5cb65d07a8f50f6"
	type Binding interface {
		InternalKeys() []Field
		ExternalModel() Model
		ExternalKeys() []Field
		Limit() int64
		ExtraField
	}
	// get all internal Binding
	getInternalValues := func(field Binding) [][]interface{} {
		result := make([][]interface{}, 0)
		if items, ok := data.([]map[string]interface{}); ok && items != nil {
			for _, item := range items {
				key := make([]interface{}, 0)
				for _, internalKey := range field.InternalKeys() {
					val, ok := item[internalKey.Name()]
					if !ok || val == nil {
						break
					}
					key = append(key, val)
				}
				if l := len(key); l == len(field.InternalKeys()) && l == len(field.ExternalKeys()) {
					result = append(result, key)
				}
			}
		} else if item, ok := data.(map[string]interface{}); ok && items != nil {
			key := make([]interface{}, 0)
			for _, internalKey := range field.InternalKeys() {
				val, ok := item[internalKey.Name()]
				if !ok || val == nil {
					break
				}
				key = append(key, val)
			}
			if l := len(key); l == len(field.InternalKeys()) && l == len(field.ExternalKeys()) {
				result = append(result, key)
			}
		}
		return result
	}
	// get all external values
	getExternalValues := func(field Binding, internalValues [][]interface{}) []map[string]interface{} {
		if len(internalValues) == 0 {
			return make([]map[string]interface{}, 0)
		}
		if f := getModelFields(field.ExternalModel(), this.User.Role(), usr.ALEVEL_READ); f != nil && len(f) > 0 {
			fields := make([]db.SQLField, 0)
			table := db.NewSQLTable(field.ExternalModel().Table())
			where := make([]db.SQLWhere, 0)
			for i, externalKey := range field.ExternalKeys() {
				where = append(where, db.NewSQLWhere(externalKey.Name(), internalValues[i], "in"))
				fields = append(fields, db.NewSQLField(fmt.Sprintf("%s as %s_%d", externalKey.Name(), interimKeyName, i), nil))
			}
			for _, field := range f {
				fields = append(fields, db.NewSQLField(field.Name(), nil))
			}
			var limit db.SQLLimit = nil
			if l := field.Limit(); l >= 0 {
				limit = db.NewSQLLimit(l)
			}
			if res, err := this.DB.Select(table, fields, where, nil, nil, nil, limit, nil); err == nil && res != nil {
				return res
			}
		}
		return make([]map[string]interface{}, 0)
	}
	// get all external value by internal key
	getExternalValueBy := func(field Binding, internalValue []interface{}, values []map[string]interface{}) interface{} {
		if internalValue == nil {
			return make([]map[string]interface{}, 0)
		}
		result := make([]map[string]interface{}, 0)
		for _, value := range values {
			// check all keys (internal[i] == external[i])
			check := true
			for i, iValue := range internalValue {
				fname := fmt.Sprintf("%s_%d", interimKeyName, i)
				if v, ok := value[fname]; !ok || v != iValue {
					check = false
					break
				}
				delete(value, fname)
			}
			// save
			if check == true {
				result = append(result, value)
			}
		}
		if len(result) > 0 {
			// рекурсия для всех под модулей
			result = this.expand(field.ExternalModel(), result).([]map[string]interface{})
		}
		if field.Limit() == 1 {
			if len(result) > 0 {
				return result[0]
			} else {
				return make(map[string]interface{}, 0)
			}
		}
		return result
	}
	// binding internal with external keys/values
	extraFields := getModelExtraFields(model, this.User.Role())
	if extraFields != nil && len(extraFields) > 0 {
		for _, f := range extraFields {
			if field, ok := f.(Binding); ok && field != nil {
				internalValues := getInternalValues(field)
				externalValues := getExternalValues(field, internalValues)
				if items, ok := data.([]map[string]interface{}); ok && items != nil {
					for i, item := range items {
						vals := make([]interface{}, 0)
						for _, key := range field.InternalKeys() {
							if val, ok := item[key.Name()]; ok && val != nil {
								vals = append(vals, val)
							}
						}
						item[field.Name()] = getExternalValueBy(field, vals, externalValues)
						items[i] = item
					}
					data = items
				} else if item, ok := data.(map[string]interface{}); ok && items != nil {
					vals := make([]interface{}, 0)
					for _, key := range field.InternalKeys() {
						if val, ok := item[key.Name()]; ok && val != nil {
							vals = append(vals, val)
						}
					}
					item[field.Name()] = getExternalValueBy(field, vals, externalValues)
					data = item
				}
			}
		}
	}
	return data
}

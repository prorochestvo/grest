package grest

import (
	"github.com/prorochestvo/grest/internal"
	"github.com/prorochestvo/grest/usr"
)

/*
 * Интерфейс модели реализуемый пользователем
 */
type Model interface {
	Table() string
	Fields() []Field
}

type ModelWithExtraFields interface {
	ExtraFields() []ExtraField
	Model
}

func NewModel(table string, fields []Field, extraFields ...ExtraField) Model {
	result := model{}
	result.table = table
	result.fields = fields
	result.extraFields = extraFields
	return &result
}

type model struct {
	table       string
	fields      []Field
	extraFields []ExtraField
}

func (this *model) Table() string {
	return this.table
}

func (this *model) Fields() []Field {
	return this.fields
}

func (this *model) ExtraFields() []ExtraField {
	return this.extraFields
}

func getModelField(model Model, name string) Field {
	var result Field = nil
	for _, field := range model.Fields() {
		if field.Name() != name {
			continue
		}
		result = field
	}
	return result
}

func getModelFields(model Model, role usr.Role, level ...internal.AccessLevel) map[string]Field {
	result := make(map[string]Field, 0)
	for _, field := range model.Fields() {
		if field == nil {
			continue
		}
		// find role
		if r := field.Roles(level...); r == nil || len(r) == 0 || r.IndexOf(role) < 0 {
			continue
		}
		result[field.Name()] = field
	}
	return result
}

func getModelExtraFields(model Model, role usr.Role) map[string]ExtraField {
	result := make(map[string]ExtraField, 0)
	if m, ok := model.(ModelWithExtraFields); ok && m != nil {
		for _, field := range m.ExtraFields() {
			if field == nil {
				continue
			}
			// find role
			if r := field.Roles(); r != nil && len(r) > 0 && r.IndexOf(role) < 0 {
				continue
			}
			result[field.Name()] = field
		}
	}
	return result
}

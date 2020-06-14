package grest

import (
	"github.com/prorochestvo/grest/internal"
	"github.com/prorochestvo/grest/usr"
	"strconv"
	"time"
)

func FIELD(name string, parser func(value string) (interface{}, error), permission ...usr.Permission) FieldEx {
	return newField(name, parser, nil, permission...)
}

func TEXT(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		return value, nil
	}
	return newField(name, parser, nil, permission...)
}

func INT8(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		i, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return nil, err
		}
		return int8(i), nil
	}
	return newField(name, parser, nil, permission...)
}

func INT16(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		i, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, err
		}
		return int16(i), nil
	}
	return newField(name, parser, nil, permission...)
}

func INT32(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, err
		}
		return int32(i), nil
	}
	return newField(name, parser, nil, permission...)
}

func INT64(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	}
	return newField(name, parser, nil, permission...)
}

func UINT8(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		u, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, err
		}
		return uint8(u), nil
	}
	return newField(name, parser, nil, permission...)
}

func UINT16(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		u, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return nil, err
		}
		return uint16(u), nil
	}
	return newField(name, parser, nil, permission...)
}

func UINT32(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		u, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, err
		}
		return uint32(u), nil
	}
	return newField(name, parser, nil, permission...)
}

func UINT64(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return u, nil
	}
	return newField(name, parser, nil, permission...)
}

func FLOAT32(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return float32(f), nil
	}
	return newField(name, parser, nil, permission...)
}

func FLOAT64(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
	return newField(name, parser, nil, permission...)
}

func BOOLEAN(name string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return newField(name, parser, nil, permission...)
}

func DATETIME(name string, format string, permission ...usr.Permission) FieldEx {
	parser := func(value string) (interface{}, error) {
		b, err := time.Parse(format, value)
		if err != nil {
			return nil, err
		}
		return b.Format(format), nil
	}
	return newField(name, parser, nil, permission...)
}

func EXPAND(name string, internalKeys []Field, externalModel Model, externalKeys []Field, limit int64, role ...usr.Role) ExtraField {
	result := binding{}
	result.name = name
	result.internalKeys = internalKeys
	result.externalModel = externalModel
	result.externalKeys = externalKeys
	result.limit = limit
	result.roles = role
	return &result
}

/**
 * Field
 */
type Field interface {
	Name() string
	Validate(value interface{}) bool
	Parser(value string) (interface{}, error)
	Roles(accessLevel ...internal.AccessLevel) usr.Roles
}

type FieldEx interface {
	SetValidate(value func(interface{}) bool)
	Field
}

func newField(name string, parser func(value string) (interface{}, error), validator func(interface{}) bool, permission ...usr.Permission) *field {
	result := field{}
	result.name = name
	result.parser = parser
	result.validator = validator
	result.permission = make([]usr.Permission, 0)
	for _, p := range permission {
		result.permission = append(result.permission, p)
	}
	return &result
}

type field struct {
	name       string
	parser     func(value string) (interface{}, error)
	validator  func(interface{}) bool
	permission []usr.Permission
}

func (this *field) Name() string {
	return this.name
}

func (this *field) Validate(value interface{}) bool {
	if this.validator == nil {
		return true
	}
	return this.validator(value)
}

func (this *field) SetValidate(value func(interface{}) bool) {
	this.validator = value
}

func (this *field) Parser(value string) (interface{}, error) {
	if this.parser == nil {
		return nil, nil
	}
	return this.parser(value)
}

func (this *field) Roles(accessLevel ...internal.AccessLevel) usr.Roles {
	result := make([]usr.Role, 0)
	if accessLevel == nil || len(accessLevel) == 0 {
		for _, permission := range this.permission {
			result = append(result, permission.Role())
		}
	} else {
		for _, permission := range this.permission {
			if permission == nil {
				continue
			}
			for _, level := range accessLevel {
				if !permission.Access(level) {
					continue
				}
				result = append(result, permission.Role())
				break
			}
		}
	}
	return result
}

/**
 * Extra field
 */
type ExtraField interface {
	Name() string
	Roles() usr.Roles
}

// binding model[keys] <-> model[keys]
type binding struct {
	name          string
	roles         []usr.Role
	internalKeys  []Field
	externalModel Model
	externalKeys  []Field
	limit         int64
}

func (this *binding) Name() string { return this.name }

func (this *binding) Roles() usr.Roles {
	if this.roles == nil {
		return make([]usr.Role, 0)
	}
	return this.roles
}

func (this *binding) InternalKeys() []Field { return this.internalKeys }

func (this *binding) ExternalModel() Model { return this.externalModel }

func (this *binding) ExternalKeys() []Field { return this.externalKeys }

func (this *binding) Limit() int64 { return this.limit }

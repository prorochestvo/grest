package db

type SQLField interface {
	Name() string
	Value() interface{}
}

func NewSQLField(name string, value interface{}) SQLField {
	result := sqlField{}
	result.name = name
	result.value = value
	return &result
}

type sqlField struct {
	name  string
	value interface{}
}

func (this *sqlField) Name() string {
	return this.name
}

func (this *sqlField) Value() interface{} {
	return this.value
}

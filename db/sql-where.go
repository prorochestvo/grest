package db

import (
	"strings"
)

type SQLWhere interface {
	Instruction() string
	Field() string
	Separator() string
	Value() interface{}
	Negative() bool
}

func NewSQLWhere(field string, value interface{}, instruction ...string) SQLWhere {
	result := sqlWhere{instruction: "", separator: "AND"}
	if instruction != nil && len(instruction) > 0 {
		for _, o := range instruction {
			s := strings.Trim(strings.ToUpper(o), "\t\n\r ")
			if s == "OR" || s == "AND" {
				result.separator = s
			} else {
				result.instruction = o
			}
		}
	}
	result.field = field
	result.value = value
	return &result
}

type sqlWhere struct {
	instruction string
	field       string
	separator   string
	value       interface{}
	negative    bool
}

func (this *sqlWhere) Instruction() string {
	return this.instruction
}

func (this *sqlWhere) Field() string {
	return this.field
}

func (this *sqlWhere) Separator() string {
	return this.separator
}

func (this *sqlWhere) Value() interface{} {
	return this.value
}

func (this *sqlWhere) Negative() bool {
	return this.negative
}

package db

import (
	"strings"
)

type SQLHaving interface {
	Instruction() string
	Field() string
	Separator() string
	Value() interface{}
	Negative() bool
}

func NewSQLHaving(field string, value interface{}, instruction ...string) SQLHaving {
	result := sqlHaving{}
	result.sqlWhere = sqlWhere{instruction: "", separator: "AND"}
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

type sqlHaving struct {
	sqlWhere
}

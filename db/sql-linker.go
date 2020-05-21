package db

import (
	"fmt"
	"strings"
)

type SQLLinker interface {
	Select(table SQLTable, fields []SQLField, where []SQLWhere, groupBy []SQLGroupBy, having []SQLHaving, orderBy []SQLOrderBy, limit SQLLimit, offset SQLOffset) string
	Insert(table SQLTable, fields []SQLField) string
	Update(table SQLTable, fields []SQLField, where []SQLWhere) string
	Delete(table SQLTable, where []SQLWhere) string
}

func NewSQLLinker() SQLLinker {
	result := sqlLinker{}
	return &result
}

type sqlLinker struct {
}

func (this *sqlLinker) Select(table SQLTable, fields []SQLField, where []SQLWhere, groupBy []SQLGroupBy, having []SQLHaving, orderBy []SQLOrderBy, limit SQLLimit, offset SQLOffset) string {
	f := ""
	if fields != nil && len(fields) > 0 {
		for i, field := range fields {
			if i == 0 {
				f += fmt.Sprintf("%s", field.Name())
			} else {
				f += fmt.Sprintf(", %s", field.Name())
			}
		}
	}
	result := fmt.Sprintf("SELECT %s\nFROM %s", f, table.Name())
	// sql where
	if where != nil && len(where) > 0 {
		result += "\n" + this.parser(where)
	}
	// sql group-by
	if groupBy != nil && len(groupBy) > 0 {
		result += "\n" + this.parser(groupBy)
	}
	// sql having
	if having != nil && len(having) > 0 {
		result += "\n" + this.parser(having)
	}
	// sql order-by
	if orderBy != nil && len(orderBy) > 0 {
		result += "\n" + this.parser(orderBy)
	}
	// sql limit
	if limit != nil && limit.Count() >= 0 {
		result += "\n" + this.parser(limit)
	}
	// sql offset
	if offset != nil && offset.Rows() >= 0 {
		result += "\n" + this.parser(offset)
	}
	result += ";"
	return result
}

func (this *sqlLinker) Insert(table SQLTable, fields []SQLField) string {
	f := ""
	v := ""
	if fields != nil && len(fields) > 0 {
		for i, field := range fields {
			if i == 0 {
				f += fmt.Sprintf("%s", field.Name())
				v += fmt.Sprintf("%s", SQLEscape(field.Value()))
			} else {
				f += fmt.Sprintf(", %s", field.Name())
				v += fmt.Sprintf(", %s", SQLEscape(field.Value()))
			}
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s)\nVALUES (%s);", table.Name(), f, v)
}

func (this *sqlLinker) Update(table SQLTable, fields []SQLField, where []SQLWhere) string {
	f := ""
	if fields != nil && len(fields) > 0 {
		for i, field := range fields {
			if i == 0 {
				f += fmt.Sprintf("%s = %s", field.Name(), SQLEscape(field.Value()))
			} else {
				f += fmt.Sprintf(", %s = %s", field.Name(), SQLEscape(field.Value()))
			}
		}
	}
	result := fmt.Sprintf("UPDATE %s\nSET %s", table.Name(), f)
	// sql where
	if where != nil && len(where) > 0 {
		result += "\n" + this.parser(where)
	} else {
		result += "\nWHERE (false)"
	}
	result += ";"
	return result
}

func (this *sqlLinker) Delete(table SQLTable, where []SQLWhere) string {
	result := fmt.Sprintf("DELETE FROM %s", table.Name())
	// sql where
	if where != nil && len(where) > 0 {
		result += "\n" + this.parser(where)
	} else {
		result += "\nWHERE (false)"
	}
	result += ";"
	return result
}

func (this *sqlLinker) parser(value interface{}) string {
	result := ""
	// where
	if where, ok := value.([]SQLWhere); ok && where != nil && len(where) > 0 {
		tmp := ""
		for _, w := range where {
			q := ""
			if o := strings.ToLower(w.Instruction()); o == "between" {
				if v, ok := w.Value().([]interface{}); ok && v != nil && len(v) == 2 {
					q = fmt.Sprintf("%s BETWEEN %s AND %s", w.Field(), SQLEscape(v[0]), SQLEscape(v[1]))
					if w.Negative() {
						q = fmt.Sprintf("NOT(%s)", q)
					}
				}
			} else if o == "in" {
				if v, ok := w.Value().([]interface{}); ok && v != nil && len(v) > 0 {
					tmp := make([]string, 0)
					for _, val := range v {
						tmp = append(tmp, SQLEscape(val))
					}
					q = fmt.Sprintf("%s IN (%s)", w.Field(), strings.Join(tmp, ", "))
					if w.Negative() {
						q = fmt.Sprintf("NOT(%s)", q)
					}
				}
			} else if o == "like" {
				q = fmt.Sprintf("%s LIKE %s", w.Field(), SQLEscape(w.Value()))
			} else if o == "is_null" {
				if w.Negative() {
					q = fmt.Sprintf("IS NOT NULL %s", w.Field())
				} else {
					q = fmt.Sprintf("IS NULL %s", w.Field())
				}
			} else if o == "<" {
				q = fmt.Sprintf("%s < %s", w.Field(), SQLEscape(w.Value()))
			} else if o == "<=" {
				q = fmt.Sprintf("%s <= %s", w.Field(), SQLEscape(w.Value()))
			} else if o == ">" {
				q = fmt.Sprintf("%s > %s", w.Field(), SQLEscape(w.Value()))
			} else if o == ">=" {
				q = fmt.Sprintf("%s >= %s", w.Field(), SQLEscape(w.Value()))
			} else if len(o) == 0 {
				if w.Negative() {
					q = fmt.Sprintf("%s <> %s", w.Field(), SQLEscape(w.Value()))
				} else {
					q = fmt.Sprintf("%s = %s", w.Field(), SQLEscape(w.Value()))
				}
			}
			if s := w.Separator(); len(tmp) == 0 {
				tmp = fmt.Sprintf("(%s)", q)
			} else if s == "OR" {
				tmp = fmt.Sprintf("%s %s (%s)", tmp, s, q)
			} else if s == "AND" {
				tmp = fmt.Sprintf("%s %s (%s)", tmp, s, q)
			}
		}
		if len(tmp) > 0 {
			result = fmt.Sprintf("WHERE %s", tmp)
		}
	} else
	// group-by
	if groupBy, ok := value.([]SQLGroupBy); ok && groupBy != nil && len(groupBy) > 0 {
		tmp := ""
		for _, g := range groupBy {
			if len(g.Field()) == 0 {
				continue
			}
			if len(tmp) == 0 {
				tmp = fmt.Sprintf("%s", g.Field())
			} else {
				tmp = fmt.Sprintf("%s, %s", tmp, g.Field())
			}
		}
		if len(tmp) > 0 {
			result = fmt.Sprintf("GROUP BY %s", tmp)
		}
	} else
	// []db.SQLHaving
	if having, ok := value.([]SQLHaving); ok && having != nil && len(having) > 0 {
		tmp := ""
		for _, w := range having {
			q := ""
			if o := strings.ToLower(w.Instruction()); o == "between" {
				if v, ok := w.Value().([]interface{}); ok && v != nil && len(v) == 2 {
					q = fmt.Sprintf("%s BETWEEN %s AND %s", w.Field(), SQLEscape(v[0]), SQLEscape(v[1]))
					if w.Negative() {
						q = fmt.Sprintf("NOT(%s)", q)
					}
				}
			} else if o == "like" {
				q = fmt.Sprintf("%s LIKE %s", w.Field(), SQLEscape(w.Value()))
			} else if o == "is_null" {
				if w.Negative() {
					q = fmt.Sprintf("IS NOT NULL %s", w.Field())
				} else {
					q = fmt.Sprintf("IS NULL %s", w.Field())
				}
			} else if o == "<" {
				q = fmt.Sprintf("%s < %s", w.Field(), SQLEscape(w.Value()))
			} else if o == "<=" {
				q = fmt.Sprintf("%s <= %s", w.Field(), SQLEscape(w.Value()))
			} else if o == ">" {
				q = fmt.Sprintf("%s > %s", w.Field(), SQLEscape(w.Value()))
			} else if o == ">=" {
				q = fmt.Sprintf("%s >= %s", w.Field(), SQLEscape(w.Value()))
			} else if len(o) == 0 {
				if w.Negative() {
					q = fmt.Sprintf("%s <> %s", w.Field(), SQLEscape(w.Value()))
				} else {
					q = fmt.Sprintf("%s = %s", w.Field(), SQLEscape(w.Value()))
				}
			}
			if s := w.Separator(); len(tmp) == 0 {
				tmp = fmt.Sprintf("(%s)", q)
			} else if s == "OR" {
				tmp = fmt.Sprintf("%s %s (%s)", tmp, s, q)
			} else if s == "AND" {
				tmp = fmt.Sprintf("%s %s (%s)", tmp, s, q)
			}
		}
		if len(tmp) > 0 {
			result = fmt.Sprintf("HAVING %s", tmp)
		}
	} else
	// order-by
	if orderBy, ok := value.([]SQLOrderBy); ok && orderBy != nil && len(orderBy) > 0 {
		tmp := ""
		for _, o := range orderBy {
			if len(o.Field()) == 0 {
				continue
			}
			sort := "ASC"
			if len(o.Sort()) > 0 {
				sort = o.Sort()
			}
			if len(tmp) == 0 {
				tmp = fmt.Sprintf("%s %s", o.Field(), sort)
			} else {
				tmp = fmt.Sprintf("%s, %s %s", tmp, o.Field(), sort)
			}
		}
		if len(tmp) > 0 {
			result = fmt.Sprintf("ORDER BY %s", tmp)
		}
	} else
	// limit
	if limit, ok := value.(SQLLimit); ok && limit != nil && limit.Count() >= 0 {
		result = fmt.Sprintf("LIMIT %d", limit.Count())
	} else
	// offset
	if offset, ok := value.(SQLOffset); ok && offset != nil && offset.Rows() >= 0 {
		result = fmt.Sprintf("OFFSET %d", offset.Rows())
	}
	return result
}

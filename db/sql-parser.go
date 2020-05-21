package db

import (
	"grest/internal/helper"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//func SQLSelect(r *http.Request) string {
//	result := ""
//	return result
//}

var instructions = []string{
	"between",
	"is_null",
	"like",
	"cmp_be", ">=",
	"cmp_b", ">",
	"cmp_l", "<",
	"cmp_le", "<=",
	"group",
	"sort",
	"offset",
	"limit",
	"in",
}

func SQLParserEx(r *http.Request, parsers map[string]func(value string) (interface{}, error)) (where []SQLWhere, groupBy []SQLGroupBy, orderBy []SQLOrderBy, having []SQLHaving, limit SQLLimit, offset SQLOffset) {
	where, groupBy, orderBy, having, limit, offset = SQLParser(r.URL.Query())
	if where != nil && len(where) > 0 {
		tmp := make([]SQLWhere, 0)
		for _, w := range where {
			// check field name
			parser, ok := parsers[w.Field()]
			if !ok || parser == nil {
				continue
			}
			// check field value
			var value interface{} = nil
			if text, ok := w.Value().(string); ok {
				val, err := parser(text)
				if err != nil {
					continue
				}
				value = val
			} else if slice, ok := w.Value().([]interface{}); ok && slice != nil && len(slice) > 0 {
				val := make([]interface{}, 0)
				for _, s := range slice {
					if text, ok := s.(string); ok {
						v, err := parser(text)
						if err != nil {
							continue
						}
						val = append(val, v)
					}
				}
				if l := len(val); len(slice) != l {
					continue
				}
				value = val
			} else {
				continue
			}
			// save new instruction
			tmp = append(tmp, &sqlWhere{
				instruction: w.Instruction(),
				field:       w.Field(),
				separator:   w.Separator(),
				value:       value,
				negative:    w.Negative(),
			})
		}
		where = tmp
	}
	if groupBy != nil && len(groupBy) > 0 {
		tmp := make([]SQLGroupBy, 0)
		for _, g := range groupBy {
			// check field name
			parser, ok := parsers[g.Field()]
			if !ok || parser == nil {
				continue
			}
			// save new instruction
			tmp = append(tmp, g)
		}
		groupBy = tmp
	}
	if orderBy != nil && len(orderBy) > 0 {
		tmp := make([]SQLOrderBy, 0)
		for _, o := range orderBy {
			// check field name
			parser, ok := parsers[o.Field()]
			if !ok || parser == nil {
				continue
			}
			// save new instruction
			tmp = append(tmp, o)
		}
		orderBy = tmp
	}
	if having != nil && len(having) > 0 {
		tmp := make([]SQLHaving, 0)
		for _, h := range having {
			// check field name
			parser, ok := parsers[h.Field()]
			if !ok || parser == nil {
				continue
			}
			// check field value
			var value interface{} = nil
			if text, ok := h.Value().(string); ok {
				val, err := parser(text)
				if err != nil {
					continue
				}
				value = val
			} else if slice, ok := h.Value().([]interface{}); ok && slice != nil && len(slice) > 0 {
				val := make([]interface{}, 0)
				for _, s := range slice {
					if text, ok := s.(string); ok {
						v, err := parser(text)
						if err != nil {
							continue
						}
						val = append(val, v)
					}
				}
				if len(slice) != len(val) {
					continue
				}
				value = val
			} else {
				continue
			}
			// save new instruction
			tmp = append(tmp, &sqlHaving{
				sqlWhere: sqlWhere{
					instruction: h.Instruction(),
					field:       h.Field(),
					separator:   h.Separator(),
					value:       value,
					negative:    h.Negative(),
				},
			})
		}
		having = tmp
	}
	return where, groupBy, orderBy, having, limit, offset
}

/*
 * http://127.0.0.1:8080/user?:test=1&:|arg[]=t1&:|arg[]=t2&0:!|between[FIELD][]=1&0:!|between[FIELD][]=2
 *
 * SELECT fields
 * FROM table
 * WHERE
 *
 *   COMMAND                             // AND
 *   |COMMAND                            // OR
 *   !COMMAND                            // NOT
 *   NUM:COMMAND                         // NUM сордировки
 *
 *   FIELD_NAME=VALUE                     // field=value
 *   !FIELD_NAME=VALUE                    // field<>value
 *   FIELD_NAME[]=VALUE                   // field in (value)
 *   !FIELD_NAME[]=VALUE                  // NOT(field IN (value))
 *   :between[FIELD_NAME][]=VALUE         // field BETWEEN value:first AND value:last
 *   :!between[FIELD_NAME][]=VALUE        // NOT(field BETWEEN value:first AND value:last)
 *   :is_null[FIELD_NAME]                 // IS NULL field
 *   :!is_null[FIELD_NAME]                // IS NOT NULL field
 *   :like[FIELD_NAME]=VALUE              // field LIKE value
 *   :!like[FIELD_NAME]=VALUE             // NOT(field LIKE value)
 *   :cmp_be[FIELD_NAME]=VALUE            // field >= value
 *   :!cmp_be[FIELD_NAME]=VALUE           // NOT(field >= value)
 *   :cmp_b[FIELD_NAME]=VALUE             // field > value
 *   :!cmp_b[FIELD_NAME]=VALUE            // NOT(field > value)
 *   :cmp_l[FIELD_NAME]=VALUE             // field < value
 *   :!cmp_l[FIELD_NAME]=VALUE            // NOT(field < value)
 *   :cmp_le[FIELD_NAME]=VALUE            // field <= value
 *   :!cmp_le[FIELD_NAME]=VALUE           // NOT(field <= value)
 * GROUP BY
 *   :group[FIELD_NAME]                   // Групперовать по FIELD и включая его в SELECT
 * ORDER BY
 *   :sort[FIELD_NAME]=ASC|DESC
 * OFFSET
 *   :offset=VALUE
 * LIMIT
 *   :limit=VALUE
 */
func SQLParser(query url.Values) (where []SQLWhere, groupBy []SQLGroupBy, orderBy []SQLOrderBy, having []SQLHaving, limit SQLLimit, offset SQLOffset) {
	where = make([]SQLWhere, 0)
	groupBy = make([]SQLGroupBy, 0)
	orderBy = make([]SQLOrderBy, 0)
	having = make([]SQLHaving, 0)
	limit = nil
	offset = nil
	// parser url values
	rx := regexp.MustCompile(`(?i)^(\d+)*([:!|]*)([0-9A-Za-z<_=->]+)(?:\[(.*)\])*$`)
	var index uint64 = 0xFFFFFFFFFFFFFFFF
	options := make([]struct {
		Number      uint64
		Separator   string
		Instruction string
		Negative    bool
		Field       string
		Value       interface{}
	}, 0)
	for key, val := range query {
		var number uint64
		var separator string
		var negative bool
		var instruction string
		var field string
		var value interface{}
		for _, match := range rx.FindAllStringSubmatch(key, -1) {
			if len(match) != 5 {
				continue
			}
			if n, err := strconv.ParseUint(match[1], 10, 64); len(match[1]) > 0 && err == nil {
				number = n
			} else {
				index--
				number = index
			}
			separator = "AND"
			negative = false
			for _, c := range match[2] {
				switch c {
				case '!':
					negative = true
				case '|':
					separator = "OR"
				case ':':
					fields := strings.Split(match[4], "][")
					if n := strings.ToLower(match[3]); helper.StringsIndexOf(instructions, n) >= 0 {
						// :between[FIELD_NAME][]=VALUE
						// :!between[FIELD_NAME][]=VALUE
						if n == "between" {
							if len(val) >= 2 && len(fields) > 0 && len(fields[0]) > 0 {
								instruction = n
								value = []interface{}{val[0], val[len(val)-1]}
								field = fields[0]
							}
						} else
						// :is_null[FIELD_NAME]
						// :!is_null[FIELD_NAME]
						if n == "is_null" {
							if len(val) > 0 && len(fields) > 0 && len(fields[0]) > 0 {
								instruction = n
								value = val[0]
								field = fields[0]
							}
						} else
						// :like[FIELD_NAME]=VALUE
						// :!like[FIELD_NAME]=VALUE
						if n == "like" {
							if len(val) > 0 && len(fields) > 0 && len(fields[0]) > 0 {
								instruction = n
								value = val[0]
								field = fields[0]
							}
						} else
						// :cmp_be[FIELD_NAME]=VALUE
						// :!cmp_be[FIELD_NAME]=VALUE
						if n == "cmp_be" || n == ">=" {
							if len(val) > 0 && len(fields) > 0 && len(fields[0]) > 0 {
								instruction = ">="
								value = val[0]
								field = fields[0]
							}
						} else
						// :cmp_b[FIELD_NAME]=VALUE
						// :!cmp_b[FIELD_NAME]=VALUE
						if n == "cmp_b" || n == ">" {
							if len(val) > 0 && len(fields) > 0 && len(fields[0]) > 0 {
								instruction = ">"
								value = val[0]
								field = fields[0]
							}
						} else
						// :cmp_l[FIELD_NAME]=VALUE
						// :!cmp_l[FIELD_NAME]=VALUE
						if n == "cmp_l" || n == "<" {
							if len(val) > 0 && len(fields) > 0 && len(fields[0]) > 0 {
								instruction = "<"
								value = val[0]
								field = fields[0]
							}
						} else
						// :cmp_le[FIELD_NAME]=VALUE
						// :!cmp_le[FIELD_NAME]=VALUE
						if n == "cmp_le" || n == "<=" {
							if len(val) > 0 && len(fields) > 0 && len(fields[0]) > 0 {
								instruction = "<="
								value = val[0]
								field = fields[0]
							}
						} else
						// :group[FIELD_NAME]
						// :group[FIELD_NAME]=VALUE
						if n == "group" || n == "group-by" {
							if len(val) > 0 && len(fields) > 0 && len(fields[0]) > 0 {
								instruction = "group"
								field = fields[0]
							}
							if val != nil && len(val) > 0 {
								value = val[0]
							}
						} else
						// :sort[FIELD_NAME]=ASC|DESC
						if n == "sort" || n == "order" || n == "order-by" {
							if len(val) == 0 {
								val = []string{"ASC"}
							} else if val[0] != "ASC" && val[0] != "DESC" {
								val[0] = "ASC"
							}
							if len(fields) > 0 && len(fields[0]) > 0 {
								instruction = "sort"
								value = val[0]
								field = fields[0]
							}
						} else
						// :offset=VALUE
						if n == "offset" && val != nil && len(val[0]) > 0 {
							if v, err := strconv.ParseInt(val[0], 10, 64); err == nil && v >= 0 {
								instruction = n
								value = v
							}
						} else
						// :limit=VALUE
						if n == "limit" && val != nil && len(val[0]) > 0 {
							if v, err := strconv.ParseInt(val[0], 10, 64); err == nil && v >= 0 {
								instruction = n
								value = v
							}
						} else
						// :IN[FIELD_NAME][]=VALUE
						// :!IN[FIELD_NAME][]=VALUE
						if n == "in" {
							if len(val) > 0 && len(fields) > 0 && len(fields[0]) > 0 {
								tmp := make([]interface{}, 0)
								for _, v := range val {
									tmp = append(tmp, v)
								}
								instruction = n
								value = tmp
								field = fields[0]
							}
						}
					}
				}
			}
			if strings.Index(match[2], ":") < 0 && val != nil && len(val) > 0 {
				field = match[3]
				value = val[0]
			}
			if len(instruction) == 0 && len(field) == 0 {
				continue
			}
			options = append(options, struct {
				Number      uint64
				Separator   string
				Instruction string
				Negative    bool
				Field       string
				Value       interface{}
			}{
				Number:      number,
				Separator:   separator,
				Instruction: instruction,
				Negative:    negative,
				Field:       field,
				Value:       value,
			})
		}
	}
	// sort options
	sort.Slice(options, func(i, j int) bool {
		return options[i].Number < options[j].Number
	})
	// parser options
	for _, option := range options {
		if len(option.Instruction) > 0 && helper.StringsIndexOf(instructions, option.Instruction) >= 0 {
			if option.Instruction == "group" {
				g := &sqlGroupBy{}
				g.field = option.Field
				groupBy = append(groupBy, g)
				g.field = option.Field
				w := &sqlHaving{}
				w.instruction = option.Instruction
				w.field = option.Field
				w.separator = option.Separator
				w.value = option.Value
				w.negative = option.Negative
				having = append(having, w)
			} else if option.Instruction == "sort" {
				if t, ok := option.Value.(string); ok && len(t) > 0 {
					s := &sqlOrderBy{}
					s.sort = t
					s.field = option.Field
					orderBy = append(orderBy, s)
				}
			} else if option.Instruction == "offset" {
				if i, ok := option.Value.(int64); ok && i > 0 {
					o := &sqlOffset{}
					o.value = i
					offset = o
				}
			} else if option.Instruction == "limit" {
				if i, ok := option.Value.(int64); ok && i > 0 {
					l := &sqlLimit{}
					l.value = i
					limit = l
				}
			} else {
				w := &sqlWhere{}
				w.instruction = option.Instruction
				w.field = option.Field
				w.separator = option.Separator
				w.value = option.Value
				w.negative = option.Negative
				where = append(where, w)
			}
		} else if len(option.Instruction) == 0 && len(option.Field) > 0 && option.Value != nil {
			w := &sqlWhere{}
			w.field = option.Field
			w.separator = option.Separator
			w.value = option.Value
			w.negative = option.Negative
			where = append(where, w)
		}
	}
	return
}

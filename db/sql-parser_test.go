package db

import (
	"fmt"
	"github.com/prorochestvo/grest/internal/helper"
	"net/url"
	"strings"
	"testing"
)

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
 *   :IN[FIELD_NAME][]=VALUE              // field IN value
 *   :!IN[FIELD_NAME][]=VALUE             // NOT(field IN value)
 * GROUP BY
 *   :group[FIELD_NAME]                   // Групперовать по FIELD и включая его в SELECT
 * ORDER BY
 *   :sort[FIELD_NAME]=ASC|DESC
 * OFFSET
 *   :offset=VALUE
 * LIMIT
 *   :limit=VALUE
 */
func TestSQLParser(t *testing.T) {
	// build normal instructions
	query := url.Values{}
	for i, instruction := range instructions {
		if instruction == "cmp_be" {
			i = helper.StringsIndexOf(instructions, ">=")
		} else if instruction == "cmp_b" {
			i = helper.StringsIndexOf(instructions, ">")
		} else if instruction == "cmp_l" {
			i = helper.StringsIndexOf(instructions, "<")
		} else if instruction == "cmp_le" {
			i = helper.StringsIndexOf(instructions, "<=")
		}
		instruction = strings.ToUpper(instruction)
		if (i % 2) == 0 {
			instruction = fmt.Sprintf("!%s", instruction)
		}
		if (i % 3) == 0 {
			instruction = fmt.Sprintf("|%s", instruction)
		}
		query[fmt.Sprintf(":%s[F%0.3d]", instruction, i+1)] = []string{fmt.Sprintf("%0.3d", i+1), fmt.Sprintf("%0.3d", i+1)}
	}
	where, groupBy, orderBy, having, limit, offset := SQLParser(query)
	if where == nil || len(where) == 0 {
		t.Errorf("db[normal-instructions]: wrong instruction «%s»", "where")
	} else {
		for _, w := range where {
			if pos := helper.StringsIndexOf(instructions, w.Instruction()); pos < 0 {
				t.Errorf("db[normal-instructions]: not found instruction «%s» (Field: %s)", w.Instruction(), w.Field())
			} else if (pos%3) != 0 && w.Separator() != "AND" {
				t.Errorf("db[normal-instructions]: wrong separator «%s» (Separator: %s)", w.Instruction(), w.Separator())
			} else if (pos%3) == 0 && w.Separator() != "OR" {
				t.Errorf("db[normal-instructions]: wrong separator «%s» (Separator: %s)", w.Instruction(), w.Separator())
			} else if (pos%2) == 0 && w.Negative() != true {
				t.Errorf("db[normal-instructions]: wrong negative «%s» (Separator: %s)", w.Instruction(), w.Separator())
			} else if (pos%2) != 0 && w.Negative() != false {
				t.Errorf("db[normal-instructions]: wrong negative «%s» (Separator: %s)", w.Instruction(), w.Separator())
			} else if w.Field() != fmt.Sprintf("F%0.3d", pos+1) {
				t.Errorf("db[normal-instructions]: wrong field «%s» (Field: %s)", w.Instruction(), w.Field())
			} else
			// between value
			if values, ok := w.Value().([]interface{}); w.Instruction() == "between" && (!ok || values == nil || len(values) != 2 || values[0] != fmt.Sprintf("%0.3d", pos+1) || values[1] != fmt.Sprintf("%0.3d", pos+1)) {
				t.Errorf("db[normal-instructions]: wrong value «%s» (Value: %q)", w.Instruction(), w.Value())
			} else
			// in value
			if values, ok := w.Value().([]interface{}); w.Instruction() == "in" && (!ok || values == nil || len(values) != 2 || values[0] != fmt.Sprintf("%0.3d", pos+1) || values[1] != fmt.Sprintf("%0.3d", pos+1)) {
				t.Errorf("db[normal-instructions]: wrong value «%s» (Value: %q)", w.Instruction(), w.Value())
			} else
			// all value
			if w.Instruction() != "between" && w.Instruction() != "in" && w.Value() != fmt.Sprintf("%0.3d", pos+1) {
				t.Errorf("db[normal-instructions]: wrong value «%s» (Value: %q)", w.Instruction(), w.Value())
			}
		}
	}
	if groupBy == nil || len(groupBy) == 0 {
		t.Errorf("db[normal-instructions]: wrong instruction «%s»", "group-by")
	} else {
		for _, g := range groupBy {
			if pos := helper.StringsIndexOf(instructions, "group"); pos < 0 {
				t.Errorf("db[normal-instructions]: not found instruction «%s» (Field: %s)", "group", g.Field())
			} else if g.Field() != fmt.Sprintf("F%0.3d", pos+1) {
				t.Errorf("db[normal-instructions]: wrong field «%s» «%s»", "group", g.Field())
			}
		}
	}
	if having == nil || len(having) == 0 {
		t.Errorf("db[normal-instructions]: wrong instruction «%s»", "having")
	} else {
		for _, h := range having {
			if pos := helper.StringsIndexOf(instructions, "group"); pos < 0 {
				t.Errorf("db[normal-instructions]: not found instruction «%s» (Field: %s)", "group", h.Field())
			} else if h.Field() != fmt.Sprintf("F%0.3d", pos+1) {
				t.Errorf("db[normal-instructions]: wrong field «%s» «%s»", "group", h.Field())
			} else if h.Value() != fmt.Sprintf("%0.3d", pos+1) {
				t.Errorf("db[normal-instructions]: wrong having value «%s» (Value: %q)", "group", h.Value())
			}
		}
	}
	if orderBy == nil || len(orderBy) == 0 {
		t.Errorf("db[normal-instructions]: wrong instruction «%s»", "order-by")
	} else {
		for _, o := range orderBy {
			if pos := helper.StringsIndexOf(instructions, "sort"); pos < 0 {
				t.Errorf("db[normal-instructions]: not found instruction «%s» (Field: %s)", "sort", o.Field())
			} else if o.Field() != fmt.Sprintf("F%0.3d", pos+1) {
				t.Errorf("db[normal-instructions]: wrong field «%s» (Field: %s)", "sort", o.Field())
			} else if o.Sort() != "ASC" {
				t.Errorf("db[normal-instructions]: wrong sort value «%s» (Sort: %s)", "sort", o.Sort())
			}
		}
	}
	if pos := helper.StringsIndexOf(instructions, "limit"); pos < 0 || limit == nil {
		t.Errorf("db[normal-instructions]: wrong instruction «%s»", "limit")
	} else if limit.Count() != int64(pos+1) {
		t.Errorf("db[normal-instructions]:  value «%s» (Limit: %d)", "limit", limit.Count())
	}
	if pos := helper.StringsIndexOf(instructions, "offset"); pos < 0 || offset == nil {
		t.Errorf("db[normal-instructions]: wrong instruction «%s»", "offset")
	} else if offset.Rows() != int64(pos+1) {
		t.Errorf("db[normal-instructions]: wrong value «%s» (Offset: %d)", "offset", offset.Rows())
	}
	// wrong instructions
	query = url.Values{}
	query[fmt.Sprintf("1:!%s[ID%d][]", "between", 0)] = []string{"100"}
	query[fmt.Sprintf("2:!%s[ID%d][]", "between", 0)] = []string{"200"}
	where, _, _, _, _, _ = SQLParser(query)
	if where != nil && len(where) > 0 {
		names := make([]string, 0)
		for _, w := range where {
			names = append(names, w.Instruction())
		}
		t.Errorf("db[wrong-instructions]: wrong instruction «%s» (%s)", "where", strings.Join(names, ", "))
	}
	query = url.Values{}
	query[fmt.Sprintf("|%s", "login")] = []string{"root"}
	where, _, _, _, _, _ = SQLParser(query)
	if where == nil || len(where) != 1 {
		t.Errorf("db[wrong-instructions]: wrong instruction «%s»", "where")
	} else if w := where[0]; w.Field() != "login" {
		t.Errorf("db[wrong-instructions]: wrong field «%s», must be «login»", w.Field())
	} else if w.Value() != "root" {
		t.Errorf("db[wrong-instructions]: wrong value «%q», must be «root»", w.Value())
	} else if w.Separator() != "OR" {
		t.Errorf("db[wrong-instructions]: wrong separator «%s», must be «OR»", w.Separator())
	} else if len(w.Instruction()) != 0 {
		t.Errorf("db[wrong-instructions]: wrong instruction «%s», must be empty", w.Instruction())
	}
}

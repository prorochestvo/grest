package db

import (
	"fmt"
	"strings"
)

func SQLEscape(value interface{}) string {
	result := ""
	if v, ok := value.(string); ok {
		v = strings.ReplaceAll(v, "'", "''")
		result = fmt.Sprintf("'%s'", v)
	} else if v, ok := value.(int); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(int64); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(int32); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(int16); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(int8); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(uint); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(uint64); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(uint32); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(uint16); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(uint8); ok {
		result = fmt.Sprintf("%d", v)
	} else if v, ok := value.(float32); ok {
		result = fmt.Sprintf("%f", v)
	} else if v, ok := value.(float64); ok {
		result = fmt.Sprintf("%f", v)
	} else if v, ok := value.(bool); ok {
		if v == true {
			result = "TRUE"
		} else {
			result = "FALSE"
		}
	}
	return result
}

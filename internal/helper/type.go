package helper

import (
	"reflect"
	"strings"
)

func TypeName(i interface{}) string {
	if i == nil {
		return ""
	}
	result := strings.ReplaceAll(reflect.TypeOf(i).String(), "*", "")
	if pos := strings.Index(result, "."); pos > 0 {
		p := strings.ToUpper(result[0:pos])
		c := result[pos:]
		result = p + c
	}
	return result
}

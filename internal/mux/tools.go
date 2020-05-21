package mux

import (
	"fmt"
	"path"
)

func braceIndices(s string) ([]int, error) {
	result := make([]int, 0)
	level := 0
	idx := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				result = append(result, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("unbalanced braces in %s", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("unbalanced braces in %s", s)
	}
	return result, nil
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}

	return np
}

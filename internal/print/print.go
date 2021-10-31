package print

import (
	"fmt"
	"strings"
)

var rpls = map[string]string{
	"string":  "''",
	"object":  "{}",
	"array":   "[]",
	"boolean": "false",
	"number":  "0.0",
	"integer": "0",
}

// FormData prints formatted form values.
func FormData(params map[string]string) string {
	if params == nil {
		return ""
	}

	var fields []string
	for k, v := range params {
		fields = append(fields, fmt.Sprintf("%s=%s", k, rpls[v]))
	}
	return strings.Join(fields, "&")
}

// Query prints formatted query values.
func Query(params map[string]string) string {
	if params == nil {
		return ""
	}

	var fields []string
	for k, v := range params {
		fields = append(fields, fmt.Sprintf("%s=%s", k, rpls[v]))
	}
	return fmt.Sprintf("?%s", strings.Join(fields, "&"))
}

// Object prints formatted json body values.
func Object(params map[string]string) string {
	if params == nil {
		return ""
	}

	var fields []string
	for k, v := range params {
		fields = append(fields, fmt.Sprintf("'%s': %s", k, rpls[v]))
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

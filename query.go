package dbx

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func RewriteQuery(ctx context.Context, sql string, args []any) (string, []any, error) {

	if len(args) != 1 {
		return sql, args, nil
	}

	typeOf := reflect.TypeOf(args[0])
	switch {
	case typeOf.Kind() == reflect.Struct, typeOf.Kind() == reflect.Map:
	default:
		return sql, args, nil
	}

	var matcher = regexp.MustCompile(`:{1,2}[a-zA-Z_0-9]+`)

	var namedParams []string
	var newArgs []any

	sql = matcher.ReplaceAllStringFunc(sql, func(param string) string {
		if strings.HasPrefix(param, "::") {
			return param
		}
		trimmed := param[1:]
		for i, exitingParam := range namedParams {
			if trimmed == exitingParam {
				return "$" + strconv.Itoa(i+1)
			}
		}

		namedParams = append(namedParams, trimmed)
		return "$" + strconv.Itoa(len(namedParams))
	})

	value := reflect.ValueOf(args[0])
	for _, param := range namedParams {
		var val reflect.Value
		switch {
		case typeOf.Kind() == reflect.Struct:
			val = value.FieldByNameFunc(func(s string) bool {
				return s == param
			})
		case typeOf.Kind() == reflect.Map:
			iter := value.MapRange()
			for iter.Next() {
				if param == iter.Key().String() {
					val = iter.Value()
				}
			}
		}
		switch {
		case val.IsZero():
			return "", nil, fmt.Errorf("unable to find source for parameter '%v'", param)
		default:
			newArgs = append(newArgs, val.Interface())
		}
	}

	return sql, newArgs, nil
}

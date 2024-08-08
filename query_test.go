package dbx_test

import (
	"context"
	"dbx"
	"reflect"
	"testing"
)

type queryTestCase struct {
	name       string
	startQuery string
	endQuery   string
	startArgs  []any
	endArgs    []any
	shouldErr  bool
}

type simpleStruct struct {
	Qux int
}

type embededStruct struct {
	Qaz string
	simpleStruct
}

func TestQueryRewrite(t *testing.T) {
	ctx := context.TODO()

	for _, testcase := range []queryTestCase{
		{
			name:       "No arguments",
			startQuery: "SELECT * FROM foo.bar",
			endQuery:   "SELECT * FROM foo.bar",
		},
		{
			name:       "Simple numerical arguments",
			startQuery: "SELECT * FROM foo.bar where baz = $1",
			endQuery:   "SELECT * FROM foo.bar where baz = $1",
			startArgs:  []any{1},
			endArgs:    []any{1},
		},
		{
			name:       "Simple map",
			startQuery: "SELECT * FROM foo.bar where baz = :qux",
			endQuery:   "SELECT * FROM foo.bar where baz = $1",
			startArgs: []any{map[string]any{
				"qux": 1,
			}},
			endArgs: []any{1},
		},
		{
			name:       "Simple struct perfect name match",
			startQuery: "SELECT * FROM foo.bar where baz = :Qux",
			endQuery:   "SELECT * FROM foo.bar where baz = $1",
			startArgs:  []any{simpleStruct{Qux: 1}},
			endArgs:    []any{1},
		},
		{
			name:       "Embeded struct perfect name match",
			startQuery: "SELECT * FROM foo.bar where baz = :Qux AND bax = :Qaz",
			endQuery:   "SELECT * FROM foo.bar where baz = $1 AND bax = $2",
			startArgs:  []any{embededStruct{Qaz: "foo", simpleStruct: simpleStruct{Qux: 1}}},
			endArgs:    []any{1, "foo"},
		},
	} {
		endQuery, endArgs, err := dbx.RewriteQuery(ctx, testcase.startQuery, testcase.startArgs)
		if !testcase.shouldErr && err != nil {
			t.Fatalf("err encountered : %v\n%v", testcase.name, err)
		}
		if endQuery != testcase.endQuery {
			t.Fatalf("query missmatch: %v\n%v\n%v", testcase.name, endQuery, testcase.endQuery)
		}
		if !reflect.DeepEqual(endArgs, testcase.endArgs) {
			t.Fatalf("args missmatch: %v\n%v\n%v", testcase.name, endArgs, testcase.endArgs)
		}
	}
}

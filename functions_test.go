package xtemplate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoCache(t *testing.T) {

	got := NoCache("style.css")
	if !strings.HasPrefix(got, "style.css?t=") && len(got) < len("style.css?t=") {
		t.Fatal("want: style.css?t=nnnnnnn, got:", got)
	}

}

func TestIfEmpty(t *testing.T) {
	mapVal := map[string]interface{}{
		"name": "soo",
	}

	tests := []struct {
		name     string
		val      interface{}
		defa     interface{}
		expected interface{}
	}{
		{name: "string val", val: "", defa: "str", expected: "str"},
		{name: "int val", val: 0, defa: 1, expected: 1},
		{name: "map val", val: mapVal["name"], defa: "...", expected: "soo"},
		{name: "map non existent key", val: mapVal["foo"], defa: "...", expected: "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, IfEmpty(tt.val, tt.defa), tt.expected)
		})
	}
}

func TestIsEmpty(t *testing.T) {
	var (
		pstr *string
	)
	mapVal := map[string]interface{}{
		"name": "soo", "foo": "",
	}

	tests := []struct {
		name  string
		val   interface{}
		empty bool
	}{
		{name: "empty string", val: "", empty: true},
		{name: "string val", val: "ab", empty: false},
		{name: "empty *string", val: pstr, empty: true},
		{name: "empty int", val: 0, empty: true},
		{name: "int val", val: -1, empty: false},
		{name: "map str val", val: mapVal["name"], empty: false},
		{name: "map str empty", val: mapVal["foo"], empty: true},
		{name: "map nonexistent key", val: mapVal["bar"], empty: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, IsEmpty(tt.val), tt.empty)
		})
	}
}

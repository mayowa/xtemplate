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

func Test_getDefault(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
		defa interface{}
	}{
		{name: "string val", val: "", defa: "str"},
		{name: "int val", val: 0, defa: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, getDefault(tt.val, tt.defa), tt.defa)
		})
	}
}

func TestIsEmpty(t *testing.T) {
	var (
		pstr *string
	)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, IsEmpty(tt.val), tt.empty)
		})
	}
}

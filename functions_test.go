package xtemplate

import (
	"strings"
	"testing"
)

func TestNoCache(t *testing.T) {

	got := NoCache("style.css")
	if !strings.HasPrefix(got, "style.css?t=") && len(got) < len("style.css?t=") {
		t.Fatal("want: style.css?t=nnnnnnn, got:", got)
	}

}

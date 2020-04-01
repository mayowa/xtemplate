package xtemplate

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreCache(t *testing.T) {

	xt := New("./samples")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "plain.html", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "dinma is 18\n", buff.String())
	assert.Nil(t, xt.Lookup("plain.html"))
}

func TestCache(t *testing.T) {

	xt := New("./samples")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	assert.Nil(t, xt.ParseFile("plain.html"))

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "plain.html", data, false); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "dinma is 18\n", buff.String())
	assert.NotNil(t, xt.Lookup("plain.html"))
}

func TestMasterChild(t *testing.T) {

	xt := New("./samples")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "overlay.html", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "master with overlay body\n\n", buff.String())
}

func TestIncludes(t *testing.T) {

	xt := New("./samples")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "overlay2.html", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(
		t,
		"master \nwith overlay body<br>\n\na button submit\n\n\n\na footer\n",
		buff.String(),
	)
}

func TestSubFolderFile(t *testing.T) {

	xt := New("./samples")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "sub/plain.html", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "dinma is 18\n", buff.String())
	assert.Nil(t, xt.Lookup("plain.html"))
}

func TestFunctionSyntax(t *testing.T) {

	xt := New("./samples")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "functions.html", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "dinma is 18\nmy name is dinma\n\nmy name is dinma", buff.String())
}

func TestFunctions(t *testing.T) {

	xt := New("./samples")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}
	tpl := `hello world`
	retv, err := xt.RenderString(tpl, data)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "hello world", retv)

	tpl = `{{ upper("hello world") }}`
	retv, err = xt.RenderString(tpl, data)
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "HELLO WORLD", retv)
}

func TestRenderString(t *testing.T) {

	xt := New("./samples")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}
	tpl := `
	{{extends "master.html"}}
	{{include "extras.html"}}

	{{define "body"}}
	with overlay body<br>
	{{ template button("submit") }}
	{{end}}
	`
	retv, err := xt.RenderString(tpl, data)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "master \n\twith overlay body<br>\n\t\na button submit\n\n\t\n\na footer\n", retv)

}

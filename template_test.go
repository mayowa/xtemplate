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
		"master \nwith overlay body<br>\n\na BUTTON submit2\n\n\n\n* A\n\n* B\n\n\n\nHello mayowa, age: 18\n\n\n<div  class=\"red\">\n    <p>Hello</p>\n    <p>World</p>\n</div>\n\na footer\n",
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

	assert.Equal(t, "dinma is 18\nmy name is dinma\n\nmy name is dinma\n\ndinma is my name", buff.String())
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
	assert.Equal(t, "master \n\twith overlay body<br>\n\t\na BUTTON submit\n\n\t\n\na footer\n", retv)

}

func TestTranslateTags(t *testing.T) {

	src := []byte(`
	<tag type="input" class="red sm:red" x-data="{'a':1}"></tag>
	`)
	retv := string(translateTags(src))

	exp := "\n\t\n\t\t{{ $jputeHFozdHtml := `` }}\n\t\t{{ $jputeHFozdAttr := kwargs [\"class\" `red sm:red` \"x-data\" `{'a':1}`] }}\n\t\ttag \"input\" $jputeHFozdAttr $jputeHFozdHtml\n\t\t\n\t"
	assert.Equal(t, exp, retv)
	// assert.Greater(t, len(retv), 5)

}

func TestTagsOnly(t *testing.T) {

	xt := New("./samples")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}
	tpl := `
	<tag type="text-field" class="red sm:red" x-data="{'a':1}"></tag>
	`
	retv, err := xt.RenderString(tpl, data)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "\n\t\n\t\t\n\t\t\n\t\t<text-field  class=\"red sm:red\" x-data=\"{'a':1}\"></text-field>\n\t\t\n\t", retv)

}

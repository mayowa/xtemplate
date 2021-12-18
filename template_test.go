package xtemplate

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreCache(t *testing.T) {

	xt := New("./samples", "html")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "plain", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "dinma is 18\n", buff.String())
	assert.Nil(t, xt.Lookup("plain"))
}

func TestCache(t *testing.T) {

	xt := New("./samples", "html")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	assert.NoError(t, xt.ParseFile("plain"))

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "plain.html", data, false); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "dinma is 18\n", buff.String())
	assert.NotNil(t, xt.Lookup("plain"))
}

func TestMasterChild(t *testing.T) {

	xt := New("./samples", "html")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "overlay", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "The Base\n\n\n** Master\n\n    original master body\n    with overlay\n\n\n===\n", buff.String())
}

func TestIncludes(t *testing.T) {

	xt := New("./samples", "html")
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
		"The Base\n\n\n** Master\n\nwith overlay body<br>\n\na BUTTON submit2\n\n\n* A\n\n* B\n\n\nHello mayowa, age: 18\n<div class=\"red\">\n    <p>Hello</p>\n    <p>World</p>\n</div>\n\n\n\n\na footer\n\n",
		buff.String(),
	)
}

func TestSubFolderFile(t *testing.T) {

	xt := New("./samples", "html")
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

	xt := New("./samples", "html")
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

	xt := New("./samples", "html")
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

	xt := New("./samples", "html")
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
	assert.Equal(t, "The Base\n\n\n** Master\n\n\twith overlay body<br>\n\t\na BUTTON submit\n\n\n\na footer\n\n", retv)

}

func TestTranslateTags(t *testing.T) {

	xt := New("", "html")
	src := []byte(`
	<tag type="input" class="red sm:red"></tag>
	<tag type="input" x-data="{'a':1}"></tag>

	<tag type="p">{{.Name}}</tag>
	<tag type="div">
		<tag type="input" value="abc"></tag>
	</tag>
	`)
	retv := string(translateTags(xt, src, 1))

	exp := "\n\t<input class=\"red sm:red\"></input>\n\t<input x-data=\"{'a':1}\"></input>\n\n\t<p >{{.Name}}</p>\n\t<div >\n\t\t<input value=\"abc\"></input>\n\t</div>\n\t"
	assert.Equal(t, exp, retv)
	// assert.Greater(t, len(retv), 5)

}

func TestTagsOnly(t *testing.T) {

	xt := New("./samples", "html")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	tpl := `
	<tag type="text-field" class="red sm:red" ></tag>
	<tag type="text-field" x-data="{'b':2}">
		<p>{{.name}}</p>
		<p>{{.age}}</p>
	</tag>
	<tag type="div"><tag type="p" >{{.age}}</tag></tag>
	<tag type="field">
		<tag type="select" value=1>
			{{range .options}}
			<option value="{{.id}}">{{.label}}</option>
			{{end}}
		</tag>
	</tag>
	`

	// tpl := `
	// <tag type="field">
	// 	<tag type="select" value=1>
	// 		{{range .options}}
	// 		<option value="{{.id}}">{{.label}}</option>
	// 		{{end}}
	// 	</tag>
	// </tag>
	// `
	retv, err := xt.RenderString(tpl, data)
	if err != nil {
		t.Error(err)
		return
	}
	exp := "\n\t<text-field class=\"red sm:red\"></text-field>\n\t<text-field x-data=\"{'b':2}\">\n\t\t<p>dinma</p>\n\t\t<p>18</p>\n\t</text-field>\n\t<div ><p >18</p></div>\n\t<field >\n\t\t<select value=\"1\">\n\t\t\t\n\t\t</select>\n\t</field>\n\t"
	assert.Equal(t, exp, retv)

}

func TestExtractedTemplates(t *testing.T) {

	xt := New("./samples", "html")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "tplAction.html", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(
		t,
		"\n\n\n\n\n\n\n\na BUTTON submit2\n\n\n* A\n\n* B\n\n\nHello mayowa, age: 18\nIm just text",
		buff.String(),
	)
}

func TestPartials(t *testing.T) {

	xt := New("./samples", "html")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "partials", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "\n<div>label: dinma</div>\n\n\n<div>box: 18</div>\n\n", buff.String())
	assert.Nil(t, xt.Lookup("plain.html"))
}

func TestTemplateInclude(t *testing.T) {

	xt := New("./samples", "html")
	data := map[string]interface{}{
		"name": "dinma", "age": 18,
	}

	buff := bytes.NewBufferString("")
	if err := xt.Render(buff, "overlay3", data, true); err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "The Base\n\n\n<h1>The body</h1>\n<header>\n\t<h2>To be included</h2>\n\t<input type=\"text\" value=\"search\">\n</header>\n\n", buff.String())
	assert.Nil(t, xt.Lookup("plain.html"))
}

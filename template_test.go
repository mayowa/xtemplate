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

package xtemplate

import (
	"testing"
)

const tplSource = `
﹟Event Shorthand
Stimulus lets you shorten the action descriptors for some common element/event pairs, 
such as the button/click pair above, by omitting the event name:

<button data-action="gallery#next">…</button>

{{define "component--card"}}
	Hello World
	{{block "slot--header" .}}Title{{end}}

	{{block "slot--body" .}}
		Chidinma is a fine girl!
		{{define "component--article"}}
			{{block "slot--body" .}}
				{{range $i := .Names}}
			In the beginning was the word, and the {{.}} was with God and the word was God!
				{{end}}
			{{end}}
		{{end}}
	{{end}}
{{end}}

`

func TestLexicon_ParseActions(t *testing.T) {
	lex := Lexicon{src: []byte(tplSource)}
	lex.ParseActions()
	if len(lex.actions) == 0 {
		t.Fatalf("parseAction failed: nothing was parsed")
	}

	expected := []ActionItem{
		{name: "define", tokenType: ActionTypeBlockStart, parameters: "component--card"},
		{name: "block", tokenType: ActionTypeBlockStart, parameters: "slot--header"},
		{name: "end", tokenType: ActionTypeBlockEnd},
		{name: "block", tokenType: ActionTypeBlockStart, parameters: "slot--body"},
		{name: "define", tokenType: ActionTypeBlockStart, parameters: "component--article"},
		{name: "block", tokenType: ActionTypeBlockStart, parameters: "slot--body"},
		{name: "range", tokenType: ActionTypeBlockStart, parameters: "$i := .Names"},
		{name: "end", tokenType: ActionTypeBlockEnd},
		{name: "end", tokenType: ActionTypeBlockEnd},
		{name: "end", tokenType: ActionTypeBlockEnd},
		{name: "end", tokenType: ActionTypeBlockEnd},
		{name: "end", tokenType: ActionTypeBlockEnd},
	}

	for i, a := range lex.actions {
		t.Log(i, ":", a.name, a.tokenType.String(), a.parameters)

		if a.name != expected[i].name && a.Type() != expected[i].Type() {
			t.Errorf("parse error - expecting:%s, got:%s", expected[i].name, a.name)
		}
	}

	// t.Fatal("stop")
}

func TestLexicon_ParseComponents(t *testing.T) {
	lex := Lexicon{src: []byte(tplSource)}
	lex.ParseComponents()
	if len(lex.components) == 0 {
		t.Fatalf("parseComponents failed: nothing was parsed")
	}

	expected := []Component{
		{Block: &Block{name: "define", parameters: "\"component--card\""}, isChild: false, children: []*Component{&Component{}, &Component{}}},
		{Block: &Block{name: "define", parameters: "\"component--article\""}, isChild: true},
	}

	for i, c := range lex.components {
		t.Log(i, ":", c.id, c.name, c.parameters, c.isChild, len(c.children))

		if c.name != expected[i].name && c.parameters != expected[i].parameters && len(c.children) != len(expected[i].children) {
			t.Errorf("parse error - expecting:%s, got:%s", expected[i].name, c.name)
		}
	}

	// t.Fatal("stop")
}

func TestBlock_UpdateParameters(t *testing.T) {
	src := []byte(`
	{{define "component--card"}}
		a card
	{{end}}
	`)

	lex := Lexicon{src: src}
	lex.ParseComponents()
	b := lex.blocks[0]

	b.UpdateParameters("\"component__1123__card\"")
	if string(b.src[b.localStartPos():b.localContentStart()]) != "{{define \"component__1123__card\"}}" {
		t.Fatal("header:", string(b.src[b.localStartPos():b.localContentStart()]))
	}

	newContent := "\n\tsome cards\n\t"
	b.UpdateContent([]byte(newContent))
	if string(b.content) != newContent {
		t.Fatal("content:", string(b.content))
	}
}

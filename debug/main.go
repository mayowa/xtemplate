package main

import (
	"fmt"
	"github.com/mayowa/xtemplate"
)

const tplSource = `
﹟Event Shorthand
Stimulus lets you shorten the action descriptors for some common element/event pairs, 
such as the button/click pair above, by omitting the event name:

<button data-action="gallery#next">…</button>

{{#component "card" id}}
	Hello World
	{{#slot "header" card}}Title{{end}}

	{{#slot "body" card}}
		Chidinma is a fine girl!
		{{#component "article"}}
			{{#slot "body" article}}
			In the beginning was the word, and the Word was with God and the word was God!
			{{end}}
		{{end}}
	{{end}}
{{end}}
`

/*
{{#component "table" id}}

{{#slot "tab1" table}}
tab 1
{{end}}

{{#slot "tab2" table}}
tab 2
{{end}}
{{end}}*/

func main() {
	println(tplSource)
	err := xtemplate.Transform([]byte(tplSource))
	if err != nil {
		fmt.Println(err)
		return
	}
}

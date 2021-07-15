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

{{#component 123 "card"}}
	Hello World
	{{#slot "header"}}Title{{end}}

	{{#slot "body"}}
	Chidinma is a fine girl!
	{{end}}
{{end}}

<table>
<thead>
<tr>
<th>Element</th>
<th>Default Event</th>
</tr>
</thead>
<tbody>
<tr>
<td>a</td>
<td>click</td>
</tr>
<tr>
<td>button</td>
<td>click</td>
</tr>
<tr>
<td>form</td>
<td>submit</td>
</tr>
<tr>
<td>input</td>
<td>input</td>
</tr>
<tr>
<td>input type=submit</td>
<td>click</td>
</tr>
<tr>
<td>select</td>
<td>change</td>
</tr>
<tr>
<td>textarea</td>
<td>input</td>
</tr>
</tbody>
</table>
`
func main() {
	err := xtemplate.Transform([]byte(tplSource))
	if err != nil {
		fmt.Println(err)
		return 
	}
}

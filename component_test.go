package xtemplate

import (
	"strings"
	"testing"

	"github.com/andreyvit/diff"
)

func TestTransformer_transformComponentBlock(t *testing.T) {

	tests := []struct {
		name     string
		source   []byte
		expected []byte
	}{
		{
			name: "test 1",
			source: []byte(`
			{{define "component--card"}}
				{{block "slot--header" .}}Title{{end}}
			{{end}}			
			`),
			expected: []byte(`{{define "component__1__card" }}
  <div class="card">
    <div class="card__header">
    {{block "card__1__header" .}}Title{{end}}
    </div>
    <div class="card_body">
    {{block "card__1__body" .}}
      a body
    {{end}}
    </div>
  </div> 
{{end}}
{{template "component__1__card" .}}
`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Document(tt.source)
			trf := NewTransformer(&doc, "./samples", "tmpl")

			c := trf.components[0]
			cName := getComponentName(c)
			cTpl, err := trf.getComponentTpl(cName)
			if err != nil {
				t.Fatalf("component %s not found", cName)
			}

			got := trf.transformComponentBlock(cName, c, cTpl, false)

			sg := strings.TrimSpace(string(got))
			se := strings.TrimSpace(string(tt.expected))
			if sg != se {
				t.Errorf("wrong output:\n%s", diff.LineDiff(sg, se))
			}
		})
	}

}

func TestTransformer_transformSubcomponentCall(t *testing.T) {

	doc := Document(tplSource)
	trf := NewTransformer(&doc, "./samples", "tmpl")

	t.Run("test 1", func(t *testing.T) {
		c := trf.components[0]

		slot := c.children[0].Clone()

		if got := trf.transformSubcomponentCall(slot); got != nil {
			t.Errorf("\n%s", string(got))
		}
	})

}

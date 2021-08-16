package xtemplate

import (
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
			expected: []byte(`
			{{define "component__1__card" }}
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

		{
			name: "test 2",
			source: []byte(`
			{{define "component--card"}}
				{{block "slot--body" .}}
					{{define "component--article"}}
						{{block "slot--body" .}}
							lorem ipsum
						{{end}}
					{{end}}
				{{end}}
			{{end}}			
			`),

			expected: []byte(`
      {{define "component__1__card" }}
        <div class="card">
          <div class="card__header">
          {{block "card__1__header" .}}
            a header
          {{end}}
          </div>
          <div class="card_body">
          {{block "card__1__body" .}}
            {{block "component__3__article" . }}
            <div class="article">
              {{block "article__3__body" .}}
                lorem ipsum
              {{end}}
            </div>
            {{end}}
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

			sg := diff.TrimLinesInString(string(got))
			se := diff.TrimLinesInString(string(tt.expected))
			if sg != se {
				t.Errorf("wrong output:\n%s\n\n%s", string(got), string(tt.expected))
			}
		})
	}

}

func TestTransformer_transformSubcomponentCall(t *testing.T) {

	doc := Document(tplSource)
	trf := NewTransformer(&doc, "./samples", "tmpl")

	expected := `
  {{block "slot--body" .}}
      Chidinma is a fine girl!
    {{block "component__3__article" . }}
    <div class="article">
      {{block "article__3__body" .}}
        {{range $i := .Names}}
        In the beginning was the word, and the {{.}} was with God and the word was God!
        {{end}}
      {{end}}
    </div>
    {{end}}
  {{end}}	
	`

	t.Run("test 1", func(t *testing.T) {
		c := trf.components[0]

		slot := c.children[0].Clone()

		got := trf.transformSubcomponentCall(slot)

		if diff.TrimLinesInString(string(got)) != diff.TrimLinesInString(expected) {
			t.Errorf("got:\n%s\n\nwant\n%s", string(got), expected)
		}
	})

}

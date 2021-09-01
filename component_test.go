package xtemplate

import (
	"testing"

	"github.com/andreyvit/diff"
)

func Test_translateComponents(t *testing.T) {

	tests := []struct {
		name string
		src  string
		out  string
	}{
		{
			name: "test 1",
			src: `
			<div>
				<component type="card">
					<slot name="body">
						<component type="article">
							<slot name="body">
							article
							</slot>
						</component>
					</slot>
				</component>
			</div>			
			`,
			out: `        
       			<div>
        				{{- block "component__card__1" . -}}
        {{- $props := (kwargs "_blank" "") -}}
        <div class="card">
        	<div class="card__header">
        	{{block "card__1__header" .}}
        		a header
        	{{end}}
        	</div>
        	<div class="card__body">
        	{{block "card__1__body" .}}
        {{- $props := (kwargs "_blank" "") -}}
        
        						{{- block "component__article__2" . -}}
        {{- $props := (kwargs "_blank" "") -}}
        <div class="article">
        {{block "article__2__body" .}}
        {{- $props := (kwargs "_blank" "") -}}
        
        							article
        							{{end}}
        </div>
        
        {{end -}}
        					{{end}}
        	</div>
        </div> 
        
        {{end -}}
        			</div>	
		`,
		},
		{name: "test 2",
			src: `
			<component type="card">
				<slot name="body">
					<aside>
						a side menu
					</aside>
					<component type="article2">
						<slot name="body"> lorem ipsum </slot>
					</component>
				</slot>
			</component>
		`,
			out: `
        			{{- block "component__card__1" . -}}
        {{- $props := (kwargs "_blank" "") -}}
        <div class="card">
        	<div class="card__header">
        	{{block "card__1__header" .}}
        		a header
        	{{end}}
        	</div>
        	<div class="card__body">
        	{{block "card__1__body" .}}
        {{- $props := (kwargs "_blank" "") -}}
        
        					<aside>
        						a side menu
        					</aside>
        					{{- block "component__article2__2" . -}}
        {{- $props := (kwargs "_blank" "") -}}
        {{- block "component__card__3" . -}}
        {{- $props := (kwargs "_blank" "") -}}
        <div class="card">
        	<div class="card__header">
        	{{block "card__3__header" .}}
        		a header
        	{{end}}
        	</div>
        	<div class="card__body">
        	{{block "card__3__body" .}}
        {{- $props := (kwargs "_blank" "") -}}
        
        			{{block "article2__2__body" .}}
        {{- $props := (kwargs "_blank" "") -}}
         lorem ipsum {{end}}
        		{{end}}
        	</div>
        </div> 
        
        {{end -}}
        
        {{end -}}
        				{{end}}
        	</div>
        </div> 
        
        {{end -}}
		`,
		},
		{
			name: "test default",
			src: `
			<component type="box">
				whats in a box?
			</component>
			`,
			out: `
        {{- block "component__box__1" . -}}
        {{- $props := (kwargs "_blank" "") -}}
        <div class="box">
        	{{block "box__1__default" .}}
        {{- $props := (kwargs "_blank" "") -}}
        
        				whats in a box?
        			{{end}}
        </div>
        
        {{end -}}
			`,
		},
		{
			name: "test html attributes",
			src: `
			<component type="box" class="red">
				whats in a box?
			</component>
			`,
			out: `
      {{- block "component__box__1" . -}}
        {{- $props := (kwargs "class" "red") -}}
        <div class="box">
        	{{block "box__1__default" .}}
        {{- $props := (kwargs "class" "red") -}}
        
        				whats in a box?
        			{{end}}
        </div>
        
        {{end -}}
			`,
		},
	}

	xt := New("./samples", "html")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := translateComponents(xt, Document(tt.src))
			if err != nil {
				t.Fatal(err)
			}

			if diff.TrimLinesInString(tt.out) != diff.TrimLinesInString(string(src)) {
				t.Errorf("got:\n%s, \nwant:\n%s", string(src), tt.out)
			}
		})
	}

}

func Test_renderComponents(t *testing.T) {

	tests := []struct {
		name string
		src  string
		out  string
	}{
		{
			name: "test 1",
			src: `
			<component type="box" class="red">
				whats in a box {{$props.class}} {{.name}}?
			</component>
			`,
			out: `
        <div class="box">
        	whats in a box red bread?
        			
        </div>			
			`,
		},
	}

	xt := New("./samples", "html")
	data := map[string]interface{}{"name": "bread"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			src, err := xt.RenderString(tt.src, data)
			if err != nil {
				t.Fatal(err)
			}

			if diff.TrimLinesInString(tt.out) != diff.TrimLinesInString(src) {
				t.Errorf("got:\n%s, \nwant:\n%s", src, tt.out)
			}
		})
	}

}

func Test_listComponents(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected int
		wantErr  bool
	}{
		{
			name: "test 1",
			src: `
			<component type="box">
				whats in a box?
				<component type="article">
				article
				</component>
			</component>			
			
			<component type="box">box</component>
			`,
			expected: 3,
		},
		{
			name: "test mismatched",
			src: `
			<component type="box">
				whats in a box?
				<component type="article">
				article
			</component>			
			
			<component type="box">box</component>
			`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retv, err := listComponents([]byte(tt.src))
			if err != nil && !tt.wantErr {
				t.Fatal(err)
			} else if tt.wantErr && err == nil {
				t.Fatal("expected err got nil")
			}

			if tt.wantErr {
				return
			}

			if len(retv) != tt.expected {
				t.Fatalf("expected %d components got %d", tt.expected, len(retv))
			}
		})

	}
}

func Test_popAction(t *testing.T) {
	as := []Action{
		{ID: "a"}, {ID: "b"}, {ID: "c"},
	}

	popAction(&as, "a")
	popAction(&as, "a")
	popAction(&as, "b")
	popAction(&as, "b")
	popAction(&as, "c")
	popAction(&as, "c")

	if len(as) != 0 {
		t.Fatal("list should be empty")
	}
}

func Test_listActionSlots(t *testing.T) {
	var source = []byte(`
<div class="card">
	<div class="card__header">
	{{block "#slot--header" .}}
		a header
	{{end}}
	</div>
	<div class="card__body">
	{{block "#slot--body" .}}
		a body
	{{end}}
	</div>
</div> 
`)

	retv, err := listActionSlots(source)
	if err != nil {
		t.Fatal(err)
	}

	if len(retv) != 2 {
		t.Fatal("actions not found")
	}

	if retv[0].ID != "#slot--header" && retv[1].Name != "#slot--body" {
		t.Error("wrong actions returned")
		for _, s := range retv {
			t.Error("action :", s.ID)
		}
	}
}

func Test_listComponentSlots(t *testing.T) {
	var source = []byte(`
<div>
	<component type="card">
		<slot name="header">
			a header
		</slot>
		<slot name="body">
			<component type="article">
				<slot name="aBody">inner</slot>
			</component>
		</slot>
	</component>
</div>
`)

	retv, err := listComponentSlots(source, "card")
	if err != nil {
		t.Fatal(err)
	}

	if len(retv) == 0 {
		t.Fatal("components not detected in source")
	}

	if retv[0].Name != "header" && retv[0].Name != "body" {
		t.Error("wrong slots returned")
		for _, s := range retv {
			t.Error("slot :", s.Name)
		}
	}
}

func Test_findTag(t *testing.T) {
	var source = []byte(`
<div>
	<component id="card">
		<p>
		<slot Name="body">
			<component id="article">inner</component>
		</slot>
	</component>
</div>
`)

	tests := []struct {
		name    string
		tag     string
		id      string
		tname   string
		wantErr bool
	}{
		{name: "find component", tag: "component", id: "card"},
		{name: "find slot", tag: "slot", tname: "body"},
		{name: "find p", tag: "p", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := findTag(source, tt.tag)
			if tt.wantErr && err == nil {
				t.Fatalf("want error but got %s", c.Element)
			}

			if c == nil && err == nil {
				t.Fatalf("nil returned expecting %s", tt.tag)
			}

			if c != nil && tt.id != c.ID && tt.tname != c.Name {
				t.Fatalf("expected: id=%s, name=%s, got:i d=%s, name=%s",
					tt.id, tt.tname, c.ID, c.Name)
			}

		})
	}
}

func Test_findAction(t *testing.T) {
	var source = []byte(`
{{block "card" .}}
<div class="w-100 bg-reg-800">
	{{define "card__body"}}
	Hello World
	{{template "foo" . }}
	{{block "inner" .}}inner{{end}}
	{{end}}
</div>
{{end}}
`)

	tests := []struct {
		name    string
		tag     string
		id      string
		wantErr bool
	}{
		{name: "find block", tag: "block", id: "card"},
		{name: "find define", tag: "define", id: "card__body"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := findAction(source, tt.tag, tt.id)
			if tt.wantErr && err == nil {
				t.Fatalf("want error but got %s", c.Name)
			}

			if c == nil && err == nil {
				t.Fatalf("nil returned expecting %s", tt.tag)
			}

			if c != nil && tt.id != c.ID {
				t.Fatalf("expected: id=%s, got:id=%s",
					tt.id, c.ID)
			}

		})
	}
}

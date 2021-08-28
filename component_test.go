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
				<component id="card">
					<slot name="body">
						<component id="article">
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
        				{{block "card_1" .}}
        <div class="card">
        	<div class="card__header">
        	{{block "card__1__header" .}}
        		a header
        	{{end}}
        	</div>
        	<div class="card__body">
        	{{block "card__1__body" .}}
        						{{block "article_2" .}}
        <div class="article">
        {{block "article__2__body" .}}
        							article
        							{{end}}
        </div>
        
        {{end}}
        					{{end}}
        	</div>
        </div> 
        
        {{end}}
        			</div>	
		`,
		},
		{name: "test 2",
			src: `
			<component id="card">
				<slot name="body">
					<aside>
						a side menu
					</aside>
					<component id="article2">
						<slot name="body"> lorem ipsum </slot>
					</component>
				</slot>
			</component>
		`,
			out: `
        			{{block "card_1" .}}
        <div class="card">
        	<div class="card__header">
        	{{block "card__1__header" .}}
        		a header
        	{{end}}
        	</div>
        	<div class="card__body">
        	{{block "card__1__body" .}}
        					<aside>
        						a side menu
        					</aside>
        					{{block "article2_2" .}}
        {{block "card_3" .}}
        <div class="card">
        	<div class="card__header">
        	{{block "card__3__header" .}}
        		a header
        	{{end}}
        	</div>
        	<div class="card__body">
        	{{block "card__3__body" .}}
        	{{block "article2__2__body" .}} lorem ipsum {{end}}
        	{{end}}
        	</div>
        </div> 
        
        {{end}}
        
        {{end}}
        				{{end}}
        	</div>
        </div> 
        
        {{end}}
		
		`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := Document(tt.src)
			translateComponents(&src, "./samples")

			if diff.TrimLinesInString(tt.out) != diff.TrimLinesInString(string(src)) {
				t.Errorf("got:\n%s, \nwant:\n%s", string(src), tt.out)
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
	<component id="card">
		<slot name="header">
			a header
		</slot>
		<slot name="body">
			<component id="article">
				<slot name="abody">inner</slot>
			</component>
		</slot>
	</component>
</div>
`)

	retv, err := listComponentSlots(source, "card")
	if err != nil {
		t.Fatal(err)
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

var tTpl = `
<div>
	<component id="card">
		<p>
		<slot Name="body">
			<component id="article">inner</component>
		</slot>
	</component>
</div>
`

var cTpl = `
<div class="w-100 bg-reg-800">
	{{block "body" . }}
	Hello World
	{{end}}
</div>
`

/*
{{block "card" .}}
<div class="w-100 bg-reg-800">
	{{block "card__body" .}}
	Hello World
	{{end}}
</div>
{{end}}
*/

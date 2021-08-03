package xtemplate

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var cmpBlockStart = regexp.MustCompile(`{{-*\s*([#\w]+)\s+"([\w.]*?)"[\s\w".$]*?\s*-*}}`)

type ComponentTemplate struct {
	Name string
	Lex  Lexicon
}

var components map[string]*ComponentTemplate

func init() {
	components = make(map[string]*ComponentTemplate)
}

type Transformer struct {
	componentTemplates map[string]*ComponentTemplate
	components         []*Component
	componentFolder    string
	fileExt            string
	lex                *Lexicon
}

func NewTransformer(content *Document, templateFolder, ext string) *Transformer {
	t := new(Transformer)
	t.componentTemplates = make(map[string]*ComponentTemplate)
	t.componentFolder = filepath.Join(templateFolder, "_components")
	t.fileExt = ext

	lex := Lexicon{src: *content}
	lex.ParseComponents()
	t.lex = &lex
	t.components = lex.components

	return t
}

func (t *Transformer) TransformComponents(content *Document, componentFolder, ext string) error {
	if t.componentTemplates == nil {
		t.componentTemplates = make(map[string]*ComponentTemplate)
	}

	t.componentFolder = filepath.Join(componentFolder, "_components")
	t.fileExt = ext

	for _, c := range t.components {
		cName := getComponentName(c)
		if cName == "" {
			// component has no parameters (its malformed), cut it out
			content.Cut(c.startPos, c.endPos)
			continue
		}
		cTpl, err := t.getComponentTpl(cName)
		if err != nil {
			// component template file not found or some other error, cut!
			content.Cut(c.startPos, c.endPos)
			continue
		}

		buff := bytes.NewBuffer([]byte{})
		out := t.transformComponentBlock(cName, c, cTpl, false)
		buff.Write(out)

		fmt.Println(c.String(), "\n", "== debug ==")
		fmt.Println(buff.String())
		fmt.Println("\n", "== end ==")
	}

	return nil
}

func (t *Transformer) transformComponentBlock(cName string, c *Component, cTpl *ComponentTemplate, isChild bool) []byte {
	// make a copy of the template component
	tc := cTpl.Lex.components[0].Clone()
	// make a copy of the component. (Do NOT modify the source component c *Component)
	wc := c.Clone()

	// update component name
	if !isChild {
		name := fmt.Sprintf(`"component__%d__%s"`, wc.id, cName)
		tc.UpdateParameters(name)
	} else {
		// for child components: define --> block
		name := fmt.Sprintf(`"component__%d__%s" .`, wc.id, cName)
		tc.UpdateParameters(name)
		tc.Replace([]byte("define"), []byte("block"))
	}

	for _, slot := range tc.children {
		if slot.blockType != SlotAction {
			continue
		}

		// rename child slot
		slotName := getComponentName(slot)
		param := fmt.Sprintf(`"%s__%d__%s" .`, cName, c.id, slotName)

		tc.Replace([]byte(slot.parameters), []byte(param))
		slot.UpdateParameters(param)

		// update slot content
		cSlot, exists := c.childMap[slotName]
		if exists {
			oldSrc := slot.src[:]
			slot.UpdateContent(cSlot.content)
			tc.Replace(oldSrc, slot.src)
		}
		oldSrc := slot.src[:]
		newSrc := t.transformSubcomponentCall(slot)
		if newSrc != nil {
			tc.Replace(oldSrc, newSrc)
		}
	}

	retv := tc.src
	if !isChild {
		tplCall := fmt.Sprintf("\n{{template \"component__%d__%s\" .}}\n", wc.id, cName)
		retv = append(retv, []byte(tplCall)...)
	}
	return retv
}

func (t *Transformer) transformSubcomponentCall(c *Component) []byte {
	lex := Lexicon{src: c.content}
	lex.ParseComponents()
	if len(lex.components) == 0 {
		return nil
	}

	wc := c.Clone()

	for i := 0; i < len(lex.components); i++ {
		sc := lex.components[i]
		cc := findComponentByParameter(sc.parameters, t.components)
		if cc == nil {
			continue
		}

		cName := getComponentName(cc)
		cTpl, err := t.getComponentTpl(cName)
		if err != nil {
			return nil
		}

		retv := t.transformComponentBlock(cName, cc, cTpl, true)
		if retv == nil {
			continue
		}

		oldSrc := cc.src
		cc.Replace(cc.src, retv)
		wc.Replace(oldSrc, cc.src)
	}

	return wc.src
}

func getComponentName(c *Component) string {
	p := strings.Split(strings.TrimSpace(c.parameters), " ")
	if len(p) == 0 {
		return ""
	}

	retv := stripQuotes(p[0])
	p = strings.Split(retv, "--")
	if len(p) < 2 {
		return ""
	}

	return p[1]
}

func (t *Transformer) getComponentTpl(cName string) (*ComponentTemplate, error) {
	ct, inCache := t.componentTemplates[cName]
	if !inCache {
		cnFile := filepath.Join(t.componentFolder, cName+"."+t.fileExt)
		content, err := ioutil.ReadFile(cnFile)
		if err != nil {
			// Todo: log error
			return nil, err
		}

		ct = &ComponentTemplate{Name: cName}
		ct.Lex.src = content
		ct.Lex.Parse()

		t.componentTemplates[cName] = ct
	}

	return ct, nil
}

func stripQuotes(v string) string {
	return strings.Trim(v, `"`)
}

type Document []byte

func (d *Document) Cut(start, end int) {
	*d = append((*d)[:start], (*d)[end:]...)
}

func (d *Document) CutAndInsert(start, end int, data []byte) {
	d.Cut(start, end)

	// grow to accommodate new data if required
	if end-start < len(data) {
		*d = append(*d, make(Document, len(data))...)
	}

	otherHalf := (*d)[end:]
	// insert data
	copy((*d)[start+1:], data)
	// append the other half
	copy((*d)[start+len(data):], otherHalf)
	// clean up unused space
	*d = (*d)[:start+len(data)+len(otherHalf)]
}

func (d *Document) Replace(old, new []byte, n int) {
	*d = bytes.Replace(*d, old, new, n)
}

func DeepCopy(dst, src interface{}) error {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(src); err != nil {
		return err
	}

	return gob.NewDecoder(&buffer).Decode(dst)
}

package xtemplate

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	
	"github.com/jinzhu/copier"
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

func TransformComponents(content *Document, componentFolder, ext string) error {
	componentFolder = filepath.Join(componentFolder, "_components")
	
	lex := Lexicon{src: *content}
	lex.ParseComponents()
	
	for _, c := range lex.components {
		
		cName := getComponentName(c)
		if cName == "" {
			// component has no parameters (its malformed), cut it out
			content.Cut(c.startPos, c.endPos)
			continue
		}
		cTpl, err := getComponentTpl(cName+"."+ext, componentFolder)
		if err != nil {
			// component template file not found or some other error, cut!
			content.Cut(c.startPos, c.endPos)
			continue
		}
		
		buff := bytes.NewBuffer([]byte{})
		out := transformComponentBlock(cName, c, cTpl)
		buff.Write(out)
		
		fmt.Println(c.String(), "\n", "== debug ==")
		fmt.Println(buff.String())
		fmt.Println("\n", "== end ==")
	}
	
	return nil
}

func transformComponentBlock(cName string, c *Component, cTpl *ComponentTemplate) []byte {
	// make a copy of the template lexicon
	lex := Lexicon{}
	err := copier.CopyWithOption(&lex, cTpl.Lex, copier.Option{DeepCopy: true})
	if err != nil {
		return nil
	}
	tc := lex.components[0]
	
	// make a copy of the component
	// wc := Component{}
	// err = copier.CopyWithOption(&wc, c, copier.Option{DeepCopy: true})
	// if err != nil {
	// 	return nil
	// }	
	
	newName := fmt.Sprint("\"", "component__", c.id, "__", cName, "\"")
	tc.UpdateParameters(newName)
	
	for _, slot := range tc.children {
		if slot.blockType != SlotAction {
			continue
		}
		
		oldName := getComponentName(slot)
		newName := fmt.Sprint("\"", "slot__", c.id, "__", oldName, "\"")
		slot.UpdateParameters(newName)
		delete(tc.childMap, oldName)
		tc.childMap[newName] = slot
		
		cSlot, exists := c.childMap[oldName]
		if exists {
			slot.UpdateContent(cSlot.content)
			
			// transform {{define "component--xxx"}} ... {{end}} to {{template "component__n__xxx" .}}
			retv := transformSubcomponentCall(slot)
			if retv == nil {
				continue
			}
			slot.UpdateContent(retv)
		}
	}
	
	return tc.src
}

func transformSubcomponentCall(c *Component) []byte {
	lex := Lexicon{src: c.content}
	lex.ParseComponents()
	if len(lex.components) == 0 {
		return nil
	}
	
	for i := 0; i < len(lex.components); i++ {
		sc := lex.components[i]
		cc := findComponentByParameter(sc.parameters, c.children)
		if cc == nil {
			continue
		}
		
		newTrx := fmt.Sprint("{{template \"component__", cc.id, "__", getComponentName(sc), "\" .}}")
		c.Replace(sc.src, []byte(newTrx))
	}
	
	return c.src
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

func getComponentTpl(cName string, componentFolder string) (*ComponentTemplate, error) {
	ct, inCache := components[cName]
	if !inCache {
		cnFile := filepath.Join(componentFolder, cName)
		content, err := ioutil.ReadFile(cnFile)
		if err != nil {
			// Todo: log error
			return nil, err
		}
		
		ct = &ComponentTemplate{Name: cName}
		ct.Lex.src = content
		ct.Lex.Parse()
		
		components[cName] = ct
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

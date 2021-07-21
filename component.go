package xtemplate

import (
	"bytes"
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

func TransformComponents(content *Document, componentFolder, ext string) error {
	componentFolder = filepath.Join(componentFolder, "_components")

	lex := Lexicon{src: *content}
	// lex.parse(ActionFilter{ComponentAction, SlotAction})
	lex.parseActions()
	lex.parseComponents()
	for _, c := range lex.components {
		fmt.Println(c.String(), "\n", "== debug ==")

		cName := getComponentName(c.parameters)
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
		tComponent := cTpl.Lex.components[0]

		buff := bytes.NewBuffer([]byte{})
		for _, slot := range c.children {
			if slot.blockType != SlotAction {
				// only slots are processed other content/tags are ignored
				continue
			}

			// get matching slot from the component template
			sName := getComponentName(slot.parameters)
			tSlot := findSlotByName(sName, tComponent)
			if tSlot == nil {
				// slot not found in component template
				continue
			}

			// #slot
			txc := transformSlotToDefine(cName, c.id, slot, tSlot)
			buff.Write(txc)
			fmt.Println(buff.String())

		}

		fmt.Println("\n", "=======")
	}

	return nil
}

func getComponentHeader(c *Component) []byte {
	return c.src[c.localStartPos():c.localContentStart()]
}

func transformSlotToDefine(cName string, id int, slot, tSlot *Component) []byte {
	/*
		{{#slot "header" .}}   			{{define "component_card_1_header" .}}
			Hello Header       -->			Hello Header
		{{end}}											{{end}}
	*/

	out := bytes.NewBufferString("")
	header := Document(tSlot.src[tSlot.localStartPos():tSlot.localContentStart()])
	parts := cmpBlockStart.FindSubmatch(header)
	
	newName := fmt.Sprintf("component_%s_%d_%s", cName,  id, string(parts[2]))
	newHeader := fmt.Sprintf(`{{define "%s"}}`, newName)
	out.WriteString(newHeader)
	
	// slot.content -> tSlot.content
	if slot != nil {
		out.Write(slot.content)
	} else {
		out.Write(tSlot.content)
	}
	out.WriteString("{{end}}")

	return out.Bytes()
}

func findSlotByName(name string, component *Component) *Component {
	if component == nil {
		return nil
	}

	for _, s := range component.children {
		sName := getComponentName(s.parameters)
		if sName == name {
			return s
		}
	}

	return nil
}

func getComponentName(params string) string {
	p := strings.Split(strings.TrimSpace(params), " ")
	if len(p) == 0 {
		return ""
	}

	return stripQuotes(p[0])
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
		ct.Lex.parse()

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

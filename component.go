package xtemplate

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

/*
	- A component is an HTML tag <component id="card"></component>
	- Every valid component must have an id attribute
	- A template file whose name is the value of id attribute must exist in the component template folder
	- A component template must be a valid go text/template
		when a component template is found its translated into a block action
		e.g
		<component id="card"><slot name="body">A new world</slot></component>
		-- _components/card.tmpl
		<div>
			{{block "#slot--body" . }}
			hello world
			{{end}}
		</div>
		== >
		{{block "component_1_card" .}}
		<div>
			{{block "card__1__body" . }}
			hello world
			{{end}}
		</div>
		{{end}}

	- A component can have at least one child <slot name="body"></slot> tag
		if a block action whose first parameter is the name of a slot exists in the component template
		the body of the slot tag will replace the body of the block action
		e.g:
		<slot name="title">A slot title</slot> && {{block "#slot--title" . }}title{{end}}
		 ==> {{block "title" . }}a slot title{{end}}
	- If a component has no child slot but the component template has a default block
		the body if the component replaces the body of the default block
		e.g:
		<component id="article">Hello world</component>
		-- _components/article.tmpl
		{{block "#slot--default" .}}{{end}}

		==> {{block "component_1_article" .}}
					{{block "article__1__default".}}Hello world{{
				{{end}}

*/

// var componentStartRe = regexp.MustCompile(`(?i)<(component|slot)\s+(id|name)="([a-zA-Z0-9\-\_]*?)"\s*>`)
// var componentEndRe = regexp.MustCompile(`</(component|slot)>`)
var attrRe = regexp.MustCompile(`(?i)(id|name)="([a-zA-Z0-9\-_]*?)"`)
var htmlTagRe = regexp.MustCompile(`</*([a-zA-Z]+)([\s="a-zA-Z0-9\-_]*?)>`)
var actionTagRe = regexp.MustCompile(`{{-*\s*([\w]+)\s?([\s\w"-.$:=]*?)\s*-*}}`)
var inQuotes = regexp.MustCompile(`"([\s\w#-.$:=]*?)"`)

type tagType int

const (
	OpeningTag tagType = iota
	ClosingTag
)

type Tag struct {
	Element  string
	ID       string
	Name     string
	StartPos int
	EndPos   int
	Type     tagType
	Body     string
}

func (t Tag) getSrc(src []byte) []byte {
	return src[t.StartPos:t.EndPos]
}

func translateComponents(src Document, tplFolder string) ([]byte, error) {

	cCount := 0
	tplFolder = filepath.Join(tplFolder, "_components")

	for {
		// log.Println(src.String())
		tag, _ := findTag(src, "component")
		if tag == nil {
			break
		}
		cCount++

		cTpl, err := getComponentTemplate(tag.ID, tplFolder)
		if err != nil {
			continue
		}

		cBlock := Document(fmt.Sprintf("{{block \"%s_%d\" .}}\n", tag.ID, cCount))
		cBlock.Append(cTpl, []byte("\n{{end}}"))

		slots, err := listComponentSlots(tag.getSrc(src), tag.ID)
		if err != nil {
			return nil, err
		}
		actions, err := listActionSlots(cBlock)
		if err != nil {
			return nil, err
		}

		// substitute slot content
		for _, slot := range slots {
			action := popAction(&actions, "#slot--"+slot.Name)
			if action == nil {
				continue
			}

			action, err := findAction(cBlock, "block", action.ID)
			if err != nil || action == nil {
				return nil, err
			}

			swapContent(&cBlock, action, tag.ID, cCount, slot.Name, slot.Body)

		}

		// process unused slots
		for _, a := range actions {
			action, err := findAction(cBlock, "block", a.ID)
			if err != nil || action == nil {
				return nil, err
			}

			if action.ID == "#slot--default" && len(slots) == 0 {
				swapContent(&cBlock, action, tag.ID, cCount, "default", tag.Body)
				continue
			}

			swapContent(&cBlock, action, tag.ID, cCount, action.ID, "")
		}

		// replace tag in src
		src.Replace(tag.getSrc(src), cBlock, 1)
	}

	return src, nil
}

func swapContent(cBlock *Document, action *Action, tagID string, cCount int, slotName, slotBody string) {
	oldAction := (*cBlock)[action.StartPos:action.EndPos]
	newAction := Document(oldAction)
	// prefix slot block name with component id
	if strings.HasPrefix(slotName, "#slot--") {
		slotName = strings.TrimPrefix(slotName, "#slot--")
	}
	sn := fmt.Sprintf("%s__%d__%s", tagID, cCount, slotName)
	newAction.Replace([]byte(action.ID), []byte(sn), 1)

	// replace body
	if len(slotBody) > 0 {
		newAction.Replace([]byte(action.Body), []byte(slotBody), 1)
	}

	// replace template action
	cBlock.Replace(oldAction, newAction, 1)
}

func popAction(actions *[]Action, id string) *Action {
	var a Action
	for i := 0; i < len(*actions); i++ {
		a = (*actions)[i]
		if a.ID == id {
			*actions = append((*actions)[:i], (*actions)[i+1:]...)
			return &a
		}
	}

	return nil
}

func findTag(src []byte, name string) (*Tag, error) {
	name = strings.ToLower(name)
	tags := htmlTagRe.FindAllSubmatchIndex(src, -1)
	stack := []Tag{}

	for i := 0; i < len(tags); i++ {
		tag := getTag(src, tags[i])
		// log.Println(tag.Name)
		if name == tag.Element && tag.Type == OpeningTag {
			stack = append(stack, tag)
			continue
		}

		if name == tag.Element && tag.Type == ClosingTag {
			var sTag Tag
			sTag, stack = stack[len(stack)-1], stack[:len(stack)-1]

			if len(stack) == 0 {
				sTag.Body = string(src[sTag.EndPos:tag.StartPos])
				sTag.EndPos = tag.EndPos

				return &sTag, nil
			}
		}
	}
	var err error
	if len(stack) > 0 {
		err = fmt.Errorf("cant find closing tag for:%s", name)
	}
	return nil, err
}

func listComponentSlots(src []byte, id string) ([]Tag, error) {
	id = strings.ToLower(id)
	tags := htmlTagRe.FindAllSubmatchIndex(src, -1)
	parentFound := false
	stack := make([]Tag, 0)
	retv := make([]Tag, 0)

	for i := 0; i < len(tags); i++ {
		tag := getTag(src, tags[i])
		// log.Println(tag.Name)
		if tag.Element == "component" && tag.ID == id && tag.Type == OpeningTag {
			parentFound = true
			continue
		}
		if !parentFound {
			continue
		}

		if tag.Element == "slot" && tag.Type == OpeningTag {
			stack = append(stack, tag)
			continue
		}

		if tag.Element == "slot" && tag.Type == ClosingTag {
			var sTag Tag
			sTag, stack = stack[len(stack)-1], stack[:len(stack)-1]

			if len(stack) == 0 {
				sTag.Body = string(src[sTag.EndPos:tag.StartPos])
				sTag.EndPos = tag.EndPos

				retv = append(retv, sTag)
			}
		}
	}

	var err error
	if len(stack) > 0 {
		err = fmt.Errorf("cant find closing tag for:%s", id)
	}

	return retv, err
}

func getTag(src []byte, location []int) Tag {
	tag := Tag{
		Element:  strings.ToLower(string(src[location[2]:location[3]])),
		StartPos: location[0],
		EndPos:   location[1],
		Type:     OpeningTag,
	}

	if src[tag.StartPos+1] == '/' {
		tag.Type = ClosingTag
	}

	if tag.Type == OpeningTag {
		tagSrc := src[tag.StartPos:tag.EndPos]
		// log.Println(string(tagSrc))
		groups := attrRe.FindAllSubmatchIndex(tagSrc, -1)
		if len(groups) > 0 && len(groups[0]) > 5 {
			key := strings.ToLower(string(tagSrc[groups[0][2]:groups[0][3]]))
			val := string(tagSrc[groups[0][4]:groups[0][5]])
			if key == "id" {
				tag.ID = val
			} else {
				tag.Name = val
			}

			// log.Println(key, val)
		}

	}

	return tag
}

type actionType int

const (
	OpeningAction actionType = iota
	ClosingAction
	SingleAction
)

type Action struct {
	Name     string
	Attr     string
	ID       string
	StartPos int
	EndPos   int
	Body     string
	Type     actionType
}

func findAction(src []byte, name, id string) (*Action, error) {
	name = strings.ToLower(name)
	actions := actionTagRe.FindAllSubmatchIndex(src, -1)
	stack := make([]Action, 0)

	for i := 0; i < len(actions); i++ {
		action := getAction(src, actions[i])
		// log.Println(tag.Name)
		if action.Type == OpeningAction {
			stack = append(stack, action)
			continue
		}

		if action.Type == ClosingAction {
			var sAction Action
			if len(stack) == 0 {
				return nil, fmt.Errorf("found closing action without a coresponding opening action")
			}

			sAction, stack = stack[len(stack)-1], stack[:len(stack)-1]

			if sAction.Name == name && sAction.ID == id {
				sAction.Body = string(src[sAction.EndPos:action.StartPos])
				sAction.EndPos = action.EndPos

				return &sAction, nil
			}
		}
	}

	var err error
	if len(stack) > 0 {
		err = fmt.Errorf("cant find closing action for:%s", name)
	}
	return nil, err
}

func listActionSlots(src []byte) ([]Action, error) {
	name := "block"
	slotPrefix := "#slot--"

	actions := actionTagRe.FindAllSubmatchIndex(src, -1)
	stack := make([]Action, 0)
	slots := make([]Action, 0)

	for i := 0; i < len(actions); i++ {
		action := getAction(src, actions[i])
		// log.Println(tag.Name)
		if action.Type == OpeningAction {
			stack = append(stack, action)
			continue
		}

		if action.Type == ClosingAction {
			var sAction Action
			if len(stack) == 0 {
				return nil, fmt.Errorf("found closing action without a coresponding opening action")
			}

			sAction, stack = stack[len(stack)-1], stack[:len(stack)-1]

			if sAction.Name == name && strings.HasPrefix(sAction.ID, slotPrefix) {
				sAction.Body = string(src[sAction.EndPos:action.StartPos])
				sAction.EndPos = action.EndPos

				slots = append(slots, sAction)
				continue
			}
		}
	}

	var err error
	if len(stack) > 0 {
		err = fmt.Errorf("cant find closing action for:%s", name)
	}
	return slots, err
}

func getAction(src []byte, location []int) Action {
	act := Action{
		Name:     strings.ToLower(string(src[location[2]:location[3]])),
		Attr:     string(src[location[4]:location[5]]),
		StartPos: location[0],
		EndPos:   location[1],
		Type:     OpeningAction,
	}

	if act.Name == "end" {
		act.Type = ClosingAction
	} else if !StrListIncludes(act.Name, []string{"define", "if", "range", "block", "with"}) {
		act.Type = SingleAction
	}

	if act.Type == OpeningAction || act.Type == SingleAction {
		groups := inQuotes.FindAllStringSubmatch(act.Attr, -1)
		if len(groups) > 0 {
			act.ID = groups[0][1]
		}
	}

	return act
}

func getComponentTemplate(name, folder string) (contents []byte, err error) {
	fleName := filepath.Join(folder, name+".tmpl")
	contents, err = ioutil.ReadFile(fleName)

	return
}

func StrListIncludes(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}

	return false
}

type Document []byte

func (d *Document) String() string {
	return string(*d)
}

func (d *Document) Append(docs ...[]byte) {
	for _, doc := range docs {
		if doc == nil {
			continue
		}
		*d = append(*d, doc...)
	}
}

func (d *Document) AppendString(docs ...string) {
	for _, doc := range docs {
		d.Append([]byte(doc))
	}
}

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

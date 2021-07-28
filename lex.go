package xtemplate

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var lexAction = regexp.MustCompile(`{{-*\s*([#\w]+)\s?([\s\w"-.$:=]*?)\s*-*}}`)
var lexComponentPrefix = regexp.MustCompile(`(?mi)"component--`)
var lexSlotPrefix = regexp.MustCompile(`(?mi)"slot--`)

type ActionType int

func (a ActionType) String() string {
	str := ""
	switch a {
	case ComponentAction:
		str = "#component"
	case SlotAction:
		str = "#slot"
	case IfAction:
		str = "if"
	case RangeAction:
		str = "range"
	case WithAction:
		str = "with"
	case BlockAction:
		str = "block"
	case DefineAction:
		str = "define"
	case ElseAction:
		str = "else"
	case EndAction:
		str = "end"
	default:
		str = "other"
	}

	return str
}

func StrToBaseAction(val string) ActionType {
	var action ActionType
	switch val {
	case "#component":
		action = ComponentAction
	case "#slot":
		action = SlotAction
	case "if":
		action = IfAction
	case "range":
		action = RangeAction
	case "with":
		action = WithAction
	case "block":
		action = BlockAction
	case "define":
		action = DefineAction
	case "else":
		action = ElseAction
	case "end":
		action = EndAction
	default:
		action = OtherAction
	}

	return action
}

const (
	OtherAction ActionType = iota
	ComponentAction
	SlotAction
	IfAction
	RangeAction
	WithAction
	BlockAction
	DefineAction
	ElseAction
	EndAction
)

type TokenType int

func (t TokenType) String() string {
	switch t {
	case ActionTypeBlockStart:
		return "blockStart"
	case ActionTypeBlockElse:
		return "blockElse"
	case ActionTypeBlockEnd:
		return "blockEnd"
	default:
		return "single"
	}
}

const (
	ActionTypeSingle TokenType = iota
	ActionTypeBlockStart
	ActionTypeBlockElse
	ActionTypeBlockEnd
)

type ActionItem struct {
	name       string
	parameters string
	startPos   int
	endPos     int
	src        []byte
	tokenType  TokenType
}

type Block struct {
	name         string
	parameters   string
	startPos     int
	endPos       int
	blockType    ActionType
	src          []byte
	content      []byte
	contentStart int
	contentEnd   int
}

type Blocks []Block

func (b *Blocks) Len() int      { return len(*b) }
func (b *Blocks) Swap(i, j int) { (*b)[i], (*b)[j] = (*b)[j], (*b)[i] }
func (b *Blocks) Less(i, j int) bool {
	return (*b)[i].endPos < (*b)[j].endPos
}

func (b Block) String() string {
	return fmt.Sprint(b.name, "(", b.parameters, " ) :", string(b.content))
}

func (b Block) localStartPos() int {
	return 0
}
func (b Block) localEndPos() int {
	return b.endPos - b.startPos
}
func (b Block) localContentStart() int {
	return b.contentStart - b.startPos
}
func (b Block) localContentEnd() int {
	return b.contentEnd - b.startPos
}

func (b *Block) UpdateParameters(val string) {
	b.Replace([]byte(b.parameters), []byte(val))
	b.parameters = val
}

func (b *Block) UpdateContent(new []byte) {
	b.Replace(b.content, new)
}

func (b *Block) Replace(old, new []byte) {
	// find position of snippet to replace
	idx := bytes.Index(b.src, old)
	if idx == -1 {
		return
	}

	oldLen := len(old)
	newLen := len(new)
	diff := oldLen - newLen
	if diff < 0 {
		diff *= -1
	}

	b.src = bytes.Replace(b.src, old, new, 1)

	if idx < b.localContentStart() {
		if oldLen < newLen {
			b.endPos += diff
			b.contentStart += diff
			b.contentEnd += diff
		} else if oldLen > newLen {
			b.endPos -= diff
			b.contentStart -= diff
			b.contentEnd -= diff
		}
	} else {
		if oldLen < newLen {
			b.contentEnd += diff
		} else if oldLen > newLen {
			b.contentEnd -= diff
		}
	}

	b.content = b.src[b.localContentStart():b.localContentEnd()]
}

func (i ActionItem) Type() ActionType {
	at := StrToBaseAction(i.name)
	if i.tokenType == ActionTypeBlockStart {
		if at == DefineAction && lexComponentPrefix.Find([]byte(i.parameters)) != nil {
			return ComponentAction
		} else if at == BlockAction && lexSlotPrefix.Find([]byte(i.parameters)) != nil {
			return SlotAction
		}
	}

	return at
}

type Lexicon struct {
	currentPos int64
	src        []byte
	lastAction int
	actions    []ActionItem
	blocks     Blocks
	components []*Component
}

func (l *Lexicon) Parse() {
	l.ParseComponents()
}

func (l *Lexicon) ParseActions() {
	locations := lexAction.FindAllSubmatchIndex(l.src, -1)
	l.actions = make([]ActionItem, 0)
	actionBlocks := make([]*ActionItem, 0)

	for i := 0; i < len(locations); i++ {
		loc := locations[i]
		action := ActionItem{}

		action.startPos = loc[0]
		action.endPos = loc[1]
		action.src = l.src[loc[0]:loc[1]]
		action.name = strings.TrimSpace(string(l.src[loc[2]:loc[3]]))

		if action.Type() != ElseAction && action.Type() != EndAction {
			if action.Type() != OtherAction {
				action.tokenType = ActionTypeBlockStart
				actionBlocks = append(actionBlocks, &action)
			} else {
				action.tokenType = ActionTypeSingle
			}
		} else if action.Type() == ElseAction {
			action.tokenType = ActionTypeBlockElse
		} else {
			action.tokenType = ActionTypeBlockEnd
			if len(actionBlocks) != 0 {
				var act *ActionItem
				// pop last item in actionBlocks
				act, actionBlocks = actionBlocks[len(actionBlocks)-1], actionBlocks[:len(actionBlocks)-1]

				blk := Block{
					name:         act.name,
					parameters:   act.parameters,
					blockType:    act.Type(),
					startPos:     act.startPos,
					endPos:       action.endPos,
					content:      l.src[act.endPos:action.startPos],
					contentStart: act.endPos,
					contentEnd:   action.startPos,
				}
				blk.src = l.src[blk.startPos:blk.endPos]
				l.blocks = append(l.blocks, blk)
			}
		}

		if len(loc) > 3 {
			action.parameters = string(l.src[loc[4]:loc[5]])
		}

		l.actions = append(l.actions, action)
		// fmt.Println(i, ":", action.name, action.tokenType.String(), action.parameters)
	}
}

type Component struct {
	*Block
	children []*Component
	childMap map[string]*Component
	id       int
	isChild  bool
}

func (c Component) String(depth ...int) string {
	d := 0
	if len(depth) > 0 {
		d = depth[0]
	}

	retv := strings.Builder{}
	retv.WriteString(fmt.Sprintln(strings.Repeat("\t", d), c.id, ":", c.name, "(", c.parameters, ")", "{[", c.startPos, "]"))
	for _, s := range c.children {
		retv.WriteString(fmt.Sprintln(strings.Repeat("\t", d+1), s.id, ":", s.name, s.parameters, s.startPos, s.endPos, len(s.children)))
		for _, t := range s.children {
			retv.WriteString(strings.Repeat("\t\t", d) + t.String(d+1))
		}
	}
	retv.WriteString(fmt.Sprintln(strings.Repeat("\t", d), "[", c.endPos, "]}"))

	return retv.String()
}

func (l *Lexicon) ParseComponents() {
	if len(l.blocks) == 0 {
		// src hasn't been parsed
		l.ParseActions()
	}

	components := make([]*Component, 0)
	for i := len(l.blocks) - 1; i >= 0; i-- {
		b := l.blocks[i]
		if b.blockType == ComponentAction {
			c := &Component{
				Block:    &b,
				children: make([]*Component, 0),
				childMap: make(map[string]*Component),
			}
			components = append(components, c)
			c.id = len(components)
			c.parameters = strings.ToLower(c.parameters)

			bParent := l.findParentBlock(&b)
			if bParent != nil {
				parent := l.findMatchingComponent(bParent, components)
				if parent != nil {
					parent.children = append(parent.children, c)
					c.isChild = true
					name := getComponentName(c)
					parent.childMap[name] = c
				}
			}

			l.components = append(l.components, c)
			continue
		}

		if b.blockType == SlotAction {
			c := &Component{
				Block:    &b,
				children: make([]*Component, 0),
				childMap: make(map[string]*Component),
			}
			components = append(components, c)
			c.id = len(components)
			c.parameters = strings.ToLower(c.parameters)

			bParent := l.findParentBlock(&b)
			if bParent != nil {
				parent := l.findMatchingComponent(bParent, components)
				if parent != nil {
					parent.children = append(parent.children, c)
					name := getComponentName(c)
					parent.childMap[name] = c
					continue
				}
			}
			continue
		}
	}
}

func (l *Lexicon) findParentBlock(b *Block) (parent *Block) {
	for i := 0; i < len(l.blocks); i++ {
		parent = &l.blocks[i]
		if b.startPos > parent.startPos && b.endPos > parent.startPos && b.endPos < parent.endPos {
			return
		}
	}

	return nil
}

func (l *Lexicon) findParentComponent(c *Component) (parent *Component) {
	for i := 0; i < len(l.components); i++ {
		parent = l.components[i]
		if c.startPos > parent.startPos && c.endPos > parent.startPos && c.endPos < parent.endPos {
			return
		}
	}

	return nil
}

func (l *Lexicon) findMatchingComponent(b *Block, list []*Component) *Component {

	for _, c := range list {
		if b.startPos == c.startPos && b.endPos == c.endPos {
			return c
		}
	}

	return nil
}

func (l *Lexicon) findChildComponentByParameter(b *Block) *Component {

	for _, c := range l.components {
		if !c.isChild {
			continue
		}
		if b.parameters == c.parameters {
			return c
		}
	}

	return nil
}

func findComponentByParameter(param string, list []*Component) *Component {

	for _, c := range list {
		if c.parameters == param {
			return c
		}
	}

	return nil
}

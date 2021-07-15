package xtemplate

import (
	"fmt"
	"regexp"
	"strings"
)

var lexAction = regexp.MustCompile(`{{-*\s*([#\w]+)([\s\w".$]*?)\s*-*}}`)

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

func StrToAction(val string) ActionType {
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

type BlockItem struct {
	name       string
	parameters string
	startPos   int
	endPos     int
	src        []byte
	content    []byte
	contentStart int
	contentEnd int
}

func (b BlockItem) String() string {
	return fmt.Sprint(b.name, "(", b.parameters, " ) :", string(b.content))
}

func (i ActionItem) Type() ActionType {
	return StrToAction(i.name)
}

type ActionFilter []ActionType

func (f ActionFilter) Includes(v ActionType) bool {
	for _, i := range f {
		if i == v {
			return true
		}
	}

	return false
}

func Transform(content []byte) error {
	lex := Lexicon{
		src: content,
	}
	
	lex.parse(ActionFilter{ComponentAction, SlotAction})
	for _, i := range lex.blocks {
		fmt.Println(i.name, i.parameters)
	}
	
	return nil
}

type Lexicon struct {
	currentPos int64
	src        []byte
	lastAction int
	actions    []ActionItem
	blocks     []BlockItem
	filter     ActionFilter
}

func (l *Lexicon) parse(filter ActionFilter) {
	l.filter = filter
	if l.filter == nil {
		l.filter = ActionFilter{
			ComponentAction, SlotAction,
			IfAction, RangeAction, WithAction,
			BlockAction, DefineAction, ElseAction,
			EndAction, OtherAction,
		}

	}
	l.parseActions()
	l.parseBlocks(0)
}

func (l *Lexicon) parseActions() {
	locations := lexAction.FindAllSubmatchIndex(l.src, -1)
	l.actions = make([]ActionItem, 0)

	for i := l.lastAction; i < len(locations); i++ {
		loc := locations[i]
		action := ActionItem{}

		action.startPos = loc[0]
		action.endPos = loc[1]
		action.src = l.src[loc[0]:loc[1]]
		action.name = strings.TrimSpace(string(l.src[loc[2]:loc[3]]))
		if action.Type() != ElseAction && action.Type() != EndAction {
			if action.Type() != OtherAction {
				action.tokenType = ActionTypeBlockStart
			} else {
				action.tokenType = ActionTypeSingle
			}
		} else if action.Type() == ElseAction {
			action.tokenType = ActionTypeBlockElse
		} else {
			action.tokenType = ActionTypeBlockEnd
		}

		if len(loc) > 3 {
			action.parameters = string(l.src[loc[4]:loc[5]])
		}

		l.actions = append(l.actions, action)
		// fmt.Println(i, ":", action.name, action.Type().String(), action.tokenType.String())
	}
}

func (l *Lexicon) parseBlocks(idx int) int {
	foundBlkStart := -1
	for i := idx; i < len(l.actions); i++ {
		a := l.actions[i]

		if a.tokenType == ActionTypeSingle || a.tokenType == ActionTypeBlockElse {
			continue
		}

		if a.tokenType == ActionTypeBlockStart {
			if foundBlkStart != -1 {
				// found another blockStart while in a block
				i = l.parseBlocks(i)
				continue
			}

			foundBlkStart = i
			continue
		}

		if a.tokenType == ActionTypeBlockEnd {
			act := l.actions[foundBlkStart]
			blk := BlockItem{
				name:       act.name,
				parameters: act.parameters,
				startPos:   act.startPos,
				endPos:     a.endPos,
				content:    l.src[act.endPos:a.startPos],
				contentStart: act.endPos,
				contentEnd: a.startPos,
			}

			blk.src = l.src[blk.startPos:blk.endPos]
			foundBlkStart = -1

			if l.filter.Includes(act.Type()) {
				l.blocks = append(l.blocks, blk)
				// fmt.Println(blk)
			}

			if idx != 0 {
				// exit recursion
				return i
			}
		}
	}

	return len(l.actions) - 1
}

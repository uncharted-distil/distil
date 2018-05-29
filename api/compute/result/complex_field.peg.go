package compute

//go:generate peg complex_field.peg

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleComplexField
	rulearray
	ruleitem
	rulestring
	ruleescaped_string
	rulenonescaped_string
	rulews
	rulecomma
	rulelf
	rulecr
	ruleescdquote
	rulesquote
	ruleobracket
	rulecbracket
	rulevalue
	ruletextdata
	rulePegText
	ruleAction0
	ruleAction1
	ruleAction2
)

var rul3s = [...]string{
	"Unknown",
	"ComplexField",
	"array",
	"item",
	"string",
	"escaped_string",
	"nonescaped_string",
	"ws",
	"comma",
	"lf",
	"cr",
	"escdquote",
	"squote",
	"obracket",
	"cbracket",
	"value",
	"textdata",
	"PegText",
	"Action0",
	"Action1",
	"Action2",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type ComplexField struct {
	arrayElements

	Buffer string
	buffer []rune
	rules  [21]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *ComplexField) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *ComplexField) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *ComplexField
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *ComplexField) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *ComplexField) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.addElement(buffer[begin:end])
		case ruleAction1:
			p.addElement(buffer[begin:end])
		case ruleAction2:
			p.addElement(buffer[begin:end])

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *ComplexField) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 ComplexField <- <(array !.)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[rulearray]() {
					goto l0
				}
				{
					position2, tokenIndex2 := position, tokenIndex
					if !matchDot() {
						goto l2
					}
					goto l0
				l2:
					position, tokenIndex = position2, tokenIndex2
				}
				add(ruleComplexField, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 array <- <(ws* obracket ws* (item ws* (comma ws* item ws*)*)? cbracket)> */
		func() bool {
			position3, tokenIndex3 := position, tokenIndex
			{
				position4 := position
			l5:
				{
					position6, tokenIndex6 := position, tokenIndex
					if !_rules[rulews]() {
						goto l6
					}
					goto l5
				l6:
					position, tokenIndex = position6, tokenIndex6
				}
				if !_rules[ruleobracket]() {
					goto l3
				}
			l7:
				{
					position8, tokenIndex8 := position, tokenIndex
					if !_rules[rulews]() {
						goto l8
					}
					goto l7
				l8:
					position, tokenIndex = position8, tokenIndex8
				}
				{
					position9, tokenIndex9 := position, tokenIndex
					if !_rules[ruleitem]() {
						goto l9
					}
				l11:
					{
						position12, tokenIndex12 := position, tokenIndex
						if !_rules[rulews]() {
							goto l12
						}
						goto l11
					l12:
						position, tokenIndex = position12, tokenIndex12
					}
				l13:
					{
						position14, tokenIndex14 := position, tokenIndex
						if !_rules[rulecomma]() {
							goto l14
						}
					l15:
						{
							position16, tokenIndex16 := position, tokenIndex
							if !_rules[rulews]() {
								goto l16
							}
							goto l15
						l16:
							position, tokenIndex = position16, tokenIndex16
						}
						if !_rules[ruleitem]() {
							goto l14
						}
					l17:
						{
							position18, tokenIndex18 := position, tokenIndex
							if !_rules[rulews]() {
								goto l18
							}
							goto l17
						l18:
							position, tokenIndex = position18, tokenIndex18
						}
						goto l13
					l14:
						position, tokenIndex = position14, tokenIndex14
					}
					goto l10
				l9:
					position, tokenIndex = position9, tokenIndex9
				}
			l10:
				if !_rules[rulecbracket]() {
					goto l3
				}
				add(rulearray, position4)
			}
			return true
		l3:
			position, tokenIndex = position3, tokenIndex3
			return false
		},
		/* 2 item <- <(string / (<value*> Action0))> */
		func() bool {
			position19, tokenIndex19 := position, tokenIndex
			{
				position20 := position
				{
					position21, tokenIndex21 := position, tokenIndex
					if !_rules[rulestring]() {
						goto l22
					}
					goto l21
				l22:
					position, tokenIndex = position21, tokenIndex21
					{
						position23 := position
					l24:
						{
							position25, tokenIndex25 := position, tokenIndex
							if !_rules[rulevalue]() {
								goto l25
							}
							goto l24
						l25:
							position, tokenIndex = position25, tokenIndex25
						}
						add(rulePegText, position23)
					}
					if !_rules[ruleAction0]() {
						goto l19
					}
				}
			l21:
				add(ruleitem, position20)
			}
			return true
		l19:
			position, tokenIndex = position19, tokenIndex19
			return false
		},
		/* 3 string <- <(escaped_string / nonescaped_string)> */
		func() bool {
			position26, tokenIndex26 := position, tokenIndex
			{
				position27 := position
				{
					position28, tokenIndex28 := position, tokenIndex
					if !_rules[ruleescaped_string]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex = position28, tokenIndex28
					if !_rules[rulenonescaped_string]() {
						goto l26
					}
				}
			l28:
				add(rulestring, position27)
			}
			return true
		l26:
			position, tokenIndex = position26, tokenIndex26
			return false
		},
		/* 4 escaped_string <- <(escdquote <(textdata / squote / lf / cr / obracket / cbracket)*> escdquote Action1)> */
		func() bool {
			position30, tokenIndex30 := position, tokenIndex
			{
				position31 := position
				if !_rules[ruleescdquote]() {
					goto l30
				}
				{
					position32 := position
				l33:
					{
						position34, tokenIndex34 := position, tokenIndex
						{
							position35, tokenIndex35 := position, tokenIndex
							if !_rules[ruletextdata]() {
								goto l36
							}
							goto l35
						l36:
							position, tokenIndex = position35, tokenIndex35
							if !_rules[rulesquote]() {
								goto l37
							}
							goto l35
						l37:
							position, tokenIndex = position35, tokenIndex35
							if !_rules[rulelf]() {
								goto l38
							}
							goto l35
						l38:
							position, tokenIndex = position35, tokenIndex35
							if !_rules[rulecr]() {
								goto l39
							}
							goto l35
						l39:
							position, tokenIndex = position35, tokenIndex35
							if !_rules[ruleobracket]() {
								goto l40
							}
							goto l35
						l40:
							position, tokenIndex = position35, tokenIndex35
							if !_rules[rulecbracket]() {
								goto l34
							}
						}
					l35:
						goto l33
					l34:
						position, tokenIndex = position34, tokenIndex34
					}
					add(rulePegText, position32)
				}
				if !_rules[ruleescdquote]() {
					goto l30
				}
				if !_rules[ruleAction1]() {
					goto l30
				}
				add(ruleescaped_string, position31)
			}
			return true
		l30:
			position, tokenIndex = position30, tokenIndex30
			return false
		},
		/* 5 nonescaped_string <- <(squote <textdata*> squote Action2)> */
		func() bool {
			position41, tokenIndex41 := position, tokenIndex
			{
				position42 := position
				if !_rules[rulesquote]() {
					goto l41
				}
				{
					position43 := position
				l44:
					{
						position45, tokenIndex45 := position, tokenIndex
						if !_rules[ruletextdata]() {
							goto l45
						}
						goto l44
					l45:
						position, tokenIndex = position45, tokenIndex45
					}
					add(rulePegText, position43)
				}
				if !_rules[rulesquote]() {
					goto l41
				}
				if !_rules[ruleAction2]() {
					goto l41
				}
				add(rulenonescaped_string, position42)
			}
			return true
		l41:
			position, tokenIndex = position41, tokenIndex41
			return false
		},
		/* 6 ws <- <' '> */
		func() bool {
			position46, tokenIndex46 := position, tokenIndex
			{
				position47 := position
				if buffer[position] != rune(' ') {
					goto l46
				}
				position++
				add(rulews, position47)
			}
			return true
		l46:
			position, tokenIndex = position46, tokenIndex46
			return false
		},
		/* 7 comma <- <','> */
		func() bool {
			position48, tokenIndex48 := position, tokenIndex
			{
				position49 := position
				if buffer[position] != rune(',') {
					goto l48
				}
				position++
				add(rulecomma, position49)
			}
			return true
		l48:
			position, tokenIndex = position48, tokenIndex48
			return false
		},
		/* 8 lf <- <'\n'> */
		func() bool {
			position50, tokenIndex50 := position, tokenIndex
			{
				position51 := position
				if buffer[position] != rune('\n') {
					goto l50
				}
				position++
				add(rulelf, position51)
			}
			return true
		l50:
			position, tokenIndex = position50, tokenIndex50
			return false
		},
		/* 9 cr <- <'\r'> */
		func() bool {
			position52, tokenIndex52 := position, tokenIndex
			{
				position53 := position
				if buffer[position] != rune('\r') {
					goto l52
				}
				position++
				add(rulecr, position53)
			}
			return true
		l52:
			position, tokenIndex = position52, tokenIndex52
			return false
		},
		/* 10 escdquote <- <'"'> */
		func() bool {
			position54, tokenIndex54 := position, tokenIndex
			{
				position55 := position
				if buffer[position] != rune('"') {
					goto l54
				}
				position++
				add(ruleescdquote, position55)
			}
			return true
		l54:
			position, tokenIndex = position54, tokenIndex54
			return false
		},
		/* 11 squote <- <'\''> */
		func() bool {
			position56, tokenIndex56 := position, tokenIndex
			{
				position57 := position
				if buffer[position] != rune('\'') {
					goto l56
				}
				position++
				add(rulesquote, position57)
			}
			return true
		l56:
			position, tokenIndex = position56, tokenIndex56
			return false
		},
		/* 12 obracket <- <'['> */
		func() bool {
			position58, tokenIndex58 := position, tokenIndex
			{
				position59 := position
				if buffer[position] != rune('[') {
					goto l58
				}
				position++
				add(ruleobracket, position59)
			}
			return true
		l58:
			position, tokenIndex = position58, tokenIndex58
			return false
		},
		/* 13 cbracket <- <']'> */
		func() bool {
			position60, tokenIndex60 := position, tokenIndex
			{
				position61 := position
				if buffer[position] != rune(']') {
					goto l60
				}
				position++
				add(rulecbracket, position61)
			}
			return true
		l60:
			position, tokenIndex = position60, tokenIndex60
			return false
		},
		/* 14 value <- <([a-z] / [A-Z] / [0-9])> */
		func() bool {
			position62, tokenIndex62 := position, tokenIndex
			{
				position63 := position
				{
					position64, tokenIndex64 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l65
					}
					position++
					goto l64
				l65:
					position, tokenIndex = position64, tokenIndex64
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l66
					}
					position++
					goto l64
				l66:
					position, tokenIndex = position64, tokenIndex64
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l62
					}
					position++
				}
			l64:
				add(rulevalue, position63)
			}
			return true
		l62:
			position, tokenIndex = position62, tokenIndex62
			return false
		},
		/* 15 textdata <- <([a-z] / [A-Z] / [0-9] / ' ' / '!' / '#' / '$' / '&' / '%' / '(' / ')' / '*' / '+' / '-' / '.' / '/' / ':' / ';' / [<->] / '?' / '\\' / '^' / '_' / '`' / '{' / '|' / '}' / '~')> */
		func() bool {
			position67, tokenIndex67 := position, tokenIndex
			{
				position68 := position
				{
					position69, tokenIndex69 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l70
					}
					position++
					goto l69
				l70:
					position, tokenIndex = position69, tokenIndex69
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l71
					}
					position++
					goto l69
				l71:
					position, tokenIndex = position69, tokenIndex69
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l72
					}
					position++
					goto l69
				l72:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune(' ') {
						goto l73
					}
					position++
					goto l69
				l73:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('!') {
						goto l74
					}
					position++
					goto l69
				l74:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('#') {
						goto l75
					}
					position++
					goto l69
				l75:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('$') {
						goto l76
					}
					position++
					goto l69
				l76:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('&') {
						goto l77
					}
					position++
					goto l69
				l77:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('%') {
						goto l78
					}
					position++
					goto l69
				l78:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('(') {
						goto l79
					}
					position++
					goto l69
				l79:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune(')') {
						goto l80
					}
					position++
					goto l69
				l80:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('*') {
						goto l81
					}
					position++
					goto l69
				l81:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('+') {
						goto l82
					}
					position++
					goto l69
				l82:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('-') {
						goto l83
					}
					position++
					goto l69
				l83:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('.') {
						goto l84
					}
					position++
					goto l69
				l84:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('/') {
						goto l85
					}
					position++
					goto l69
				l85:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune(':') {
						goto l86
					}
					position++
					goto l69
				l86:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune(';') {
						goto l87
					}
					position++
					goto l69
				l87:
					position, tokenIndex = position69, tokenIndex69
					if c := buffer[position]; c < rune('<') || c > rune('>') {
						goto l88
					}
					position++
					goto l69
				l88:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('?') {
						goto l89
					}
					position++
					goto l69
				l89:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('\\') {
						goto l90
					}
					position++
					goto l69
				l90:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('^') {
						goto l91
					}
					position++
					goto l69
				l91:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('_') {
						goto l92
					}
					position++
					goto l69
				l92:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('`') {
						goto l93
					}
					position++
					goto l69
				l93:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('{') {
						goto l94
					}
					position++
					goto l69
				l94:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('|') {
						goto l95
					}
					position++
					goto l69
				l95:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('}') {
						goto l96
					}
					position++
					goto l69
				l96:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('~') {
						goto l67
					}
					position++
				}
			l69:
				add(ruletextdata, position68)
			}
			return true
		l67:
			position, tokenIndex = position67, tokenIndex67
			return false
		},
		nil,
		/* 18 Action0 <- <{ p.addElement(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 19 Action1 <- <{ p.addElement(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 20 Action2 <- <{ p.addElement(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
	}
	p.rules = _rules
}

package compute

//go:generate peg -inline ./api/compute/result/complex_field.peg

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
	ruledquote_string
	rulesquote_string
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
	ruleAction0
	ruleAction1
	rulePegText
	ruleAction2
	ruleAction3
	ruleAction4
)

var rul3s = [...]string{
	"Unknown",
	"ComplexField",
	"array",
	"item",
	"string",
	"dquote_string",
	"squote_string",
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
	"Action0",
	"Action1",
	"PegText",
	"Action2",
	"Action3",
	"Action4",
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
	rules  [23]func() bool
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
			p.pushArray()
		case ruleAction1:
			p.popArray()
		case ruleAction2:
			p.addElement(buffer[begin:end])
		case ruleAction3:
			p.addElement(buffer[begin:end])
		case ruleAction4:
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
		/* 1 array <- <(ws* obracket Action0 ws* (item ws* (comma ws* item ws*)*)? cbracket Action1)> */
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
				{
					add(ruleAction0, position)
				}
			l8:
				{
					position9, tokenIndex9 := position, tokenIndex
					if !_rules[rulews]() {
						goto l9
					}
					goto l8
				l9:
					position, tokenIndex = position9, tokenIndex9
				}
				{
					position10, tokenIndex10 := position, tokenIndex
					if !_rules[ruleitem]() {
						goto l10
					}
				l12:
					{
						position13, tokenIndex13 := position, tokenIndex
						if !_rules[rulews]() {
							goto l13
						}
						goto l12
					l13:
						position, tokenIndex = position13, tokenIndex13
					}
				l14:
					{
						position15, tokenIndex15 := position, tokenIndex
						{
							position16 := position
							if buffer[position] != rune(',') {
								goto l15
							}
							position++
							add(rulecomma, position16)
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
						if !_rules[ruleitem]() {
							goto l15
						}
					l19:
						{
							position20, tokenIndex20 := position, tokenIndex
							if !_rules[rulews]() {
								goto l20
							}
							goto l19
						l20:
							position, tokenIndex = position20, tokenIndex20
						}
						goto l14
					l15:
						position, tokenIndex = position15, tokenIndex15
					}
					goto l11
				l10:
					position, tokenIndex = position10, tokenIndex10
				}
			l11:
				if !_rules[rulecbracket]() {
					goto l3
				}
				{
					add(ruleAction1, position)
				}
				add(rulearray, position4)
			}
			return true
		l3:
			position, tokenIndex = position3, tokenIndex3
			return false
		},
		/* 2 item <- <(array / string / (<value*> Action2))> */
		func() bool {
			{
				position23 := position
				{
					position24, tokenIndex24 := position, tokenIndex
					if !_rules[rulearray]() {
						goto l25
					}
					goto l24
				l25:
					position, tokenIndex = position24, tokenIndex24
					{
						position27 := position
						{
							position28, tokenIndex28 := position, tokenIndex
							{
								position30 := position
								if !_rules[ruleescdquote]() {
									goto l29
								}
								{
									position31 := position
								l32:
									{
										position33, tokenIndex33 := position, tokenIndex
										{
											position34, tokenIndex34 := position, tokenIndex
											if !_rules[ruletextdata]() {
												goto l35
											}
											goto l34
										l35:
											position, tokenIndex = position34, tokenIndex34
											if !_rules[rulesquote]() {
												goto l36
											}
											goto l34
										l36:
											position, tokenIndex = position34, tokenIndex34
											if !_rules[rulelf]() {
												goto l37
											}
											goto l34
										l37:
											position, tokenIndex = position34, tokenIndex34
											if !_rules[rulecr]() {
												goto l38
											}
											goto l34
										l38:
											position, tokenIndex = position34, tokenIndex34
											if !_rules[ruleobracket]() {
												goto l39
											}
											goto l34
										l39:
											position, tokenIndex = position34, tokenIndex34
											if !_rules[rulecbracket]() {
												goto l33
											}
										}
									l34:
										goto l32
									l33:
										position, tokenIndex = position33, tokenIndex33
									}
									add(rulePegText, position31)
								}
								if !_rules[ruleescdquote]() {
									goto l29
								}
								{
									add(ruleAction3, position)
								}
								add(ruledquote_string, position30)
							}
							goto l28
						l29:
							position, tokenIndex = position28, tokenIndex28
							{
								position41 := position
								if !_rules[rulesquote]() {
									goto l26
								}
								{
									position42 := position
								l43:
									{
										position44, tokenIndex44 := position, tokenIndex
										{
											position45, tokenIndex45 := position, tokenIndex
											if !_rules[ruletextdata]() {
												goto l46
											}
											goto l45
										l46:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[ruleescdquote]() {
												goto l47
											}
											goto l45
										l47:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[rulelf]() {
												goto l48
											}
											goto l45
										l48:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[rulecr]() {
												goto l49
											}
											goto l45
										l49:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[ruleobracket]() {
												goto l50
											}
											goto l45
										l50:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[rulecbracket]() {
												goto l44
											}
										}
									l45:
										goto l43
									l44:
										position, tokenIndex = position44, tokenIndex44
									}
									add(rulePegText, position42)
								}
								if !_rules[rulesquote]() {
									goto l26
								}
								{
									add(ruleAction4, position)
								}
								add(rulesquote_string, position41)
							}
						}
					l28:
						add(rulestring, position27)
					}
					goto l24
				l26:
					position, tokenIndex = position24, tokenIndex24
					{
						position52 := position
					l53:
						{
							position54, tokenIndex54 := position, tokenIndex
							{
								position55 := position
								{
									position56, tokenIndex56 := position, tokenIndex
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l57
									}
									position++
									goto l56
								l57:
									position, tokenIndex = position56, tokenIndex56
									if c := buffer[position]; c < rune('A') || c > rune('Z') {
										goto l58
									}
									position++
									goto l56
								l58:
									position, tokenIndex = position56, tokenIndex56
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l54
									}
									position++
								}
							l56:
								add(rulevalue, position55)
							}
							goto l53
						l54:
							position, tokenIndex = position54, tokenIndex54
						}
						add(rulePegText, position52)
					}
					{
						add(ruleAction2, position)
					}
				}
			l24:
				add(ruleitem, position23)
			}
			return true
		},
		/* 3 string <- <(dquote_string / squote_string)> */
		nil,
		/* 4 dquote_string <- <(escdquote <(textdata / squote / lf / cr / obracket / cbracket)*> escdquote Action3)> */
		nil,
		/* 5 squote_string <- <(squote <(textdata / escdquote / lf / cr / obracket / cbracket)*> squote Action4)> */
		nil,
		/* 6 ws <- <' '> */
		func() bool {
			position63, tokenIndex63 := position, tokenIndex
			{
				position64 := position
				if buffer[position] != rune(' ') {
					goto l63
				}
				position++
				add(rulews, position64)
			}
			return true
		l63:
			position, tokenIndex = position63, tokenIndex63
			return false
		},
		/* 7 comma <- <','> */
		nil,
		/* 8 lf <- <'\n'> */
		func() bool {
			position66, tokenIndex66 := position, tokenIndex
			{
				position67 := position
				if buffer[position] != rune('\n') {
					goto l66
				}
				position++
				add(rulelf, position67)
			}
			return true
		l66:
			position, tokenIndex = position66, tokenIndex66
			return false
		},
		/* 9 cr <- <'\r'> */
		func() bool {
			position68, tokenIndex68 := position, tokenIndex
			{
				position69 := position
				if buffer[position] != rune('\r') {
					goto l68
				}
				position++
				add(rulecr, position69)
			}
			return true
		l68:
			position, tokenIndex = position68, tokenIndex68
			return false
		},
		/* 10 escdquote <- <'"'> */
		func() bool {
			position70, tokenIndex70 := position, tokenIndex
			{
				position71 := position
				if buffer[position] != rune('"') {
					goto l70
				}
				position++
				add(ruleescdquote, position71)
			}
			return true
		l70:
			position, tokenIndex = position70, tokenIndex70
			return false
		},
		/* 11 squote <- <'\''> */
		func() bool {
			position72, tokenIndex72 := position, tokenIndex
			{
				position73 := position
				if buffer[position] != rune('\'') {
					goto l72
				}
				position++
				add(rulesquote, position73)
			}
			return true
		l72:
			position, tokenIndex = position72, tokenIndex72
			return false
		},
		/* 12 obracket <- <'['> */
		func() bool {
			position74, tokenIndex74 := position, tokenIndex
			{
				position75 := position
				if buffer[position] != rune('[') {
					goto l74
				}
				position++
				add(ruleobracket, position75)
			}
			return true
		l74:
			position, tokenIndex = position74, tokenIndex74
			return false
		},
		/* 13 cbracket <- <']'> */
		func() bool {
			position76, tokenIndex76 := position, tokenIndex
			{
				position77 := position
				if buffer[position] != rune(']') {
					goto l76
				}
				position++
				add(rulecbracket, position77)
			}
			return true
		l76:
			position, tokenIndex = position76, tokenIndex76
			return false
		},
		/* 14 value <- <([a-z] / [A-Z] / [0-9])> */
		nil,
		/* 15 textdata <- <([a-z] / [A-Z] / [0-9] / ' ' / '!' / '#' / '$' / '&' / '%' / '(' / ')' / '*' / '+' / '-' / '.' / '/' / ':' / ';' / [<->] / '?' / '\\' / '^' / '_' / '`' / '{' / '|' / '}' / '~')> */
		func() bool {
			position79, tokenIndex79 := position, tokenIndex
			{
				position80 := position
				{
					position81, tokenIndex81 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l82
					}
					position++
					goto l81
				l82:
					position, tokenIndex = position81, tokenIndex81
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l83
					}
					position++
					goto l81
				l83:
					position, tokenIndex = position81, tokenIndex81
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l84
					}
					position++
					goto l81
				l84:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune(' ') {
						goto l85
					}
					position++
					goto l81
				l85:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('!') {
						goto l86
					}
					position++
					goto l81
				l86:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('#') {
						goto l87
					}
					position++
					goto l81
				l87:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('$') {
						goto l88
					}
					position++
					goto l81
				l88:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('&') {
						goto l89
					}
					position++
					goto l81
				l89:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('%') {
						goto l90
					}
					position++
					goto l81
				l90:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('(') {
						goto l91
					}
					position++
					goto l81
				l91:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune(')') {
						goto l92
					}
					position++
					goto l81
				l92:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('*') {
						goto l93
					}
					position++
					goto l81
				l93:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('+') {
						goto l94
					}
					position++
					goto l81
				l94:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('-') {
						goto l95
					}
					position++
					goto l81
				l95:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('.') {
						goto l96
					}
					position++
					goto l81
				l96:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('/') {
						goto l97
					}
					position++
					goto l81
				l97:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune(':') {
						goto l98
					}
					position++
					goto l81
				l98:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune(';') {
						goto l99
					}
					position++
					goto l81
				l99:
					position, tokenIndex = position81, tokenIndex81
					if c := buffer[position]; c < rune('<') || c > rune('>') {
						goto l100
					}
					position++
					goto l81
				l100:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('?') {
						goto l101
					}
					position++
					goto l81
				l101:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('\\') {
						goto l102
					}
					position++
					goto l81
				l102:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('^') {
						goto l103
					}
					position++
					goto l81
				l103:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('_') {
						goto l104
					}
					position++
					goto l81
				l104:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('`') {
						goto l105
					}
					position++
					goto l81
				l105:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('{') {
						goto l106
					}
					position++
					goto l81
				l106:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('|') {
						goto l107
					}
					position++
					goto l81
				l107:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('}') {
						goto l108
					}
					position++
					goto l81
				l108:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('~') {
						goto l79
					}
					position++
				}
			l81:
				add(ruletextdata, position80)
			}
			return true
		l79:
			position, tokenIndex = position79, tokenIndex79
			return false
		},
		/* 17 Action0 <- <{ p.pushArray() }> */
		nil,
		/* 18 Action1 <- <{ p.popArray() }> */
		nil,
		nil,
		/* 20 Action2 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
		/* 21 Action3 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
		/* 22 Action4 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}

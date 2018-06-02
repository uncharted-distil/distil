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
	ruleesc
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
	"esc",
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
	rules  [24]func() bool
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
						if !_rules[rulecomma]() {
							goto l15
						}
					l16:
						{
							position17, tokenIndex17 := position, tokenIndex
							if !_rules[rulews]() {
								goto l17
							}
							goto l16
						l17:
							position, tokenIndex = position17, tokenIndex17
						}
						if !_rules[ruleitem]() {
							goto l15
						}
					l18:
						{
							position19, tokenIndex19 := position, tokenIndex
							if !_rules[rulews]() {
								goto l19
							}
							goto l18
						l19:
							position, tokenIndex = position19, tokenIndex19
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
				position22 := position
				{
					position23, tokenIndex23 := position, tokenIndex
					if !_rules[rulearray]() {
						goto l24
					}
					goto l23
				l24:
					position, tokenIndex = position23, tokenIndex23
					{
						position26 := position
						{
							position27, tokenIndex27 := position, tokenIndex
							{
								position29 := position
								if !_rules[ruleescdquote]() {
									goto l28
								}
								{
									position30 := position
								l31:
									{
										position32, tokenIndex32 := position, tokenIndex
										{
											position33, tokenIndex33 := position, tokenIndex
											if !_rules[ruletextdata]() {
												goto l34
											}
											goto l33
										l34:
											position, tokenIndex = position33, tokenIndex33
											if !_rules[rulesquote]() {
												goto l35
											}
											goto l33
										l35:
											position, tokenIndex = position33, tokenIndex33
											if !_rules[rulelf]() {
												goto l36
											}
											goto l33
										l36:
											position, tokenIndex = position33, tokenIndex33
											if !_rules[rulecr]() {
												goto l37
											}
											goto l33
										l37:
											position, tokenIndex = position33, tokenIndex33
											if !_rules[ruleobracket]() {
												goto l38
											}
											goto l33
										l38:
											position, tokenIndex = position33, tokenIndex33
											if !_rules[rulecbracket]() {
												goto l39
											}
											goto l33
										l39:
											position, tokenIndex = position33, tokenIndex33
											if !_rules[rulecomma]() {
												goto l32
											}
										}
									l33:
										goto l31
									l32:
										position, tokenIndex = position32, tokenIndex32
									}
									add(rulePegText, position30)
								}
								if !_rules[ruleescdquote]() {
									goto l28
								}
								{
									add(ruleAction3, position)
								}
								add(ruledquote_string, position29)
							}
							goto l27
						l28:
							position, tokenIndex = position27, tokenIndex27
							{
								position41 := position
								if !_rules[rulesquote]() {
									goto l25
								}
								{
									position42 := position
								l43:
									{
										position44, tokenIndex44 := position, tokenIndex
										{
											position45, tokenIndex45 := position, tokenIndex
											{
												position47 := position
												if buffer[position] != rune('\\') {
													goto l46
												}
												position++
												add(ruleesc, position47)
											}
											if !_rules[rulesquote]() {
												goto l46
											}
											goto l45
										l46:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[ruleescdquote]() {
												goto l48
											}
											goto l45
										l48:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[ruletextdata]() {
												goto l49
											}
											goto l45
										l49:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[rulelf]() {
												goto l50
											}
											goto l45
										l50:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[rulecr]() {
												goto l51
											}
											goto l45
										l51:
											position, tokenIndex = position45, tokenIndex45
											if !_rules[ruleobracket]() {
												goto l52
											}
											goto l45
										l52:
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
									goto l25
								}
								{
									add(ruleAction4, position)
								}
								add(rulesquote_string, position41)
							}
						}
					l27:
						add(rulestring, position26)
					}
					goto l23
				l25:
					position, tokenIndex = position23, tokenIndex23
					{
						position54 := position
					l55:
						{
							position56, tokenIndex56 := position, tokenIndex
							{
								position57 := position
								{
									position58, tokenIndex58 := position, tokenIndex
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l59
									}
									position++
									goto l58
								l59:
									position, tokenIndex = position58, tokenIndex58
									if c := buffer[position]; c < rune('A') || c > rune('Z') {
										goto l60
									}
									position++
									goto l58
								l60:
									position, tokenIndex = position58, tokenIndex58
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l56
									}
									position++
								}
							l58:
								add(rulevalue, position57)
							}
							goto l55
						l56:
							position, tokenIndex = position56, tokenIndex56
						}
						add(rulePegText, position54)
					}
					{
						add(ruleAction2, position)
					}
				}
			l23:
				add(ruleitem, position22)
			}
			return true
		},
		/* 3 string <- <(dquote_string / squote_string)> */
		nil,
		/* 4 dquote_string <- <(escdquote <(textdata / squote / lf / cr / obracket / cbracket / comma)*> escdquote Action3)> */
		nil,
		/* 5 squote_string <- <(squote <((esc squote) / escdquote / textdata / lf / cr / obracket / cbracket)*> squote Action4)> */
		nil,
		/* 6 ws <- <' '> */
		func() bool {
			position65, tokenIndex65 := position, tokenIndex
			{
				position66 := position
				if buffer[position] != rune(' ') {
					goto l65
				}
				position++
				add(rulews, position66)
			}
			return true
		l65:
			position, tokenIndex = position65, tokenIndex65
			return false
		},
		/* 7 comma <- <','> */
		func() bool {
			position67, tokenIndex67 := position, tokenIndex
			{
				position68 := position
				if buffer[position] != rune(',') {
					goto l67
				}
				position++
				add(rulecomma, position68)
			}
			return true
		l67:
			position, tokenIndex = position67, tokenIndex67
			return false
		},
		/* 8 lf <- <'\n'> */
		func() bool {
			position69, tokenIndex69 := position, tokenIndex
			{
				position70 := position
				if buffer[position] != rune('\n') {
					goto l69
				}
				position++
				add(rulelf, position70)
			}
			return true
		l69:
			position, tokenIndex = position69, tokenIndex69
			return false
		},
		/* 9 cr <- <'\r'> */
		func() bool {
			position71, tokenIndex71 := position, tokenIndex
			{
				position72 := position
				if buffer[position] != rune('\r') {
					goto l71
				}
				position++
				add(rulecr, position72)
			}
			return true
		l71:
			position, tokenIndex = position71, tokenIndex71
			return false
		},
		/* 10 esc <- <'\\'> */
		nil,
		/* 11 escdquote <- <'"'> */
		func() bool {
			position74, tokenIndex74 := position, tokenIndex
			{
				position75 := position
				if buffer[position] != rune('"') {
					goto l74
				}
				position++
				add(ruleescdquote, position75)
			}
			return true
		l74:
			position, tokenIndex = position74, tokenIndex74
			return false
		},
		/* 12 squote <- <'\''> */
		func() bool {
			position76, tokenIndex76 := position, tokenIndex
			{
				position77 := position
				if buffer[position] != rune('\'') {
					goto l76
				}
				position++
				add(rulesquote, position77)
			}
			return true
		l76:
			position, tokenIndex = position76, tokenIndex76
			return false
		},
		/* 13 obracket <- <'['> */
		func() bool {
			position78, tokenIndex78 := position, tokenIndex
			{
				position79 := position
				if buffer[position] != rune('[') {
					goto l78
				}
				position++
				add(ruleobracket, position79)
			}
			return true
		l78:
			position, tokenIndex = position78, tokenIndex78
			return false
		},
		/* 14 cbracket <- <']'> */
		func() bool {
			position80, tokenIndex80 := position, tokenIndex
			{
				position81 := position
				if buffer[position] != rune(']') {
					goto l80
				}
				position++
				add(rulecbracket, position81)
			}
			return true
		l80:
			position, tokenIndex = position80, tokenIndex80
			return false
		},
		/* 15 value <- <([a-z] / [A-Z] / [0-9])> */
		nil,
		/* 16 textdata <- <([a-z] / [A-Z] / [0-9] / ' ' / '!' / '#' / '$' / '&' / '%' / '(' / ')' / '*' / '+' / '-' / '.' / '/' / ':' / ';' / [<->] / '?' / '\\' / '^' / '_' / '`' / '{' / '|' / '}' / '~')> */
		func() bool {
			position83, tokenIndex83 := position, tokenIndex
			{
				position84 := position
				{
					position85, tokenIndex85 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l86
					}
					position++
					goto l85
				l86:
					position, tokenIndex = position85, tokenIndex85
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l87
					}
					position++
					goto l85
				l87:
					position, tokenIndex = position85, tokenIndex85
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l88
					}
					position++
					goto l85
				l88:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune(' ') {
						goto l89
					}
					position++
					goto l85
				l89:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('!') {
						goto l90
					}
					position++
					goto l85
				l90:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('#') {
						goto l91
					}
					position++
					goto l85
				l91:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('$') {
						goto l92
					}
					position++
					goto l85
				l92:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('&') {
						goto l93
					}
					position++
					goto l85
				l93:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('%') {
						goto l94
					}
					position++
					goto l85
				l94:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('(') {
						goto l95
					}
					position++
					goto l85
				l95:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune(')') {
						goto l96
					}
					position++
					goto l85
				l96:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('*') {
						goto l97
					}
					position++
					goto l85
				l97:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('+') {
						goto l98
					}
					position++
					goto l85
				l98:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('-') {
						goto l99
					}
					position++
					goto l85
				l99:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('.') {
						goto l100
					}
					position++
					goto l85
				l100:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('/') {
						goto l101
					}
					position++
					goto l85
				l101:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune(':') {
						goto l102
					}
					position++
					goto l85
				l102:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune(';') {
						goto l103
					}
					position++
					goto l85
				l103:
					position, tokenIndex = position85, tokenIndex85
					if c := buffer[position]; c < rune('<') || c > rune('>') {
						goto l104
					}
					position++
					goto l85
				l104:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('?') {
						goto l105
					}
					position++
					goto l85
				l105:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('\\') {
						goto l106
					}
					position++
					goto l85
				l106:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('^') {
						goto l107
					}
					position++
					goto l85
				l107:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('_') {
						goto l108
					}
					position++
					goto l85
				l108:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('`') {
						goto l109
					}
					position++
					goto l85
				l109:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('{') {
						goto l110
					}
					position++
					goto l85
				l110:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('|') {
						goto l111
					}
					position++
					goto l85
				l111:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('}') {
						goto l112
					}
					position++
					goto l85
				l112:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('~') {
						goto l83
					}
					position++
				}
			l85:
				add(ruletextdata, position84)
			}
			return true
		l83:
			position, tokenIndex = position83, tokenIndex83
			return false
		},
		/* 18 Action0 <- <{ p.pushArray() }> */
		nil,
		/* 19 Action1 <- <{ p.popArray() }> */
		nil,
		nil,
		/* 21 Action2 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
		/* 22 Action3 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
		/* 23 Action4 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}

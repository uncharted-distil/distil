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
				{
					position2 := position
				l3:
					{
						position4, tokenIndex4 := position, tokenIndex
						if !_rules[rulews]() {
							goto l4
						}
						goto l3
					l4:
						position, tokenIndex = position4, tokenIndex4
					}
					if !_rules[ruleobracket]() {
						goto l0
					}
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
					{
						position7, tokenIndex7 := position, tokenIndex
						if !_rules[ruleitem]() {
							goto l7
						}
					l9:
						{
							position10, tokenIndex10 := position, tokenIndex
							if !_rules[rulews]() {
								goto l10
							}
							goto l9
						l10:
							position, tokenIndex = position10, tokenIndex10
						}
					l11:
						{
							position12, tokenIndex12 := position, tokenIndex
							{
								position13 := position
								if buffer[position] != rune(',') {
									goto l12
								}
								position++
								add(rulecomma, position13)
							}
						l14:
							{
								position15, tokenIndex15 := position, tokenIndex
								if !_rules[rulews]() {
									goto l15
								}
								goto l14
							l15:
								position, tokenIndex = position15, tokenIndex15
							}
							if !_rules[ruleitem]() {
								goto l12
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
							goto l11
						l12:
							position, tokenIndex = position12, tokenIndex12
						}
						goto l8
					l7:
						position, tokenIndex = position7, tokenIndex7
					}
				l8:
					if !_rules[rulecbracket]() {
						goto l0
					}
					add(rulearray, position2)
				}
				{
					position18, tokenIndex18 := position, tokenIndex
					if !matchDot() {
						goto l18
					}
					goto l0
				l18:
					position, tokenIndex = position18, tokenIndex18
				}
				add(ruleComplexField, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 array <- <(ws* obracket ws* (item ws* (comma ws* item ws*)*)? cbracket)> */
		nil,
		/* 2 item <- <(string / (<value*> Action0))> */
		func() bool {
			{
				position21 := position
				{
					position22, tokenIndex22 := position, tokenIndex
					{
						position24 := position
						{
							position25, tokenIndex25 := position, tokenIndex
							{
								position27 := position
								if !_rules[ruleescdquote]() {
									goto l26
								}
								{
									position28 := position
								l29:
									{
										position30, tokenIndex30 := position, tokenIndex
										{
											position31, tokenIndex31 := position, tokenIndex
											if !_rules[ruletextdata]() {
												goto l32
											}
											goto l31
										l32:
											position, tokenIndex = position31, tokenIndex31
											if !_rules[rulesquote]() {
												goto l33
											}
											goto l31
										l33:
											position, tokenIndex = position31, tokenIndex31
											{
												position35 := position
												if buffer[position] != rune('\n') {
													goto l34
												}
												position++
												add(rulelf, position35)
											}
											goto l31
										l34:
											position, tokenIndex = position31, tokenIndex31
											{
												position37 := position
												if buffer[position] != rune('\r') {
													goto l36
												}
												position++
												add(rulecr, position37)
											}
											goto l31
										l36:
											position, tokenIndex = position31, tokenIndex31
											if !_rules[ruleobracket]() {
												goto l38
											}
											goto l31
										l38:
											position, tokenIndex = position31, tokenIndex31
											if !_rules[rulecbracket]() {
												goto l30
											}
										}
									l31:
										goto l29
									l30:
										position, tokenIndex = position30, tokenIndex30
									}
									add(rulePegText, position28)
								}
								if !_rules[ruleescdquote]() {
									goto l26
								}
								{
									add(ruleAction1, position)
								}
								add(ruleescaped_string, position27)
							}
							goto l25
						l26:
							position, tokenIndex = position25, tokenIndex25
							{
								position40 := position
								if !_rules[rulesquote]() {
									goto l23
								}
								{
									position41 := position
								l42:
									{
										position43, tokenIndex43 := position, tokenIndex
										if !_rules[ruletextdata]() {
											goto l43
										}
										goto l42
									l43:
										position, tokenIndex = position43, tokenIndex43
									}
									add(rulePegText, position41)
								}
								if !_rules[rulesquote]() {
									goto l23
								}
								{
									add(ruleAction2, position)
								}
								add(rulenonescaped_string, position40)
							}
						}
					l25:
						add(rulestring, position24)
					}
					goto l22
				l23:
					position, tokenIndex = position22, tokenIndex22
					{
						position45 := position
					l46:
						{
							position47, tokenIndex47 := position, tokenIndex
							{
								position48 := position
								{
									position49, tokenIndex49 := position, tokenIndex
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l50
									}
									position++
									goto l49
								l50:
									position, tokenIndex = position49, tokenIndex49
									if c := buffer[position]; c < rune('A') || c > rune('Z') {
										goto l51
									}
									position++
									goto l49
								l51:
									position, tokenIndex = position49, tokenIndex49
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l47
									}
									position++
								}
							l49:
								add(rulevalue, position48)
							}
							goto l46
						l47:
							position, tokenIndex = position47, tokenIndex47
						}
						add(rulePegText, position45)
					}
					{
						add(ruleAction0, position)
					}
				}
			l22:
				add(ruleitem, position21)
			}
			return true
		},
		/* 3 string <- <(escaped_string / nonescaped_string)> */
		nil,
		/* 4 escaped_string <- <(escdquote <(textdata / squote / lf / cr / obracket / cbracket)*> escdquote Action1)> */
		nil,
		/* 5 nonescaped_string <- <(squote <textdata*> squote Action2)> */
		nil,
		/* 6 ws <- <' '> */
		func() bool {
			position56, tokenIndex56 := position, tokenIndex
			{
				position57 := position
				if buffer[position] != rune(' ') {
					goto l56
				}
				position++
				add(rulews, position57)
			}
			return true
		l56:
			position, tokenIndex = position56, tokenIndex56
			return false
		},
		/* 7 comma <- <','> */
		nil,
		/* 8 lf <- <'\n'> */
		nil,
		/* 9 cr <- <'\r'> */
		nil,
		/* 10 escdquote <- <'"'> */
		func() bool {
			position61, tokenIndex61 := position, tokenIndex
			{
				position62 := position
				if buffer[position] != rune('"') {
					goto l61
				}
				position++
				add(ruleescdquote, position62)
			}
			return true
		l61:
			position, tokenIndex = position61, tokenIndex61
			return false
		},
		/* 11 squote <- <'\''> */
		func() bool {
			position63, tokenIndex63 := position, tokenIndex
			{
				position64 := position
				if buffer[position] != rune('\'') {
					goto l63
				}
				position++
				add(rulesquote, position64)
			}
			return true
		l63:
			position, tokenIndex = position63, tokenIndex63
			return false
		},
		/* 12 obracket <- <'['> */
		func() bool {
			position65, tokenIndex65 := position, tokenIndex
			{
				position66 := position
				if buffer[position] != rune('[') {
					goto l65
				}
				position++
				add(ruleobracket, position66)
			}
			return true
		l65:
			position, tokenIndex = position65, tokenIndex65
			return false
		},
		/* 13 cbracket <- <']'> */
		func() bool {
			position67, tokenIndex67 := position, tokenIndex
			{
				position68 := position
				if buffer[position] != rune(']') {
					goto l67
				}
				position++
				add(rulecbracket, position68)
			}
			return true
		l67:
			position, tokenIndex = position67, tokenIndex67
			return false
		},
		/* 14 value <- <([a-z] / [A-Z] / [0-9])> */
		nil,
		/* 15 textdata <- <([a-z] / [A-Z] / [0-9] / ' ' / '!' / '#' / '$' / '&' / '%' / '(' / ')' / '*' / '+' / '-' / '.' / '/' / ':' / ';' / [<->] / '?' / '\\' / '^' / '_' / '`' / '{' / '|' / '}' / '~')> */
		func() bool {
			position70, tokenIndex70 := position, tokenIndex
			{
				position71 := position
				{
					position72, tokenIndex72 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l73
					}
					position++
					goto l72
				l73:
					position, tokenIndex = position72, tokenIndex72
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l74
					}
					position++
					goto l72
				l74:
					position, tokenIndex = position72, tokenIndex72
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l75
					}
					position++
					goto l72
				l75:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune(' ') {
						goto l76
					}
					position++
					goto l72
				l76:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('!') {
						goto l77
					}
					position++
					goto l72
				l77:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('#') {
						goto l78
					}
					position++
					goto l72
				l78:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('$') {
						goto l79
					}
					position++
					goto l72
				l79:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('&') {
						goto l80
					}
					position++
					goto l72
				l80:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('%') {
						goto l81
					}
					position++
					goto l72
				l81:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('(') {
						goto l82
					}
					position++
					goto l72
				l82:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune(')') {
						goto l83
					}
					position++
					goto l72
				l83:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('*') {
						goto l84
					}
					position++
					goto l72
				l84:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('+') {
						goto l85
					}
					position++
					goto l72
				l85:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('-') {
						goto l86
					}
					position++
					goto l72
				l86:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('.') {
						goto l87
					}
					position++
					goto l72
				l87:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('/') {
						goto l88
					}
					position++
					goto l72
				l88:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune(':') {
						goto l89
					}
					position++
					goto l72
				l89:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune(';') {
						goto l90
					}
					position++
					goto l72
				l90:
					position, tokenIndex = position72, tokenIndex72
					if c := buffer[position]; c < rune('<') || c > rune('>') {
						goto l91
					}
					position++
					goto l72
				l91:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('?') {
						goto l92
					}
					position++
					goto l72
				l92:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('\\') {
						goto l93
					}
					position++
					goto l72
				l93:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('^') {
						goto l94
					}
					position++
					goto l72
				l94:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('_') {
						goto l95
					}
					position++
					goto l72
				l95:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('`') {
						goto l96
					}
					position++
					goto l72
				l96:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('{') {
						goto l97
					}
					position++
					goto l72
				l97:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('|') {
						goto l98
					}
					position++
					goto l72
				l98:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('}') {
						goto l99
					}
					position++
					goto l72
				l99:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('~') {
						goto l70
					}
					position++
				}
			l72:
				add(ruletextdata, position71)
			}
			return true
		l70:
			position, tokenIndex = position70, tokenIndex70
			return false
		},
		nil,
		/* 18 Action0 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
		/* 19 Action1 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
		/* 20 Action2 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}

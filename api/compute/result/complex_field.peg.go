package result

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
	rulevalue
	rulews
	rulecomma
	rulelf
	rulecr
	ruleescdquote
	ruleescsquote
	rulesquote
	ruleobracket
	rulecbracket
	ruleoparen
	rulecparen
	rulenumber
	rulenegative
	ruledecimal_point
	ruletextdata
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	rulePegText
	ruleAction4
	ruleAction5
	ruleAction6
)

var rul3s = [...]string{
	"Unknown",
	"ComplexField",
	"array",
	"item",
	"string",
	"dquote_string",
	"squote_string",
	"value",
	"ws",
	"comma",
	"lf",
	"cr",
	"escdquote",
	"escsquote",
	"squote",
	"obracket",
	"cbracket",
	"oparen",
	"cparen",
	"number",
	"negative",
	"decimal_point",
	"textdata",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"PegText",
	"Action4",
	"Action5",
	"Action6",
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
	rules  [31]func() bool
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
			p.pushArray()
		case ruleAction3:
			p.popArray()
		case ruleAction4:
			p.addElement(buffer[begin:end])
		case ruleAction5:
			p.addElement(buffer[begin:end])
		case ruleAction6:
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
		/* 1 array <- <((ws* obracket Action0 ws* (item ws* (comma ws* item ws*)*)? cbracket Action1) / (ws* oparen Action2 ws* (item ws* (comma ws* item ws*)*)? cparen Action3))> */
		func() bool {
			position3, tokenIndex3 := position, tokenIndex
			{
				position4 := position
				{
					position5, tokenIndex5 := position, tokenIndex
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
					if !_rules[ruleobracket]() {
						goto l6
					}
					{
						add(ruleAction0, position)
					}
				l10:
					{
						position11, tokenIndex11 := position, tokenIndex
						if !_rules[rulews]() {
							goto l11
						}
						goto l10
					l11:
						position, tokenIndex = position11, tokenIndex11
					}
					{
						position12, tokenIndex12 := position, tokenIndex
						if !_rules[ruleitem]() {
							goto l12
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
					l16:
						{
							position17, tokenIndex17 := position, tokenIndex
							if !_rules[rulecomma]() {
								goto l17
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
							if !_rules[ruleitem]() {
								goto l17
							}
						l20:
							{
								position21, tokenIndex21 := position, tokenIndex
								if !_rules[rulews]() {
									goto l21
								}
								goto l20
							l21:
								position, tokenIndex = position21, tokenIndex21
							}
							goto l16
						l17:
							position, tokenIndex = position17, tokenIndex17
						}
						goto l13
					l12:
						position, tokenIndex = position12, tokenIndex12
					}
				l13:
					if !_rules[rulecbracket]() {
						goto l6
					}
					{
						add(ruleAction1, position)
					}
					goto l5
				l6:
					position, tokenIndex = position5, tokenIndex5
				l23:
					{
						position24, tokenIndex24 := position, tokenIndex
						if !_rules[rulews]() {
							goto l24
						}
						goto l23
					l24:
						position, tokenIndex = position24, tokenIndex24
					}
					if !_rules[ruleoparen]() {
						goto l3
					}
					{
						add(ruleAction2, position)
					}
				l26:
					{
						position27, tokenIndex27 := position, tokenIndex
						if !_rules[rulews]() {
							goto l27
						}
						goto l26
					l27:
						position, tokenIndex = position27, tokenIndex27
					}
					{
						position28, tokenIndex28 := position, tokenIndex
						if !_rules[ruleitem]() {
							goto l28
						}
					l30:
						{
							position31, tokenIndex31 := position, tokenIndex
							if !_rules[rulews]() {
								goto l31
							}
							goto l30
						l31:
							position, tokenIndex = position31, tokenIndex31
						}
					l32:
						{
							position33, tokenIndex33 := position, tokenIndex
							if !_rules[rulecomma]() {
								goto l33
							}
						l34:
							{
								position35, tokenIndex35 := position, tokenIndex
								if !_rules[rulews]() {
									goto l35
								}
								goto l34
							l35:
								position, tokenIndex = position35, tokenIndex35
							}
							if !_rules[ruleitem]() {
								goto l33
							}
						l36:
							{
								position37, tokenIndex37 := position, tokenIndex
								if !_rules[rulews]() {
									goto l37
								}
								goto l36
							l37:
								position, tokenIndex = position37, tokenIndex37
							}
							goto l32
						l33:
							position, tokenIndex = position33, tokenIndex33
						}
						goto l29
					l28:
						position, tokenIndex = position28, tokenIndex28
					}
				l29:
					if !_rules[rulecparen]() {
						goto l3
					}
					{
						add(ruleAction3, position)
					}
				}
			l5:
				add(rulearray, position4)
			}
			return true
		l3:
			position, tokenIndex = position3, tokenIndex3
			return false
		},
		/* 2 item <- <(array / string / (<value*> Action4))> */
		func() bool {
			{
				position40 := position
				{
					position41, tokenIndex41 := position, tokenIndex
					if !_rules[rulearray]() {
						goto l42
					}
					goto l41
				l42:
					position, tokenIndex = position41, tokenIndex41
					{
						position44 := position
						{
							position45, tokenIndex45 := position, tokenIndex
							{
								position47 := position
								if !_rules[ruleescdquote]() {
									goto l46
								}
								{
									position48 := position
								l49:
									{
										position50, tokenIndex50 := position, tokenIndex
										{
											position51, tokenIndex51 := position, tokenIndex
											if !_rules[ruletextdata]() {
												goto l52
											}
											goto l51
										l52:
											position, tokenIndex = position51, tokenIndex51
											if !_rules[rulesquote]() {
												goto l53
											}
											goto l51
										l53:
											position, tokenIndex = position51, tokenIndex51
											if !_rules[rulelf]() {
												goto l54
											}
											goto l51
										l54:
											position, tokenIndex = position51, tokenIndex51
											if !_rules[rulecr]() {
												goto l55
											}
											goto l51
										l55:
											position, tokenIndex = position51, tokenIndex51
											if !_rules[ruleobracket]() {
												goto l56
											}
											goto l51
										l56:
											position, tokenIndex = position51, tokenIndex51
											if !_rules[rulecbracket]() {
												goto l57
											}
											goto l51
										l57:
											position, tokenIndex = position51, tokenIndex51
											if !_rules[ruleoparen]() {
												goto l58
											}
											goto l51
										l58:
											position, tokenIndex = position51, tokenIndex51
											if !_rules[rulecparen]() {
												goto l59
											}
											goto l51
										l59:
											position, tokenIndex = position51, tokenIndex51
											if !_rules[rulecomma]() {
												goto l50
											}
										}
									l51:
										goto l49
									l50:
										position, tokenIndex = position50, tokenIndex50
									}
									add(rulePegText, position48)
								}
								if !_rules[ruleescdquote]() {
									goto l46
								}
								{
									add(ruleAction5, position)
								}
								add(ruledquote_string, position47)
							}
							goto l45
						l46:
							position, tokenIndex = position45, tokenIndex45
							{
								position61 := position
								if !_rules[rulesquote]() {
									goto l43
								}
								{
									position62 := position
								l63:
									{
										position64, tokenIndex64 := position, tokenIndex
										{
											position65, tokenIndex65 := position, tokenIndex
											{
												position67 := position
												if buffer[position] != rune('\\') {
													goto l66
												}
												position++
												if buffer[position] != rune('\'') {
													goto l66
												}
												position++
												add(ruleescsquote, position67)
											}
											goto l65
										l66:
											position, tokenIndex = position65, tokenIndex65
											if !_rules[ruleescdquote]() {
												goto l68
											}
											goto l65
										l68:
											position, tokenIndex = position65, tokenIndex65
											if !_rules[ruletextdata]() {
												goto l69
											}
											goto l65
										l69:
											position, tokenIndex = position65, tokenIndex65
											if !_rules[rulelf]() {
												goto l70
											}
											goto l65
										l70:
											position, tokenIndex = position65, tokenIndex65
											if !_rules[rulecr]() {
												goto l71
											}
											goto l65
										l71:
											position, tokenIndex = position65, tokenIndex65
											if !_rules[ruleobracket]() {
												goto l72
											}
											goto l65
										l72:
											position, tokenIndex = position65, tokenIndex65
											if !_rules[rulecbracket]() {
												goto l73
											}
											goto l65
										l73:
											position, tokenIndex = position65, tokenIndex65
											if !_rules[ruleoparen]() {
												goto l74
											}
											goto l65
										l74:
											position, tokenIndex = position65, tokenIndex65
											if !_rules[rulecparen]() {
												goto l64
											}
										}
									l65:
										goto l63
									l64:
										position, tokenIndex = position64, tokenIndex64
									}
									add(rulePegText, position62)
								}
								if !_rules[rulesquote]() {
									goto l43
								}
								{
									add(ruleAction6, position)
								}
								add(rulesquote_string, position61)
							}
						}
					l45:
						add(rulestring, position44)
					}
					goto l41
				l43:
					position, tokenIndex = position41, tokenIndex41
					{
						position76 := position
					l77:
						{
							position78, tokenIndex78 := position, tokenIndex
							{
								position79 := position
								{
									position80, tokenIndex80 := position, tokenIndex
									{
										position82 := position
										if buffer[position] != rune('-') {
											goto l80
										}
										position++
										add(rulenegative, position82)
									}
									goto l81
								l80:
									position, tokenIndex = position80, tokenIndex80
								}
							l81:
								if !_rules[rulenumber]() {
									goto l78
								}
							l83:
								{
									position84, tokenIndex84 := position, tokenIndex
									if !_rules[rulenumber]() {
										goto l84
									}
									goto l83
								l84:
									position, tokenIndex = position84, tokenIndex84
								}
								{
									position85, tokenIndex85 := position, tokenIndex
									{
										position87 := position
										if buffer[position] != rune('.') {
											goto l85
										}
										position++
										add(ruledecimal_point, position87)
									}
									if !_rules[rulenumber]() {
										goto l85
									}
								l88:
									{
										position89, tokenIndex89 := position, tokenIndex
										if !_rules[rulenumber]() {
											goto l89
										}
										goto l88
									l89:
										position, tokenIndex = position89, tokenIndex89
									}
									goto l86
								l85:
									position, tokenIndex = position85, tokenIndex85
								}
							l86:
								add(rulevalue, position79)
							}
							goto l77
						l78:
							position, tokenIndex = position78, tokenIndex78
						}
						add(rulePegText, position76)
					}
					{
						add(ruleAction4, position)
					}
				}
			l41:
				add(ruleitem, position40)
			}
			return true
		},
		/* 3 string <- <(dquote_string / squote_string)> */
		nil,
		/* 4 dquote_string <- <(escdquote <(textdata / squote / lf / cr / obracket / cbracket / oparen / cparen / comma)*> escdquote Action5)> */
		nil,
		/* 5 squote_string <- <(squote <(escsquote / escdquote / textdata / lf / cr / obracket / cbracket / oparen / cparen)*> squote Action6)> */
		nil,
		/* 6 value <- <(negative? number+ (decimal_point number+)?)> */
		nil,
		/* 7 ws <- <' '> */
		func() bool {
			position95, tokenIndex95 := position, tokenIndex
			{
				position96 := position
				if buffer[position] != rune(' ') {
					goto l95
				}
				position++
				add(rulews, position96)
			}
			return true
		l95:
			position, tokenIndex = position95, tokenIndex95
			return false
		},
		/* 8 comma <- <','> */
		func() bool {
			position97, tokenIndex97 := position, tokenIndex
			{
				position98 := position
				if buffer[position] != rune(',') {
					goto l97
				}
				position++
				add(rulecomma, position98)
			}
			return true
		l97:
			position, tokenIndex = position97, tokenIndex97
			return false
		},
		/* 9 lf <- <'\n'> */
		func() bool {
			position99, tokenIndex99 := position, tokenIndex
			{
				position100 := position
				if buffer[position] != rune('\n') {
					goto l99
				}
				position++
				add(rulelf, position100)
			}
			return true
		l99:
			position, tokenIndex = position99, tokenIndex99
			return false
		},
		/* 10 cr <- <'\r'> */
		func() bool {
			position101, tokenIndex101 := position, tokenIndex
			{
				position102 := position
				if buffer[position] != rune('\r') {
					goto l101
				}
				position++
				add(rulecr, position102)
			}
			return true
		l101:
			position, tokenIndex = position101, tokenIndex101
			return false
		},
		/* 11 escdquote <- <'"'> */
		func() bool {
			position103, tokenIndex103 := position, tokenIndex
			{
				position104 := position
				if buffer[position] != rune('"') {
					goto l103
				}
				position++
				add(ruleescdquote, position104)
			}
			return true
		l103:
			position, tokenIndex = position103, tokenIndex103
			return false
		},
		/* 12 escsquote <- <('\\' '\'')> */
		nil,
		/* 13 squote <- <'\''> */
		func() bool {
			position106, tokenIndex106 := position, tokenIndex
			{
				position107 := position
				if buffer[position] != rune('\'') {
					goto l106
				}
				position++
				add(rulesquote, position107)
			}
			return true
		l106:
			position, tokenIndex = position106, tokenIndex106
			return false
		},
		/* 14 obracket <- <'['> */
		func() bool {
			position108, tokenIndex108 := position, tokenIndex
			{
				position109 := position
				if buffer[position] != rune('[') {
					goto l108
				}
				position++
				add(ruleobracket, position109)
			}
			return true
		l108:
			position, tokenIndex = position108, tokenIndex108
			return false
		},
		/* 15 cbracket <- <']'> */
		func() bool {
			position110, tokenIndex110 := position, tokenIndex
			{
				position111 := position
				if buffer[position] != rune(']') {
					goto l110
				}
				position++
				add(rulecbracket, position111)
			}
			return true
		l110:
			position, tokenIndex = position110, tokenIndex110
			return false
		},
		/* 16 oparen <- <'('> */
		func() bool {
			position112, tokenIndex112 := position, tokenIndex
			{
				position113 := position
				if buffer[position] != rune('(') {
					goto l112
				}
				position++
				add(ruleoparen, position113)
			}
			return true
		l112:
			position, tokenIndex = position112, tokenIndex112
			return false
		},
		/* 17 cparen <- <')'> */
		func() bool {
			position114, tokenIndex114 := position, tokenIndex
			{
				position115 := position
				if buffer[position] != rune(')') {
					goto l114
				}
				position++
				add(rulecparen, position115)
			}
			return true
		l114:
			position, tokenIndex = position114, tokenIndex114
			return false
		},
		/* 18 number <- <([a-z] / [A-Z] / [0-9])> */
		func() bool {
			position116, tokenIndex116 := position, tokenIndex
			{
				position117 := position
				{
					position118, tokenIndex118 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l119
					}
					position++
					goto l118
				l119:
					position, tokenIndex = position118, tokenIndex118
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l120
					}
					position++
					goto l118
				l120:
					position, tokenIndex = position118, tokenIndex118
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l116
					}
					position++
				}
			l118:
				add(rulenumber, position117)
			}
			return true
		l116:
			position, tokenIndex = position116, tokenIndex116
			return false
		},
		/* 19 negative <- <'-'> */
		nil,
		/* 20 decimal_point <- <'.'> */
		nil,
		/* 21 textdata <- <([a-z] / [A-Z] / [0-9] / ' ' / '!' / '#' / '$' / '&' / '%' / '*' / '+' / '-' / '.' / '/' / ':' / ';' / [<->] / '?' / '\\' / '^' / '_' / '`' / '{' / '|' / '}' / '~')> */
		func() bool {
			position123, tokenIndex123 := position, tokenIndex
			{
				position124 := position
				{
					position125, tokenIndex125 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l126
					}
					position++
					goto l125
				l126:
					position, tokenIndex = position125, tokenIndex125
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l127
					}
					position++
					goto l125
				l127:
					position, tokenIndex = position125, tokenIndex125
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l128
					}
					position++
					goto l125
				l128:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune(' ') {
						goto l129
					}
					position++
					goto l125
				l129:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('!') {
						goto l130
					}
					position++
					goto l125
				l130:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('#') {
						goto l131
					}
					position++
					goto l125
				l131:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('$') {
						goto l132
					}
					position++
					goto l125
				l132:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('&') {
						goto l133
					}
					position++
					goto l125
				l133:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('%') {
						goto l134
					}
					position++
					goto l125
				l134:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('*') {
						goto l135
					}
					position++
					goto l125
				l135:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('+') {
						goto l136
					}
					position++
					goto l125
				l136:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('-') {
						goto l137
					}
					position++
					goto l125
				l137:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('.') {
						goto l138
					}
					position++
					goto l125
				l138:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('/') {
						goto l139
					}
					position++
					goto l125
				l139:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune(':') {
						goto l140
					}
					position++
					goto l125
				l140:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune(';') {
						goto l141
					}
					position++
					goto l125
				l141:
					position, tokenIndex = position125, tokenIndex125
					if c := buffer[position]; c < rune('<') || c > rune('>') {
						goto l142
					}
					position++
					goto l125
				l142:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('?') {
						goto l143
					}
					position++
					goto l125
				l143:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('\\') {
						goto l144
					}
					position++
					goto l125
				l144:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('^') {
						goto l145
					}
					position++
					goto l125
				l145:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('_') {
						goto l146
					}
					position++
					goto l125
				l146:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('`') {
						goto l147
					}
					position++
					goto l125
				l147:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('{') {
						goto l148
					}
					position++
					goto l125
				l148:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('|') {
						goto l149
					}
					position++
					goto l125
				l149:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('}') {
						goto l150
					}
					position++
					goto l125
				l150:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('~') {
						goto l123
					}
					position++
				}
			l125:
				add(ruletextdata, position124)
			}
			return true
		l123:
			position, tokenIndex = position123, tokenIndex123
			return false
		},
		/* 23 Action0 <- <{ p.pushArray() }> */
		nil,
		/* 24 Action1 <- <{ p.popArray() }> */
		nil,
		/* 25 Action2 <- <{ p.pushArray() }> */
		nil,
		/* 26 Action3 <- <{ p.popArray() }> */
		nil,
		nil,
		/* 28 Action4 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
		/* 29 Action5 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
		/* 30 Action6 <- <{ p.addElement(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}

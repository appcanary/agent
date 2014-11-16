package gemfile

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleGemfile
	ruleSource
	ruleGit
	ruleGem
	rulePlatforms
	ruleDependencies
	ruleOption
	ruleSpecs
	ruleSpec
	ruleSpecDep
	ruleDependency
	ruleGemVersion
	ruleGemName
	ruleVersion
	ruleConstraint
	ruleVersionOp
	rulePlatform
	ruleEq
	ruleNeq
	ruleLeq
	ruleLt
	ruleGeq
	ruleGt
	ruleTwiddleWakka
	ruleSpace
	ruleSpaces
	ruleIndent2
	ruleIndent4
	ruleIndent6
	ruleEndOfLine
	ruleNotWhitespace
	ruleLineEnd
	ruleEndOfFile
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
	ruleAction3
	ruleAction4

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"Gemfile",
	"Source",
	"Git",
	"Gem",
	"Platforms",
	"Dependencies",
	"Option",
	"Specs",
	"Spec",
	"SpecDep",
	"Dependency",
	"GemVersion",
	"GemName",
	"Version",
	"Constraint",
	"VersionOp",
	"Platform",
	"Eq",
	"Neq",
	"Leq",
	"Lt",
	"Geq",
	"Gt",
	"TwiddleWakka",
	"Space",
	"Spaces",
	"Indent2",
	"Indent4",
	"Indent6",
	"EndOfLine",
	"NotWhitespace",
	"LineEnd",
	"EndOfFile",
	"Action0",
	"Action1",
	"Action2",
	"PegText",
	"Action3",
	"Action4",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(buffer[node.begin:node.end]))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	pegRule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
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
	return stack.node
}

func (t *tokens16) PreOrder() (<-chan state16, [][]token16) {
	s, ordered := make(chan state16, 6), t.Order()
	go func() {
		var states [8]state16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token16{pegRule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
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
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token32{pegRule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type GemfileParser struct {
	Gemfile
	ParserState

	Buffer string
	buffer []rune
	rules  [40]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
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
	p *GemfileParser
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *GemfileParser) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *GemfileParser) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *GemfileParser) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
		case ruleAction0:
			p.setState(ParsingSpec)
		case ruleAction1:
			p.setState(ParsingSpecDep)
		case ruleAction2:
			p.setState(ParsingDependency)
		case ruleAction3:
			p.addGem(buffer[begin:end])
		case ruleAction4:
			p.addVersion(buffer[begin:end])

		}
	}
}

func (p *GemfileParser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
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
		/* 0 Gemfile <- <(Source* Platforms Dependencies EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					if !_rules[ruleSource]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position3, tokenIndex3, depth3
				}
				if !_rules[rulePlatforms]() {
					goto l0
				}
				if !_rules[ruleDependencies]() {
					goto l0
				}
				if !_rules[ruleEndOfFile]() {
					goto l0
				}
				depth--
				add(ruleGemfile, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Source <- <(Git / Gem)> */
		func() bool {
			position4, tokenIndex4, depth4 := position, tokenIndex, depth
			{
				position5 := position
				depth++
				{
					position6, tokenIndex6, depth6 := position, tokenIndex, depth
					if !_rules[ruleGit]() {
						goto l7
					}
					goto l6
				l7:
					position, tokenIndex, depth = position6, tokenIndex6, depth6
					if !_rules[ruleGem]() {
						goto l4
					}
				}
			l6:
				depth--
				add(ruleSource, position5)
			}
			return true
		l4:
			position, tokenIndex, depth = position4, tokenIndex4, depth4
			return false
		},
		/* 2 Git <- <('G' 'I' 'T' LineEnd Option* Specs LineEnd)> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				if buffer[position] != rune('G') {
					goto l8
				}
				position++
				if buffer[position] != rune('I') {
					goto l8
				}
				position++
				if buffer[position] != rune('T') {
					goto l8
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l8
				}
			l10:
				{
					position11, tokenIndex11, depth11 := position, tokenIndex, depth
					if !_rules[ruleOption]() {
						goto l11
					}
					goto l10
				l11:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
				}
				if !_rules[ruleSpecs]() {
					goto l8
				}
				if !_rules[ruleLineEnd]() {
					goto l8
				}
				depth--
				add(ruleGit, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 3 Gem <- <('G' 'E' 'M' LineEnd Option* Specs LineEnd)> */
		func() bool {
			position12, tokenIndex12, depth12 := position, tokenIndex, depth
			{
				position13 := position
				depth++
				if buffer[position] != rune('G') {
					goto l12
				}
				position++
				if buffer[position] != rune('E') {
					goto l12
				}
				position++
				if buffer[position] != rune('M') {
					goto l12
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l12
				}
			l14:
				{
					position15, tokenIndex15, depth15 := position, tokenIndex, depth
					if !_rules[ruleOption]() {
						goto l15
					}
					goto l14
				l15:
					position, tokenIndex, depth = position15, tokenIndex15, depth15
				}
				if !_rules[ruleSpecs]() {
					goto l12
				}
				if !_rules[ruleLineEnd]() {
					goto l12
				}
				depth--
				add(ruleGem, position13)
			}
			return true
		l12:
			position, tokenIndex, depth = position12, tokenIndex12, depth12
			return false
		},
		/* 4 Platforms <- <('P' 'L' 'A' 'T' 'F' 'O' 'R' 'M' 'S' LineEnd Platform+ LineEnd)> */
		func() bool {
			position16, tokenIndex16, depth16 := position, tokenIndex, depth
			{
				position17 := position
				depth++
				if buffer[position] != rune('P') {
					goto l16
				}
				position++
				if buffer[position] != rune('L') {
					goto l16
				}
				position++
				if buffer[position] != rune('A') {
					goto l16
				}
				position++
				if buffer[position] != rune('T') {
					goto l16
				}
				position++
				if buffer[position] != rune('F') {
					goto l16
				}
				position++
				if buffer[position] != rune('O') {
					goto l16
				}
				position++
				if buffer[position] != rune('R') {
					goto l16
				}
				position++
				if buffer[position] != rune('M') {
					goto l16
				}
				position++
				if buffer[position] != rune('S') {
					goto l16
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l16
				}
				if !_rules[rulePlatform]() {
					goto l16
				}
			l18:
				{
					position19, tokenIndex19, depth19 := position, tokenIndex, depth
					if !_rules[rulePlatform]() {
						goto l19
					}
					goto l18
				l19:
					position, tokenIndex, depth = position19, tokenIndex19, depth19
				}
				if !_rules[ruleLineEnd]() {
					goto l16
				}
				depth--
				add(rulePlatforms, position17)
			}
			return true
		l16:
			position, tokenIndex, depth = position16, tokenIndex16, depth16
			return false
		},
		/* 5 Dependencies <- <('D' 'E' 'P' 'E' 'N' 'D' 'E' 'N' 'C' 'I' 'E' 'S' LineEnd Dependency+)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				if buffer[position] != rune('D') {
					goto l20
				}
				position++
				if buffer[position] != rune('E') {
					goto l20
				}
				position++
				if buffer[position] != rune('P') {
					goto l20
				}
				position++
				if buffer[position] != rune('E') {
					goto l20
				}
				position++
				if buffer[position] != rune('N') {
					goto l20
				}
				position++
				if buffer[position] != rune('D') {
					goto l20
				}
				position++
				if buffer[position] != rune('E') {
					goto l20
				}
				position++
				if buffer[position] != rune('N') {
					goto l20
				}
				position++
				if buffer[position] != rune('C') {
					goto l20
				}
				position++
				if buffer[position] != rune('I') {
					goto l20
				}
				position++
				if buffer[position] != rune('E') {
					goto l20
				}
				position++
				if buffer[position] != rune('S') {
					goto l20
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l20
				}
				if !_rules[ruleDependency]() {
					goto l20
				}
			l22:
				{
					position23, tokenIndex23, depth23 := position, tokenIndex, depth
					if !_rules[ruleDependency]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex, depth = position23, tokenIndex23, depth23
				}
				depth--
				add(ruleDependencies, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 6 Option <- <(Indent2 ([a-z] / [A-Z])+ (':' ' ') (!EndOfLine .)* LineEnd)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l24
				}
				{
					position28, tokenIndex28, depth28 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l29
					}
					position++
					goto l28
				l29:
					position, tokenIndex, depth = position28, tokenIndex28, depth28
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l24
					}
					position++
				}
			l28:
			l26:
				{
					position27, tokenIndex27, depth27 := position, tokenIndex, depth
					{
						position30, tokenIndex30, depth30 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l31
						}
						position++
						goto l30
					l31:
						position, tokenIndex, depth = position30, tokenIndex30, depth30
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l27
						}
						position++
					}
				l30:
					goto l26
				l27:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
				}
				if buffer[position] != rune(':') {
					goto l24
				}
				position++
				if buffer[position] != rune(' ') {
					goto l24
				}
				position++
			l32:
				{
					position33, tokenIndex33, depth33 := position, tokenIndex, depth
					{
						position34, tokenIndex34, depth34 := position, tokenIndex, depth
						if !_rules[ruleEndOfLine]() {
							goto l34
						}
						goto l33
					l34:
						position, tokenIndex, depth = position34, tokenIndex34, depth34
					}
					if !matchDot() {
						goto l33
					}
					goto l32
				l33:
					position, tokenIndex, depth = position33, tokenIndex33, depth33
				}
				if !_rules[ruleLineEnd]() {
					goto l24
				}
				depth--
				add(ruleOption, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 7 Specs <- <(Indent2 ('s' 'p' 'e' 'c' 's' ':') LineEnd Spec+)> */
		func() bool {
			position35, tokenIndex35, depth35 := position, tokenIndex, depth
			{
				position36 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l35
				}
				if buffer[position] != rune('s') {
					goto l35
				}
				position++
				if buffer[position] != rune('p') {
					goto l35
				}
				position++
				if buffer[position] != rune('e') {
					goto l35
				}
				position++
				if buffer[position] != rune('c') {
					goto l35
				}
				position++
				if buffer[position] != rune('s') {
					goto l35
				}
				position++
				if buffer[position] != rune(':') {
					goto l35
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l35
				}
				if !_rules[ruleSpec]() {
					goto l35
				}
			l37:
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					if !_rules[ruleSpec]() {
						goto l38
					}
					goto l37
				l38:
					position, tokenIndex, depth = position38, tokenIndex38, depth38
				}
				depth--
				add(ruleSpecs, position36)
			}
			return true
		l35:
			position, tokenIndex, depth = position35, tokenIndex35, depth35
			return false
		},
		/* 8 Spec <- <(Indent4 Action0 GemVersion SpecDep*)> */
		func() bool {
			position39, tokenIndex39, depth39 := position, tokenIndex, depth
			{
				position40 := position
				depth++
				if !_rules[ruleIndent4]() {
					goto l39
				}
				if !_rules[ruleAction0]() {
					goto l39
				}
				if !_rules[ruleGemVersion]() {
					goto l39
				}
			l41:
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !_rules[ruleSpecDep]() {
						goto l42
					}
					goto l41
				l42:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
				}
				depth--
				add(ruleSpec, position40)
			}
			return true
		l39:
			position, tokenIndex, depth = position39, tokenIndex39, depth39
			return false
		},
		/* 9 SpecDep <- <(Indent6 Action1 GemVersion)> */
		func() bool {
			position43, tokenIndex43, depth43 := position, tokenIndex, depth
			{
				position44 := position
				depth++
				if !_rules[ruleIndent6]() {
					goto l43
				}
				if !_rules[ruleAction1]() {
					goto l43
				}
				if !_rules[ruleGemVersion]() {
					goto l43
				}
				depth--
				add(ruleSpecDep, position44)
			}
			return true
		l43:
			position, tokenIndex, depth = position43, tokenIndex43, depth43
			return false
		},
		/* 10 Dependency <- <(Indent2 Action2 GemVersion)> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l45
				}
				if !_rules[ruleAction2]() {
					goto l45
				}
				if !_rules[ruleGemVersion]() {
					goto l45
				}
				depth--
				add(ruleDependency, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 11 GemVersion <- <(GemName Spaces Version? LineEnd)> */
		func() bool {
			position47, tokenIndex47, depth47 := position, tokenIndex, depth
			{
				position48 := position
				depth++
				if !_rules[ruleGemName]() {
					goto l47
				}
				if !_rules[ruleSpaces]() {
					goto l47
				}
				{
					position49, tokenIndex49, depth49 := position, tokenIndex, depth
					if !_rules[ruleVersion]() {
						goto l49
					}
					goto l50
				l49:
					position, tokenIndex, depth = position49, tokenIndex49, depth49
				}
			l50:
				if !_rules[ruleLineEnd]() {
					goto l47
				}
				depth--
				add(ruleGemVersion, position48)
			}
			return true
		l47:
			position, tokenIndex, depth = position47, tokenIndex47, depth47
			return false
		},
		/* 12 GemName <- <(<([a-z] / [A-Z] / '-' / '_' / '!' / [0-9])+> Action3)> */
		func() bool {
			position51, tokenIndex51, depth51 := position, tokenIndex, depth
			{
				position52 := position
				depth++
				{
					position53 := position
					depth++
					{
						position56, tokenIndex56, depth56 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l57
						}
						position++
						goto l56
					l57:
						position, tokenIndex, depth = position56, tokenIndex56, depth56
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l58
						}
						position++
						goto l56
					l58:
						position, tokenIndex, depth = position56, tokenIndex56, depth56
						if buffer[position] != rune('-') {
							goto l59
						}
						position++
						goto l56
					l59:
						position, tokenIndex, depth = position56, tokenIndex56, depth56
						if buffer[position] != rune('_') {
							goto l60
						}
						position++
						goto l56
					l60:
						position, tokenIndex, depth = position56, tokenIndex56, depth56
						if buffer[position] != rune('!') {
							goto l61
						}
						position++
						goto l56
					l61:
						position, tokenIndex, depth = position56, tokenIndex56, depth56
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l51
						}
						position++
					}
				l56:
				l54:
					{
						position55, tokenIndex55, depth55 := position, tokenIndex, depth
						{
							position62, tokenIndex62, depth62 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l63
							}
							position++
							goto l62
						l63:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l64
							}
							position++
							goto l62
						l64:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if buffer[position] != rune('-') {
								goto l65
							}
							position++
							goto l62
						l65:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if buffer[position] != rune('_') {
								goto l66
							}
							position++
							goto l62
						l66:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if buffer[position] != rune('!') {
								goto l67
							}
							position++
							goto l62
						l67:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l55
							}
							position++
						}
					l62:
						goto l54
					l55:
						position, tokenIndex, depth = position55, tokenIndex55, depth55
					}
					depth--
					add(rulePegText, position53)
				}
				if !_rules[ruleAction3]() {
					goto l51
				}
				depth--
				add(ruleGemName, position52)
			}
			return true
		l51:
			position, tokenIndex, depth = position51, tokenIndex51, depth51
			return false
		},
		/* 13 Version <- <(<('(' Constraint (',' ' ' Constraint)* ')')> Action4)> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				{
					position70 := position
					depth++
					if buffer[position] != rune('(') {
						goto l68
					}
					position++
					if !_rules[ruleConstraint]() {
						goto l68
					}
				l71:
					{
						position72, tokenIndex72, depth72 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l72
						}
						position++
						if buffer[position] != rune(' ') {
							goto l72
						}
						position++
						if !_rules[ruleConstraint]() {
							goto l72
						}
						goto l71
					l72:
						position, tokenIndex, depth = position72, tokenIndex72, depth72
					}
					if buffer[position] != rune(')') {
						goto l68
					}
					position++
					depth--
					add(rulePegText, position70)
				}
				if !_rules[ruleAction4]() {
					goto l68
				}
				depth--
				add(ruleVersion, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 14 Constraint <- <(VersionOp? Spaces [0-9]+ ('.' [0-9]+)*)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				{
					position75, tokenIndex75, depth75 := position, tokenIndex, depth
					if !_rules[ruleVersionOp]() {
						goto l75
					}
					goto l76
				l75:
					position, tokenIndex, depth = position75, tokenIndex75, depth75
				}
			l76:
				if !_rules[ruleSpaces]() {
					goto l73
				}
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l73
				}
				position++
			l77:
				{
					position78, tokenIndex78, depth78 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l78
					}
					position++
					goto l77
				l78:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
				}
			l79:
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l80
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l80
					}
					position++
				l81:
					{
						position82, tokenIndex82, depth82 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l82
						}
						position++
						goto l81
					l82:
						position, tokenIndex, depth = position82, tokenIndex82, depth82
					}
					goto l79
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
				}
				depth--
				add(ruleConstraint, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 15 VersionOp <- <(Eq / Neq / Leq / Lt / Geq / Gt / TwiddleWakka)> */
		func() bool {
			position83, tokenIndex83, depth83 := position, tokenIndex, depth
			{
				position84 := position
				depth++
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if !_rules[ruleEq]() {
						goto l86
					}
					goto l85
				l86:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if !_rules[ruleNeq]() {
						goto l87
					}
					goto l85
				l87:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if !_rules[ruleLeq]() {
						goto l88
					}
					goto l85
				l88:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if !_rules[ruleLt]() {
						goto l89
					}
					goto l85
				l89:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if !_rules[ruleGeq]() {
						goto l90
					}
					goto l85
				l90:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if !_rules[ruleGt]() {
						goto l91
					}
					goto l85
				l91:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if !_rules[ruleTwiddleWakka]() {
						goto l83
					}
				}
			l85:
				depth--
				add(ruleVersionOp, position84)
			}
			return true
		l83:
			position, tokenIndex, depth = position83, tokenIndex83, depth83
			return false
		},
		/* 16 Platform <- <(Indent2 NotWhitespace+ LineEnd)> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l92
				}
				if !_rules[ruleNotWhitespace]() {
					goto l92
				}
			l94:
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					if !_rules[ruleNotWhitespace]() {
						goto l95
					}
					goto l94
				l95:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
				}
				if !_rules[ruleLineEnd]() {
					goto l92
				}
				depth--
				add(rulePlatform, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 17 Eq <- <'='> */
		func() bool {
			position96, tokenIndex96, depth96 := position, tokenIndex, depth
			{
				position97 := position
				depth++
				if buffer[position] != rune('=') {
					goto l96
				}
				position++
				depth--
				add(ruleEq, position97)
			}
			return true
		l96:
			position, tokenIndex, depth = position96, tokenIndex96, depth96
			return false
		},
		/* 18 Neq <- <('!' '=')> */
		func() bool {
			position98, tokenIndex98, depth98 := position, tokenIndex, depth
			{
				position99 := position
				depth++
				if buffer[position] != rune('!') {
					goto l98
				}
				position++
				if buffer[position] != rune('=') {
					goto l98
				}
				position++
				depth--
				add(ruleNeq, position99)
			}
			return true
		l98:
			position, tokenIndex, depth = position98, tokenIndex98, depth98
			return false
		},
		/* 19 Leq <- <('<' '=')> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				if buffer[position] != rune('<') {
					goto l100
				}
				position++
				if buffer[position] != rune('=') {
					goto l100
				}
				position++
				depth--
				add(ruleLeq, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 20 Lt <- <'<'> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				if buffer[position] != rune('<') {
					goto l102
				}
				position++
				depth--
				add(ruleLt, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 21 Geq <- <('>' '=')> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				if buffer[position] != rune('>') {
					goto l104
				}
				position++
				if buffer[position] != rune('=') {
					goto l104
				}
				position++
				depth--
				add(ruleGeq, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 22 Gt <- <'>'> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				if buffer[position] != rune('>') {
					goto l106
				}
				position++
				depth--
				add(ruleGt, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 23 TwiddleWakka <- <('~' '>')> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				if buffer[position] != rune('~') {
					goto l108
				}
				position++
				if buffer[position] != rune('>') {
					goto l108
				}
				position++
				depth--
				add(ruleTwiddleWakka, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 24 Space <- <(' ' / '\t')> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				{
					position112, tokenIndex112, depth112 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l113
					}
					position++
					goto l112
				l113:
					position, tokenIndex, depth = position112, tokenIndex112, depth112
					if buffer[position] != rune('\t') {
						goto l110
					}
					position++
				}
			l112:
				depth--
				add(ruleSpace, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 25 Spaces <- <Space*> */
		func() bool {
			{
				position115 := position
				depth++
			l116:
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if !_rules[ruleSpace]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
				}
				depth--
				add(ruleSpaces, position115)
			}
			return true
		},
		/* 26 Indent2 <- <(Space Space)> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				if !_rules[ruleSpace]() {
					goto l118
				}
				if !_rules[ruleSpace]() {
					goto l118
				}
				depth--
				add(ruleIndent2, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 27 Indent4 <- <(Space Space Space Space)> */
		func() bool {
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				if !_rules[ruleSpace]() {
					goto l120
				}
				if !_rules[ruleSpace]() {
					goto l120
				}
				if !_rules[ruleSpace]() {
					goto l120
				}
				if !_rules[ruleSpace]() {
					goto l120
				}
				depth--
				add(ruleIndent4, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 28 Indent6 <- <(Space Space Space Space Space Space)> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				if !_rules[ruleSpace]() {
					goto l122
				}
				if !_rules[ruleSpace]() {
					goto l122
				}
				if !_rules[ruleSpace]() {
					goto l122
				}
				if !_rules[ruleSpace]() {
					goto l122
				}
				if !_rules[ruleSpace]() {
					goto l122
				}
				if !_rules[ruleSpace]() {
					goto l122
				}
				depth--
				add(ruleIndent6, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 29 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position124, tokenIndex124, depth124 := position, tokenIndex, depth
			{
				position125 := position
				depth++
				{
					position126, tokenIndex126, depth126 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l127
					}
					position++
					if buffer[position] != rune('\n') {
						goto l127
					}
					position++
					goto l126
				l127:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if buffer[position] != rune('\n') {
						goto l128
					}
					position++
					goto l126
				l128:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if buffer[position] != rune('\r') {
						goto l124
					}
					position++
				}
			l126:
				depth--
				add(ruleEndOfLine, position125)
			}
			return true
		l124:
			position, tokenIndex, depth = position124, tokenIndex124, depth124
			return false
		},
		/* 30 NotWhitespace <- <(!(Space / EndOfLine) .)> */
		func() bool {
			position129, tokenIndex129, depth129 := position, tokenIndex, depth
			{
				position130 := position
				depth++
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					{
						position132, tokenIndex132, depth132 := position, tokenIndex, depth
						if !_rules[ruleSpace]() {
							goto l133
						}
						goto l132
					l133:
						position, tokenIndex, depth = position132, tokenIndex132, depth132
						if !_rules[ruleEndOfLine]() {
							goto l131
						}
					}
				l132:
					goto l129
				l131:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
				}
				if !matchDot() {
					goto l129
				}
				depth--
				add(ruleNotWhitespace, position130)
			}
			return true
		l129:
			position, tokenIndex, depth = position129, tokenIndex129, depth129
			return false
		},
		/* 31 LineEnd <- <(Spaces EndOfLine)> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				if !_rules[ruleSpaces]() {
					goto l134
				}
				if !_rules[ruleEndOfLine]() {
					goto l134
				}
				depth--
				add(ruleLineEnd, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 32 EndOfFile <- <(EndOfLine* !.)> */
		func() bool {
			position136, tokenIndex136, depth136 := position, tokenIndex, depth
			{
				position137 := position
				depth++
			l138:
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					if !_rules[ruleEndOfLine]() {
						goto l139
					}
					goto l138
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
				{
					position140, tokenIndex140, depth140 := position, tokenIndex, depth
					if !matchDot() {
						goto l140
					}
					goto l136
				l140:
					position, tokenIndex, depth = position140, tokenIndex140, depth140
				}
				depth--
				add(ruleEndOfFile, position137)
			}
			return true
		l136:
			position, tokenIndex, depth = position136, tokenIndex136, depth136
			return false
		},
		/* 34 Action0 <- <{p.setState(ParsingSpec)}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 35 Action1 <- <{p.setState(ParsingSpecDep)}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 36 Action2 <- <{ p.setState(ParsingDependency) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		nil,
		/* 38 Action3 <- <{ p.addGem(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 39 Action4 <- <{ p.addVersion(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
	}
	p.rules = _rules
}

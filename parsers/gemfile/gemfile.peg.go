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
	ruleGit
	ruleGem
	rulePlatforms
	ruleDependencies
	ruleRemote
	ruleRevision
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
	ruleURL
	ruleSHA
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
	"Git",
	"Gem",
	"Platforms",
	"Dependencies",
	"Remote",
	"Revision",
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
	"URL",
	"SHA",
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
	rules  [42]func() bool
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
		/* 0 Gemfile <- <((Git / Gem)+ Platforms Dependencies EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position4, tokenIndex4, depth4 := position, tokenIndex, depth
					if !_rules[ruleGit]() {
						goto l5
					}
					goto l4
				l5:
					position, tokenIndex, depth = position4, tokenIndex4, depth4
					if !_rules[ruleGem]() {
						goto l0
					}
				}
			l4:
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					{
						position6, tokenIndex6, depth6 := position, tokenIndex, depth
						if !_rules[ruleGit]() {
							goto l7
						}
						goto l6
					l7:
						position, tokenIndex, depth = position6, tokenIndex6, depth6
						if !_rules[ruleGem]() {
							goto l3
						}
					}
				l6:
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
		/* 1 Git <- <('G' 'I' 'T' LineEnd Remote Revision Specs LineEnd)> */
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
				if !_rules[ruleRemote]() {
					goto l8
				}
				if !_rules[ruleRevision]() {
					goto l8
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
		/* 2 Gem <- <('G' 'E' 'M' LineEnd Remote Specs LineEnd)> */
		func() bool {
			position10, tokenIndex10, depth10 := position, tokenIndex, depth
			{
				position11 := position
				depth++
				if buffer[position] != rune('G') {
					goto l10
				}
				position++
				if buffer[position] != rune('E') {
					goto l10
				}
				position++
				if buffer[position] != rune('M') {
					goto l10
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l10
				}
				if !_rules[ruleRemote]() {
					goto l10
				}
				if !_rules[ruleSpecs]() {
					goto l10
				}
				if !_rules[ruleLineEnd]() {
					goto l10
				}
				depth--
				add(ruleGem, position11)
			}
			return true
		l10:
			position, tokenIndex, depth = position10, tokenIndex10, depth10
			return false
		},
		/* 3 Platforms <- <('P' 'L' 'A' 'T' 'F' 'O' 'R' 'M' 'S' LineEnd Platform+ LineEnd)> */
		func() bool {
			position12, tokenIndex12, depth12 := position, tokenIndex, depth
			{
				position13 := position
				depth++
				if buffer[position] != rune('P') {
					goto l12
				}
				position++
				if buffer[position] != rune('L') {
					goto l12
				}
				position++
				if buffer[position] != rune('A') {
					goto l12
				}
				position++
				if buffer[position] != rune('T') {
					goto l12
				}
				position++
				if buffer[position] != rune('F') {
					goto l12
				}
				position++
				if buffer[position] != rune('O') {
					goto l12
				}
				position++
				if buffer[position] != rune('R') {
					goto l12
				}
				position++
				if buffer[position] != rune('M') {
					goto l12
				}
				position++
				if buffer[position] != rune('S') {
					goto l12
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l12
				}
				if !_rules[rulePlatform]() {
					goto l12
				}
			l14:
				{
					position15, tokenIndex15, depth15 := position, tokenIndex, depth
					if !_rules[rulePlatform]() {
						goto l15
					}
					goto l14
				l15:
					position, tokenIndex, depth = position15, tokenIndex15, depth15
				}
				if !_rules[ruleLineEnd]() {
					goto l12
				}
				depth--
				add(rulePlatforms, position13)
			}
			return true
		l12:
			position, tokenIndex, depth = position12, tokenIndex12, depth12
			return false
		},
		/* 4 Dependencies <- <('D' 'E' 'P' 'E' 'N' 'D' 'E' 'N' 'C' 'I' 'E' 'S' LineEnd Dependency+)> */
		func() bool {
			position16, tokenIndex16, depth16 := position, tokenIndex, depth
			{
				position17 := position
				depth++
				if buffer[position] != rune('D') {
					goto l16
				}
				position++
				if buffer[position] != rune('E') {
					goto l16
				}
				position++
				if buffer[position] != rune('P') {
					goto l16
				}
				position++
				if buffer[position] != rune('E') {
					goto l16
				}
				position++
				if buffer[position] != rune('N') {
					goto l16
				}
				position++
				if buffer[position] != rune('D') {
					goto l16
				}
				position++
				if buffer[position] != rune('E') {
					goto l16
				}
				position++
				if buffer[position] != rune('N') {
					goto l16
				}
				position++
				if buffer[position] != rune('C') {
					goto l16
				}
				position++
				if buffer[position] != rune('I') {
					goto l16
				}
				position++
				if buffer[position] != rune('E') {
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
				if !_rules[ruleDependency]() {
					goto l16
				}
			l18:
				{
					position19, tokenIndex19, depth19 := position, tokenIndex, depth
					if !_rules[ruleDependency]() {
						goto l19
					}
					goto l18
				l19:
					position, tokenIndex, depth = position19, tokenIndex19, depth19
				}
				depth--
				add(ruleDependencies, position17)
			}
			return true
		l16:
			position, tokenIndex, depth = position16, tokenIndex16, depth16
			return false
		},
		/* 5 Remote <- <(Indent2 ('r' 'e' 'm' 'o' 't' 'e' ':') Spaces URL LineEnd)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l20
				}
				if buffer[position] != rune('r') {
					goto l20
				}
				position++
				if buffer[position] != rune('e') {
					goto l20
				}
				position++
				if buffer[position] != rune('m') {
					goto l20
				}
				position++
				if buffer[position] != rune('o') {
					goto l20
				}
				position++
				if buffer[position] != rune('t') {
					goto l20
				}
				position++
				if buffer[position] != rune('e') {
					goto l20
				}
				position++
				if buffer[position] != rune(':') {
					goto l20
				}
				position++
				if !_rules[ruleSpaces]() {
					goto l20
				}
				if !_rules[ruleURL]() {
					goto l20
				}
				if !_rules[ruleLineEnd]() {
					goto l20
				}
				depth--
				add(ruleRemote, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 6 Revision <- <(Indent2 ('r' 'e' 'v' 'i' 's' 'i' 'o' 'n' ':') Spaces SHA LineEnd)> */
		func() bool {
			position22, tokenIndex22, depth22 := position, tokenIndex, depth
			{
				position23 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l22
				}
				if buffer[position] != rune('r') {
					goto l22
				}
				position++
				if buffer[position] != rune('e') {
					goto l22
				}
				position++
				if buffer[position] != rune('v') {
					goto l22
				}
				position++
				if buffer[position] != rune('i') {
					goto l22
				}
				position++
				if buffer[position] != rune('s') {
					goto l22
				}
				position++
				if buffer[position] != rune('i') {
					goto l22
				}
				position++
				if buffer[position] != rune('o') {
					goto l22
				}
				position++
				if buffer[position] != rune('n') {
					goto l22
				}
				position++
				if buffer[position] != rune(':') {
					goto l22
				}
				position++
				if !_rules[ruleSpaces]() {
					goto l22
				}
				if !_rules[ruleSHA]() {
					goto l22
				}
				if !_rules[ruleLineEnd]() {
					goto l22
				}
				depth--
				add(ruleRevision, position23)
			}
			return true
		l22:
			position, tokenIndex, depth = position22, tokenIndex22, depth22
			return false
		},
		/* 7 Specs <- <(Indent2 ('s' 'p' 'e' 'c' 's' ':') LineEnd Spec+)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l24
				}
				if buffer[position] != rune('s') {
					goto l24
				}
				position++
				if buffer[position] != rune('p') {
					goto l24
				}
				position++
				if buffer[position] != rune('e') {
					goto l24
				}
				position++
				if buffer[position] != rune('c') {
					goto l24
				}
				position++
				if buffer[position] != rune('s') {
					goto l24
				}
				position++
				if buffer[position] != rune(':') {
					goto l24
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l24
				}
				if !_rules[ruleSpec]() {
					goto l24
				}
			l26:
				{
					position27, tokenIndex27, depth27 := position, tokenIndex, depth
					if !_rules[ruleSpec]() {
						goto l27
					}
					goto l26
				l27:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
				}
				depth--
				add(ruleSpecs, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 8 Spec <- <(Indent4 Action0 GemVersion SpecDep*)> */
		func() bool {
			position28, tokenIndex28, depth28 := position, tokenIndex, depth
			{
				position29 := position
				depth++
				if !_rules[ruleIndent4]() {
					goto l28
				}
				if !_rules[ruleAction0]() {
					goto l28
				}
				if !_rules[ruleGemVersion]() {
					goto l28
				}
			l30:
				{
					position31, tokenIndex31, depth31 := position, tokenIndex, depth
					if !_rules[ruleSpecDep]() {
						goto l31
					}
					goto l30
				l31:
					position, tokenIndex, depth = position31, tokenIndex31, depth31
				}
				depth--
				add(ruleSpec, position29)
			}
			return true
		l28:
			position, tokenIndex, depth = position28, tokenIndex28, depth28
			return false
		},
		/* 9 SpecDep <- <(Indent6 Action1 GemVersion)> */
		func() bool {
			position32, tokenIndex32, depth32 := position, tokenIndex, depth
			{
				position33 := position
				depth++
				if !_rules[ruleIndent6]() {
					goto l32
				}
				if !_rules[ruleAction1]() {
					goto l32
				}
				if !_rules[ruleGemVersion]() {
					goto l32
				}
				depth--
				add(ruleSpecDep, position33)
			}
			return true
		l32:
			position, tokenIndex, depth = position32, tokenIndex32, depth32
			return false
		},
		/* 10 Dependency <- <(Indent2 Action2 GemVersion)> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l34
				}
				if !_rules[ruleAction2]() {
					goto l34
				}
				if !_rules[ruleGemVersion]() {
					goto l34
				}
				depth--
				add(ruleDependency, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 11 GemVersion <- <(GemName Spaces Version? LineEnd)> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if !_rules[ruleGemName]() {
					goto l36
				}
				if !_rules[ruleSpaces]() {
					goto l36
				}
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					if !_rules[ruleVersion]() {
						goto l38
					}
					goto l39
				l38:
					position, tokenIndex, depth = position38, tokenIndex38, depth38
				}
			l39:
				if !_rules[ruleLineEnd]() {
					goto l36
				}
				depth--
				add(ruleGemVersion, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 12 GemName <- <(<([a-z] / [A-Z] / '-' / '_' / '!' / [0-9])+> Action3)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				{
					position42 := position
					depth++
					{
						position45, tokenIndex45, depth45 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l46
						}
						position++
						goto l45
					l46:
						position, tokenIndex, depth = position45, tokenIndex45, depth45
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l47
						}
						position++
						goto l45
					l47:
						position, tokenIndex, depth = position45, tokenIndex45, depth45
						if buffer[position] != rune('-') {
							goto l48
						}
						position++
						goto l45
					l48:
						position, tokenIndex, depth = position45, tokenIndex45, depth45
						if buffer[position] != rune('_') {
							goto l49
						}
						position++
						goto l45
					l49:
						position, tokenIndex, depth = position45, tokenIndex45, depth45
						if buffer[position] != rune('!') {
							goto l50
						}
						position++
						goto l45
					l50:
						position, tokenIndex, depth = position45, tokenIndex45, depth45
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l40
						}
						position++
					}
				l45:
				l43:
					{
						position44, tokenIndex44, depth44 := position, tokenIndex, depth
						{
							position51, tokenIndex51, depth51 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l52
							}
							position++
							goto l51
						l52:
							position, tokenIndex, depth = position51, tokenIndex51, depth51
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l53
							}
							position++
							goto l51
						l53:
							position, tokenIndex, depth = position51, tokenIndex51, depth51
							if buffer[position] != rune('-') {
								goto l54
							}
							position++
							goto l51
						l54:
							position, tokenIndex, depth = position51, tokenIndex51, depth51
							if buffer[position] != rune('_') {
								goto l55
							}
							position++
							goto l51
						l55:
							position, tokenIndex, depth = position51, tokenIndex51, depth51
							if buffer[position] != rune('!') {
								goto l56
							}
							position++
							goto l51
						l56:
							position, tokenIndex, depth = position51, tokenIndex51, depth51
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l44
							}
							position++
						}
					l51:
						goto l43
					l44:
						position, tokenIndex, depth = position44, tokenIndex44, depth44
					}
					depth--
					add(rulePegText, position42)
				}
				if !_rules[ruleAction3]() {
					goto l40
				}
				depth--
				add(ruleGemName, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 13 Version <- <(<('(' Constraint (',' ' ' Constraint)* ')')> Action4)> */
		func() bool {
			position57, tokenIndex57, depth57 := position, tokenIndex, depth
			{
				position58 := position
				depth++
				{
					position59 := position
					depth++
					if buffer[position] != rune('(') {
						goto l57
					}
					position++
					if !_rules[ruleConstraint]() {
						goto l57
					}
				l60:
					{
						position61, tokenIndex61, depth61 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l61
						}
						position++
						if buffer[position] != rune(' ') {
							goto l61
						}
						position++
						if !_rules[ruleConstraint]() {
							goto l61
						}
						goto l60
					l61:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
					}
					if buffer[position] != rune(')') {
						goto l57
					}
					position++
					depth--
					add(rulePegText, position59)
				}
				if !_rules[ruleAction4]() {
					goto l57
				}
				depth--
				add(ruleVersion, position58)
			}
			return true
		l57:
			position, tokenIndex, depth = position57, tokenIndex57, depth57
			return false
		},
		/* 14 Constraint <- <(VersionOp? Spaces [0-9]+ ('.' [0-9]+)*)> */
		func() bool {
			position62, tokenIndex62, depth62 := position, tokenIndex, depth
			{
				position63 := position
				depth++
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					if !_rules[ruleVersionOp]() {
						goto l64
					}
					goto l65
				l64:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
				}
			l65:
				if !_rules[ruleSpaces]() {
					goto l62
				}
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l62
				}
				position++
			l66:
				{
					position67, tokenIndex67, depth67 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l67
					}
					position++
					goto l66
				l67:
					position, tokenIndex, depth = position67, tokenIndex67, depth67
				}
			l68:
				{
					position69, tokenIndex69, depth69 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l69
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l69
					}
					position++
				l70:
					{
						position71, tokenIndex71, depth71 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l71
						}
						position++
						goto l70
					l71:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
					}
					goto l68
				l69:
					position, tokenIndex, depth = position69, tokenIndex69, depth69
				}
				depth--
				add(ruleConstraint, position63)
			}
			return true
		l62:
			position, tokenIndex, depth = position62, tokenIndex62, depth62
			return false
		},
		/* 15 VersionOp <- <(Eq / Neq / Leq / Lt / Geq / Gt / TwiddleWakka)> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				{
					position74, tokenIndex74, depth74 := position, tokenIndex, depth
					if !_rules[ruleEq]() {
						goto l75
					}
					goto l74
				l75:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleNeq]() {
						goto l76
					}
					goto l74
				l76:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleLeq]() {
						goto l77
					}
					goto l74
				l77:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleLt]() {
						goto l78
					}
					goto l74
				l78:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleGeq]() {
						goto l79
					}
					goto l74
				l79:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleGt]() {
						goto l80
					}
					goto l74
				l80:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleTwiddleWakka]() {
						goto l72
					}
				}
			l74:
				depth--
				add(ruleVersionOp, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 16 Platform <- <(Indent2 NotWhitespace+ LineEnd)> */
		func() bool {
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l81
				}
				if !_rules[ruleNotWhitespace]() {
					goto l81
				}
			l83:
				{
					position84, tokenIndex84, depth84 := position, tokenIndex, depth
					if !_rules[ruleNotWhitespace]() {
						goto l84
					}
					goto l83
				l84:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
				}
				if !_rules[ruleLineEnd]() {
					goto l81
				}
				depth--
				add(rulePlatform, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 17 URL <- <NotWhitespace+> */
		func() bool {
			position85, tokenIndex85, depth85 := position, tokenIndex, depth
			{
				position86 := position
				depth++
				if !_rules[ruleNotWhitespace]() {
					goto l85
				}
			l87:
				{
					position88, tokenIndex88, depth88 := position, tokenIndex, depth
					if !_rules[ruleNotWhitespace]() {
						goto l88
					}
					goto l87
				l88:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
				}
				depth--
				add(ruleURL, position86)
			}
			return true
		l85:
			position, tokenIndex, depth = position85, tokenIndex85, depth85
			return false
		},
		/* 18 SHA <- <([a-z] / [A-Z] / [0-9])+> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				{
					position93, tokenIndex93, depth93 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l94
					}
					position++
					goto l93
				l94:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l95
					}
					position++
					goto l93
				l95:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l89
					}
					position++
				}
			l93:
			l91:
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					{
						position96, tokenIndex96, depth96 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l97
						}
						position++
						goto l96
					l97:
						position, tokenIndex, depth = position96, tokenIndex96, depth96
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l98
						}
						position++
						goto l96
					l98:
						position, tokenIndex, depth = position96, tokenIndex96, depth96
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l92
						}
						position++
					}
				l96:
					goto l91
				l92:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
				}
				depth--
				add(ruleSHA, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 19 Eq <- <'='> */
		func() bool {
			position99, tokenIndex99, depth99 := position, tokenIndex, depth
			{
				position100 := position
				depth++
				if buffer[position] != rune('=') {
					goto l99
				}
				position++
				depth--
				add(ruleEq, position100)
			}
			return true
		l99:
			position, tokenIndex, depth = position99, tokenIndex99, depth99
			return false
		},
		/* 20 Neq <- <('!' '=')> */
		func() bool {
			position101, tokenIndex101, depth101 := position, tokenIndex, depth
			{
				position102 := position
				depth++
				if buffer[position] != rune('!') {
					goto l101
				}
				position++
				if buffer[position] != rune('=') {
					goto l101
				}
				position++
				depth--
				add(ruleNeq, position102)
			}
			return true
		l101:
			position, tokenIndex, depth = position101, tokenIndex101, depth101
			return false
		},
		/* 21 Leq <- <('<' '=')> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				if buffer[position] != rune('<') {
					goto l103
				}
				position++
				if buffer[position] != rune('=') {
					goto l103
				}
				position++
				depth--
				add(ruleLeq, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 22 Lt <- <'<'> */
		func() bool {
			position105, tokenIndex105, depth105 := position, tokenIndex, depth
			{
				position106 := position
				depth++
				if buffer[position] != rune('<') {
					goto l105
				}
				position++
				depth--
				add(ruleLt, position106)
			}
			return true
		l105:
			position, tokenIndex, depth = position105, tokenIndex105, depth105
			return false
		},
		/* 23 Geq <- <('>' '=')> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				if buffer[position] != rune('>') {
					goto l107
				}
				position++
				if buffer[position] != rune('=') {
					goto l107
				}
				position++
				depth--
				add(ruleGeq, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 24 Gt <- <'>'> */
		func() bool {
			position109, tokenIndex109, depth109 := position, tokenIndex, depth
			{
				position110 := position
				depth++
				if buffer[position] != rune('>') {
					goto l109
				}
				position++
				depth--
				add(ruleGt, position110)
			}
			return true
		l109:
			position, tokenIndex, depth = position109, tokenIndex109, depth109
			return false
		},
		/* 25 TwiddleWakka <- <('~' '>')> */
		func() bool {
			position111, tokenIndex111, depth111 := position, tokenIndex, depth
			{
				position112 := position
				depth++
				if buffer[position] != rune('~') {
					goto l111
				}
				position++
				if buffer[position] != rune('>') {
					goto l111
				}
				position++
				depth--
				add(ruleTwiddleWakka, position112)
			}
			return true
		l111:
			position, tokenIndex, depth = position111, tokenIndex111, depth111
			return false
		},
		/* 26 Space <- <(' ' / '\t')> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l116
					}
					position++
					goto l115
				l116:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
					if buffer[position] != rune('\t') {
						goto l113
					}
					position++
				}
			l115:
				depth--
				add(ruleSpace, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 27 Spaces <- <Space*> */
		func() bool {
			{
				position118 := position
				depth++
			l119:
				{
					position120, tokenIndex120, depth120 := position, tokenIndex, depth
					if !_rules[ruleSpace]() {
						goto l120
					}
					goto l119
				l120:
					position, tokenIndex, depth = position120, tokenIndex120, depth120
				}
				depth--
				add(ruleSpaces, position118)
			}
			return true
		},
		/* 28 Indent2 <- <(Space Space)> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				if !_rules[ruleSpace]() {
					goto l121
				}
				if !_rules[ruleSpace]() {
					goto l121
				}
				depth--
				add(ruleIndent2, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 29 Indent4 <- <(Space Space Space Space)> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				if !_rules[ruleSpace]() {
					goto l123
				}
				if !_rules[ruleSpace]() {
					goto l123
				}
				if !_rules[ruleSpace]() {
					goto l123
				}
				if !_rules[ruleSpace]() {
					goto l123
				}
				depth--
				add(ruleIndent4, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 30 Indent6 <- <(Space Space Space Space Space Space)> */
		func() bool {
			position125, tokenIndex125, depth125 := position, tokenIndex, depth
			{
				position126 := position
				depth++
				if !_rules[ruleSpace]() {
					goto l125
				}
				if !_rules[ruleSpace]() {
					goto l125
				}
				if !_rules[ruleSpace]() {
					goto l125
				}
				if !_rules[ruleSpace]() {
					goto l125
				}
				if !_rules[ruleSpace]() {
					goto l125
				}
				if !_rules[ruleSpace]() {
					goto l125
				}
				depth--
				add(ruleIndent6, position126)
			}
			return true
		l125:
			position, tokenIndex, depth = position125, tokenIndex125, depth125
			return false
		},
		/* 31 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				{
					position129, tokenIndex129, depth129 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l130
					}
					position++
					if buffer[position] != rune('\n') {
						goto l130
					}
					position++
					goto l129
				l130:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if buffer[position] != rune('\n') {
						goto l131
					}
					position++
					goto l129
				l131:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if buffer[position] != rune('\r') {
						goto l127
					}
					position++
				}
			l129:
				depth--
				add(ruleEndOfLine, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 32 NotWhitespace <- <(!(Space / EndOfLine) .)> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				{
					position134, tokenIndex134, depth134 := position, tokenIndex, depth
					{
						position135, tokenIndex135, depth135 := position, tokenIndex, depth
						if !_rules[ruleSpace]() {
							goto l136
						}
						goto l135
					l136:
						position, tokenIndex, depth = position135, tokenIndex135, depth135
						if !_rules[ruleEndOfLine]() {
							goto l134
						}
					}
				l135:
					goto l132
				l134:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
				}
				if !matchDot() {
					goto l132
				}
				depth--
				add(ruleNotWhitespace, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 33 LineEnd <- <(Spaces EndOfLine)> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				if !_rules[ruleSpaces]() {
					goto l137
				}
				if !_rules[ruleEndOfLine]() {
					goto l137
				}
				depth--
				add(ruleLineEnd, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 34 EndOfFile <- <(EndOfLine* !.)> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
			l141:
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					if !_rules[ruleEndOfLine]() {
						goto l142
					}
					goto l141
				l142:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
				}
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if !matchDot() {
						goto l143
					}
					goto l139
				l143:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
				}
				depth--
				add(ruleEndOfFile, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 36 Action0 <- <{p.setState(ParsingSpec)}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 37 Action1 <- <{p.setState(ParsingSpecDep)}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 38 Action2 <- <{ p.setState(ParsingDependency) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		nil,
		/* 40 Action3 <- <{ p.addGem(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 41 Action4 <- <{ p.addVersion(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
	}
	p.rules = _rules
}

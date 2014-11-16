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
	ruleGem
	ruleGit
	ruleSVN
	rulePath
	ruleSource
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
	ruleAction3
	rulePegText
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"Gemfile",
	"Gem",
	"Git",
	"SVN",
	"Path",
	"Source",
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
	"Action3",
	"PegText",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",

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
	rules  [47]func() bool
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
			p.addSource(RubyGems)
		case ruleAction1:
			p.addSource(Git)
		case ruleAction2:
			p.addSource(SVN)
		case ruleAction3:
			p.addSource(Path)
		case ruleAction4:
			p.addOption(buffer[begin:end])
		case ruleAction5:
			p.setState(ParsingSpec)
		case ruleAction6:
			p.setState(ParsingSpecDep)
		case ruleAction7:
			p.setState(ParsingDependency)
		case ruleAction8:
			p.addGem(buffer[begin:end])
		case ruleAction9:
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
		/* 0 Gemfile <- <((Gem / Git / SVN / Path)* Platforms Dependencies EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					{
						position4, tokenIndex4, depth4 := position, tokenIndex, depth
						if !_rules[ruleGem]() {
							goto l5
						}
						goto l4
					l5:
						position, tokenIndex, depth = position4, tokenIndex4, depth4
						if !_rules[ruleGit]() {
							goto l6
						}
						goto l4
					l6:
						position, tokenIndex, depth = position4, tokenIndex4, depth4
						if !_rules[ruleSVN]() {
							goto l7
						}
						goto l4
					l7:
						position, tokenIndex, depth = position4, tokenIndex4, depth4
						if !_rules[rulePath]() {
							goto l3
						}
					}
				l4:
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
		/* 1 Gem <- <('G' 'E' 'M' LineEnd Action0 Source)> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				if buffer[position] != rune('G') {
					goto l8
				}
				position++
				if buffer[position] != rune('E') {
					goto l8
				}
				position++
				if buffer[position] != rune('M') {
					goto l8
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l8
				}
				if !_rules[ruleAction0]() {
					goto l8
				}
				if !_rules[ruleSource]() {
					goto l8
				}
				depth--
				add(ruleGem, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 2 Git <- <('G' 'I' 'T' LineEnd Action1 Source)> */
		func() bool {
			position10, tokenIndex10, depth10 := position, tokenIndex, depth
			{
				position11 := position
				depth++
				if buffer[position] != rune('G') {
					goto l10
				}
				position++
				if buffer[position] != rune('I') {
					goto l10
				}
				position++
				if buffer[position] != rune('T') {
					goto l10
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l10
				}
				if !_rules[ruleAction1]() {
					goto l10
				}
				if !_rules[ruleSource]() {
					goto l10
				}
				depth--
				add(ruleGit, position11)
			}
			return true
		l10:
			position, tokenIndex, depth = position10, tokenIndex10, depth10
			return false
		},
		/* 3 SVN <- <('S' 'V' 'N' LineEnd Action2 Source)> */
		func() bool {
			position12, tokenIndex12, depth12 := position, tokenIndex, depth
			{
				position13 := position
				depth++
				if buffer[position] != rune('S') {
					goto l12
				}
				position++
				if buffer[position] != rune('V') {
					goto l12
				}
				position++
				if buffer[position] != rune('N') {
					goto l12
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l12
				}
				if !_rules[ruleAction2]() {
					goto l12
				}
				if !_rules[ruleSource]() {
					goto l12
				}
				depth--
				add(ruleSVN, position13)
			}
			return true
		l12:
			position, tokenIndex, depth = position12, tokenIndex12, depth12
			return false
		},
		/* 4 Path <- <('P' 'A' 'T' 'H' LineEnd Action3 Source)> */
		func() bool {
			position14, tokenIndex14, depth14 := position, tokenIndex, depth
			{
				position15 := position
				depth++
				if buffer[position] != rune('P') {
					goto l14
				}
				position++
				if buffer[position] != rune('A') {
					goto l14
				}
				position++
				if buffer[position] != rune('T') {
					goto l14
				}
				position++
				if buffer[position] != rune('H') {
					goto l14
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l14
				}
				if !_rules[ruleAction3]() {
					goto l14
				}
				if !_rules[ruleSource]() {
					goto l14
				}
				depth--
				add(rulePath, position15)
			}
			return true
		l14:
			position, tokenIndex, depth = position14, tokenIndex14, depth14
			return false
		},
		/* 5 Source <- <(Option* Specs LineEnd)> */
		func() bool {
			position16, tokenIndex16, depth16 := position, tokenIndex, depth
			{
				position17 := position
				depth++
			l18:
				{
					position19, tokenIndex19, depth19 := position, tokenIndex, depth
					if !_rules[ruleOption]() {
						goto l19
					}
					goto l18
				l19:
					position, tokenIndex, depth = position19, tokenIndex19, depth19
				}
				if !_rules[ruleSpecs]() {
					goto l16
				}
				if !_rules[ruleLineEnd]() {
					goto l16
				}
				depth--
				add(ruleSource, position17)
			}
			return true
		l16:
			position, tokenIndex, depth = position16, tokenIndex16, depth16
			return false
		},
		/* 6 Platforms <- <('P' 'L' 'A' 'T' 'F' 'O' 'R' 'M' 'S' LineEnd Platform+ LineEnd)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				if buffer[position] != rune('P') {
					goto l20
				}
				position++
				if buffer[position] != rune('L') {
					goto l20
				}
				position++
				if buffer[position] != rune('A') {
					goto l20
				}
				position++
				if buffer[position] != rune('T') {
					goto l20
				}
				position++
				if buffer[position] != rune('F') {
					goto l20
				}
				position++
				if buffer[position] != rune('O') {
					goto l20
				}
				position++
				if buffer[position] != rune('R') {
					goto l20
				}
				position++
				if buffer[position] != rune('M') {
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
				if !_rules[rulePlatform]() {
					goto l20
				}
			l22:
				{
					position23, tokenIndex23, depth23 := position, tokenIndex, depth
					if !_rules[rulePlatform]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex, depth = position23, tokenIndex23, depth23
				}
				if !_rules[ruleLineEnd]() {
					goto l20
				}
				depth--
				add(rulePlatforms, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 7 Dependencies <- <('D' 'E' 'P' 'E' 'N' 'D' 'E' 'N' 'C' 'I' 'E' 'S' LineEnd Dependency+)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if buffer[position] != rune('D') {
					goto l24
				}
				position++
				if buffer[position] != rune('E') {
					goto l24
				}
				position++
				if buffer[position] != rune('P') {
					goto l24
				}
				position++
				if buffer[position] != rune('E') {
					goto l24
				}
				position++
				if buffer[position] != rune('N') {
					goto l24
				}
				position++
				if buffer[position] != rune('D') {
					goto l24
				}
				position++
				if buffer[position] != rune('E') {
					goto l24
				}
				position++
				if buffer[position] != rune('N') {
					goto l24
				}
				position++
				if buffer[position] != rune('C') {
					goto l24
				}
				position++
				if buffer[position] != rune('I') {
					goto l24
				}
				position++
				if buffer[position] != rune('E') {
					goto l24
				}
				position++
				if buffer[position] != rune('S') {
					goto l24
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l24
				}
				if !_rules[ruleDependency]() {
					goto l24
				}
			l26:
				{
					position27, tokenIndex27, depth27 := position, tokenIndex, depth
					if !_rules[ruleDependency]() {
						goto l27
					}
					goto l26
				l27:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
				}
				depth--
				add(ruleDependencies, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 8 Option <- <(Indent2 <(([a-z] / [A-Z])+ (':' ' ') (!EndOfLine .)*)> LineEnd Action4)> */
		func() bool {
			position28, tokenIndex28, depth28 := position, tokenIndex, depth
			{
				position29 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l28
				}
				{
					position30 := position
					depth++
					{
						position33, tokenIndex33, depth33 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l34
						}
						position++
						goto l33
					l34:
						position, tokenIndex, depth = position33, tokenIndex33, depth33
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l28
						}
						position++
					}
				l33:
				l31:
					{
						position32, tokenIndex32, depth32 := position, tokenIndex, depth
						{
							position35, tokenIndex35, depth35 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l36
							}
							position++
							goto l35
						l36:
							position, tokenIndex, depth = position35, tokenIndex35, depth35
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l32
							}
							position++
						}
					l35:
						goto l31
					l32:
						position, tokenIndex, depth = position32, tokenIndex32, depth32
					}
					if buffer[position] != rune(':') {
						goto l28
					}
					position++
					if buffer[position] != rune(' ') {
						goto l28
					}
					position++
				l37:
					{
						position38, tokenIndex38, depth38 := position, tokenIndex, depth
						{
							position39, tokenIndex39, depth39 := position, tokenIndex, depth
							if !_rules[ruleEndOfLine]() {
								goto l39
							}
							goto l38
						l39:
							position, tokenIndex, depth = position39, tokenIndex39, depth39
						}
						if !matchDot() {
							goto l38
						}
						goto l37
					l38:
						position, tokenIndex, depth = position38, tokenIndex38, depth38
					}
					depth--
					add(rulePegText, position30)
				}
				if !_rules[ruleLineEnd]() {
					goto l28
				}
				if !_rules[ruleAction4]() {
					goto l28
				}
				depth--
				add(ruleOption, position29)
			}
			return true
		l28:
			position, tokenIndex, depth = position28, tokenIndex28, depth28
			return false
		},
		/* 9 Specs <- <(Indent2 ('s' 'p' 'e' 'c' 's' ':') LineEnd Spec+)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l40
				}
				if buffer[position] != rune('s') {
					goto l40
				}
				position++
				if buffer[position] != rune('p') {
					goto l40
				}
				position++
				if buffer[position] != rune('e') {
					goto l40
				}
				position++
				if buffer[position] != rune('c') {
					goto l40
				}
				position++
				if buffer[position] != rune('s') {
					goto l40
				}
				position++
				if buffer[position] != rune(':') {
					goto l40
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l40
				}
				if !_rules[ruleSpec]() {
					goto l40
				}
			l42:
				{
					position43, tokenIndex43, depth43 := position, tokenIndex, depth
					if !_rules[ruleSpec]() {
						goto l43
					}
					goto l42
				l43:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
				}
				depth--
				add(ruleSpecs, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 10 Spec <- <(Indent4 Action5 GemVersion SpecDep*)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				if !_rules[ruleIndent4]() {
					goto l44
				}
				if !_rules[ruleAction5]() {
					goto l44
				}
				if !_rules[ruleGemVersion]() {
					goto l44
				}
			l46:
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
					if !_rules[ruleSpecDep]() {
						goto l47
					}
					goto l46
				l47:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
				}
				depth--
				add(ruleSpec, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 11 SpecDep <- <(Indent6 Action6 GemVersion)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if !_rules[ruleIndent6]() {
					goto l48
				}
				if !_rules[ruleAction6]() {
					goto l48
				}
				if !_rules[ruleGemVersion]() {
					goto l48
				}
				depth--
				add(ruleSpecDep, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 12 Dependency <- <(Indent2 Action7 GemVersion)> */
		func() bool {
			position50, tokenIndex50, depth50 := position, tokenIndex, depth
			{
				position51 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l50
				}
				if !_rules[ruleAction7]() {
					goto l50
				}
				if !_rules[ruleGemVersion]() {
					goto l50
				}
				depth--
				add(ruleDependency, position51)
			}
			return true
		l50:
			position, tokenIndex, depth = position50, tokenIndex50, depth50
			return false
		},
		/* 13 GemVersion <- <(GemName Spaces Version? LineEnd)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if !_rules[ruleGemName]() {
					goto l52
				}
				if !_rules[ruleSpaces]() {
					goto l52
				}
				{
					position54, tokenIndex54, depth54 := position, tokenIndex, depth
					if !_rules[ruleVersion]() {
						goto l54
					}
					goto l55
				l54:
					position, tokenIndex, depth = position54, tokenIndex54, depth54
				}
			l55:
				if !_rules[ruleLineEnd]() {
					goto l52
				}
				depth--
				add(ruleGemVersion, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 14 GemName <- <(<([a-z] / [A-Z] / '-' / '_' / '!' / [0-9])+> Action8)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				{
					position58 := position
					depth++
					{
						position61, tokenIndex61, depth61 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l62
						}
						position++
						goto l61
					l62:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l63
						}
						position++
						goto l61
					l63:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if buffer[position] != rune('-') {
							goto l64
						}
						position++
						goto l61
					l64:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if buffer[position] != rune('_') {
							goto l65
						}
						position++
						goto l61
					l65:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if buffer[position] != rune('!') {
							goto l66
						}
						position++
						goto l61
					l66:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l56
						}
						position++
					}
				l61:
				l59:
					{
						position60, tokenIndex60, depth60 := position, tokenIndex, depth
						{
							position67, tokenIndex67, depth67 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l68
							}
							position++
							goto l67
						l68:
							position, tokenIndex, depth = position67, tokenIndex67, depth67
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l69
							}
							position++
							goto l67
						l69:
							position, tokenIndex, depth = position67, tokenIndex67, depth67
							if buffer[position] != rune('-') {
								goto l70
							}
							position++
							goto l67
						l70:
							position, tokenIndex, depth = position67, tokenIndex67, depth67
							if buffer[position] != rune('_') {
								goto l71
							}
							position++
							goto l67
						l71:
							position, tokenIndex, depth = position67, tokenIndex67, depth67
							if buffer[position] != rune('!') {
								goto l72
							}
							position++
							goto l67
						l72:
							position, tokenIndex, depth = position67, tokenIndex67, depth67
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l60
							}
							position++
						}
					l67:
						goto l59
					l60:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
					}
					depth--
					add(rulePegText, position58)
				}
				if !_rules[ruleAction8]() {
					goto l56
				}
				depth--
				add(ruleGemName, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 15 Version <- <(<('(' Constraint (',' ' ' Constraint)* ')')> Action9)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				{
					position75 := position
					depth++
					if buffer[position] != rune('(') {
						goto l73
					}
					position++
					if !_rules[ruleConstraint]() {
						goto l73
					}
				l76:
					{
						position77, tokenIndex77, depth77 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l77
						}
						position++
						if buffer[position] != rune(' ') {
							goto l77
						}
						position++
						if !_rules[ruleConstraint]() {
							goto l77
						}
						goto l76
					l77:
						position, tokenIndex, depth = position77, tokenIndex77, depth77
					}
					if buffer[position] != rune(')') {
						goto l73
					}
					position++
					depth--
					add(rulePegText, position75)
				}
				if !_rules[ruleAction9]() {
					goto l73
				}
				depth--
				add(ruleVersion, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 16 Constraint <- <(VersionOp? Spaces [0-9]+ ('.' [0-9]+)*)> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if !_rules[ruleVersionOp]() {
						goto l80
					}
					goto l81
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
				}
			l81:
				if !_rules[ruleSpaces]() {
					goto l78
				}
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l78
				}
				position++
			l82:
				{
					position83, tokenIndex83, depth83 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l83
					}
					position++
					goto l82
				l83:
					position, tokenIndex, depth = position83, tokenIndex83, depth83
				}
			l84:
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l85
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l85
					}
					position++
				l86:
					{
						position87, tokenIndex87, depth87 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l87
						}
						position++
						goto l86
					l87:
						position, tokenIndex, depth = position87, tokenIndex87, depth87
					}
					goto l84
				l85:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
				}
				depth--
				add(ruleConstraint, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 17 VersionOp <- <(Eq / Neq / Leq / Lt / Geq / Gt / TwiddleWakka)> */
		func() bool {
			position88, tokenIndex88, depth88 := position, tokenIndex, depth
			{
				position89 := position
				depth++
				{
					position90, tokenIndex90, depth90 := position, tokenIndex, depth
					if !_rules[ruleEq]() {
						goto l91
					}
					goto l90
				l91:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleNeq]() {
						goto l92
					}
					goto l90
				l92:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleLeq]() {
						goto l93
					}
					goto l90
				l93:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleLt]() {
						goto l94
					}
					goto l90
				l94:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleGeq]() {
						goto l95
					}
					goto l90
				l95:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleGt]() {
						goto l96
					}
					goto l90
				l96:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleTwiddleWakka]() {
						goto l88
					}
				}
			l90:
				depth--
				add(ruleVersionOp, position89)
			}
			return true
		l88:
			position, tokenIndex, depth = position88, tokenIndex88, depth88
			return false
		},
		/* 18 Platform <- <(Indent2 NotWhitespace+ LineEnd)> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l97
				}
				if !_rules[ruleNotWhitespace]() {
					goto l97
				}
			l99:
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					if !_rules[ruleNotWhitespace]() {
						goto l100
					}
					goto l99
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
				}
				if !_rules[ruleLineEnd]() {
					goto l97
				}
				depth--
				add(rulePlatform, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 19 Eq <- <'='> */
		func() bool {
			position101, tokenIndex101, depth101 := position, tokenIndex, depth
			{
				position102 := position
				depth++
				if buffer[position] != rune('=') {
					goto l101
				}
				position++
				depth--
				add(ruleEq, position102)
			}
			return true
		l101:
			position, tokenIndex, depth = position101, tokenIndex101, depth101
			return false
		},
		/* 20 Neq <- <('!' '=')> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				if buffer[position] != rune('!') {
					goto l103
				}
				position++
				if buffer[position] != rune('=') {
					goto l103
				}
				position++
				depth--
				add(ruleNeq, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 21 Leq <- <('<' '=')> */
		func() bool {
			position105, tokenIndex105, depth105 := position, tokenIndex, depth
			{
				position106 := position
				depth++
				if buffer[position] != rune('<') {
					goto l105
				}
				position++
				if buffer[position] != rune('=') {
					goto l105
				}
				position++
				depth--
				add(ruleLeq, position106)
			}
			return true
		l105:
			position, tokenIndex, depth = position105, tokenIndex105, depth105
			return false
		},
		/* 22 Lt <- <'<'> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				if buffer[position] != rune('<') {
					goto l107
				}
				position++
				depth--
				add(ruleLt, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 23 Geq <- <('>' '=')> */
		func() bool {
			position109, tokenIndex109, depth109 := position, tokenIndex, depth
			{
				position110 := position
				depth++
				if buffer[position] != rune('>') {
					goto l109
				}
				position++
				if buffer[position] != rune('=') {
					goto l109
				}
				position++
				depth--
				add(ruleGeq, position110)
			}
			return true
		l109:
			position, tokenIndex, depth = position109, tokenIndex109, depth109
			return false
		},
		/* 24 Gt <- <'>'> */
		func() bool {
			position111, tokenIndex111, depth111 := position, tokenIndex, depth
			{
				position112 := position
				depth++
				if buffer[position] != rune('>') {
					goto l111
				}
				position++
				depth--
				add(ruleGt, position112)
			}
			return true
		l111:
			position, tokenIndex, depth = position111, tokenIndex111, depth111
			return false
		},
		/* 25 TwiddleWakka <- <('~' '>')> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				if buffer[position] != rune('~') {
					goto l113
				}
				position++
				if buffer[position] != rune('>') {
					goto l113
				}
				position++
				depth--
				add(ruleTwiddleWakka, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 26 Space <- <(' ' / '\t')> */
		func() bool {
			position115, tokenIndex115, depth115 := position, tokenIndex, depth
			{
				position116 := position
				depth++
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l118
					}
					position++
					goto l117
				l118:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if buffer[position] != rune('\t') {
						goto l115
					}
					position++
				}
			l117:
				depth--
				add(ruleSpace, position116)
			}
			return true
		l115:
			position, tokenIndex, depth = position115, tokenIndex115, depth115
			return false
		},
		/* 27 Spaces <- <Space*> */
		func() bool {
			{
				position120 := position
				depth++
			l121:
				{
					position122, tokenIndex122, depth122 := position, tokenIndex, depth
					if !_rules[ruleSpace]() {
						goto l122
					}
					goto l121
				l122:
					position, tokenIndex, depth = position122, tokenIndex122, depth122
				}
				depth--
				add(ruleSpaces, position120)
			}
			return true
		},
		/* 28 Indent2 <- <(Space Space)> */
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
				depth--
				add(ruleIndent2, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 29 Indent4 <- <(Space Space Space Space)> */
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
				depth--
				add(ruleIndent4, position126)
			}
			return true
		l125:
			position, tokenIndex, depth = position125, tokenIndex125, depth125
			return false
		},
		/* 30 Indent6 <- <(Space Space Space Space Space Space)> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				if !_rules[ruleSpace]() {
					goto l127
				}
				if !_rules[ruleSpace]() {
					goto l127
				}
				if !_rules[ruleSpace]() {
					goto l127
				}
				if !_rules[ruleSpace]() {
					goto l127
				}
				if !_rules[ruleSpace]() {
					goto l127
				}
				if !_rules[ruleSpace]() {
					goto l127
				}
				depth--
				add(ruleIndent6, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 31 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position129, tokenIndex129, depth129 := position, tokenIndex, depth
			{
				position130 := position
				depth++
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l132
					}
					position++
					if buffer[position] != rune('\n') {
						goto l132
					}
					position++
					goto l131
				l132:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if buffer[position] != rune('\n') {
						goto l133
					}
					position++
					goto l131
				l133:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if buffer[position] != rune('\r') {
						goto l129
					}
					position++
				}
			l131:
				depth--
				add(ruleEndOfLine, position130)
			}
			return true
		l129:
			position, tokenIndex, depth = position129, tokenIndex129, depth129
			return false
		},
		/* 32 NotWhitespace <- <(!(Space / EndOfLine) .)> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				{
					position136, tokenIndex136, depth136 := position, tokenIndex, depth
					{
						position137, tokenIndex137, depth137 := position, tokenIndex, depth
						if !_rules[ruleSpace]() {
							goto l138
						}
						goto l137
					l138:
						position, tokenIndex, depth = position137, tokenIndex137, depth137
						if !_rules[ruleEndOfLine]() {
							goto l136
						}
					}
				l137:
					goto l134
				l136:
					position, tokenIndex, depth = position136, tokenIndex136, depth136
				}
				if !matchDot() {
					goto l134
				}
				depth--
				add(ruleNotWhitespace, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 33 LineEnd <- <(Spaces EndOfLine)> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				if !_rules[ruleSpaces]() {
					goto l139
				}
				if !_rules[ruleEndOfLine]() {
					goto l139
				}
				depth--
				add(ruleLineEnd, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 34 EndOfFile <- <(EndOfLine* !.)> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
			l143:
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if !_rules[ruleEndOfLine]() {
						goto l144
					}
					goto l143
				l144:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
				}
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					if !matchDot() {
						goto l145
					}
					goto l141
				l145:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
				}
				depth--
				add(ruleEndOfFile, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 36 Action0 <- <{ p.addSource(RubyGems) }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 37 Action1 <- <{ p.addSource(Git) }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 38 Action2 <- <{ p.addSource(SVN) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 39 Action3 <- <{ p.addSource(Path) }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		nil,
		/* 41 Action4 <- <{ p.addOption(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 42 Action5 <- <{p.setState(ParsingSpec)}> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 43 Action6 <- <{p.setState(ParsingSpecDep)}> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 44 Action7 <- <{ p.setState(ParsingDependency) }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 45 Action8 <- <{ p.addGem(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 46 Action9 <- <{ p.addVersion(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
	}
	p.rules = _rules
}

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
	ruleDependency
	ruleGemName
	ruleVersion
	ruleConstraint
	ruleVersionOp
	ruleEq
	ruleNeq
	ruleLeq
	ruleLt
	ruleGeq
	ruleGt
	ruleTwiddleWakka
	rulePlatform
	ruleURL
	ruleSHA
	ruleNL
	ruleNotNL
	ruleNotSP
	ruleLineEnd
	ruleSpace
	ruleSpaces
	ruleEndOfLine
	ruleEndOfFile
	ruleIndent2
	ruleAction0
	rulePegText
	ruleAction1

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
	"Dependency",
	"GemName",
	"Version",
	"Constraint",
	"VersionOp",
	"Eq",
	"Neq",
	"Leq",
	"Lt",
	"Geq",
	"Gt",
	"TwiddleWakka",
	"Platform",
	"URL",
	"SHA",
	"NL",
	"NotNL",
	"NotSP",
	"LineEnd",
	"Space",
	"Spaces",
	"EndOfLine",
	"EndOfFile",
	"Indent2",
	"Action0",
	"PegText",
	"Action1",

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

type GemfileGrammar struct {
	Gemfile

	Buffer string
	buffer []rune
	rules  [37]func() bool
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
	p *GemfileGrammar
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

func (p *GemfileGrammar) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *GemfileGrammar) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *GemfileGrammar) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
		case ruleAction0:
			p.addSpec(buffer[begin:end])
		case ruleAction1:
			p.addSpecVersion(buffer[begin:end])

		}
	}
}

func (p *GemfileGrammar) Init() {
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
		/* 0 Gemfile <- <(Git? Gem Platforms Dependencies EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !_rules[ruleGit]() {
						goto l2
					}
					goto l3
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
			l3:
				if !_rules[ruleGem]() {
					goto l0
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
			position4, tokenIndex4, depth4 := position, tokenIndex, depth
			{
				position5 := position
				depth++
				if buffer[position] != rune('G') {
					goto l4
				}
				position++
				if buffer[position] != rune('I') {
					goto l4
				}
				position++
				if buffer[position] != rune('T') {
					goto l4
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l4
				}
				if !_rules[ruleRemote]() {
					goto l4
				}
				if !_rules[ruleRevision]() {
					goto l4
				}
				if !_rules[ruleSpecs]() {
					goto l4
				}
				if !_rules[ruleLineEnd]() {
					goto l4
				}
				depth--
				add(ruleGit, position5)
			}
			return true
		l4:
			position, tokenIndex, depth = position4, tokenIndex4, depth4
			return false
		},
		/* 2 Gem <- <('G' 'E' 'M' LineEnd Remote Specs LineEnd)> */
		func() bool {
			position6, tokenIndex6, depth6 := position, tokenIndex, depth
			{
				position7 := position
				depth++
				if buffer[position] != rune('G') {
					goto l6
				}
				position++
				if buffer[position] != rune('E') {
					goto l6
				}
				position++
				if buffer[position] != rune('M') {
					goto l6
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l6
				}
				if !_rules[ruleRemote]() {
					goto l6
				}
				if !_rules[ruleSpecs]() {
					goto l6
				}
				if !_rules[ruleLineEnd]() {
					goto l6
				}
				depth--
				add(ruleGem, position7)
			}
			return true
		l6:
			position, tokenIndex, depth = position6, tokenIndex6, depth6
			return false
		},
		/* 3 Platforms <- <('P' 'L' 'A' 'T' 'F' 'O' 'R' 'M' 'S' LineEnd Platform+ LineEnd)> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				if buffer[position] != rune('P') {
					goto l8
				}
				position++
				if buffer[position] != rune('L') {
					goto l8
				}
				position++
				if buffer[position] != rune('A') {
					goto l8
				}
				position++
				if buffer[position] != rune('T') {
					goto l8
				}
				position++
				if buffer[position] != rune('F') {
					goto l8
				}
				position++
				if buffer[position] != rune('O') {
					goto l8
				}
				position++
				if buffer[position] != rune('R') {
					goto l8
				}
				position++
				if buffer[position] != rune('M') {
					goto l8
				}
				position++
				if buffer[position] != rune('S') {
					goto l8
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l8
				}
				if !_rules[rulePlatform]() {
					goto l8
				}
			l10:
				{
					position11, tokenIndex11, depth11 := position, tokenIndex, depth
					if !_rules[rulePlatform]() {
						goto l11
					}
					goto l10
				l11:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
				}
				if !_rules[ruleLineEnd]() {
					goto l8
				}
				depth--
				add(rulePlatforms, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 4 Dependencies <- <('D' 'E' 'P' 'E' 'N' 'D' 'E' 'N' 'C' 'I' 'E' 'S' LineEnd Dependency+)> */
		func() bool {
			position12, tokenIndex12, depth12 := position, tokenIndex, depth
			{
				position13 := position
				depth++
				if buffer[position] != rune('D') {
					goto l12
				}
				position++
				if buffer[position] != rune('E') {
					goto l12
				}
				position++
				if buffer[position] != rune('P') {
					goto l12
				}
				position++
				if buffer[position] != rune('E') {
					goto l12
				}
				position++
				if buffer[position] != rune('N') {
					goto l12
				}
				position++
				if buffer[position] != rune('D') {
					goto l12
				}
				position++
				if buffer[position] != rune('E') {
					goto l12
				}
				position++
				if buffer[position] != rune('N') {
					goto l12
				}
				position++
				if buffer[position] != rune('C') {
					goto l12
				}
				position++
				if buffer[position] != rune('I') {
					goto l12
				}
				position++
				if buffer[position] != rune('E') {
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
				if !_rules[ruleDependency]() {
					goto l12
				}
			l14:
				{
					position15, tokenIndex15, depth15 := position, tokenIndex, depth
					if !_rules[ruleDependency]() {
						goto l15
					}
					goto l14
				l15:
					position, tokenIndex, depth = position15, tokenIndex15, depth15
				}
				depth--
				add(ruleDependencies, position13)
			}
			return true
		l12:
			position, tokenIndex, depth = position12, tokenIndex12, depth12
			return false
		},
		/* 5 Remote <- <(Indent2 ('r' 'e' 'm' 'o' 't' 'e' ':') Spaces URL LineEnd)> */
		func() bool {
			position16, tokenIndex16, depth16 := position, tokenIndex, depth
			{
				position17 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l16
				}
				if buffer[position] != rune('r') {
					goto l16
				}
				position++
				if buffer[position] != rune('e') {
					goto l16
				}
				position++
				if buffer[position] != rune('m') {
					goto l16
				}
				position++
				if buffer[position] != rune('o') {
					goto l16
				}
				position++
				if buffer[position] != rune('t') {
					goto l16
				}
				position++
				if buffer[position] != rune('e') {
					goto l16
				}
				position++
				if buffer[position] != rune(':') {
					goto l16
				}
				position++
				if !_rules[ruleSpaces]() {
					goto l16
				}
				if !_rules[ruleURL]() {
					goto l16
				}
				if !_rules[ruleLineEnd]() {
					goto l16
				}
				depth--
				add(ruleRemote, position17)
			}
			return true
		l16:
			position, tokenIndex, depth = position16, tokenIndex16, depth16
			return false
		},
		/* 6 Revision <- <(Indent2 ('r' 'e' 'v' 'i' 's' 'i' 'o' 'n' ':') Spaces SHA LineEnd)> */
		func() bool {
			position18, tokenIndex18, depth18 := position, tokenIndex, depth
			{
				position19 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l18
				}
				if buffer[position] != rune('r') {
					goto l18
				}
				position++
				if buffer[position] != rune('e') {
					goto l18
				}
				position++
				if buffer[position] != rune('v') {
					goto l18
				}
				position++
				if buffer[position] != rune('i') {
					goto l18
				}
				position++
				if buffer[position] != rune('s') {
					goto l18
				}
				position++
				if buffer[position] != rune('i') {
					goto l18
				}
				position++
				if buffer[position] != rune('o') {
					goto l18
				}
				position++
				if buffer[position] != rune('n') {
					goto l18
				}
				position++
				if buffer[position] != rune(':') {
					goto l18
				}
				position++
				if !_rules[ruleSpaces]() {
					goto l18
				}
				if !_rules[ruleSHA]() {
					goto l18
				}
				if !_rules[ruleLineEnd]() {
					goto l18
				}
				depth--
				add(ruleRevision, position19)
			}
			return true
		l18:
			position, tokenIndex, depth = position18, tokenIndex18, depth18
			return false
		},
		/* 7 Specs <- <(Indent2 ('s' 'p' 'e' 'c' 's' ':') LineEnd Spec+)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l20
				}
				if buffer[position] != rune('s') {
					goto l20
				}
				position++
				if buffer[position] != rune('p') {
					goto l20
				}
				position++
				if buffer[position] != rune('e') {
					goto l20
				}
				position++
				if buffer[position] != rune('c') {
					goto l20
				}
				position++
				if buffer[position] != rune('s') {
					goto l20
				}
				position++
				if buffer[position] != rune(':') {
					goto l20
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l20
				}
				if !_rules[ruleSpec]() {
					goto l20
				}
			l22:
				{
					position23, tokenIndex23, depth23 := position, tokenIndex, depth
					if !_rules[ruleSpec]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex, depth = position23, tokenIndex23, depth23
				}
				depth--
				add(ruleSpecs, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 8 Spec <- <(Indent2+ GemName Action0 Spaces Version? LineEnd)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l24
				}
			l26:
				{
					position27, tokenIndex27, depth27 := position, tokenIndex, depth
					if !_rules[ruleIndent2]() {
						goto l27
					}
					goto l26
				l27:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
				}
				if !_rules[ruleGemName]() {
					goto l24
				}
				if !_rules[ruleAction0]() {
					goto l24
				}
				if !_rules[ruleSpaces]() {
					goto l24
				}
				{
					position28, tokenIndex28, depth28 := position, tokenIndex, depth
					if !_rules[ruleVersion]() {
						goto l28
					}
					goto l29
				l28:
					position, tokenIndex, depth = position28, tokenIndex28, depth28
				}
			l29:
				if !_rules[ruleLineEnd]() {
					goto l24
				}
				depth--
				add(ruleSpec, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 9 Dependency <- <(Indent2 GemName Spaces Version? LineEnd)> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l30
				}
				if !_rules[ruleGemName]() {
					goto l30
				}
				if !_rules[ruleSpaces]() {
					goto l30
				}
				{
					position32, tokenIndex32, depth32 := position, tokenIndex, depth
					if !_rules[ruleVersion]() {
						goto l32
					}
					goto l33
				l32:
					position, tokenIndex, depth = position32, tokenIndex32, depth32
				}
			l33:
				if !_rules[ruleLineEnd]() {
					goto l30
				}
				depth--
				add(ruleDependency, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 10 GemName <- <<([a-z] / [A-Z] / '-' / '_' / '!' / [0-9])+>> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				{
					position36 := position
					depth++
					{
						position39, tokenIndex39, depth39 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l40
						}
						position++
						goto l39
					l40:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l41
						}
						position++
						goto l39
					l41:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
						if buffer[position] != rune('-') {
							goto l42
						}
						position++
						goto l39
					l42:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
						if buffer[position] != rune('_') {
							goto l43
						}
						position++
						goto l39
					l43:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
						if buffer[position] != rune('!') {
							goto l44
						}
						position++
						goto l39
					l44:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l34
						}
						position++
					}
				l39:
				l37:
					{
						position38, tokenIndex38, depth38 := position, tokenIndex, depth
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
								goto l38
							}
							position++
						}
					l45:
						goto l37
					l38:
						position, tokenIndex, depth = position38, tokenIndex38, depth38
					}
					depth--
					add(rulePegText, position36)
				}
				depth--
				add(ruleGemName, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 11 Version <- <(<('(' Constraint (',' ' ' Constraint)* ')')> Action1)> */
		func() bool {
			position51, tokenIndex51, depth51 := position, tokenIndex, depth
			{
				position52 := position
				depth++
				{
					position53 := position
					depth++
					if buffer[position] != rune('(') {
						goto l51
					}
					position++
					if !_rules[ruleConstraint]() {
						goto l51
					}
				l54:
					{
						position55, tokenIndex55, depth55 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l55
						}
						position++
						if buffer[position] != rune(' ') {
							goto l55
						}
						position++
						if !_rules[ruleConstraint]() {
							goto l55
						}
						goto l54
					l55:
						position, tokenIndex, depth = position55, tokenIndex55, depth55
					}
					if buffer[position] != rune(')') {
						goto l51
					}
					position++
					depth--
					add(rulePegText, position53)
				}
				if !_rules[ruleAction1]() {
					goto l51
				}
				depth--
				add(ruleVersion, position52)
			}
			return true
		l51:
			position, tokenIndex, depth = position51, tokenIndex51, depth51
			return false
		},
		/* 12 Constraint <- <(VersionOp? Spaces [0-9]+ ('.' [0-9]+)*)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				{
					position58, tokenIndex58, depth58 := position, tokenIndex, depth
					if !_rules[ruleVersionOp]() {
						goto l58
					}
					goto l59
				l58:
					position, tokenIndex, depth = position58, tokenIndex58, depth58
				}
			l59:
				if !_rules[ruleSpaces]() {
					goto l56
				}
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l56
				}
				position++
			l60:
				{
					position61, tokenIndex61, depth61 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l61
					}
					position++
					goto l60
				l61:
					position, tokenIndex, depth = position61, tokenIndex61, depth61
				}
			l62:
				{
					position63, tokenIndex63, depth63 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l63
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l63
					}
					position++
				l64:
					{
						position65, tokenIndex65, depth65 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l65
						}
						position++
						goto l64
					l65:
						position, tokenIndex, depth = position65, tokenIndex65, depth65
					}
					goto l62
				l63:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
				}
				depth--
				add(ruleConstraint, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 13 VersionOp <- <(Eq / Neq / Leq / Lt / Geq / Gt / TwiddleWakka)> */
		func() bool {
			position66, tokenIndex66, depth66 := position, tokenIndex, depth
			{
				position67 := position
				depth++
				{
					position68, tokenIndex68, depth68 := position, tokenIndex, depth
					if !_rules[ruleEq]() {
						goto l69
					}
					goto l68
				l69:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
					if !_rules[ruleNeq]() {
						goto l70
					}
					goto l68
				l70:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
					if !_rules[ruleLeq]() {
						goto l71
					}
					goto l68
				l71:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
					if !_rules[ruleLt]() {
						goto l72
					}
					goto l68
				l72:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
					if !_rules[ruleGeq]() {
						goto l73
					}
					goto l68
				l73:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
					if !_rules[ruleGt]() {
						goto l74
					}
					goto l68
				l74:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
					if !_rules[ruleTwiddleWakka]() {
						goto l66
					}
				}
			l68:
				depth--
				add(ruleVersionOp, position67)
			}
			return true
		l66:
			position, tokenIndex, depth = position66, tokenIndex66, depth66
			return false
		},
		/* 14 Eq <- <'='> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				if buffer[position] != rune('=') {
					goto l75
				}
				position++
				depth--
				add(ruleEq, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 15 Neq <- <('!' '=')> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if buffer[position] != rune('!') {
					goto l77
				}
				position++
				if buffer[position] != rune('=') {
					goto l77
				}
				position++
				depth--
				add(ruleNeq, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 16 Leq <- <('<' '=')> */
		func() bool {
			position79, tokenIndex79, depth79 := position, tokenIndex, depth
			{
				position80 := position
				depth++
				if buffer[position] != rune('<') {
					goto l79
				}
				position++
				if buffer[position] != rune('=') {
					goto l79
				}
				position++
				depth--
				add(ruleLeq, position80)
			}
			return true
		l79:
			position, tokenIndex, depth = position79, tokenIndex79, depth79
			return false
		},
		/* 17 Lt <- <'<'> */
		func() bool {
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				if buffer[position] != rune('<') {
					goto l81
				}
				position++
				depth--
				add(ruleLt, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 18 Geq <- <('>' '=')> */
		func() bool {
			position83, tokenIndex83, depth83 := position, tokenIndex, depth
			{
				position84 := position
				depth++
				if buffer[position] != rune('>') {
					goto l83
				}
				position++
				if buffer[position] != rune('=') {
					goto l83
				}
				position++
				depth--
				add(ruleGeq, position84)
			}
			return true
		l83:
			position, tokenIndex, depth = position83, tokenIndex83, depth83
			return false
		},
		/* 19 Gt <- <'>'> */
		func() bool {
			position85, tokenIndex85, depth85 := position, tokenIndex, depth
			{
				position86 := position
				depth++
				if buffer[position] != rune('>') {
					goto l85
				}
				position++
				depth--
				add(ruleGt, position86)
			}
			return true
		l85:
			position, tokenIndex, depth = position85, tokenIndex85, depth85
			return false
		},
		/* 20 TwiddleWakka <- <('~' '>')> */
		func() bool {
			position87, tokenIndex87, depth87 := position, tokenIndex, depth
			{
				position88 := position
				depth++
				if buffer[position] != rune('~') {
					goto l87
				}
				position++
				if buffer[position] != rune('>') {
					goto l87
				}
				position++
				depth--
				add(ruleTwiddleWakka, position88)
			}
			return true
		l87:
			position, tokenIndex, depth = position87, tokenIndex87, depth87
			return false
		},
		/* 21 Platform <- <(NotNL+ LineEnd)> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if !_rules[ruleNotNL]() {
					goto l89
				}
			l91:
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if !_rules[ruleNotNL]() {
						goto l92
					}
					goto l91
				l92:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
				}
				if !_rules[ruleLineEnd]() {
					goto l89
				}
				depth--
				add(rulePlatform, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 22 URL <- <NotSP+> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if !_rules[ruleNotSP]() {
					goto l93
				}
			l95:
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					if !_rules[ruleNotSP]() {
						goto l96
					}
					goto l95
				l96:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
				}
				depth--
				add(ruleURL, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 23 SHA <- <([a-z] / [A-Z] / [0-9])+> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l102
					}
					position++
					goto l101
				l102:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l103
					}
					position++
					goto l101
				l103:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l97
					}
					position++
				}
			l101:
			l99:
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					{
						position104, tokenIndex104, depth104 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l105
						}
						position++
						goto l104
					l105:
						position, tokenIndex, depth = position104, tokenIndex104, depth104
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l106
						}
						position++
						goto l104
					l106:
						position, tokenIndex, depth = position104, tokenIndex104, depth104
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l100
						}
						position++
					}
				l104:
					goto l99
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
				}
				depth--
				add(ruleSHA, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 24 NL <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				{
					position109, tokenIndex109, depth109 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l110
					}
					position++
					if buffer[position] != rune('\n') {
						goto l110
					}
					position++
					goto l109
				l110:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if buffer[position] != rune('\n') {
						goto l111
					}
					position++
					goto l109
				l111:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if buffer[position] != rune('\r') {
						goto l107
					}
					position++
				}
			l109:
				depth--
				add(ruleNL, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 25 NotNL <- <(!NL .)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if !_rules[ruleNL]() {
						goto l114
					}
					goto l112
				l114:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
				}
				if !matchDot() {
					goto l112
				}
				depth--
				add(ruleNotNL, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 26 NotSP <- <(!(Space / NL) .)> */
		func() bool {
			position115, tokenIndex115, depth115 := position, tokenIndex, depth
			{
				position116 := position
				depth++
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					{
						position118, tokenIndex118, depth118 := position, tokenIndex, depth
						if !_rules[ruleSpace]() {
							goto l119
						}
						goto l118
					l119:
						position, tokenIndex, depth = position118, tokenIndex118, depth118
						if !_rules[ruleNL]() {
							goto l117
						}
					}
				l118:
					goto l115
				l117:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
				}
				if !matchDot() {
					goto l115
				}
				depth--
				add(ruleNotSP, position116)
			}
			return true
		l115:
			position, tokenIndex, depth = position115, tokenIndex115, depth115
			return false
		},
		/* 27 LineEnd <- <(Spaces EndOfLine)> */
		func() bool {
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				if !_rules[ruleSpaces]() {
					goto l120
				}
				if !_rules[ruleEndOfLine]() {
					goto l120
				}
				depth--
				add(ruleLineEnd, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 28 Space <- <(' ' / '\t')> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				{
					position124, tokenIndex124, depth124 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l125
					}
					position++
					goto l124
				l125:
					position, tokenIndex, depth = position124, tokenIndex124, depth124
					if buffer[position] != rune('\t') {
						goto l122
					}
					position++
				}
			l124:
				depth--
				add(ruleSpace, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 29 Spaces <- <Space*> */
		func() bool {
			{
				position127 := position
				depth++
			l128:
				{
					position129, tokenIndex129, depth129 := position, tokenIndex, depth
					if !_rules[ruleSpace]() {
						goto l129
					}
					goto l128
				l129:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
				}
				depth--
				add(ruleSpaces, position127)
			}
			return true
		},
		/* 30 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l133
					}
					position++
					if buffer[position] != rune('\n') {
						goto l133
					}
					position++
					goto l132
				l133:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if buffer[position] != rune('\n') {
						goto l134
					}
					position++
					goto l132
				l134:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if buffer[position] != rune('\r') {
						goto l130
					}
					position++
				}
			l132:
				depth--
				add(ruleEndOfLine, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 31 EndOfFile <- <(EndOfLine* !.)> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
			l137:
				{
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					if !_rules[ruleEndOfLine]() {
						goto l138
					}
					goto l137
				l138:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
				}
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					if !matchDot() {
						goto l139
					}
					goto l135
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
				depth--
				add(ruleEndOfFile, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 32 Indent2 <- <(' ' ' ')> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				if buffer[position] != rune(' ') {
					goto l140
				}
				position++
				if buffer[position] != rune(' ') {
					goto l140
				}
				position++
				depth--
				add(ruleIndent2, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 34 Action0 <- <{ p.addSpec(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		nil,
		/* 36 Action1 <- <{ p.addSpecVersion(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
	}
	p.rules = _rules
}

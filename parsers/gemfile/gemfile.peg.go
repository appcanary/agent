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
	ruleGemName
	ruleSpecVersion
	ruleSpecDepVersion
	ruleDependencyVersion
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
	ruleIndent4
	ruleIndent6
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
	ruleAction3
	ruleAction4
	ruleAction5

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
	"GemName",
	"SpecVersion",
	"SpecDepVersion",
	"DependencyVersion",
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
	"Indent4",
	"Indent6",
	"Action0",
	"Action1",
	"Action2",
	"PegText",
	"Action3",
	"Action4",
	"Action5",

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
			p.addSpecDep(buffer[begin:end])
		case ruleAction2:
			p.addDependency(buffer[begin:end])
		case ruleAction3:
			p.addSpecVersion(buffer[begin:end])
		case ruleAction4:
			p.addSpecDepVersion(buffer[begin:end])
		case ruleAction5:
			p.addDependencyVersion(buffer[begin:end])

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
		/* 0 Gemfile <- <((Git / Gem / Platforms)+ Dependencies EndOfFile)> */
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
						goto l6
					}
					goto l4
				l6:
					position, tokenIndex, depth = position4, tokenIndex4, depth4
					if !_rules[rulePlatforms]() {
						goto l0
					}
				}
			l4:
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					{
						position7, tokenIndex7, depth7 := position, tokenIndex, depth
						if !_rules[ruleGit]() {
							goto l8
						}
						goto l7
					l8:
						position, tokenIndex, depth = position7, tokenIndex7, depth7
						if !_rules[ruleGem]() {
							goto l9
						}
						goto l7
					l9:
						position, tokenIndex, depth = position7, tokenIndex7, depth7
						if !_rules[rulePlatforms]() {
							goto l3
						}
					}
				l7:
					goto l2
				l3:
					position, tokenIndex, depth = position3, tokenIndex3, depth3
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
				if !_rules[ruleRemote]() {
					goto l10
				}
				if !_rules[ruleRevision]() {
					goto l10
				}
				if !_rules[ruleSpecs]() {
					goto l10
				}
				if !_rules[ruleLineEnd]() {
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
		/* 2 Gem <- <('G' 'E' 'M' LineEnd Remote Specs LineEnd)> */
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
				if !_rules[ruleRemote]() {
					goto l12
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
		/* 3 Platforms <- <('P' 'L' 'A' 'T' 'F' 'O' 'R' 'M' 'S' LineEnd Platform+ LineEnd)> */
		func() bool {
			position14, tokenIndex14, depth14 := position, tokenIndex, depth
			{
				position15 := position
				depth++
				if buffer[position] != rune('P') {
					goto l14
				}
				position++
				if buffer[position] != rune('L') {
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
				if buffer[position] != rune('F') {
					goto l14
				}
				position++
				if buffer[position] != rune('O') {
					goto l14
				}
				position++
				if buffer[position] != rune('R') {
					goto l14
				}
				position++
				if buffer[position] != rune('M') {
					goto l14
				}
				position++
				if buffer[position] != rune('S') {
					goto l14
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l14
				}
				if !_rules[rulePlatform]() {
					goto l14
				}
			l16:
				{
					position17, tokenIndex17, depth17 := position, tokenIndex, depth
					if !_rules[rulePlatform]() {
						goto l17
					}
					goto l16
				l17:
					position, tokenIndex, depth = position17, tokenIndex17, depth17
				}
				if !_rules[ruleLineEnd]() {
					goto l14
				}
				depth--
				add(rulePlatforms, position15)
			}
			return true
		l14:
			position, tokenIndex, depth = position14, tokenIndex14, depth14
			return false
		},
		/* 4 Dependencies <- <('D' 'E' 'P' 'E' 'N' 'D' 'E' 'N' 'C' 'I' 'E' 'S' LineEnd Dependency+)> */
		func() bool {
			position18, tokenIndex18, depth18 := position, tokenIndex, depth
			{
				position19 := position
				depth++
				if buffer[position] != rune('D') {
					goto l18
				}
				position++
				if buffer[position] != rune('E') {
					goto l18
				}
				position++
				if buffer[position] != rune('P') {
					goto l18
				}
				position++
				if buffer[position] != rune('E') {
					goto l18
				}
				position++
				if buffer[position] != rune('N') {
					goto l18
				}
				position++
				if buffer[position] != rune('D') {
					goto l18
				}
				position++
				if buffer[position] != rune('E') {
					goto l18
				}
				position++
				if buffer[position] != rune('N') {
					goto l18
				}
				position++
				if buffer[position] != rune('C') {
					goto l18
				}
				position++
				if buffer[position] != rune('I') {
					goto l18
				}
				position++
				if buffer[position] != rune('E') {
					goto l18
				}
				position++
				if buffer[position] != rune('S') {
					goto l18
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l18
				}
				if !_rules[ruleDependency]() {
					goto l18
				}
			l20:
				{
					position21, tokenIndex21, depth21 := position, tokenIndex, depth
					if !_rules[ruleDependency]() {
						goto l21
					}
					goto l20
				l21:
					position, tokenIndex, depth = position21, tokenIndex21, depth21
				}
				depth--
				add(ruleDependencies, position19)
			}
			return true
		l18:
			position, tokenIndex, depth = position18, tokenIndex18, depth18
			return false
		},
		/* 5 Remote <- <(Indent2 ('r' 'e' 'm' 'o' 't' 'e' ':') Spaces URL LineEnd)> */
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
				if buffer[position] != rune('m') {
					goto l22
				}
				position++
				if buffer[position] != rune('o') {
					goto l22
				}
				position++
				if buffer[position] != rune('t') {
					goto l22
				}
				position++
				if buffer[position] != rune('e') {
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
				if !_rules[ruleURL]() {
					goto l22
				}
				if !_rules[ruleLineEnd]() {
					goto l22
				}
				depth--
				add(ruleRemote, position23)
			}
			return true
		l22:
			position, tokenIndex, depth = position22, tokenIndex22, depth22
			return false
		},
		/* 6 Revision <- <(Indent2 ('r' 'e' 'v' 'i' 's' 'i' 'o' 'n' ':') Spaces SHA LineEnd)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l24
				}
				if buffer[position] != rune('r') {
					goto l24
				}
				position++
				if buffer[position] != rune('e') {
					goto l24
				}
				position++
				if buffer[position] != rune('v') {
					goto l24
				}
				position++
				if buffer[position] != rune('i') {
					goto l24
				}
				position++
				if buffer[position] != rune('s') {
					goto l24
				}
				position++
				if buffer[position] != rune('i') {
					goto l24
				}
				position++
				if buffer[position] != rune('o') {
					goto l24
				}
				position++
				if buffer[position] != rune('n') {
					goto l24
				}
				position++
				if buffer[position] != rune(':') {
					goto l24
				}
				position++
				if !_rules[ruleSpaces]() {
					goto l24
				}
				if !_rules[ruleSHA]() {
					goto l24
				}
				if !_rules[ruleLineEnd]() {
					goto l24
				}
				depth--
				add(ruleRevision, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 7 Specs <- <(Indent2 ('s' 'p' 'e' 'c' 's' ':') LineEnd Spec+)> */
		func() bool {
			position26, tokenIndex26, depth26 := position, tokenIndex, depth
			{
				position27 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l26
				}
				if buffer[position] != rune('s') {
					goto l26
				}
				position++
				if buffer[position] != rune('p') {
					goto l26
				}
				position++
				if buffer[position] != rune('e') {
					goto l26
				}
				position++
				if buffer[position] != rune('c') {
					goto l26
				}
				position++
				if buffer[position] != rune('s') {
					goto l26
				}
				position++
				if buffer[position] != rune(':') {
					goto l26
				}
				position++
				if !_rules[ruleLineEnd]() {
					goto l26
				}
				if !_rules[ruleSpec]() {
					goto l26
				}
			l28:
				{
					position29, tokenIndex29, depth29 := position, tokenIndex, depth
					if !_rules[ruleSpec]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex, depth = position29, tokenIndex29, depth29
				}
				depth--
				add(ruleSpecs, position27)
			}
			return true
		l26:
			position, tokenIndex, depth = position26, tokenIndex26, depth26
			return false
		},
		/* 8 Spec <- <(Indent4 GemName Action0 Spaces SpecVersion? LineEnd SpecDep*)> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				if !_rules[ruleIndent4]() {
					goto l30
				}
				if !_rules[ruleGemName]() {
					goto l30
				}
				if !_rules[ruleAction0]() {
					goto l30
				}
				if !_rules[ruleSpaces]() {
					goto l30
				}
				{
					position32, tokenIndex32, depth32 := position, tokenIndex, depth
					if !_rules[ruleSpecVersion]() {
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
			l34:
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !_rules[ruleSpecDep]() {
						goto l35
					}
					goto l34
				l35:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
				}
				depth--
				add(ruleSpec, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 9 SpecDep <- <(Indent6 GemName Action1 Spaces SpecDepVersion? LineEnd)> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if !_rules[ruleIndent6]() {
					goto l36
				}
				if !_rules[ruleGemName]() {
					goto l36
				}
				if !_rules[ruleAction1]() {
					goto l36
				}
				if !_rules[ruleSpaces]() {
					goto l36
				}
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					if !_rules[ruleSpecDepVersion]() {
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
				add(ruleSpecDep, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 10 Dependency <- <(Indent2 GemName Action2 Spaces DependencyVersion? LineEnd)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if !_rules[ruleIndent2]() {
					goto l40
				}
				if !_rules[ruleGemName]() {
					goto l40
				}
				if !_rules[ruleAction2]() {
					goto l40
				}
				if !_rules[ruleSpaces]() {
					goto l40
				}
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !_rules[ruleDependencyVersion]() {
						goto l42
					}
					goto l43
				l42:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
				}
			l43:
				if !_rules[ruleLineEnd]() {
					goto l40
				}
				depth--
				add(ruleDependency, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 11 GemName <- <<([a-z] / [A-Z] / '-' / '_' / '!' / [0-9])+>> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				{
					position46 := position
					depth++
					{
						position49, tokenIndex49, depth49 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l50
						}
						position++
						goto l49
					l50:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l51
						}
						position++
						goto l49
					l51:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if buffer[position] != rune('-') {
							goto l52
						}
						position++
						goto l49
					l52:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if buffer[position] != rune('_') {
							goto l53
						}
						position++
						goto l49
					l53:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if buffer[position] != rune('!') {
							goto l54
						}
						position++
						goto l49
					l54:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l44
						}
						position++
					}
				l49:
				l47:
					{
						position48, tokenIndex48, depth48 := position, tokenIndex, depth
						{
							position55, tokenIndex55, depth55 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l56
							}
							position++
							goto l55
						l56:
							position, tokenIndex, depth = position55, tokenIndex55, depth55
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l57
							}
							position++
							goto l55
						l57:
							position, tokenIndex, depth = position55, tokenIndex55, depth55
							if buffer[position] != rune('-') {
								goto l58
							}
							position++
							goto l55
						l58:
							position, tokenIndex, depth = position55, tokenIndex55, depth55
							if buffer[position] != rune('_') {
								goto l59
							}
							position++
							goto l55
						l59:
							position, tokenIndex, depth = position55, tokenIndex55, depth55
							if buffer[position] != rune('!') {
								goto l60
							}
							position++
							goto l55
						l60:
							position, tokenIndex, depth = position55, tokenIndex55, depth55
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l48
							}
							position++
						}
					l55:
						goto l47
					l48:
						position, tokenIndex, depth = position48, tokenIndex48, depth48
					}
					depth--
					add(rulePegText, position46)
				}
				depth--
				add(ruleGemName, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 12 SpecVersion <- <(<Version> Action3)> */
		func() bool {
			position61, tokenIndex61, depth61 := position, tokenIndex, depth
			{
				position62 := position
				depth++
				{
					position63 := position
					depth++
					if !_rules[ruleVersion]() {
						goto l61
					}
					depth--
					add(rulePegText, position63)
				}
				if !_rules[ruleAction3]() {
					goto l61
				}
				depth--
				add(ruleSpecVersion, position62)
			}
			return true
		l61:
			position, tokenIndex, depth = position61, tokenIndex61, depth61
			return false
		},
		/* 13 SpecDepVersion <- <(<Version> Action4)> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				{
					position66 := position
					depth++
					if !_rules[ruleVersion]() {
						goto l64
					}
					depth--
					add(rulePegText, position66)
				}
				if !_rules[ruleAction4]() {
					goto l64
				}
				depth--
				add(ruleSpecDepVersion, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 14 DependencyVersion <- <(<Version> Action5)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				{
					position69 := position
					depth++
					if !_rules[ruleVersion]() {
						goto l67
					}
					depth--
					add(rulePegText, position69)
				}
				if !_rules[ruleAction5]() {
					goto l67
				}
				depth--
				add(ruleDependencyVersion, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 15 Version <- <('(' Constraint (',' ' ' Constraint)* ')')> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				if buffer[position] != rune('(') {
					goto l70
				}
				position++
				if !_rules[ruleConstraint]() {
					goto l70
				}
			l72:
				{
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l73
					}
					position++
					if buffer[position] != rune(' ') {
						goto l73
					}
					position++
					if !_rules[ruleConstraint]() {
						goto l73
					}
					goto l72
				l73:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
				}
				if buffer[position] != rune(')') {
					goto l70
				}
				position++
				depth--
				add(ruleVersion, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 16 Constraint <- <(VersionOp? Spaces [0-9]+ ('.' [0-9]+)*)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				{
					position76, tokenIndex76, depth76 := position, tokenIndex, depth
					if !_rules[ruleVersionOp]() {
						goto l76
					}
					goto l77
				l76:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
				}
			l77:
				if !_rules[ruleSpaces]() {
					goto l74
				}
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l74
				}
				position++
			l78:
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l79
					}
					position++
					goto l78
				l79:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
				}
			l80:
				{
					position81, tokenIndex81, depth81 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l81
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l81
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
					goto l80
				l81:
					position, tokenIndex, depth = position81, tokenIndex81, depth81
				}
				depth--
				add(ruleConstraint, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 17 VersionOp <- <(Eq / Neq / Leq / Lt / Geq / Gt / TwiddleWakka)> */
		func() bool {
			position84, tokenIndex84, depth84 := position, tokenIndex, depth
			{
				position85 := position
				depth++
				{
					position86, tokenIndex86, depth86 := position, tokenIndex, depth
					if !_rules[ruleEq]() {
						goto l87
					}
					goto l86
				l87:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if !_rules[ruleNeq]() {
						goto l88
					}
					goto l86
				l88:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if !_rules[ruleLeq]() {
						goto l89
					}
					goto l86
				l89:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if !_rules[ruleLt]() {
						goto l90
					}
					goto l86
				l90:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if !_rules[ruleGeq]() {
						goto l91
					}
					goto l86
				l91:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if !_rules[ruleGt]() {
						goto l92
					}
					goto l86
				l92:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if !_rules[ruleTwiddleWakka]() {
						goto l84
					}
				}
			l86:
				depth--
				add(ruleVersionOp, position85)
			}
			return true
		l84:
			position, tokenIndex, depth = position84, tokenIndex84, depth84
			return false
		},
		/* 18 Eq <- <'='> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if buffer[position] != rune('=') {
					goto l93
				}
				position++
				depth--
				add(ruleEq, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 19 Neq <- <('!' '=')> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				if buffer[position] != rune('!') {
					goto l95
				}
				position++
				if buffer[position] != rune('=') {
					goto l95
				}
				position++
				depth--
				add(ruleNeq, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 20 Leq <- <('<' '=')> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				if buffer[position] != rune('<') {
					goto l97
				}
				position++
				if buffer[position] != rune('=') {
					goto l97
				}
				position++
				depth--
				add(ruleLeq, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 21 Lt <- <'<'> */
		func() bool {
			position99, tokenIndex99, depth99 := position, tokenIndex, depth
			{
				position100 := position
				depth++
				if buffer[position] != rune('<') {
					goto l99
				}
				position++
				depth--
				add(ruleLt, position100)
			}
			return true
		l99:
			position, tokenIndex, depth = position99, tokenIndex99, depth99
			return false
		},
		/* 22 Geq <- <('>' '=')> */
		func() bool {
			position101, tokenIndex101, depth101 := position, tokenIndex, depth
			{
				position102 := position
				depth++
				if buffer[position] != rune('>') {
					goto l101
				}
				position++
				if buffer[position] != rune('=') {
					goto l101
				}
				position++
				depth--
				add(ruleGeq, position102)
			}
			return true
		l101:
			position, tokenIndex, depth = position101, tokenIndex101, depth101
			return false
		},
		/* 23 Gt <- <'>'> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				if buffer[position] != rune('>') {
					goto l103
				}
				position++
				depth--
				add(ruleGt, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 24 TwiddleWakka <- <('~' '>')> */
		func() bool {
			position105, tokenIndex105, depth105 := position, tokenIndex, depth
			{
				position106 := position
				depth++
				if buffer[position] != rune('~') {
					goto l105
				}
				position++
				if buffer[position] != rune('>') {
					goto l105
				}
				position++
				depth--
				add(ruleTwiddleWakka, position106)
			}
			return true
		l105:
			position, tokenIndex, depth = position105, tokenIndex105, depth105
			return false
		},
		/* 25 Platform <- <(NotNL+ LineEnd)> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				if !_rules[ruleNotNL]() {
					goto l107
				}
			l109:
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					if !_rules[ruleNotNL]() {
						goto l110
					}
					goto l109
				l110:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
				}
				if !_rules[ruleLineEnd]() {
					goto l107
				}
				depth--
				add(rulePlatform, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 26 URL <- <NotSP+> */
		func() bool {
			position111, tokenIndex111, depth111 := position, tokenIndex, depth
			{
				position112 := position
				depth++
				if !_rules[ruleNotSP]() {
					goto l111
				}
			l113:
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if !_rules[ruleNotSP]() {
						goto l114
					}
					goto l113
				l114:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
				}
				depth--
				add(ruleURL, position112)
			}
			return true
		l111:
			position, tokenIndex, depth = position111, tokenIndex111, depth111
			return false
		},
		/* 27 SHA <- <([a-z] / [A-Z] / [0-9])+> */
		func() bool {
			position115, tokenIndex115, depth115 := position, tokenIndex, depth
			{
				position116 := position
				depth++
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l120
					}
					position++
					goto l119
				l120:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l121
					}
					position++
					goto l119
				l121:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l115
					}
					position++
				}
			l119:
			l117:
				{
					position118, tokenIndex118, depth118 := position, tokenIndex, depth
					{
						position122, tokenIndex122, depth122 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l123
						}
						position++
						goto l122
					l123:
						position, tokenIndex, depth = position122, tokenIndex122, depth122
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l124
						}
						position++
						goto l122
					l124:
						position, tokenIndex, depth = position122, tokenIndex122, depth122
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l118
						}
						position++
					}
				l122:
					goto l117
				l118:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
				}
				depth--
				add(ruleSHA, position116)
			}
			return true
		l115:
			position, tokenIndex, depth = position115, tokenIndex115, depth115
			return false
		},
		/* 28 NL <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position125, tokenIndex125, depth125 := position, tokenIndex, depth
			{
				position126 := position
				depth++
				{
					position127, tokenIndex127, depth127 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l128
					}
					position++
					if buffer[position] != rune('\n') {
						goto l128
					}
					position++
					goto l127
				l128:
					position, tokenIndex, depth = position127, tokenIndex127, depth127
					if buffer[position] != rune('\n') {
						goto l129
					}
					position++
					goto l127
				l129:
					position, tokenIndex, depth = position127, tokenIndex127, depth127
					if buffer[position] != rune('\r') {
						goto l125
					}
					position++
				}
			l127:
				depth--
				add(ruleNL, position126)
			}
			return true
		l125:
			position, tokenIndex, depth = position125, tokenIndex125, depth125
			return false
		},
		/* 29 NotNL <- <(!NL .)> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					if !_rules[ruleNL]() {
						goto l132
					}
					goto l130
				l132:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
				}
				if !matchDot() {
					goto l130
				}
				depth--
				add(ruleNotNL, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 30 NotSP <- <(!(Space / NL) .)> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					{
						position136, tokenIndex136, depth136 := position, tokenIndex, depth
						if !_rules[ruleSpace]() {
							goto l137
						}
						goto l136
					l137:
						position, tokenIndex, depth = position136, tokenIndex136, depth136
						if !_rules[ruleNL]() {
							goto l135
						}
					}
				l136:
					goto l133
				l135:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
				}
				if !matchDot() {
					goto l133
				}
				depth--
				add(ruleNotSP, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 31 LineEnd <- <(Spaces EndOfLine)> */
		func() bool {
			position138, tokenIndex138, depth138 := position, tokenIndex, depth
			{
				position139 := position
				depth++
				if !_rules[ruleSpaces]() {
					goto l138
				}
				if !_rules[ruleEndOfLine]() {
					goto l138
				}
				depth--
				add(ruleLineEnd, position139)
			}
			return true
		l138:
			position, tokenIndex, depth = position138, tokenIndex138, depth138
			return false
		},
		/* 32 Space <- <(' ' / '\t')> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l143
					}
					position++
					goto l142
				l143:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
					if buffer[position] != rune('\t') {
						goto l140
					}
					position++
				}
			l142:
				depth--
				add(ruleSpace, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 33 Spaces <- <Space*> */
		func() bool {
			{
				position145 := position
				depth++
			l146:
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if !_rules[ruleSpace]() {
						goto l147
					}
					goto l146
				l147:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
				}
				depth--
				add(ruleSpaces, position145)
			}
			return true
		},
		/* 34 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position148, tokenIndex148, depth148 := position, tokenIndex, depth
			{
				position149 := position
				depth++
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l151
					}
					position++
					if buffer[position] != rune('\n') {
						goto l151
					}
					position++
					goto l150
				l151:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
					if buffer[position] != rune('\n') {
						goto l152
					}
					position++
					goto l150
				l152:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
					if buffer[position] != rune('\r') {
						goto l148
					}
					position++
				}
			l150:
				depth--
				add(ruleEndOfLine, position149)
			}
			return true
		l148:
			position, tokenIndex, depth = position148, tokenIndex148, depth148
			return false
		},
		/* 35 EndOfFile <- <(EndOfLine* !.)> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
			l155:
				{
					position156, tokenIndex156, depth156 := position, tokenIndex, depth
					if !_rules[ruleEndOfLine]() {
						goto l156
					}
					goto l155
				l156:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
				}
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					if !matchDot() {
						goto l157
					}
					goto l153
				l157:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
				}
				depth--
				add(ruleEndOfFile, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 36 Indent2 <- <(' ' ' ')> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				if buffer[position] != rune(' ') {
					goto l158
				}
				position++
				if buffer[position] != rune(' ') {
					goto l158
				}
				position++
				depth--
				add(ruleIndent2, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 37 Indent4 <- <(' ' ' ' ' ' ' ')> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				if buffer[position] != rune(' ') {
					goto l160
				}
				position++
				if buffer[position] != rune(' ') {
					goto l160
				}
				position++
				if buffer[position] != rune(' ') {
					goto l160
				}
				position++
				if buffer[position] != rune(' ') {
					goto l160
				}
				position++
				depth--
				add(ruleIndent4, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 38 Indent6 <- <(' ' ' ' ' ' ' ' ' ' ' ')> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				if buffer[position] != rune(' ') {
					goto l162
				}
				position++
				if buffer[position] != rune(' ') {
					goto l162
				}
				position++
				if buffer[position] != rune(' ') {
					goto l162
				}
				position++
				if buffer[position] != rune(' ') {
					goto l162
				}
				position++
				if buffer[position] != rune(' ') {
					goto l162
				}
				position++
				if buffer[position] != rune(' ') {
					goto l162
				}
				position++
				depth--
				add(ruleIndent6, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 40 Action0 <- <{ p.addSpec(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 41 Action1 <- <{ p.addSpecDep(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 42 Action2 <- <{ p.addDependency(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		nil,
		/* 44 Action3 <- <{ p.addSpecVersion(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 45 Action4 <- <{ p.addSpecDepVersion(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 46 Action5 <- <{ p.addDependencyVersion(buffer[begin:end])}> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
	}
	p.rules = _rules
}

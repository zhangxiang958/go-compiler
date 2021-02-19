package main

import (
	"fmt"
	"log"
	"strings"
)

type tokenType string

const (
	// Parenthesis 括号类型
	Parenthesis tokenType = "paren"
	// Numeric 数字类型
	Numeric tokenType = "number"
	// Name 名字类型
	Name tokenType = "name"
	// Character 字符类型
	Character tokenType = "character"
)

type token struct {
	Kind  tokenType
	Value string
}

func tokenizer(input string) []token {
	input += "\n"

	current := 0

	tokens := []token{}

	for current < len([]rune(input)) {
		char := string([]rune(input)[current])

		if char == "(" {
			tokens = append(tokens, token{
				Kind:  Parenthesis,
				Value: "(",
			})
			current++
			continue
		}
		if char == ")" {
			tokens = append(tokens, token{
				Kind:  Parenthesis,
				Value: ")",
			})
			current++
			continue
		}
		if char == " " {
			current++
			continue
		}
		if isNumber(char) {
			value := ""
			for isNumber(char) {
				value += char
				current++
				char = string([]rune(input)[current])
			}
			tokens = append(tokens, token{
				Kind:  Numeric,
				Value: value,
			})
			continue
		}
		if isLetter(char) {
			value := ""
			for isLetter(char) {
				value += char
				current++
				char = string([]rune(input)[current])
			}
			tokens = append(tokens, token{
				Kind:  Name,
				Value: value,
			})
			continue
		}
		break
	}

	return tokens
}

// NodeType node 类型
type NodeType string

const (
	// Program node 类型
	Program NodeType = "program"
	// NumberLiteral node 类型
	NumberLiteral NodeType = "numberLiteral"
	// CallExpression node 类型
	CallExpression NodeType = "callexpression"
	// Identifier node 类型
	Identifier NodeType = "identifier"
	// ExpressionStatement node 类型
	ExpressionStatement NodeType = "expressStatement"
)

type node struct {
	Kind       NodeType
	Value      string
	Name       string
	Callee     *node
	Expression *node
	Body       []node
	Params     []node
	Arguments  *[]node
	Context    *[]node
}

type ast node

var pc int
var pt []token

func parser(tokens []token) ast {
	pc = 0
	pt = tokens

	ast := ast{
		Kind: Program,
		Body: []node{},
	}

	for pc < len(pt) {
		ast.Body = append(ast.Body, walk())
	}
	return ast
}

func walk() node {
	token := pt[pc]
	if token.Kind == Numeric {
		pc++
		return node{
			Kind:  NumberLiteral,
			Value: token.Value,
		}
	}
	if token.Kind == Parenthesis && token.Value == "(" {
		pc++
		token = pt[pc]

		n := node{
			Kind:   CallExpression,
			Value:  token.Value,
			Params: []node{},
		}

		pc++
		token = pt[pc]
		for token.Kind != Parenthesis || (token.Kind == Parenthesis && token.Value != ")") {
			n.Params = append(n.Params, walk())
			token = pt[pc]
		}

		pc++
		return n
	}

	log.Fatal(token.Kind)
	return node{}
}

type visitor map[string]struct {
	Enter func(n *node, parent node) int
	Leave func(n *node, parent node, idx int)
}

func traverseNode(n, p node, v visitor) {
	nodeVisitor, exist := v[string(n.Kind)]
	idx = -1
	if exist {
		idx = nodeVisitor.Enter(&n, p)
	}

	switch n.Kind {
	case Program:
		traverseArray(n.Body, n, v)
	case CallExpression:
		traverseArray(n.Params, n, v)
	case NumberLiteral:
		break
	default:
		log.Fatal(n.Kind)
	}

	if nodeVisitor.Leave != nil && idx != -1 {
		nodeVisitor.Leave(&n, p, idx)
	}
}

func traverseArray(a []node, p node, v visitor) {
	for _, child := range a {
		traverseNode(child, p, v)
	}
}

func traverser(a ast, v visitor) {
	traverseNode(node(a), node{}, v)
}

var idx int

func transformer(a ast) ast {
	newAst := ast{
		Kind: Program,
		Body: []node{},
	}

	traverseStack := make([]*[]node, 0)
	// 这里都是向父节点插入数据的
	traverseStack = append(traverseStack, &newAst.Body)
	traverser(a, map[string]struct {
		Enter func(n *node, parent node) int
		Leave func(n *node, parent node, idx int)
	}{
		string(NumberLiteral): struct {
			Enter func(n *node, parent node) int
			Leave func(n *node, parent node, idx int)
		}{
			Enter: func(n *node, parent node) int {
				currentStack := traverseStack[len(traverseStack)-1]
				*currentStack = append(*currentStack, node{
					Kind:  NumberLiteral,
					Value: n.Value,
				})
				return -1
			},
			Leave: nil,
		},
		string(CallExpression): struct {
			Enter func(n *node, parent node) int
			Leave func(n *node, parent node, idx int)
		}{
			Enter: func(n *node, parent node) int {
				arguments := make([]node, 0)
				e := node{
					Kind: CallExpression,
					Callee: &node{
						Kind: Identifier,
						Name: n.Value,
					},
					Arguments: &arguments,
				}

				if parent.Kind != CallExpression {
					es := node{
						Kind:       ExpressionStatement,
						Expression: &e,
					}
					currentStack := traverseStack[len(traverseStack)-1]
					*currentStack = append(*currentStack, es)
				} else {
					currentStack := traverseStack[len(traverseStack)-1]
					*currentStack = append(*currentStack, e)
				}

				traverseStack = append(traverseStack, e.Arguments)
				return len(traverseStack) - 1
			},
			Leave: func(n *node, parent node, idx int) {
				traverseStack = append(traverseStack[:idx], traverseStack[idx+1:]...)
			},
		},
	})
	return newAst
}

func codeGenerator(n node) string {
	switch n.Kind {
	case Program:
		var r []string
		for _, no := range n.Body {
			r = append(r, codeGenerator(no))
		}
		return strings.Join(r, "\n")
	case ExpressionStatement:
		return codeGenerator(*n.Expression) + ";"
	case CallExpression:
		var ra []string
		c := codeGenerator(*n.Callee)

		for _, no := range *n.Arguments {
			ra = append(ra, codeGenerator(no))
		}

		r := strings.Join(ra, ", ")
		return c + "(" + r + ")"
	case Identifier:
		return n.Name
	case NumberLiteral:
		return n.Value
	default:
		log.Fatal("err")
		return ""
	}
}

func compiler(input string) string {
	tokens := tokenizer(input)
	ast := parser(tokens)
	newAst := transformer(ast)
	out := codeGenerator(node(newAst))
	// fmt.Printf("%#v\n", newAst)
	return out
}

func main() {
	program := "(add 10 (subtract 10 6))"
	out := compiler(program)
	fmt.Println(out)
}

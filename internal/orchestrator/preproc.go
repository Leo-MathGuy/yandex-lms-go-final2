package orchestrator

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// Constants
const NUMBER, OP, SPACE, PAR int = 0, 1, 2, 3
const OPS string = "+-*/"
const PARS string = "()"

// Misc
func get_op(x float64) string {
	return string(OPS[int(x)])
}
func get_par(x float64) string {
	return string(PARS[int(x)])
}

func is_op(x float64) bool {
	return int(x) == OP
}
func is_par(x float64) bool {
	return int(x) == PAR
}

// Utilities

func Strip(expr string) string {
	for expr[0] == ' ' {
		expr = expr[1:]
	}
	for expr[len(expr)-1] == ' ' {
		expr = expr[:len(expr)-1]
	}
	return expr
}

func SymbolType(x rune) int {
	if unicode.IsDigit(x) || x == '.' {
		return NUMBER
	}

	switch x {
	case ' ':
		return SPACE
	case '+', '-', '/', '*':
		return OP
	case '(', ')':
		return PAR
	}

	return -1
}

// // Phase 1: Validation
func Step1(expr string) ([]string, error) {
	if len(expr) == 0 {
		return []string{}, errors.New("empty expression")
	}

	// Separate symbols
	rexpr := make([]string, 0)
	expr = Strip(expr)

	for _, v := range expr {
		t := SymbolType(v)

		if t == -1 {
			return []string{}, errors.New("invalid character " + string(v))
		}

		rexpr = append(rexpr, string(v))
	}

	// Validate operators at ends
	l := len(rexpr)

	if rexpr[0] != "-" && SymbolType(rune(rexpr[0][0])) == OP {
		return []string{}, errors.New("Operator at start: " + rexpr[0])
	}

	if SymbolType(rune(rexpr[l-1][0])) == OP {
		return []string{}, errors.New("Operator at end: " + rexpr[l-1])
	}

	// Check spaces (spaces are guaranteed not to be at the ends now)
	cleared := make([]string, 0)

	for i := 0; i < len(rexpr)-1; i++ {
		if rexpr[i] == " " && rexpr[i+1] == " " {
			rexpr = append(rexpr[:i], rexpr[i+1:]...)
			i--
		}
	}

	for i, v := range rexpr {
		if v == " " {

			before := SymbolType(rune(rexpr[i-1][0]))
			after := SymbolType(rune(rexpr[i+1][0]))

			if before == after && (before != PAR) && rexpr[i+1] != "-" {
				return []string{}, errors.New("Invalid space: \"" + rexpr[i-1] + " " + rexpr[i+1] + "\"")
			}
		} else {
			cleared = append(cleared, v)
		}
	}
	rexpr = cleared

	// Check parentheses
	count := 0

	for _, v := range rexpr {
		if v == "(" {
			count++
		}

		if v == ")" {
			count--

			if count < 0 {
				return []string{}, errors.New("invalid parentheses")
			}
		}
	}

	if count != 0 {
		return []string{}, errors.New("invalid parentheses")
	}

	// Join numbers
	l = len(rexpr)

	cleared = make([]string, 0)

	skip := 0

	for i, v := range rexpr {
		if skip > 0 {
			skip--
			continue
		}

		minus_is_num := i == 0 || SymbolType(rune(rexpr[i-1][0])) == OP || rune(rexpr[i-1][0]) == '('

		if (SymbolType(rune(v[0])) != 0 && !(minus_is_num && v == "-")) || i == l-1 {
			cleared = append(cleared, v)
			continue
		}

		j := 1
		cleared = append(cleared, v)

		for SymbolType(rune(rexpr[i+j][0])) == 0 {
			cleared[len(cleared)-1] += rexpr[i+j]
			skip++
			j++

			if i+j >= l {
				break
			}
		}
	}

	rexpr = cleared

	// Decimal check
	for index, v := range rexpr {
		if v == "." {
			return []string{}, errors.New("invalid number: \".\"")
		}

		if strings.Count(v, ".") > 1 {
			return []string{}, errors.New("invalid number: " + v)
		}

		i := strings.Index(v, ".")
		if i == -1 {
			continue
		}

		if i == len(v)-1 {
			return []string{}, errors.New("invalid number: " + v)
		}

		if i == 0 {
			rexpr[index] = "0" + v
		}
	}

	return rexpr, nil
}

// // Phase 2: Tokenization
func Step2(step1 []string) ([][]float64, error) {
	fexpr := make([][]float64, 0)

	for _, v := range step1 {
		if len(v) > 1 || SymbolType(rune(v[0])) == NUMBER {
			num, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return [][]float64{}, errors.New("Error during number parse: " + v + " - " + err.Error())
			}
			fexpr = append(fexpr, []float64{float64(NUMBER), num})
		} else if SymbolType(rune(v[0])) == OP {
			fexpr = append(fexpr, []float64{float64(OP), float64(strings.Index(OPS, v))})
		} else if SymbolType(rune(v[0])) == PAR {
			fexpr = append(fexpr, []float64{float64(PAR), float64(strings.Index(PARS, v))})
		}
	}

	return fexpr, nil
}

// // Phase 3: AST Generation
type ASTNode struct {
	Operator string
	Value    *float64
	Left     *ASTNode
	Right    *ASTNode
}

func parseExpression(expr [][]float64) *ASTNode {
	real := make([][]float64, 0)
	skip := false
	for i, v := range expr {
		if skip {
			skip = false
			continue
		}

		if is_op(v[0]) && get_op(v[1]) == "-" {
			if i == 0 || is_op(expr[i-1][0]) && expr[i+1][0] == float64(NUMBER) {
				real = append(real, []float64{expr[i+1][0], -expr[i+1][1]})
				skip = true
				continue
			}
		}
		real = append(real, v)
	}

	expr = real

	// Separate terms
	terms := make([][][]float64, 0)
	opps := make([]float64, 0)

	term := make([][]float64, 0)
	parens := 0
	for _, v := range expr {
		if is_par(v[0]) {
			if get_par(v[1]) == "(" {
				parens++
			} else {
				parens--
			}
		}

		if parens == 0 && is_op(v[0]) && v[1] <= 1 {
			terms = append(terms, term)
			term = [][]float64{}
			opps = append(opps, v[1])
		} else {
			term = append(term, v)
		}
	}
	terms = append(terms, term)

	if len(opps) == 0 {
		return parseTerm(terms[0])
	}

	root := &ASTNode{get_op(opps[0]), nil, parseTerm(terms[0]), parseTerm(terms[1])}
	t := 2
	for len(opps) > 1 {
		oldRoot := root
		root = &ASTNode{get_op(opps[1]), nil, oldRoot, parseTerm(terms[t])}
		t++
		opps = opps[1:]
	}

	return root
}

func parseTerm(term [][]float64) *ASTNode {
	// Separate factors
	factors := make([][][]float64, 0)
	opps := make([]float64, 0)

	factor := make([][]float64, 0)
	parens := 0
	for _, v := range term {
		if is_par(v[0]) {
			if get_par(v[1]) == "(" {
				parens++
			} else {
				parens--
			}
		}

		if parens == 0 && is_op(v[0]) && v[1] >= 2 {
			factors = append(factors, factor)
			factor = [][]float64{}
			opps = append(opps, v[1])
		} else {
			factor = append(factor, v)
		}
	}
	factors = append(factors, factor)

	if len(opps) == 0 {
		return parseFactor(factors[0])
	}

	root := &ASTNode{get_op(opps[0]), nil, parseFactor(factors[0]), parseFactor(factors[1])}
	t := 2
	for len(opps) > 1 {
		oldRoot := root
		root = &ASTNode{get_op(opps[1]), nil, oldRoot, parseFactor(factors[t])}
		t++
		opps = opps[1:]
	}

	return root
}

func parseFactor(factor [][]float64) *ASTNode {
	if len(factor) == 1 {
		if factor[0][0] != float64(NUMBER) {
			panic("UHOH")
		}

		return &ASTNode{"", floatPtr(factor[0][1]), nil, nil}
	}

	if is_par(factor[0][0]) && is_par(factor[len(factor)-1][0]) && factor[0][1]+factor[len(factor)-1][1] == 1 {
		factor = factor[1 : len(factor)-1]
		if len(factor) == 0 {
			panic("no")
		}
		return parseExpression(factor)
	}

	panic("Oh nou")
}

func floatPtr(f float64) *float64 {
	return &f
}

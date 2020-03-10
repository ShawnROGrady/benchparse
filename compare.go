package benchparse

import (
	"errors"
	"fmt"
	"strings"
)

// Comparison represents a comparison operation.
type Comparison string

// The available comparison operations.
const (
	Eq Comparison = "=="
	Ne Comparison = "!="
	Lt Comparison = "<"
	Gt Comparison = ">"
	Le Comparison = "<="
	Ge Comparison = ">="
)

func (c Comparison) description() string {
	switch c {
	case Eq:
		return "eq"
	case Ne:
		return "ne"
	case Lt:
		return "lt"
	case Gt:
		return "gt"
	case Le:
		return "le"
	case Ge:
		return "ge"
	default:
		return ""
	}
}

// Possible comparison errors.
var (
	errOperationNotDefined = errors.New("operation not defined for values")
	errNonComparable       = errors.New("values cannot be compared")
	errDifferentNames      = errors.New("variables have different names")
	errInvalidOperation    = errors.New("invalid comparison operation")
	errMalformedFilter     = errors.New("filter expression not of form 'var_name==var_value'")
)

type compareErr struct {
	val1       BenchVarValue
	val2       BenchVarValue
	comparison Comparison
	err        error
}

func (c compareErr) Error() string {
	return fmt.Sprintf("cannot evaluate (%s)%s(%s): %s", c.val1, c.comparison, c.val2, c.err)
}

func (c compareErr) Unwrap() error {
	return c.err
}

func (c Comparison) compare(v1, v2 BenchVarValue) (bool, error) {
	switch c {
	case Eq:
		eq, err := v1.equal(v2)
		if err != nil {
			return false, compareErr{val1: v1, val2: v2, comparison: c, err: err}
		}
		return eq, nil
	case Ne:
		eq, err := v1.equal(v2)
		if err != nil {
			return false, compareErr{val1: v1, val2: v2, comparison: c, err: err}
		}
		return !eq, nil
	case Lt:
		less, err := v1.less(v2)
		if err != nil {
			return false, compareErr{val1: v1, val2: v2, comparison: c, err: err}
		}
		return less, nil
	case Gt:
		eq, err := v1.equal(v2)
		if err != nil {
			return false, compareErr{val1: v1, val2: v2, comparison: c, err: err}
		}
		less, err := v1.less(v2)
		if err != nil {
			return false, compareErr{val1: v1, val2: v2, comparison: c, err: err}
		}
		return !(eq || less), nil
	case Le:
		eq, err := v1.equal(v2)
		if err != nil {
			return false, compareErr{val1: v1, val2: v2, comparison: c, err: err}
		}
		less, err := v1.less(v2)
		if err != nil {
			return false, compareErr{val1: v1, val2: v2, comparison: c, err: err}
		}
		return eq || less, nil
	case Ge:
		less, err := v1.less(v2)
		if err != nil {
			return false, compareErr{val1: v1, val2: v2, comparison: c, err: err}
		}
		return !less, nil
	default:
		return false, compareErr{val1: v1, val2: v2, comparison: c, err: errInvalidOperation}
	}
}

type varValComp struct {
	varValue BenchVarValue
	cmp      Comparison
}

func (v varValComp) String() string {
	return fmt.Sprintf("%s%s%v", v.varValue.Name, v.cmp, v.varValue.Value)
}

func parseValueComparison(in string) (varValComp, error) {
	cmps := []Comparison{
		Eq,
		Ne,
		Le,
		Ge,
		Lt,
		Gt,
	}
	for _, cmp := range cmps {
		split := strings.Split(in, string(cmp))
		if len(split) != 2 {
			continue
		}
		return varValComp{
			varValue: BenchVarValue{
				Name:  split[0],
				Value: value(split[1]),
			},
			cmp: cmp,
		}, nil
	}

	return varValComp{}, errMalformedFilter
}

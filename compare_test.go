package benchparse

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type compareResult struct {
	res bool
	err error
}

var compareTests = map[string]struct {
	v1       BenchVarValue
	v2       BenchVarValue
	expectEq compareResult
	expectNe compareResult
	expectLt compareResult
	expectGt compareResult
	expectLe compareResult
	expectGe compareResult
}{
	"same_name_equal_int_values": {
		v1:       BenchVarValue{Name: "var1", Value: 12},
		v2:       BenchVarValue{Name: "var1", Value: 12},
		expectEq: compareResult{res: true},
		expectNe: compareResult{res: false},
		expectLt: compareResult{res: false},
		expectGt: compareResult{res: false},
		expectLe: compareResult{res: true},
		expectGe: compareResult{res: true},
	},
	"same_name_string_values_v1_greater_than_v2": {
		v1:       BenchVarValue{Name: "var1", Value: "case2"},
		v2:       BenchVarValue{Name: "var1", Value: "case1"},
		expectEq: compareResult{res: false},
		expectNe: compareResult{res: true},
		expectLt: compareResult{res: false},
		expectGt: compareResult{res: true},
		expectLe: compareResult{res: false},
		expectGe: compareResult{res: true},
	},
	"same_name_float_values_v1_less_than_v2": {
		v1:       BenchVarValue{Name: "var1", Value: 1.2},
		v2:       BenchVarValue{Name: "var1", Value: 1.3},
		expectEq: compareResult{res: false},
		expectNe: compareResult{res: true},
		expectLt: compareResult{res: true},
		expectGt: compareResult{res: false},
		expectLe: compareResult{res: true},
		expectGe: compareResult{res: false},
	},
	"same_name_uint_values_v1_greater_than_v2": {
		v1:       BenchVarValue{Name: "var1", Value: uint(3)},
		v2:       BenchVarValue{Name: "var1", Value: uint(2)},
		expectEq: compareResult{res: false},
		expectNe: compareResult{res: true},
		expectLt: compareResult{res: false},
		expectGt: compareResult{res: true},
		expectLe: compareResult{res: false},
		expectGe: compareResult{res: true},
	},
	"same_name_float_v1_less_than_int_v2": {
		v1:       BenchVarValue{Name: "var1", Value: 0.9},
		v2:       BenchVarValue{Name: "var1", Value: int(1)},
		expectEq: compareResult{res: false},
		expectNe: compareResult{res: true},
		expectLt: compareResult{res: true},
		expectGt: compareResult{res: false},
		expectLe: compareResult{res: true},
		expectGe: compareResult{res: false},
	},
	"same_name_unequal_bool_values": {
		v1:       BenchVarValue{Name: "var1", Value: true},
		v2:       BenchVarValue{Name: "var1", Value: false},
		expectEq: compareResult{res: false},
		expectNe: compareResult{res: true},
		expectLt: compareResult{err: errOperationNotDefined},
		expectGt: compareResult{err: errOperationNotDefined},
		expectLe: compareResult{err: errOperationNotDefined},
		expectGe: compareResult{err: errOperationNotDefined},
	},
	"different_name_equal_int_values": {
		v1:       BenchVarValue{Name: "var1", Value: 12},
		v2:       BenchVarValue{Name: "var2", Value: 12},
		expectEq: compareResult{err: errDifferentNames},
		expectNe: compareResult{err: errDifferentNames},
		expectLt: compareResult{err: errDifferentNames},
		expectGt: compareResult{err: errDifferentNames},
		expectLe: compareResult{err: errDifferentNames},
		expectGe: compareResult{err: errDifferentNames},
	},
	"same_name_string_and_int_values": {
		v1:       BenchVarValue{Name: "var1", Value: "case1"},
		v2:       BenchVarValue{Name: "var1", Value: 3},
		expectEq: compareResult{err: errNonComparable},
		expectNe: compareResult{err: errNonComparable},
		expectLt: compareResult{err: errNonComparable},
		expectGt: compareResult{err: errNonComparable},
		expectLe: compareResult{err: errNonComparable},
		expectGe: compareResult{err: errNonComparable},
	},
}

func TestCompare(t *testing.T) {
	for testName, testCase := range compareTests {
		t.Run(testName, func(t *testing.T) {
			t.Run(string(Eq), func(t *testing.T) {
				testEq(t, testCase.v1, testCase.v2, testCase.expectEq)
			})
			t.Run(string(Ne), func(t *testing.T) {
				testNe(t, testCase.v1, testCase.v2, testCase.expectNe)
			})
			t.Run(string(Lt), func(t *testing.T) {
				testLt(t, testCase.v1, testCase.v2, testCase.expectLt)
			})
			t.Run(string(Gt), func(t *testing.T) {
				testGt(t, testCase.v1, testCase.v2, testCase.expectGt)
			})
			t.Run(string(Le), func(t *testing.T) {
				testLe(t, testCase.v1, testCase.v2, testCase.expectLe)
			})
			t.Run(string(Ge), func(t *testing.T) {
				testGe(t, testCase.v1, testCase.v2, testCase.expectGe)
			})
		})
	}
}

func testEq(t *testing.T, v1, v2 BenchVarValue, expectEq compareResult) {
	t.Helper()
	eq, err := Eq.compare(v1, v2)
	if err != nil {
		if !errors.Is(err, expectEq.err) {
			t.Errorf("unexpected error\nexpected=%s\nactual=%s", expectEq.err, err)
		}
		return
	}
	if eq != expectEq.res {
		t.Errorf("unexpected %s==%s\nexpected:%t\nactual:%t", v1, v2, expectEq, eq)
	}
}

func testNe(t *testing.T, v1, v2 BenchVarValue, expectNe compareResult) {
	t.Helper()
	ne, err := Ne.compare(v1, v2)
	if err != nil {
		if !errors.Is(err, expectNe.err) {
			t.Errorf("unexpected error\nexpected=%s\nactual=%s", expectNe.err, err)
		}
		return
	}
	if ne != expectNe.res {
		t.Errorf("unexpected %s!=%s\nexpected:%t\nactual:%t", v1, v2, expectNe, ne)
	}
}

func testLt(t *testing.T, v1, v2 BenchVarValue, expectLt compareResult) {
	t.Helper()
	lt, err := Lt.compare(v1, v2)
	if err != nil {
		if !errors.Is(err, expectLt.err) {
			t.Errorf("unexpected error\nexpected=%s\nactual=%s", expectLt.err, err)
		}
		return
	}
	if lt != expectLt.res {
		t.Errorf("unexpected %s<%s\nexpected:%t\nactual:%t", v1, v2, expectLt, lt)
	}
}

func testGt(t *testing.T, v1, v2 BenchVarValue, expectGt compareResult) {
	t.Helper()
	gt, err := Gt.compare(v1, v2)
	if err != nil {
		if !errors.Is(err, expectGt.err) {
			t.Errorf("unexpected error\nexpected=%s\nactual=%s", expectGt.err, err)
		}
		return
	}
	if gt != expectGt.res {
		t.Errorf("unexpected %s>%s\nexpected:%t\nactual:%t", v1, v2, expectGt, gt)
	}
}

func testLe(t *testing.T, v1, v2 BenchVarValue, expectLe compareResult) {
	t.Helper()
	le, err := Le.compare(v1, v2)
	if err != nil {
		if !errors.Is(err, expectLe.err) {
			t.Errorf("unexpected error\nexpected=%s\nactual=%s", expectLe.err, err)
		}
		return
	}
	if le != expectLe.res {
		t.Errorf("unexpected %s<=%s\nexpected:%t\nactual:%t", v1, v2, expectLe, le)
	}
}

func testGe(t *testing.T, v1, v2 BenchVarValue, expectGe compareResult) {
	t.Helper()
	ge, err := Ge.compare(v1, v2)
	if err != nil {
		if !errors.Is(err, expectGe.err) {
			t.Errorf("unexpected error\nexpected=%s\nactual=%s", expectGe.err, err)
		}
		return
	}
	if ge != expectGe.res {
		t.Errorf("unexpected %s>=%s\nexpected:%t\nactual:%t", v1, v2, expectGe, ge)
	}
}

var compareBenches = map[string]struct {
	v1 BenchVarValue
	v2 BenchVarValue
}{
	"equal_strings": {
		v1: BenchVarValue{Name: "var", Value: "foo"},
		v2: BenchVarValue{Name: "var", Value: "foo"},
	},
	"unequal_strings": {
		v1: BenchVarValue{Name: "var", Value: "foo"},
		v2: BenchVarValue{Name: "var", Value: "bar"},
	},
	"equal_bools": {
		v1: BenchVarValue{Name: "var", Value: true},
		v2: BenchVarValue{Name: "var", Value: true},
	},
	"unequal_bools": {
		v1: BenchVarValue{Name: "var", Value: true},
		v2: BenchVarValue{Name: "var", Value: false},
	},
	"equal_numbers": {
		v1: BenchVarValue{Name: "var", Value: uint(2)},
		v2: BenchVarValue{Name: "var", Value: int(2)},
	},
	"unequal_numbers": {
		v1: BenchVarValue{Name: "var", Value: uint(2)},
		v2: BenchVarValue{Name: "var", Value: float32(1.3)},
	},
}

var benchCompareRes error

func BenchmarkCompare(b *testing.B) {
	allComps := []Comparison{Eq, Ne, Lt, Gt, Le, Ge}
	for benchName, benchCase := range compareBenches {
		b.Run(fmt.Sprintf("case=%s", benchName), func(b *testing.B) {
			for _, comp := range allComps {
				b.Run(fmt.Sprintf("cmp=%s", comp.description()), func(b *testing.B) {
					var err error
					for i := 0; i < b.N; i++ {
						_, err = comp.compare(benchCase.v1, benchCase.v2)
					}
					benchCompareRes = err
				})
			}
		})
	}
}

func TestCompareInvalidComparison(t *testing.T) {
	var (
		v1   = BenchVarValue{Name: "var1", Value: 12}
		v2   = BenchVarValue{Name: "var1", Value: 12}
		comp = Comparison("_")
	)

	_, err := comp.compare(v1, v2)
	if err == nil {
		t.Fatalf("unexpectedly no error")
	}

	if !errors.Is(err, errInvalidOperation) {
		t.Errorf("unexpected error\nexpected=%s\nactual=%s", errInvalidOperation, err)
	}

	expectedErrString := "cannot evaluate (var1=12)_(var1=12): invalid comparison operation"
	if err.Error() != expectedErrString {
		t.Errorf("unexpected error string\nexpected=%s\nactual=%s", expectedErrString, err.Error())
	}
}

var parseValueComparisonTests = map[string]struct {
	expectedVarValCmp varValComp
	expectedString    string
	expectErr         bool
}{
	"var_1==2": {
		expectedVarValCmp: varValComp{
			varValue: BenchVarValue{Name: "var_1", Value: 2},
			cmp:      Eq,
		},
		expectedString: "var_1==2",
	},
	"var_1!=foo": {
		expectedVarValCmp: varValComp{
			varValue: BenchVarValue{Name: "var_1", Value: "foo"},
			cmp:      Ne,
		},
		expectedString: "var_1!=foo",
	},
	"var_1>2.2": {
		expectedVarValCmp: varValComp{
			varValue: BenchVarValue{Name: "var_1", Value: 2.2},
			cmp:      Gt,
		},
		expectedString: "var_1>2.2",
	},
	"var_1<1": {
		expectedVarValCmp: varValComp{
			varValue: BenchVarValue{Name: "var_1", Value: 1},
			cmp:      Lt,
		},
		expectedString: "var_1<1",
	},
	"var_1>=2.2": {
		expectedVarValCmp: varValComp{
			varValue: BenchVarValue{Name: "var_1", Value: 2.2},
			cmp:      Ge,
		},
		expectedString: "var_1>=2.2",
	},
	"var_1<=1": {
		expectedVarValCmp: varValComp{
			varValue: BenchVarValue{Name: "var_1", Value: 1},
			cmp:      Le,
		},
		expectedString: "var_1<=1",
	},
	"var1,2": {
		expectErr: true,
	},
}

func TestParseValueComparison(t *testing.T) {
	for testInput, testCase := range parseValueComparisonTests {
		t.Run(testInput, func(t *testing.T) {
			varValCmp, err := parseValueComparison(testInput)
			if err != nil {
				if !testCase.expectErr {
					t.Errorf("unexpected error: %s", err)
				}
				return
			}

			if !reflect.DeepEqual(varValCmp, testCase.expectedVarValCmp) {
				t.Errorf("unexpected parsed\nexpected:%v\nactual:%v", testCase.expectedVarValCmp, varValCmp)
			}

			if testCase.expectedString != varValCmp.String() {
				t.Errorf("unexpected parsed string\nexpected:%s\nactual:%s", testCase.expectedString, varValCmp.String())
			}
		})
	}
}

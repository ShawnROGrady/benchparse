package benchparse

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/benchmark/parse"
)

// BenchVarValue represents an input to the benchmark represented by
// a sub-benchmark with a name of the form 'var_name=var_value'.
type BenchVarValue struct {
	Name     string
	Value    interface{}
	position int
}

func (b BenchVarValue) equal(o BenchVarValue) (bool, error) {
	if b.Name != o.Name {
		return false, errDifferentNames
	}

	v1, v2 := reflect.ValueOf(b.Value), reflect.ValueOf(o.Value)
	k1, k2 := v1.Type().Kind(), v2.Type().Kind()

	if isNumeric(k1) && isNumeric(k2) {
		f1, err := getFloat(v1, k1)
		if err != nil {
			return false, err
		}
		f2, err := getFloat(v2, k2)
		if err != nil {
			return false, err
		}
		return f1 == f2, nil
	}
	if k1 != k2 {
		return false, errNonComparable
	}

	switch k1 {
	case reflect.String:
		return v1.String() == v2.String(), nil
	default:
		return b.Value == o.Value, nil
	}
}

func (b BenchVarValue) less(o BenchVarValue) (bool, error) {
	if b.Name != o.Name {
		return false, errDifferentNames
	}

	v1, v2 := reflect.ValueOf(b.Value), reflect.ValueOf(o.Value)
	k1, k2 := v1.Type().Kind(), v2.Type().Kind()

	if isNumeric(k1) && isNumeric(k2) {
		f1, err := getFloat(v1, k1)
		if err != nil {
			return false, err
		}
		f2, err := getFloat(v2, k2)
		if err != nil {
			return false, err
		}
		return f1 < f2, nil
	}
	if k1 != k2 {
		return false, errNonComparable
	}

	switch k1 {
	case reflect.String:
		return v1.String() < v2.String(), nil
	default:
		return false, errOperationNotDefined
	}
}

func isNumeric(k reflect.Kind) bool {
	numericKinds := [...]reflect.Kind{
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float64, reflect.Float32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
	}
	for _, kind := range numericKinds {
		if k == kind {
			return true
		}
	}
	return false
}

func getFloat(v reflect.Value, k reflect.Kind) (float64, error) {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Float64, reflect.Float32:
		return v.Float(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	default:
		return 0, fmt.Errorf("non-numeric type: %s", k)
	}
}

// String returns the string representation of the BenchVarValue
// with the form 'var_name=var_value'.
//
// The string representation of a BenchVarValue may vary slightly
// from the original input due to things like floating point
// precision and alternate string representations of various
// types.
//
// Currently the '%f' verb is used for floating point values
// in order to guarantee that they can be distinguished from
// integer values. For everything else the default '%v' verb
// is used for simplicities sake.
func (b BenchVarValue) String() string {
	if f, ok := b.Value.(float64); ok {
		return fmt.Sprintf("%s=%f", b.Name, f)
	}
	return fmt.Sprintf("%s=%v", b.Name, b.Value)
}

func (b BenchVarValue) pos() int {
	return b.position
}

type benchVarValues []BenchVarValue

func (b benchVarValues) String() string {
	s := make([]string, len(b))
	for i, val := range b {
		s[i] = val.String()
	}
	return strings.Join(s, ",")
}

// BenchSub represents an input to the benchmark represented
// by a sub-benchmark with a name NOT of the form 'var_name=var_value'.
type BenchSub struct {
	Name     string
	position int
}

func (b BenchSub) String() string {
	return b.Name
}

func (b BenchSub) pos() int {
	return b.position
}

type benchInput interface {
	pos() int
	fmt.Stringer
}

// BenchInputs define a sub-benchmark. For example a benchmark with
// a full name 'BenchmarkMyType/some_method/foo=2/bar=baz-4' would be
// defined by the Subs=[some_method], the VarValues=[foo=2 bar=baz],
// and MaxProcs=4.
type BenchInputs struct {
	VarValues []BenchVarValue // sub-benchmark names of the form some_var=some_val
	Subs      []BenchSub      // remaining components of a sub-benchmark
	MaxProcs  int             // the value of GOMAXPROCS when the benchmark was run
}

// String returns the string representation of the BenchInputs.
// This should be equivalent to the portion of the benchmark name
// following the name of the top-level benchmark, but formatting
// of VarValues may vary slightly.
func (b BenchInputs) String() string {
	var (
		inputs = make([]benchInput, len(b.VarValues)+len(b.Subs))
		s      strings.Builder
	)

	for i, varVal := range b.VarValues {
		inputs[i] = varVal
	}
	for i, sub := range b.Subs {
		inputs[i+len(b.VarValues)] = sub
	}
	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].pos() < inputs[j].pos()
	})

	for _, input := range inputs {
		s.WriteString("/")
		s.WriteString(input.String())
	}

	if b.MaxProcs > 1 {
		s.WriteString("-")
		s.WriteString(strconv.Itoa(b.MaxProcs))
	}
	return s.String()
}

// ErrNotMeasured indicates that a specific output
// was not measured.
var ErrNotMeasured = errors.New("not measured")

// BenchOutputs are the outputs of a single benchmark run.
//
// Since not all output values are measured on each benchmark
// run, the getter for these values will return ErrNotMeasured
// if this is the case.
type BenchOutputs interface {
	GetIterations() int
	GetNsPerOp() (float64, error)
	GetAllocedBytesPerOp() (uint64, error) // measured if either '-test.benchmem' is set of if testing.B.ReportAllocs() is called
	GetAllocsPerOp() (uint64, error)       // measured if either '-test.benchmem' is set of if testing.B.ReportAllocs() is called
	GetMBPerS() (float64, error)           // measured if testing.B.SetBytes() is called
}

func benchOutputsString(b BenchOutputs) string {
	var s strings.Builder
	s.WriteString(strconv.Itoa(b.GetIterations()))
	if nsPerOp, err := b.GetNsPerOp(); err == nil {
		fmt.Fprintf(&s, " %.2f ns/op", nsPerOp)
	}
	if mbPerS, err := b.GetMBPerS(); err == nil {
		fmt.Fprintf(&s, " %.2f MB/s", mbPerS)
	}
	if bPerOp, err := b.GetAllocedBytesPerOp(); err == nil {
		fmt.Fprintf(&s, " %d B/op", bPerOp)
	}
	if allocsPerOp, err := b.GetAllocsPerOp(); err == nil {
		fmt.Fprintf(&s, " %d allocs/op", allocsPerOp)
	}
	return s.String()
}

// parsedBenchOutputs wraps the parse.Benchmark type to
// implement the BenchOutputs interface.
type parsedBenchOutputs struct {
	parse.Benchmark
}

func (b parsedBenchOutputs) GetIterations() int {
	return b.N
}

// GetNsPerOp returns the nanoseconds per iteration.
// If not measured ErrNotMeasured is returned.
func (b parsedBenchOutputs) GetNsPerOp() (float64, error) {
	if (b.Measured & parse.NsPerOp) != 0 {
		return b.NsPerOp, nil
	}
	return 0, ErrNotMeasured
}

// GetAllocedBytesPerOp returns the bytes allocated per iteration.
// This is measured if either '-test.benchmem' is set when running
// the benchmark or if testing.B.ReportAllocs() is called.
//
// If not measured ErrNotMeasured is returned.
func (b parsedBenchOutputs) GetAllocedBytesPerOp() (uint64, error) {
	if (b.Measured & parse.AllocedBytesPerOp) != 0 {
		return b.AllocedBytesPerOp, nil
	}
	return 0, ErrNotMeasured
}

// GetAllocsPerOp returns the allocs per iteration.
// This is measured if either '-test.benchmem' is set when running
// the benchmark or if testing.B.ReportAllocs() is called.
//
// If not measured ErrNotMeasured is returned.
func (b parsedBenchOutputs) GetAllocsPerOp() (uint64, error) {
	if (b.Measured & parse.AllocsPerOp) != 0 {
		return b.AllocsPerOp, nil
	}
	return 0, ErrNotMeasured
}

// GetMBPerS returns the MB processed per second.
// This is measured if testing.B.SetBytes() is
// called.
//
// If not measured ErrNotMeasured is returned.
func (b parsedBenchOutputs) GetMBPerS() (float64, error) {
	if (b.Measured & parse.MBPerS) != 0 {
		return b.MBPerS, nil
	}
	return 0, ErrNotMeasured
}

// BenchRes represents a result from a single benchmark run.
// This corresponds to one line from the testing.B output.
type BenchRes struct {
	Inputs  BenchInputs  // the input variables
	Outputs BenchOutputs // the output result
}

// BenchResults represents a list of benchmark results
type BenchResults []BenchRes

// Filter returns a subset of the BenchResults matching
// the provided filter expr. For example filtering by the
// expression 'var1<=2' will return the results where the
// input variable named 'var1' has a value less than or
// equal to 2.
func (b BenchResults) Filter(filterExpr string) (BenchResults, error) {
	varValCmp, err := parseValueComparison(filterExpr)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", filterExpr, err)
	}

	var (
		filtered = []BenchRes{}
		cmp      = varValCmp.cmp
		value    = varValCmp.varValue
	)

	for _, res := range b {
		for _, varVal := range res.Inputs.VarValues {
			include, err := cmp.compare(varVal, value)
			if err != nil {
				if !errors.Is(err, errDifferentNames) {
					return nil, err
				}
				continue
			}
			if include {
				filtered = append(filtered, res)
				break
			}
		}
	}
	return filtered, nil
}

// Group groups a benchmarks results by a specified set of
// input variable names. For example a Benchmark with Results corresponding
// to the cases [/foo=1/bar=baz /foo=2/bar=baz /foo=1/bar=qux /foo=2/bar=qux]
// grouped by ['foo'] would have 2 groups of results (those with Inputs where
func (b BenchResults) Group(groupBy []string) GroupedResults {
	groupedResults := map[string]BenchResults{}
	if len(groupBy) == 0 {
		res := make([]BenchRes, len(b))
		copy(res, b)
		groupedResults[""] = res
		return groupedResults
	}
	for _, result := range b {
		groupVals := benchVarValues{}
		for _, varValue := range result.Inputs.VarValues {
			for _, groupName := range groupBy {
				if varValue.Name == groupName {
					groupVals = append(groupVals, varValue)
				}
			}
		}
		if len(groupVals) != len(groupBy) {
			continue
		}

		k := groupVals.String()
		if existingResults, ok := groupedResults[k]; ok {
			groupedResults[k] = append(existingResults, result)
		} else {
			groupedResults[k] = []BenchRes{result}
		}
	}
	return groupedResults
}

// GroupedResults represents a grouping of benchmark results.
type GroupedResults map[string]BenchResults

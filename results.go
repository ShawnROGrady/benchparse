package benchparse

import (
	"fmt"
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

// String returns the string representation of the BenchVarValue
// with the form 'var_name=var_value'.
// Currently the default string format ('%v') is used for the actual
// value, meaning the string representation of a BenchVarValue may
// vary slightly from the original input (for example precision of
// floating point values).
func (b BenchVarValue) String() string {
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

// BenchOutputs are the outputs of a single benchmark run.
type BenchOutputs struct {
	N                   int     // number of iterations
	nsPerOp             float64 // nanoseconds per iteration
	allocatedBytesPerOp uint64  // bytes allocated per iteration
	allocsPerOp         uint64  // allocs per iteration
	mBPerS              float64 // MB processed per second
	measured            int     // which measurements were recorded
}

// NsPerOp returns the nanoseconds per iteration.
// If not measured ErrNotMeasured is returned.
func (b BenchOutputs) NsPerOp() (float64, error) {
	if (b.measured & parse.NsPerOp) != 0 {
		return b.nsPerOp, nil
	}
	return 0, ErrNotMeasured
}

// AllocedBytesPerOp returns the bytes allocated per iteration.
// This is measured if either '-test.benchmem' is set when running
// the benchmark or if testing.B.ReportAllocs() is called.
//
// If not measured ErrNotMeasured is returned.
func (b BenchOutputs) AllocedBytesPerOp() (uint64, error) {
	if (b.measured & parse.AllocedBytesPerOp) != 0 {
		return b.allocatedBytesPerOp, nil
	}
	return 0, ErrNotMeasured
}

// AllocsPerOp returns the allocs per iteration.
// This is measured if either '-test.benchmem' is set when running
// the benchmark or if testing.B.ReportAllocs() is called.
//
// If not measured ErrNotMeasured is returned.
func (b BenchOutputs) AllocsPerOp() (uint64, error) {
	if (b.measured & parse.AllocsPerOp) != 0 {
		return b.allocsPerOp, nil
	}
	return 0, ErrNotMeasured
}

// MBPerS returns the MB processed per second.
// This is measured if testing.B.SetBytes() is
// called.
//
// If not measured ErrNotMeasured is returned.
func (b BenchOutputs) MBPerS() (float64, error) {
	if (b.measured & parse.MBPerS) != 0 {
		return b.mBPerS, nil
	}
	return 0, ErrNotMeasured
}

func (b BenchOutputs) String() string {
	// just the relevant parts of https://godoc.org/golang.org/x/tools/benchmark/parse#Benchmark.String
	var s strings.Builder
	s.WriteString(strconv.Itoa(b.N))
	if (b.measured & parse.NsPerOp) != 0 {
		fmt.Fprintf(&s, " %.2f ns/op", b.nsPerOp)
	}
	if (b.measured & parse.MBPerS) != 0 {
		fmt.Fprintf(&s, " %.2f MB/s", b.mBPerS)
	}
	if (b.measured & parse.AllocedBytesPerOp) != 0 {
		fmt.Fprintf(&s, " %d B/op", b.allocatedBytesPerOp)
	}
	if (b.measured & parse.AllocsPerOp) != 0 {
		fmt.Fprintf(&s, " %d allocs/op", b.allocsPerOp)
	}
	return s.String()
}

// BenchRes represents a result from a single benchmark run.
// This corresponds to one line from the testing.B output.
type BenchRes struct {
	Inputs  BenchInputs  // the input variables
	Outputs BenchOutputs // the output result
}

// GroupedResults represents a grouping of benchmark results.
type GroupedResults map[string][]BenchRes

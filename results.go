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

// BenchInputs define a sub-benchmark.
type BenchInputs struct {
	VarValues []BenchVarValue // sub-benchmark names of the form some_var=some_val
	Subs      []BenchSub      // remaining components of a sub-benchmark
	MaxProcs  int             // the value of GOMAXPROCS when the benchmark was run
}

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

// BenchOutputs are the outputs of a single benchmark.
// Just the relevant parts of https://godoc.org/golang.org/x/tools/benchmark/parse#Benchmark
type BenchOutputs struct {
	N                 int     // number of iterations
	NsPerOp           float64 // nanoseconds per iteration
	AllocedBytesPerOp uint64  // bytes allocated per iteration
	AllocsPerOp       uint64  // allocs per iteration
	MBPerS            float64 // MB processed per second
	Measured          int     // which measurements were recorded
}

func (b BenchOutputs) String() string {
	// just the relevant parts of https://godoc.org/golang.org/x/tools/benchmark/parse#Benchmark.String
	var s strings.Builder
	s.WriteString(strconv.Itoa(b.N))
	if (b.Measured & parse.NsPerOp) != 0 {
		fmt.Fprintf(&s, " %.2f ns/op", b.NsPerOp)
	}
	if (b.Measured & parse.MBPerS) != 0 {
		fmt.Fprintf(&s, " %.2f MB/s", b.MBPerS)
	}
	if (b.Measured & parse.AllocedBytesPerOp) != 0 {
		fmt.Fprintf(&s, " %d B/op", b.AllocedBytesPerOp)
	}
	if (b.Measured & parse.AllocsPerOp) != 0 {
		fmt.Fprintf(&s, " %d allocs/op", b.AllocsPerOp)
	}
	return s.String()
}

// BenchRes represents a result from a single benchmark run.
type BenchRes struct {
	Inputs  BenchInputs  // the input variables
	Outputs BenchOutputs // the output result
}

// GroupedResults represents a grouping of benchmark results.
type GroupedResults map[string][]BenchRes

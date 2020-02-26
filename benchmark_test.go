package benchparse

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/tools/benchmark/parse"
)

var parseBenchmarksTests = map[string]struct {
	resultSet          string
	expectedBenchmarks []Benchmark
	expectErr          bool
}{
	"1_bench_2_subs": {
		resultSet: `
			goos: darwin
			goarch: amd64
			BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2/end_x=1/abs_val=true-4         	   21801	     55357 ns/op	       0 B/op	       0 allocs/op
			BenchmarkMath/areaUnder/y=2x+3/delta=1.000000/start_x=-1/end_x=2/abs_val=false-4          	88335925	        13.3 ns/op	       0 B/op	       0 allocs/op
			BenchmarkMath/max/y=2x+3/delta=0.001000/start_x=-2/end_x=1-4                              	   56282	     20361 ns/op	       0 B/op	       0 allocs/op
			BenchmarkMath/max/y=sin(x)/delta=1.000000/start_x=-1/end_x=2-4                            	16381138	        62.7 ns/op	       0 B/op	       0 allocs/op
			PASS
			`,
		expectedBenchmarks: []Benchmark{
			{
				Name: "BenchmarkMath",
				Results: []BenchRes{
					{
						Inputs: BenchInputs{
							Subs: []BenchSub{{Name: "areaUnder", position: 1}},
							VarValues: []BenchVarValue{
								{Name: "y", Value: "sin(x)", position: 2},
								{Name: "delta", Value: 0.001, position: 3},
								{Name: "start_x", Value: -2, position: 4},
								{Name: "end_x", Value: 1, position: 5},
								{Name: "abs_val", Value: true, position: 6},
							},
							MaxProcs: 4,
						},
						Outputs: BenchOutputs{N: 21801, NsPerOp: 55357, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
					},
					{
						Inputs: BenchInputs{
							Subs: []BenchSub{{Name: "areaUnder", position: 1}},
							VarValues: []BenchVarValue{
								{Name: "y", Value: "2x+3", position: 2},
								{Name: "delta", Value: 1.0, position: 3},
								{Name: "start_x", Value: -1, position: 4},
								{Name: "end_x", Value: 2, position: 5},
								{Name: "abs_val", Value: false, position: 6},
							},
							MaxProcs: 4,
						},
						Outputs: BenchOutputs{N: 88335925, NsPerOp: 13.3, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
					},
					{
						Inputs: BenchInputs{
							Subs: []BenchSub{{Name: "max", position: 1}},
							VarValues: []BenchVarValue{
								{Name: "y", Value: "2x+3", position: 2},
								{Name: "delta", Value: 0.001, position: 3},
								{Name: "start_x", Value: -2, position: 4},
								{Name: "end_x", Value: 1, position: 5},
							},
							MaxProcs: 4,
						},
						Outputs: BenchOutputs{N: 56282, NsPerOp: 20361, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
					},
					{
						Inputs: BenchInputs{
							Subs: []BenchSub{{Name: "max", position: 1}},
							VarValues: []BenchVarValue{
								{Name: "y", Value: "sin(x)", position: 2},
								{Name: "delta", Value: 1.0, position: 3},
								{Name: "start_x", Value: -1, position: 4},
								{Name: "end_x", Value: 2, position: 5},
							},
							MaxProcs: 4,
						},
						Outputs: BenchOutputs{N: 16381138, NsPerOp: 62.7, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
					},
				},
			},
		},
	},
}

func TestParseBencharks(t *testing.T) {
	for testName, testCase := range parseBenchmarksTests {
		t.Run(testName, func(t *testing.T) {
			b := bytes.NewReader([]byte(testCase.resultSet))
			benchmarks, err := ParseBenchmarks(b)
			if err != nil {
				if !testCase.expectErr {
					t.Errorf("unexpected error: %s", err)
				}
				return
			}

			if testCase.expectErr {
				t.Fatalf("unexpectedly no error")
			}

			if !reflect.DeepEqual(benchmarks, testCase.expectedBenchmarks) {
				t.Errorf("unexpected parsed benchmarks\nexpected:\n%v\nactual:\n%v", testCase.expectedBenchmarks, benchmarks)
			}
		})
	}
}

func TestBenchmarkString(t *testing.T) {
	bench := Benchmark{
		Name: "BenchmarkMath",
		Results: []BenchRes{
			{
				Inputs: BenchInputs{
					Subs: []BenchSub{{Name: "areaUnder", position: 1}},
					VarValues: []BenchVarValue{
						{Name: "y", Value: "sin(x)", position: 2},
						{Name: "delta", Value: 0.001, position: 3},
						{Name: "start_x", Value: -2, position: 4},
						{Name: "end_x", Value: 1, position: 5},
						{Name: "abs_val", Value: true, position: 6},
					},
					MaxProcs: 4,
				},
				Outputs: BenchOutputs{N: 21801, NsPerOp: 55357, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
			},
			{
				Inputs: BenchInputs{
					Subs: []BenchSub{{Name: "areaUnder", position: 1}},
					VarValues: []BenchVarValue{
						{Name: "y", Value: "2x+3", position: 2},
						{Name: "delta", Value: 1.0, position: 3},
						{Name: "start_x", Value: -1, position: 4},
						{Name: "end_x", Value: 2, position: 5},
						{Name: "abs_val", Value: false, position: 6},
					},
					MaxProcs: 4,
				},
				Outputs: BenchOutputs{N: 88335925, NsPerOp: 13.3, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
			},
			{
				Inputs: BenchInputs{
					Subs: []BenchSub{{Name: "max", position: 1}},
					VarValues: []BenchVarValue{
						{Name: "y", Value: "2x+3", position: 2},
						{Name: "delta", Value: 0.001, position: 3},
						{Name: "start_x", Value: -2, position: 4},
						{Name: "end_x", Value: 1, position: 5},
					},
					MaxProcs: 4,
				},
				Outputs: BenchOutputs{N: 56282, NsPerOp: 20361, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
			},
			{
				Inputs: BenchInputs{
					Subs: []BenchSub{{Name: "max", position: 1}},
					VarValues: []BenchVarValue{
						{Name: "y", Value: "sin(x)", position: 2},
						{Name: "delta", Value: 1.0, position: 3},
						{Name: "start_x", Value: -1, position: 4},
						{Name: "end_x", Value: 2, position: 5},
					},
					MaxProcs: 4,
				},
				Outputs: BenchOutputs{N: 16381138, NsPerOp: 62.7, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
			},
		},
	}

	s := bench.String()
	// slightly different float precision than input
	expectedString := `BenchmarkMath/areaUnder/y=sin(x)/delta=0.001/start_x=-2/end_x=1/abs_val=true-4 21801 55357.00 ns/op 0 B/op 0 allocs/op
BenchmarkMath/areaUnder/y=2x+3/delta=1/start_x=-1/end_x=2/abs_val=false-4 88335925 13.30 ns/op 0 B/op 0 allocs/op
BenchmarkMath/max/y=2x+3/delta=0.001/start_x=-2/end_x=1-4 56282 20361.00 ns/op 0 B/op 0 allocs/op
BenchmarkMath/max/y=sin(x)/delta=1/start_x=-1/end_x=2-4 16381138 62.70 ns/op 0 B/op 0 allocs/op`
	if s != expectedString {
		t.Errorf("unexpected string\nexpected:\n%s\nactual:\n%s", expectedString, s)
	}
}

func ExampleParseBenchmarks() {
	r := strings.NewReader(`
			BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2/end_x=1/abs_val=true-4         	   21801	     55357 ns/op	       0 B/op	       0 allocs/op
			BenchmarkMath/areaUnder/y=2x+3/delta=1.000000/start_x=-1/end_x=2/abs_val=false-4          	88335925	        13.3 ns/op	       0 B/op	       0 allocs/op
			BenchmarkMath/max/y=2x+3/delta=0.001000/start_x=-2/end_x=1-4                              	   56282	     20361 ns/op	       0 B/op	       0 allocs/op
			BenchmarkMath/max/y=sin(x)/delta=1.000000/start_x=-1/end_x=2-4                            	16381138	        62.7 ns/op	       0 B/op	       0 allocs/op
			`)
	benches, err := ParseBenchmarks(r)
	if err != nil {
		log.Fatal(err)
	}

	for _, bench := range benches {
		fmt.Printf("bench name: %s\n", bench.Name)
		for _, res := range bench.Results {
			var (
				varValues = make([]string, len(res.Inputs.VarValues))
				otherSubs = make([]string, len(res.Inputs.Subs))
			)

			for i, varVal := range res.Inputs.VarValues {
				varValues[i] = varVal.String()
			}
			for i, sub := range res.Inputs.Subs {
				otherSubs[i] = sub.String()
			}

			fmt.Printf("var values = %q\n", varValues)
			fmt.Printf("other subs = %q\n", otherSubs)
			fmt.Printf("ns per op = %.2f\n", res.Outputs.NsPerOp)
		}
	}
	// Output:
	// bench name: BenchmarkMath
	// var values = ["y=sin(x)" "delta=0.001" "start_x=-2" "end_x=1" "abs_val=true"]
	// other subs = ["areaUnder"]
	// ns per op = 55357.00
	// var values = ["y=2x+3" "delta=1" "start_x=-1" "end_x=2" "abs_val=false"]
	// other subs = ["areaUnder"]
	// ns per op = 13.30
	// var values = ["y=2x+3" "delta=0.001" "start_x=-2" "end_x=1"]
	// other subs = ["max"]
	// ns per op = 20361.00
	// var values = ["y=sin(x)" "delta=1" "start_x=-1" "end_x=2"]
	// other subs = ["max"]
	// ns per op = 62.70
}

var parseBenchmarksErr error

func BenchmarkParseBenchmarks(b *testing.B) {
	var (
		allNumBenchmarks     = []int{1, 2, 3, 4, 5}
		allCasesPerBenchmark = []int{5, 10, 15, 20, 25}
	)

	for _, numBenchmarks := range allNumBenchmarks {
		b.Run(fmt.Sprintf("num_benchmarks=%d", numBenchmarks), func(b *testing.B) {
			for _, casesPerBench := range allCasesPerBenchmark {
				b.Run(fmt.Sprintf("cases_per_bench=%d", casesPerBench), func(b *testing.B) {
					benchmarkParseBenchmarks(b, numBenchmarks, casesPerBench)
				})
			}
		})
	}
}

func benchmarkParseBenchmarks(b *testing.B, numBenchmarks, casesPerBench int) {
	b.Helper()
	newReader := func() io.Reader {
		var buf bytes.Buffer
		for i := 0; i < numBenchmarks; i++ {
			benchName := fmt.Sprintf("BenchmarkMethod%d", i)
			for j := 0; j < casesPerBench; j++ {
				bench := &parse.Benchmark{
					Name:    fmt.Sprintf("%s/var1=%d/var2=%d", benchName, j, j),
					N:       j,
					NsPerOp: float64(j),
				}
				if _, err := buf.WriteString(fmt.Sprintf("%s\n", bench)); err != nil {
					b.Fatalf("err constructing input: %s", err)
				}
			}
		}
		return &buf
	}

	var err error
	var benches []Benchmark
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := newReader()
		b.StartTimer()

		benches, err = ParseBenchmarks(r)
		if err != nil {
			b.Fatalf("unexpected error: %s", err)
		}
		if len(benches) != numBenchmarks {
			b.Fatalf("unexpected number of benchmarks (expected=%d, actual=%d)", numBenchmarks, len(benches))
		}
	}
	parseBenchmarksErr = err
}

var parseInfoErr error

func BenchmarkParseInfo(b *testing.B) {
	var (
		dTypes = map[string]func(varName string) string{
			"int": func(varName string) string {
				return fmt.Sprintf("%s=%d", varName, 1)
			},
			"float64": func(varName string) string {
				return fmt.Sprintf("%s=%f", varName, 1.1)
			},
			"bool": func(varName string) string {
				return fmt.Sprintf("%s=%t", varName, true)
			},
			"string": func(varName string) string {
				return fmt.Sprintf("%s=%s", varName, "foo")
			},
		}
		allNumValues = []int{1, 2, 3, 4, 5, 10, 20}
	)

	for _, numValues := range allNumValues {
		b.Run(fmt.Sprintf("num_values=%d", numValues), func(b *testing.B) {
			for dtype, fn := range dTypes {
				b.Run(fmt.Sprintf("dtype=%s", dtype), func(b *testing.B) {
					s := make([]string, numValues+1)
					s[0] = "BenchmarkSomeMethod"
					for i := 1; i <= numValues; i++ {
						s[i] = fn(fmt.Sprintf("var%d", i))
					}
					input := strings.Join(s, "/")

					var err error
					for i := 0; i < b.N; i++ {
						_, _, err = parseInfo(input)
						if err != nil {
							b.Fatalf("unexpected error: %s", err)
						}
					}
					parseInfoErr = err
				})
			}
		})
	}
}

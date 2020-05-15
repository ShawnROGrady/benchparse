package benchparse

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/benchmark/parse"
)

func testBenchmarkEqual(t *testing.T, expected, actual Benchmark) {
	t.Helper()
	if expected.Name != actual.Name {
		t.Errorf("unexpected name (expected=%s, actual=%s)", expected.Name, actual.Name)
	}

	if len(expected.Results) == len(actual.Results) {
		for i := range expected.Results {
			testBenchResEq(t, expected.Results[i], actual.Results[i])
		}
	} else {
		t.Errorf("unexpected results\nexpected (len=%d):\n%#v\nactual (len=%d):\n%#v", len(expected.Results), expected.Results, len(actual.Results), actual.Results)
	}
}

var sampleBench = Benchmark{
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
			Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2/end_x=1/abs_val=true-4", N: 21801, NsPerOp: 55357, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp}},
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
			Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkMath/areaUnder/y=2x+3/delta=1.000000/start_x=-1/end_x=2/abs_val=false-4", N: 88335925, NsPerOp: 13.3, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp}},
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
			Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkMath/max/y=2x+3/delta=0.001000/start_x=-2/end_x=1-4", N: 56282, NsPerOp: 20361, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp}},
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
			Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkMath/max/y=sin(x)/delta=1.000000/start_x=-1/end_x=2-4", N: 16381138, NsPerOp: 62.7, Measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp}},
		},
	},
}

var parseBenchmarksTests = map[string]struct {
	resultSet          string
	expectedBenchmarks []Benchmark
	expectErr          bool
}{
	"1_bench_4_cases_benchmem_set": {
		resultSet: `
			goos: darwin
			goarch: amd64
			BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2/end_x=1/abs_val=true-4         	   21801	     55357 ns/op	       0 B/op	       0 allocs/op
			BenchmarkMath/areaUnder/y=2x+3/delta=1.000000/start_x=-1/end_x=2/abs_val=false-4          	88335925	        13.3 ns/op	       0 B/op	       0 allocs/op
			BenchmarkMath/max/y=2x+3/delta=0.001000/start_x=-2/end_x=1-4                              	   56282	     20361 ns/op	       0 B/op	       0 allocs/op
			BenchmarkMath/max/y=sin(x)/delta=1.000000/start_x=-1/end_x=2-4                            	16381138	        62.7 ns/op	       0 B/op	       0 allocs/op
			PASS
			`,
		expectedBenchmarks: []Benchmark{sampleBench},
	},
	"1_bench_2_cases_bytes_set": {
		resultSet: `
			BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=5-4              37098             31052 ns/op     5.31 MB/s
			BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=10-4             23004             52099 ns/op     6.33 MB/s
			`,
		expectedBenchmarks: []Benchmark{{
			Name: "BenchmarkParseBenchmarks",
			Results: []BenchRes{
				{
					Inputs: BenchInputs{
						VarValues: []BenchVarValue{
							{Name: "num_benchmarks", Value: 1, position: 1},
							{Name: "cases_per_bench", Value: 5, position: 2},
						},
						Subs:     []BenchSub{},
						MaxProcs: 4,
					},
					Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=5-4", N: 37098, NsPerOp: 31052, MBPerS: 5.31, Measured: parse.NsPerOp | parse.MBPerS}},
				},
				{
					Inputs: BenchInputs{
						VarValues: []BenchVarValue{
							{Name: "num_benchmarks", Value: 1, position: 1},
							{Name: "cases_per_bench", Value: 10, position: 2},
						},
						Subs:     []BenchSub{},
						MaxProcs: 4,
					},
					Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=10-4", N: 23004, NsPerOp: 52099, MBPerS: 6.33, Measured: parse.NsPerOp | parse.MBPerS}},
				},
			},
		}},
	},
	"2_benches_2_cases": {
		resultSet: `
			BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=5              37098             31052 ns/op
			BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=10             23004             52099 ns/op
			BenchmarkParseInfo/num_values=1/dtype=int                 624967              1721 ns/op
			BenchmarkParseInfo/num_values=1/dtype=float64             509164              2239 ns/op
			`,
		expectedBenchmarks: []Benchmark{
			{
				Name: "BenchmarkParseBenchmarks",
				Results: []BenchRes{
					{
						Inputs: BenchInputs{
							VarValues: []BenchVarValue{
								{Name: "num_benchmarks", Value: 1, position: 1},
								{Name: "cases_per_bench", Value: 5, position: 2},
							},
							Subs:     []BenchSub{},
							MaxProcs: 1,
						},
						Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=5", N: 37098, NsPerOp: 31052, Measured: parse.NsPerOp}},
					},
					{
						Inputs: BenchInputs{
							VarValues: []BenchVarValue{
								{Name: "num_benchmarks", Value: 1, position: 1},
								{Name: "cases_per_bench", Value: 10, position: 2},
							},
							Subs:     []BenchSub{},
							MaxProcs: 1,
						},
						Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=10", N: 23004, NsPerOp: 52099, Measured: parse.NsPerOp}},
					},
				},
			},
			{
				Name: "BenchmarkParseInfo",
				Results: []BenchRes{
					{
						Inputs: BenchInputs{
							VarValues: []BenchVarValue{
								{Name: "num_values", Value: 1, position: 1},
								{Name: "dtype", Value: "int", position: 2},
							},
							Subs:     []BenchSub{},
							MaxProcs: 1,
						},
						Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkParseInfo/num_values=1/dtype=int", N: 624967, NsPerOp: 1721, Measured: parse.NsPerOp}},
					},
					{
						Inputs: BenchInputs{
							VarValues: []BenchVarValue{
								{Name: "num_values", Value: 1, position: 1},
								{Name: "dtype", Value: "float64", position: 2},
							},
							Subs:     []BenchSub{},
							MaxProcs: 1,
						},
						Outputs: parsedBenchOutputs{parse.Benchmark{Name: "BenchmarkParseInfo/num_values=1/dtype=float64", N: 509164, NsPerOp: 2239, Measured: parse.NsPerOp}},
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

			// sort the benchmarks by name for consistent results
			sort.Slice(benchmarks, func(i, j int) bool {
				return benchmarks[i].Name < benchmarks[j].Name
			})

			if !reflect.DeepEqual(benchmarks, testCase.expectedBenchmarks) {
				t.Errorf("unexpected parsed benchmarks\nexpected:\n%v\nactual:\n%v", testCase.expectedBenchmarks, benchmarks)
			}
		})
	}
}

var parseBenchmarksFromJSONTests = map[string]struct {
	resultSet          string
	expectedBenchmarks []Benchmark
	expectErr          bool
}{
	"1_bench_4_cases_benchmem_set": {
		resultSet: `{"Time":"2020-05-13T22:50:47.859655-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"goos: darwin\n"}
{"Time":"2020-05-13T22:50:47.860205-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"goarch: amd64\n"}
{"Time":"2020-05-13T22:50:47.860222-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath\n"}
{"Time":"2020-05-13T22:50:47.860239-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/areaUnder\n"}
{"Time":"2020-05-13T22:50:47.860942-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/areaUnder/y=sin(x)\n"}
{"Time":"2020-05-13T22:50:47.861468-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000\n"}
{"Time":"2020-05-13T22:50:47.861999-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2\n"}
{"Time":"2020-05-13T22:50:47.862419-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2/end_x=1\n"}
{"Time":"2020-05-13T22:50:47.862817-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2/end_x=1/abs_val=true\n"}
{"Time":"2020-05-13T22:50:49.609057-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2/end_x=1/abs_val=true-4         \t   21801\t     55357 ns/op\t       0 B/op\t       0 allocs/op\n"}
{"Time":"2020-05-13T22:57:01.99228-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/areaUnder/y=sin(x)/delta=1.000000/start_x=-1/end_x=2/abs_val=false\n"}
{"Time":"2020-05-13T22:57:01.992288-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/areaUnder/y=2x+3/delta=1.000000/start_x=-1/end_x=2/abs_val=false-4        \t88335925\t        13.3 ns/op\t       0 B/op\t       0 allocs/op\n"}
{"Time":"2020-05-13T22:57:01.994853-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/max\n"}
{"Time":"2020-05-13T22:57:01.994961-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/max/y=2x+3\n"}
{"Time":"2020-05-13T22:57:01.994973-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/max/y=2x+3/delta=0.001000\n"}
{"Time":"2020-05-13T22:57:01.994979-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/max/y=2x+3/delta=0.001000/start_x=-2\n"}
{"Time":"2020-05-13T22:57:01.994986-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/max/y=2x+3/delta=0.001000/start_x=-2/end_x=1\n"}
{"Time":"2020-05-13T22:57:01.994993-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/max/y=2x+3/delta=0.001000/start_x=-2/end_x=1-4                            \t   56282\t     20361 ns/op\t       0 B/op\t       0 allocs/op\n"}
{"Time":"2020-05-13T22:57:01.997333-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/max/y=sin(x)/delta=1.000000/start_x=-1/end_x=2\n"}                                                                                                                                                                
{"Time":"2020-05-13T22:57:01.997344-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"BenchmarkMath/max/y=sin(x)/delta=1.000000/start_x=-1/end_x=2-4                              \t16381138\t        62.7 ns/op\t       0 B/op\t       0 allocs/op\n"}
{"Time":"2020-05-13T22:57:01.997351-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"PASS\n"}
{"Time":"2020-05-13T22:57:01.9975-05:00","Action":"output","Package":"github.com/ShawnROGrady/mathtest","Output":"ok  \tgithub.com/ShawnROGrady/mathtest\t374.272s\n"}
{"Time":"2020-05-13T22:57:01.998418-05:00","Action":"pass","Package":"github.com/ShawnROGrady/mathtest","Elapsed":374.273}`,
		expectedBenchmarks: []Benchmark{sampleBench},
	},
	"non_json": {
		resultSet: `
			goos: darwin
			goarch: amd64
			BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2/end_x=1/abs_val=true-4         	   21801	     55357 ns/op	       0 B/op	       0 allocs/op
			`,
		expectErr: true,
	},
}

func TestParseBencharksFromJSON(t *testing.T) {
	for testName, testCase := range parseBenchmarksFromJSONTests {
		t.Run(testName, func(t *testing.T) {
			b := bytes.NewReader([]byte(testCase.resultSet))
			benchmarks, err := ParseBenchmarksFromJSON(b)
			if err != nil {
				if !testCase.expectErr {
					t.Errorf("unexpected error: %s", err)
				}
				return
			}

			if testCase.expectErr {
				t.Fatalf("unexpectedly no error")
			}

			// sort the benchmarks by name for consistent results
			sort.Slice(benchmarks, func(i, j int) bool {
				return benchmarks[i].Name < benchmarks[j].Name
			})

			if !reflect.DeepEqual(benchmarks, testCase.expectedBenchmarks) {
				t.Errorf("unexpected parsed benchmarks\nexpected:\n%v\nactual:\n%v", testCase.expectedBenchmarks, benchmarks)
			}
		})
	}
}

type badReader struct{}

func (b badReader) Read([]byte) (int, error) { return 0, errors.New("test error") }

func TestParseBenchmarksReadErr(t *testing.T) {
	r := badReader{}
	_, err := ParseBenchmarks(r)
	if err == nil {
		t.Errorf("unexpectedly no error")
	}
}

var benchmarkStringTests = map[string]struct {
	bench          Benchmark
	expectedString string
}{
	"benchmem_enabled": {
		bench: sampleBench,
		// slightly different float precision than input
		expectedString: `BenchmarkMath/areaUnder/y=sin(x)/delta=0.001000/start_x=-2/end_x=1/abs_val=true-4 21801 55357.00 ns/op 0 B/op 0 allocs/op
BenchmarkMath/areaUnder/y=2x+3/delta=1.000000/start_x=-1/end_x=2/abs_val=false-4 88335925 13.30 ns/op 0 B/op 0 allocs/op
BenchmarkMath/max/y=2x+3/delta=0.001000/start_x=-2/end_x=1-4 56282 20361.00 ns/op 0 B/op 0 allocs/op
BenchmarkMath/max/y=sin(x)/delta=1.000000/start_x=-1/end_x=2-4 16381138 62.70 ns/op 0 B/op 0 allocs/op`,
	},
	"bytes_set": {
		bench: Benchmark{
			Name: "BenchmarkParseBenchmarks",
			Results: []BenchRes{
				{
					Inputs: BenchInputs{
						VarValues: []BenchVarValue{
							{Name: "num_benchmarks", Value: 1, position: 1},
							{Name: "cases_per_bench", Value: 5, position: 2},
						},
						Subs:     []BenchSub{},
						MaxProcs: 4,
					},
					Outputs: parsedBenchOutputs{parse.Benchmark{N: 37098, NsPerOp: 31052, MBPerS: 5.31, Measured: parse.NsPerOp | parse.MBPerS}},
				},
				{
					Inputs: BenchInputs{
						VarValues: []BenchVarValue{
							{Name: "num_benchmarks", Value: 1, position: 1},
							{Name: "cases_per_bench", Value: 10, position: 2},
						},
						Subs:     []BenchSub{},
						MaxProcs: 4,
					},
					Outputs: parsedBenchOutputs{parse.Benchmark{N: 23004, NsPerOp: 52099, MBPerS: 6.33, Measured: parse.NsPerOp | parse.MBPerS}},
				},
			},
		},
		expectedString: `BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=5-4 37098 31052.00 ns/op 5.31 MB/s
BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=10-4 23004 52099.00 ns/op 6.33 MB/s`,
	},
	"go_max_procs=1": {
		bench: Benchmark{
			Name: "BenchmarkParseBenchmarks",
			Results: []BenchRes{
				{
					Inputs: BenchInputs{
						VarValues: []BenchVarValue{
							{Name: "num_benchmarks", Value: 1, position: 1},
							{Name: "cases_per_bench", Value: 5, position: 2},
						},
						Subs:     []BenchSub{},
						MaxProcs: 1,
					},
					Outputs: parsedBenchOutputs{parse.Benchmark{N: 37098, NsPerOp: 31052, Measured: parse.NsPerOp}},
				},
				{
					Inputs: BenchInputs{
						VarValues: []BenchVarValue{
							{Name: "num_benchmarks", Value: 1, position: 1},
							{Name: "cases_per_bench", Value: 10, position: 2},
						},
						Subs:     []BenchSub{},
						MaxProcs: 1,
					},
					Outputs: parsedBenchOutputs{parse.Benchmark{N: 23004, NsPerOp: 52099, Measured: parse.NsPerOp}},
				},
			},
		},
		expectedString: `BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=5 37098 31052.00 ns/op
BenchmarkParseBenchmarks/num_benchmarks=1/cases_per_bench=10 23004 52099.00 ns/op`,
	},
}

func TestBenchmarkString(t *testing.T) {
	for testName, testCase := range benchmarkStringTests {
		t.Run(testName, func(t *testing.T) {
			s := testCase.bench.String()
			if s != testCase.expectedString {
				t.Errorf("unexpected string\nexpected:\n%s\nactual:\n%s", testCase.expectedString, s)
			}

			r := strings.NewReader(s)
			benches, err := ParseBenchmarks(r)
			if err != nil {
				t.Fatalf("unexpected error parsing from string: %s", err)
			}
			testBenchmarkEqual(t, testCase.bench, benches[0])
		})
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
			nsPerOp, err := res.Outputs.GetNsPerOp()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("ns per op = %.2f\n", nsPerOp)
		}
	}
	// Output:
	// bench name: BenchmarkMath
	// var values = ["y=sin(x)" "delta=0.001000" "start_x=-2" "end_x=1" "abs_val=true"]
	// other subs = ["areaUnder"]
	// ns per op = 55357.00
	// var values = ["y=2x+3" "delta=1.000000" "start_x=-1" "end_x=2" "abs_val=false"]
	// other subs = ["areaUnder"]
	// ns per op = 13.30
	// var values = ["y=2x+3" "delta=0.001000" "start_x=-2" "end_x=1"]
	// other subs = ["max"]
	// ns per op = 20361.00
	// var values = ["y=sin(x)" "delta=1.000000" "start_x=-1" "end_x=2"]
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
		b.SetBytes(int64(buf.Len()))
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

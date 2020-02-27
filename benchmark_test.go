package benchparse

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/benchmark/parse"
)

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
			Outputs: BenchOutputs{N: 21801, nsPerOp: 55357, measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
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
			Outputs: BenchOutputs{N: 88335925, nsPerOp: 13.3, measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
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
			Outputs: BenchOutputs{N: 56282, nsPerOp: 20361, measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
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
			Outputs: BenchOutputs{N: 16381138, nsPerOp: 62.7, measured: parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp},
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
					Outputs: BenchOutputs{N: 37098, nsPerOp: 31052, mBPerS: 5.31, measured: parse.NsPerOp | parse.MBPerS},
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
					Outputs: BenchOutputs{N: 23004, nsPerOp: 52099, mBPerS: 6.33, measured: parse.NsPerOp | parse.MBPerS},
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
						Outputs: BenchOutputs{N: 37098, nsPerOp: 31052, measured: parse.NsPerOp},
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
						Outputs: BenchOutputs{N: 23004, nsPerOp: 52099, measured: parse.NsPerOp},
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
						Outputs: BenchOutputs{N: 624967, nsPerOp: 1721, measured: parse.NsPerOp},
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
						Outputs: BenchOutputs{N: 509164, nsPerOp: 2239, measured: parse.NsPerOp},
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

var benchmarkStringTests = map[string]struct {
	bench          Benchmark
	expectedString string
}{
	"benchmem_enabled": {
		bench: sampleBench,
		// slightly different float precision than input
		expectedString: `BenchmarkMath/areaUnder/y=sin(x)/delta=0.001/start_x=-2/end_x=1/abs_val=true-4 21801 55357.00 ns/op 0 B/op 0 allocs/op
BenchmarkMath/areaUnder/y=2x+3/delta=1/start_x=-1/end_x=2/abs_val=false-4 88335925 13.30 ns/op 0 B/op 0 allocs/op
BenchmarkMath/max/y=2x+3/delta=0.001/start_x=-2/end_x=1-4 56282 20361.00 ns/op 0 B/op 0 allocs/op
BenchmarkMath/max/y=sin(x)/delta=1/start_x=-1/end_x=2-4 16381138 62.70 ns/op 0 B/op 0 allocs/op`,
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
					Outputs: BenchOutputs{N: 37098, nsPerOp: 31052, mBPerS: 5.31, measured: parse.NsPerOp | parse.MBPerS},
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
					Outputs: BenchOutputs{N: 23004, nsPerOp: 52099, mBPerS: 6.33, measured: parse.NsPerOp | parse.MBPerS},
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
					Outputs: BenchOutputs{N: 37098, nsPerOp: 31052, measured: parse.NsPerOp},
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
					Outputs: BenchOutputs{N: 23004, nsPerOp: 52099, measured: parse.NsPerOp},
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
			nsPerOp, err := res.Outputs.NsPerOp()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("ns per op = %.2f\n", nsPerOp)
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

var groupResultsTests = map[string]struct {
	benchmark              Benchmark
	groupBy                []string
	expectedGroupedResults GroupedResults
}{
	"group_by_1_string_var": {
		benchmark: sampleBench,
		groupBy:   []string{"y"},
		expectedGroupedResults: map[string][]BenchRes{
			"y=sin(x)": []BenchRes{
				sampleBench.Results[0],
				sampleBench.Results[3],
			},
			"y=2x+3": []BenchRes{
				sampleBench.Results[1],
				sampleBench.Results[2],
			},
		},
	},
	"no_group_by": {
		benchmark: sampleBench,
		expectedGroupedResults: map[string][]BenchRes{
			"": []BenchRes{
				sampleBench.Results[0],
				sampleBench.Results[1],
				sampleBench.Results[2],
				sampleBench.Results[3],
			},
		},
	},
	"group_by_2_vars": {
		benchmark: sampleBench,
		groupBy:   []string{"y", "delta"},
		expectedGroupedResults: map[string][]BenchRes{
			"y=sin(x),delta=0.001": []BenchRes{
				sampleBench.Results[0],
			},
			"y=2x+3,delta=1": []BenchRes{
				sampleBench.Results[1],
			},
			"y=2x+3,delta=0.001": []BenchRes{
				sampleBench.Results[2],
			},
			"y=sin(x),delta=1": []BenchRes{
				sampleBench.Results[3],
			},
		},
	},
	"group_by_sub-specific_bool_var": {
		benchmark: sampleBench,
		groupBy:   []string{"abs_val"}, // only present on half the results
		expectedGroupedResults: map[string][]BenchRes{
			"abs_val=true": []BenchRes{
				sampleBench.Results[0],
			},
			"abs_val=false": []BenchRes{
				sampleBench.Results[1],
			},
		},
	},
}

func TestGroupResults(t *testing.T) {
	for testName, testCase := range groupResultsTests {
		t.Run(testName, func(t *testing.T) {
			grouped := testCase.benchmark.GroupResults(testCase.groupBy)
			if !reflect.DeepEqual(grouped, testCase.expectedGroupedResults) {
				t.Errorf("unexpected grouped results\nexpected:\n%v\nactual:\n%v", testCase.expectedGroupedResults, grouped)
			}
		})
	}
}

func ExampleBenchmark_GroupResults() {
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

	groupedResults := benches[0].GroupResults([]string{"y"})

	// sort by key names to ensure consistent iteration order
	groupNames := make([]string, len(groupedResults))
	i := 0
	for k := range groupedResults {
		groupNames[i] = k
		i++
	}
	sort.Strings(groupNames)

	for _, k := range groupNames {
		fmt.Println(k)
		v := groupedResults[k]

		times := make([]float64, len(v))
		for i, res := range v {
			nsPerOp, err := res.Outputs.NsPerOp()
			if err != nil {
				log.Fatal(err)
			}
			times[i] = nsPerOp
		}
		fmt.Printf("ns per op = %v\n", times)
	}
	// Output:
	// y=2x+3
	// ns per op = [13.3 20361]
	// y=sin(x)
	// ns per op = [55357 62.7]
}

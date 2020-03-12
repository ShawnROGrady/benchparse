package benchparse

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/benchmark/parse"
)

func testBenchResEq(t *testing.T, expected, actual BenchRes) {
	t.Helper()
	if !reflect.DeepEqual(expected.Inputs, actual.Inputs) {
		t.Errorf("unexpected inputs\nexpected:\n%#v\nactual:\n%#v", expected.Inputs, actual.Inputs)
	}

	var (
		expectedIterations                                      = expected.Outputs.GetIterations()
		expectedNsPerOp, expectedNsPerOpErr                     = expected.Outputs.GetNsPerOp()
		expectedAllocedBytesPerOp, expectedAllocedBytesPerOpErr = expected.Outputs.GetAllocedBytesPerOp()
		expectedAllocsPerOp, expectedAllocsPerOpErr             = expected.Outputs.GetAllocsPerOp()
		expectedMBPerS, expectedMBPerSErr                       = expected.Outputs.GetMBPerS()

		actualIterations                                    = actual.Outputs.GetIterations()
		actualNsPerOp, actualNsPerOpErr                     = actual.Outputs.GetNsPerOp()
		actualAllocedBytesPerOp, actualAllocedBytesPerOpErr = actual.Outputs.GetAllocedBytesPerOp()
		actualAllocsPerOp, actualAllocsPerOpErr             = actual.Outputs.GetAllocsPerOp()
		actualMBPerS, actualMBPerSErr                       = actual.Outputs.GetMBPerS()
	)

	if expectedIterations != actualIterations {
		t.Errorf("unexpected output iterations (expected=%d, actual=%d)", expectedIterations, actualIterations)
	}

	if expectedNsPerOp != actualNsPerOp || expectedNsPerOpErr != actualNsPerOpErr {
		t.Errorf("unexpected output GetNsPerOp()\nexpected:\n%v,%s\nactual:\n%v,%s", expectedNsPerOp, expectedNsPerOpErr, actualNsPerOp, actualNsPerOpErr)
	}

	if expectedAllocedBytesPerOp != actualAllocedBytesPerOp || expectedAllocedBytesPerOpErr != actualAllocedBytesPerOpErr {
		t.Errorf("unexpected output GetAllocedBytesPerOp()\nexpected:\n%v,%s\nactual:\n%v,%s", expectedAllocedBytesPerOp, expectedAllocedBytesPerOpErr, actualAllocedBytesPerOp, actualAllocedBytesPerOpErr)
	}

	if expectedAllocsPerOp != actualAllocsPerOp || expectedAllocsPerOpErr != actualAllocsPerOpErr {
		t.Errorf("unexpected output GetAllocsPerOp()\nexpected:\n%v,%s\nactual:\n%v,%s", expectedAllocsPerOp, expectedAllocsPerOpErr, actualAllocsPerOp, actualAllocsPerOpErr)
	}

	if expectedMBPerS != actualMBPerS || expectedMBPerSErr != actualMBPerSErr {
		t.Errorf("unexpected output GetMBPerS()\nexpected:\n%v,%s\nactual:\n%v,%s", expectedMBPerS, expectedMBPerSErr, actualMBPerS, actualMBPerSErr)
	}
}

var getOutputMeasurementTests = map[string]struct {
	output                       parsedBenchOutputs
	expectedNsPerOp              float64
	expectedNsPerOpErr           error
	expectedAllocedBytesPerOp    uint64
	expectedAllocedBytesPerOpErr error
	expectedAllocsPerOp          uint64
	expectedAllocsPerOpErr       error
	expectedMBPerS               float64
	expectedMBPerSErr            error
}{
	"all_set": {
		output: parsedBenchOutputs{parse.Benchmark{
			N:                 21801,
			NsPerOp:           55357,
			AllocedBytesPerOp: 4321,
			AllocsPerOp:       21,
			MBPerS:            0.12,
			Measured:          parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp | parse.MBPerS,
		}},
		expectedNsPerOp:           55357,
		expectedAllocedBytesPerOp: 4321,
		expectedAllocsPerOp:       21,
		expectedMBPerS:            0.12,
	},
	"benchmem_not_set_with_set_bytes": {
		output: parsedBenchOutputs{parse.Benchmark{
			N:        21801,
			NsPerOp:  55357,
			MBPerS:   0.12,
			Measured: parse.NsPerOp | parse.MBPerS,
		}},
		expectedNsPerOp:              55357,
		expectedAllocedBytesPerOpErr: ErrNotMeasured,
		expectedAllocsPerOpErr:       ErrNotMeasured,
		expectedMBPerS:               0.12,
	},
	"benchmem_set_but_no_allocs": {
		output: parsedBenchOutputs{parse.Benchmark{
			N:                 21801,
			NsPerOp:           55357,
			AllocedBytesPerOp: 0,
			AllocsPerOp:       0,
			Measured:          parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp,
		}},
		expectedNsPerOp:           55357,
		expectedAllocedBytesPerOp: 0,
		expectedAllocsPerOp:       0,
		expectedMBPerSErr:         ErrNotMeasured,
	},
	"none_set": {
		output:                       parsedBenchOutputs{},
		expectedNsPerOpErr:           ErrNotMeasured,
		expectedAllocedBytesPerOpErr: ErrNotMeasured,
		expectedAllocsPerOpErr:       ErrNotMeasured,
		expectedMBPerSErr:            ErrNotMeasured,
	},
}

func TestGetOutputMeasumentTests(t *testing.T) {
	for testName, testCase := range getOutputMeasurementTests {
		t.Run(testName, func(t *testing.T) {
			t.Run("ns_per_op", func(t *testing.T) {
				testNsPerOp(t, testCase.output, testCase.expectedNsPerOp, testCase.expectedNsPerOpErr)
			})
			t.Run("allocated_bytes_per_op", func(t *testing.T) {
				testAllocedBytesPerOp(t, testCase.output, testCase.expectedAllocedBytesPerOp, testCase.expectedAllocedBytesPerOpErr)
			})
			t.Run("allocs_per_op", func(t *testing.T) {
				testAllocsPerOp(t, testCase.output, testCase.expectedAllocsPerOp, testCase.expectedAllocsPerOpErr)
			})
			t.Run("MB_per_s", func(t *testing.T) {
				testMBPerS(t, testCase.output, testCase.expectedMBPerS, testCase.expectedMBPerSErr)
			})
		})
	}
}

func testNsPerOp(t *testing.T, b parsedBenchOutputs, expectedV float64, expectedErr error) {
	t.Helper()
	ns, err := b.GetNsPerOp()
	if err != nil {
		if expectedErr != nil {
			if err != expectedErr {
				t.Errorf("unexpected error received (expected=%s, actual=%s)", expectedErr, err)
			}
		} else {
			t.Errorf("unexpected error: %s", err)
		}
		return
	}

	if expectedErr != nil {
		t.Errorf("unexpectedly no error")
	}

	if expectedV != ns {
		t.Errorf("unexpected NsPerOp (expected=%v, actual=%v)", expectedV, ns)
	}
}

func testAllocedBytesPerOp(t *testing.T, b parsedBenchOutputs, expectedV uint64, expectedErr error) {
	t.Helper()
	v, err := b.GetAllocedBytesPerOp()
	if err != nil {
		if expectedErr != nil {
			if err != expectedErr {
				t.Errorf("unexpected error received (expected=%s, actual=%s)", expectedErr, err)
			}
		} else {
			t.Errorf("unexpected error: %s", err)
		}
		return
	}

	if expectedErr != nil {
		t.Errorf("unexpectedly no error")
	}

	if expectedV != v {
		t.Errorf("unexpected AllocedBytesPerOp (expected=%v, actual=%v)", expectedV, v)
	}
}

func testAllocsPerOp(t *testing.T, b parsedBenchOutputs, expectedV uint64, expectedErr error) {
	t.Helper()
	v, err := b.GetAllocsPerOp()
	if err != nil {
		if expectedErr != nil {
			if err != expectedErr {
				t.Errorf("unexpected error received (expected=%s, actual=%s)", expectedErr, err)
			}
		} else {
			t.Errorf("unexpected error: %s", err)
		}
		return
	}

	if expectedErr != nil {
		t.Errorf("unexpectedly no error")
	}

	if expectedV != v {
		t.Errorf("unexpected AllocsPerOp (expected=%v, actual=%v)", expectedV, v)
	}
}

func testMBPerS(t *testing.T, b parsedBenchOutputs, expectedV float64, expectedErr error) {
	t.Helper()
	v, err := b.GetMBPerS()
	if err != nil {
		if expectedErr != nil {
			if err != expectedErr {
				t.Errorf("unexpected error received (expected=%s, actual=%s)", expectedErr, err)
			}
		} else {
			t.Errorf("unexpected error: %s", err)
		}
		return
	}

	if expectedErr != nil {
		t.Errorf("unexpectedly no error")
	}

	if expectedV != v {
		t.Errorf("unexpected MBPerS (expected=%v, actual=%v)", expectedV, v)
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
		expectedGroupedResults: map[string]BenchResults{
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
		expectedGroupedResults: map[string]BenchResults{
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
		expectedGroupedResults: map[string]BenchResults{
			"y=sin(x),delta=0.001000": []BenchRes{
				sampleBench.Results[0],
			},
			"y=2x+3,delta=1.000000": []BenchRes{
				sampleBench.Results[1],
			},
			"y=2x+3,delta=0.001000": []BenchRes{
				sampleBench.Results[2],
			},
			"y=sin(x),delta=1.000000": []BenchRes{
				sampleBench.Results[3],
			},
		},
	},
	"group_by_sub-specific_bool_var": {
		benchmark: sampleBench,
		groupBy:   []string{"abs_val"}, // only present on half the results
		expectedGroupedResults: map[string]BenchResults{
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
			grouped := testCase.benchmark.Results.Group(testCase.groupBy)
			if !reflect.DeepEqual(grouped, testCase.expectedGroupedResults) {
				t.Errorf("unexpected grouped results\nexpected:\n%v\nactual:\n%v", testCase.expectedGroupedResults, grouped)
			}
		})
	}
}

func ExampleBenchResults_Group() {
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

	groupedResults := benches[0].Results.Group([]string{"y"})

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
			nsPerOp, err := res.Outputs.GetNsPerOp()
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

var filterTests = map[string]struct {
	results          BenchResults
	filterExpr       string
	expectedFiltered BenchResults
	expectedErr      error
}{
	"filter_by_string_eq": {
		results:          sampleBench.Results,
		filterExpr:       "y==sin(x)",
		expectedFiltered: BenchResults{sampleBench.Results[0], sampleBench.Results[3]},
	},
	"filter_by_float_gt": {
		results:          sampleBench.Results,
		filterExpr:       "delta>0.01",
		expectedFiltered: BenchResults{sampleBench.Results[1], sampleBench.Results[3]},
	},
	"filter_by_int_lt_float_val": {
		results:          sampleBench.Results,
		filterExpr:       "delta<1",
		expectedFiltered: BenchResults{sampleBench.Results[0], sampleBench.Results[2]},
	},
	"non_comparable_values": {
		results:     sampleBench.Results,
		filterExpr:  "y==2",
		expectedErr: errNonComparable,
	},
	"invalid_filter_expr": {
		results:     sampleBench.Results,
		filterExpr:  "y,2",
		expectedErr: errMalformedFilter,
	},
}

func TestFilter(t *testing.T) {
	for testName, testCase := range filterTests {
		t.Run(testName, func(t *testing.T) {
			filtered, err := testCase.results.Filter(testCase.filterExpr)
			if err != nil {
				if !errors.Is(err, testCase.expectedErr) {
					t.Errorf("unexpected error\nexpected=%s\nactual=%s", testCase.expectedErr, err)
				}
				return
			}

			if !reflect.DeepEqual(filtered, testCase.expectedFiltered) {
				t.Errorf("unexpected filtered results\nexpected:\n%v\nactual:\n%v", testCase.expectedFiltered, filtered)
			}
		})
	}
}

func BenchmarkFilterByInt(b *testing.B) {
	var (
		allComps      = []Comparison{Eq, Ne, Lt, Gt, Le, Ge}
		allNumResults = []int{10, 20, 30}
		allNumVars    = []int{2, 3, 5, 10, 20}
	)

	for _, cmp := range allComps {
		b.Run(fmt.Sprintf("cmp=%s", cmp.description()), func(b *testing.B) {
			for _, numResults := range allNumResults {
				b.Run(fmt.Sprintf("num_results=%d", numResults), func(b *testing.B) {
					for _, numVars := range allNumVars {
						b.Run(fmt.Sprintf("num_vars=%d", numVars), func(b *testing.B) {
							benchmarkFilterByInt(b, cmp, numResults, numVars)
						})
					}
				})
			}
		})
	}
}

var filterErr error

func benchmarkFilterByInt(b *testing.B, cmp Comparison, numResults, numVars int) {
	b.Helper()
	res := make(BenchResults, numResults)
	// the index of the var value of interest
	for i := 0; i < numResults; i++ {
		varVals := make([]BenchVarValue, numVars)
		for j := 0; j < numVars; j++ {
			val := j
			if cmp == Eq {
				val = 1
			}
			varVals[j] = BenchVarValue{
				Name:  fmt.Sprintf("var%d", j),
				Value: val,
			}
		}
		res[i] = BenchRes{
			Inputs: BenchInputs{VarValues: varVals},
		}
	}
	var filterVal int
	// make sure all match filter expression - easiest way to accurately compare performance
	switch cmp {
	case Eq:
		filterVal = 1
	case Ne, Lt, Le:
		filterVal = numVars + 1
	case Gt, Ge:
		filterVal = -1
	}
	filterExpr := fmt.Sprintf("var%d%s%d", numVars-1, cmp, filterVal)

	var (
		filtered BenchResults
		err      error
	)
	for i := 0; i < b.N; i++ {
		filtered, err = res.Filter(filterExpr)
		if err != nil {
			b.Errorf("unexpected err: %s", err)
		}
		if len(filtered) != len(res) {
			b.Errorf("unexpected number of filtered results (expected=%d, actual=%d)", len(res), len(filtered))
		}
	}
	filterErr = err
}

func ExampleBenchResults_Filter() {
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

	filtered, err := benches[0].Results.Filter("y==sin(x)")
	if err != nil {
		log.Fatal(err)
	}

	for _, res := range filtered {
		nsPerOp, err := res.Outputs.GetNsPerOp()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ns per op = %v\n", nsPerOp)
	}
	// Output:
	// ns per op = 55357
	// ns per op = 62.7
}

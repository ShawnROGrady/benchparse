// Package benchparse provides utilities for parsing benchmark results.
// Parsed results are split by sub-benchmarks, with support for sub-benchmarks
// with names of the form 'var_name=var_value'
package benchparse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/benchmark/parse"
)

// Benchmark represents a single top-level benchmark and it's results.
type Benchmark struct {
	Name    string
	Results []BenchRes
}

func (b Benchmark) String() string {
	s := make([]string, len(b.Results))
	for i, res := range b.Results {
		s[i] = fmt.Sprintf("%s%s %s", b.Name, res.Inputs, res.Outputs)
	}
	return strings.Join(s, "\n")
}

// GroupResults groups a benchmarks results by a specified set of inputs.
func (b Benchmark) GroupResults(groupBy []string) GroupedResults {
	groupedResults := map[string][]BenchRes{}
	if len(groupBy) == 0 {
		res := make([]BenchRes, len(b.Results))
		copy(res, b.Results)
		groupedResults[""] = res
		return groupedResults
	}
	for _, result := range b.Results {
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

// ParseBenchmarks extracts a list of Benchmarks from testing.B output.
func ParseBenchmarks(r io.Reader) ([]Benchmark, error) {
	var (
		scanner    = bufio.NewScanner(r)
		benchmarks = map[string]Benchmark{}
	)
	for scanner.Scan() {
		parsed, err := parse.ParseLine(scanner.Text())
		if err != nil {
			// TODO: this is what ParseSet does but feels awkward - https://github.com/golang/tools/blob/master/benchmark/parse/parse.go#L114
			continue
		}

		benchName, inputs, err := parseInfo(parsed.Name)
		if err != nil {
			return nil, err
		}
		bench, ok := benchmarks[benchName]
		if !ok {
			bench = Benchmark{Name: benchName, Results: []BenchRes{}}
		}

		outputs := BenchOutputs{
			N:                 parsed.N,
			NsPerOp:           parsed.NsPerOp,
			AllocedBytesPerOp: parsed.AllocedBytesPerOp,
			AllocsPerOp:       parsed.AllocsPerOp,
			MBPerS:            parsed.MBPerS,
			Measured:          parsed.Measured,
		}

		bench.Results = append(bench.Results, BenchRes{
			Inputs:  inputs,
			Outputs: outputs,
		})

		benchmarks[benchName] = bench
	}

	parsedBenchmarks := make([]Benchmark, len(benchmarks))
	i := 0
	for _, v := range benchmarks {
		parsedBenchmarks[i] = v
		i++
	}

	return parsedBenchmarks, nil
}

// used to trim unnecessary trailing chars from benchname
var benchInfoExpr = regexp.MustCompile(`^(Benchmark.+?)(?:\-([0-9]+))?$`)

func parseInfo(s string) (string, BenchInputs, error) {
	maxProcs := 1
	submatches := benchInfoExpr.FindStringSubmatch(s)
	if len(submatches) < 1 {
		return "", BenchInputs{}, fmt.Errorf("info string '%s' didn't match regex", s)
	}
	info := submatches[1]
	// number at the end of benchmark name represents GOMAXPROCS: https://golang.org/src/testing/benchmark.go#L548
	if len(submatches) == 3 {
		var err error
		maxProcs, err = strconv.Atoi(submatches[2])
		if err != nil {
			return "", BenchInputs{}, fmt.Errorf("error parsing maxprocs: %w", err)
		}
	}
	var (
		name      string
		varValues = []BenchVarValue{}
		subs      = []BenchSub{}
		bySub     = strings.Split(info, "/")
	)

	for i, sub := range bySub {
		if i == 0 {
			name = sub
			continue
		}

		split := strings.Split(sub, "=")
		if len(split) == 2 {
			varValues = append(varValues, BenchVarValue{
				Name:     split[0],
				Value:    value(split[1]),
				position: i,
			})
		} else {
			subs = append(subs, BenchSub{
				Name:     sub,
				position: i,
			})
		}
	}

	return name, BenchInputs{VarValues: varValues, Subs: subs, MaxProcs: maxProcs}, nil
}

func value(s string) interface{} {
	convs := []func(str string) (interface{}, error){
		func(str string) (interface{}, error) {
			return strconv.Atoi(str)
		},
		func(str string) (interface{}, error) {
			return strconv.ParseFloat(str, 64)
		},
		func(str string) (interface{}, error) {
			return strconv.ParseBool(str)
		},
	}

	for _, conv := range convs {
		if res, err := conv(s); err == nil {
			return res
		}
	}

	return s
}

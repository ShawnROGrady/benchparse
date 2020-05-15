// Package benchparse provides utilities for parsing benchmark results.
// Parsed results are split by sub-benchmarks, with support for sub-benchmarks
// with names of the form 'var_name=var_value'
package benchparse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/tools/benchmark/parse"
)

// Benchmark represents a single top-level benchmark and it's results.
type Benchmark struct {
	Name    string
	Results BenchResults
}

// String returns the string representation of the benchmark.
// This follows the same format as the testing.B output.
func (b Benchmark) String() string {
	s := make([]string, len(b.Results))
	for i, res := range b.Results {
		s[i] = fmt.Sprintf("%s%s %s", b.Name, res.Inputs, benchOutputsString(res.Outputs))
	}
	return strings.Join(s, "\n")
}

// ParseBenchmarks extracts a list of Benchmarks from testing.B output.
func ParseBenchmarks(r io.Reader) ([]Benchmark, error) {
	return parseBenchmarks(r, func(line string) (string, error) {
		// line already formatted in this case
		return line, nil
	})
}

// benchEvent represents a single testing.B output with the '-json' flag
// enabled.
type benchEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

// ParseBenchmarksFromJSON extracts a list of benchmarks from testing.B output
// with the '-json' flag enabled.
func ParseBenchmarksFromJSON(r io.Reader) ([]Benchmark, error) {
	return parseBenchmarks(r, func(line string) (string, error) {
		var event benchEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return "", fmt.Errorf("unmarshal event: %s", err)
		}
		return event.Output, nil
	})
}

func parseBenchmarks(r io.Reader, fmtLine func(line string) (string, error)) ([]Benchmark, error) {
	var (
		scanner    = bufio.NewScanner(r)
		benchmarks = map[string]Benchmark{}
	)
	for scanner.Scan() {
		line, err := fmtLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		parsed, err := parse.ParseLine(line)
		if err != nil {
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

		outputs := parsedBenchOutputs{*parsed}

		bench.Results = append(bench.Results, BenchRes{
			Inputs:  inputs,
			Outputs: outputs,
		})

		benchmarks[benchName] = bench
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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
	if len(submatches) == 3 && submatches[2] != "" {
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

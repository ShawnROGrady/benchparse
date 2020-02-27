package benchparse

import (
	"testing"

	"golang.org/x/tools/benchmark/parse"
)

var getOutputMeasurementTests = map[string]struct {
	output                       BenchOutputs
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
		output: BenchOutputs{
			N:                   21801,
			nsPerOp:             55357,
			allocatedBytesPerOp: 4321,
			allocsPerOp:         21,
			mBPerS:              0.12,
			measured:            parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp | parse.MBPerS,
		},
		expectedNsPerOp:           55357,
		expectedAllocedBytesPerOp: 4321,
		expectedAllocsPerOp:       21,
		expectedMBPerS:            0.12,
	},
	"benchmem_not_set_with_set_bytes": {
		output: BenchOutputs{
			N:        21801,
			nsPerOp:  55357,
			mBPerS:   0.12,
			measured: parse.NsPerOp | parse.MBPerS,
		},
		expectedNsPerOp:              55357,
		expectedAllocedBytesPerOpErr: ErrNotMeasured,
		expectedAllocsPerOpErr:       ErrNotMeasured,
		expectedMBPerS:               0.12,
	},
	"benchmem_set_but_no_allocs": {
		output: BenchOutputs{
			N:                   21801,
			nsPerOp:             55357,
			allocatedBytesPerOp: 0,
			allocsPerOp:         0,
			measured:            parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp,
		},
		expectedNsPerOp:           55357,
		expectedAllocedBytesPerOp: 0,
		expectedAllocsPerOp:       0,
		expectedMBPerSErr:         ErrNotMeasured,
	},
	"none_set": {
		output:                       BenchOutputs{},
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

func testNsPerOp(t *testing.T, b BenchOutputs, expectedV float64, expectedErr error) {
	t.Helper()
	ns, err := b.NsPerOp()
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

func testAllocedBytesPerOp(t *testing.T, b BenchOutputs, expectedV uint64, expectedErr error) {
	t.Helper()
	v, err := b.AllocedBytesPerOp()
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

func testAllocsPerOp(t *testing.T, b BenchOutputs, expectedV uint64, expectedErr error) {
	t.Helper()
	v, err := b.AllocsPerOp()
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

func testMBPerS(t *testing.T, b BenchOutputs, expectedV float64, expectedErr error) {
	t.Helper()
	v, err := b.MBPerS()
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

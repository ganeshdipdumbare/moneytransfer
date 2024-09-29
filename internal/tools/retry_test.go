package tools

import (
	"testing"
	"time"
)

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name        string
		baseDelay   time.Duration
		maxDelay    time.Duration
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{
			name:        "First attempt",
			baseDelay:   time.Second,
			maxDelay:    time.Minute,
			attempt:     0,
			minExpected: time.Second,
			maxExpected: time.Second + 100*time.Millisecond,
		},
		{
			name:        "Second attempt",
			baseDelay:   time.Second,
			maxDelay:    time.Minute,
			attempt:     1,
			minExpected: 2 * time.Second,
			maxExpected: 2*time.Second + 200*time.Millisecond,
		},
		{
			name:        "Max delay reached",
			baseDelay:   time.Second,
			maxDelay:    10 * time.Second,
			attempt:     10,
			minExpected: 10 * time.Second,
			maxExpected: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBackoff(tt.baseDelay, tt.maxDelay, tt.attempt)
			if result < tt.minExpected || result > tt.maxExpected {
				t.Errorf("CalculateBackoff(%v, %v, %d) = %v, expected between %v and %v",
					tt.baseDelay, tt.maxDelay, tt.attempt, result, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestCalculateBackoffJitter(t *testing.T) {
	baseDelay := time.Second
	maxDelay := time.Minute
	attempt := 2

	results := make([]time.Duration, 100)
	for i := 0; i < 100; i++ {
		results[i] = CalculateBackoff(baseDelay, maxDelay, attempt)
	}

	// Check if all results are not the same (jitter is working)
	allSame := true
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Errorf("All results are the same, jitter is not working")
	}
}

package tools

import (
	"math"
	"math/rand"
	"time"
)

// CalculateBackoff calculates the backoff time based on the attempt number
// and the base delay and maximum delay.
// It uses an exponential backoff strategy with a random jitter to prevent
// thundering herd problem.
func CalculateBackoff(baseDelay, maxDelay time.Duration, attempt int) time.Duration {
	backoff := float64(baseDelay) * math.Pow(2, float64(attempt))
	jitter := rand.Float64() * 0.1 * backoff // 10% jitter
	backoff += jitter

	if backoff > float64(maxDelay) {
		backoff = float64(maxDelay)
	}

	return time.Duration(backoff)
}

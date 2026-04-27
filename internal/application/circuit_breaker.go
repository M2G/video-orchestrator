package application

import "time"

type CircuitBreaker struct {
	failures    int
	threshold   int
	resetAfter  time.Duration
	lastFailure time.Time
}

func NewCircuitBreaker(threshold int, reset time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:  threshold,
		resetAfter: reset,
	}
}

func (c *CircuitBreaker) Allow() bool {

	if c.failures < c.threshold {
		return true
	}

	if time.Since(c.lastFailure) > c.resetAfter {
		c.failures = 0
		return true
	}

	return false
}

func (c *CircuitBreaker) Fail() {
	c.failures++
	c.lastFailure = time.Now()
}

func (c *CircuitBreaker) Success() {
	c.failures = 0
}

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type CircuitBreaker struct {
	mu sync.RWMutex

	failureThreshold uint
	resetTimeout     time.Duration

	failures  uint
	lastError time.Time
	state     State
}

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Option func(*CircuitBreaker)

func WithFailureThreshold(threshold uint) Option {
	return func(cb *CircuitBreaker) {
		cb.failureThreshold = threshold
	}
}

func WithResetTimeout(timeout time.Duration) Option {
	return func(cb *CircuitBreaker) {
		cb.resetTimeout = timeout
	}
}

func NewCircuitBreaker(options ...Option) *CircuitBreaker {
	cb := &CircuitBreaker{
		failureThreshold: 5,
		resetTimeout:     10 * time.Second,
		state:            StateClosed,
	}

	for _, option := range options {
		option(cb)
	}

	return cb
}

func (cb *CircuitBreaker) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cb.AllowRequest() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "service temporarily unavailable",
			})
			c.Abort()
			return
		}

		c.Next()

		// Check if request failed
		if len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		if time.Since(cb.lastError) > cb.resetTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false

	case StateHalfOpen:
		return true

	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.failureThreshold {
			cb.state = StateOpen
			cb.lastError = time.Now()
		}

	case StateHalfOpen:
		cb.state = StateOpen
		cb.lastError = time.Now()

	default:
		// Do nothing in other states
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		cb.state = StateClosed
		cb.failures = 0
		cb.lastError = time.Time{}

	case StateClosed:
		cb.failures = 0

	default:
		// Do nothing in other states
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

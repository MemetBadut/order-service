package resilience

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type State string

const (
	StateClosed   State = "closed"
	StateOpen     State = "open"
	StateHalfOpen State = "half-open"
)

type Config struct {
	MaxFailures  uint
	ResetTimeout time.Duration
	IsFailure    func(error) bool
}

type CircuitBreaker struct {
	mu               sync.Mutex
	state            State
	failures         uint
	openedAt         time.Time
	halfOpenInFlight bool
	maxFailures      uint
	resetTimeout     time.Duration
	isFailure        func(error) bool
}

func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	if cfg.MaxFailures == 0 {
		cfg.MaxFailures = 3
	}

	if cfg.ResetTimeout <= 0 {
		cfg.ResetTimeout = 10 * time.Second
	}

	if cfg.IsFailure == nil {
		cfg.IsFailure = func(err error) bool {
			return err != nil
		}
	}

	return &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  cfg.MaxFailures,
		resetTimeout: cfg.ResetTimeout,
		isFailure:    cfg.IsFailure,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	err := fn()
	cb.afterRequest(err)

	return err
}

func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.openedAt) < cb.resetTimeout {
			return ErrCircuitOpen
		}

		cb.state = StateHalfOpen
		cb.halfOpenInFlight = true
		return nil
	case StateHalfOpen:
		if cb.halfOpenInFlight {
			return ErrCircuitOpen
		}

		cb.halfOpenInFlight = true
	}

	return nil
}

func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.halfOpenInFlight = false

		if cb.shouldCountFailure(err) {
			cb.tripOpenLocked()
			return
		}

		cb.resetLocked()
		return
	}

	if !cb.shouldCountFailure(err) {
		cb.failures = 0
		return
	}

	cb.failures++
	if cb.failures >= cb.maxFailures {
		cb.tripOpenLocked()
	}
}

func (cb *CircuitBreaker) shouldCountFailure(err error) bool {
	if err == nil {
		return false
	}

	return cb.isFailure(err)
}

func (cb *CircuitBreaker) tripOpenLocked() {
	cb.state = StateOpen
	cb.openedAt = time.Now()
	cb.halfOpenInFlight = false
}

func (cb *CircuitBreaker) resetLocked() {
	cb.state = StateClosed
	cb.failures = 0
	cb.openedAt = time.Time{}
	cb.halfOpenInFlight = false
}

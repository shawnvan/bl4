package services

import (
	"sync/atomic"
)

// Fixed atomic operations for batch processing
type AtomicCounter struct {
	value int64
}

func NewAtomicCounter(initial int64) *AtomicCounter {
	return &AtomicCounter{value: initial}
}

func (ac *AtomicCounter) Add(delta int64) int64 {
	return atomic.AddInt64(&ac.value, delta)
}

func (ac *AtomicCounter) Get() int64 {
	return atomic.LoadInt64(&ac.value)
}

func (ac *AtomicCounter) Set(value int64) {
	atomic.StoreInt64(&ac.value, value)
}

func (ac *AtomicCounter) Increment() int64 {
	return ac.Add(1)
}

func (ac *AtomicCounter) Decrement() int64 {
	return ac.Add(-1)
}

// Safe progress tracking using atomic counters
type SafeBatchProgress struct {
	Processed  *AtomicCounter
	Successful *AtomicCounter
	Failed     *AtomicCounter
	Total      int64
}

func NewSafeBatchProgress(total int64) *SafeBatchProgress {
	return &SafeBatchProgress{
		Processed:  NewAtomicCounter(0),
		Successful: NewAtomicCounter(0),
		Failed:     NewAtomicCounter(0),
		Total:      total,
	}
}

func (sbp *SafeBatchProgress) GetProgress() float64 {
	processed := sbp.Processed.Get()
	if sbp.Total <= 0 {
		return 0.0
	}
	return float64(processed) / float64(sbp.Total)
}

func (sbp *SafeBatchProgress) AddSuccess() {
	sbp.Processed.Increment()
	sbp.Successful.Increment()
}

func (sbp *SafeBatchProgress) AddFailure() {
	sbp.Processed.Increment()
	sbp.Failed.Increment()
}
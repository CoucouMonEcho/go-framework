package cache

import "time"

type RetryStrategy interface {
	// Next return retry interval and try again
	Next() (time.Duration, bool)
}

type FixedIntervalRetryStrategy struct {
	Interval time.Duration
	MaxCnt   int
	cnt      int
}

func (f *FixedIntervalRetryStrategy) Next() (time.Duration, bool) {
	if f.cnt >= f.MaxCnt {
		return 0, false
	}
	return f.Interval, true
}

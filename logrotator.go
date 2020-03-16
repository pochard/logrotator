package logrotator

import (
	strftime "github.com/lestrrat/go-strftime"
	"github.com/pkg/errors"
	"time"
)

func NewTimeBasedRotator(pattern string, period time.Duration) (*TimeBasedRotator, error) {
	strfobj, err := strftime.New(pattern)
	if err != nil {
		return nil, errors.Wrap(err, `invalid strftime pattern`)
	}

	var tw TimeBasedRotator
	tw.pattern = strfobj
	tw.period = period.Nanoseconds()
	_, offset := time.Now().Zone()
	tw.timeDiffToUTC = (time.Duration(offset) * time.Second).Nanoseconds()

	return &tw, nil
}

func NewTimeBasedCleaner(pattern string, maxAge time.Duration) (*TimeBasedCleaner, error) {
	var tc TimeBasedCleaner
	tc.pattern = pattern
	if maxAge < 0{
		return nil, errors.New(`maxAge must be greater than 0`)
	}
	tc.maxAge = maxAge

	return &tc, nil
}

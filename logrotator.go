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

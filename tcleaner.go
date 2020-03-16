package logrotator

import (
	"os"
	"path/filepath"
	"time"
	"sync"
)

type TimeBasedCleaner struct {
	pattern string
	maxAge  time.Duration
	mutex   sync.RWMutex
}

func (cleaner *TimeBasedCleaner) Clean() ([]string, error) {
	cleaner.mutex.Lock()
	defer cleaner.mutex.Unlock()

	matches, err := filepath.Glob(cleaner.pattern)
	if err != nil {
		return nil, err
	}
	//names of successfully deleted files
	dfnames := make([]string, 0, len(matches))
	cutoff := time.Now().UnixNano() - cleaner.maxAge.Nanoseconds()
	for _, m := range matches {
		fi, err2 := os.Stat(m)
		if err2 != nil {
			return dfnames, err2
		}

		if fi.ModTime().UnixNano() < cutoff {
			removeErr := os.Remove(m)
			if removeErr != nil {
				return dfnames, err
			}

			dfnames = append(dfnames, m)
		}
	}

	return dfnames, nil
}

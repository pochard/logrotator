package dailyrotator

import (
	strftime "github.com/lestrrat/go-strftime"
	"github.com/pkg/errors"
	"io"
	"os"
	"sync"
	"time"
)

// DrWriter represents a log file that gets
// automatically rotated as you write to it.
type DrWriter struct {
	timeDiffToUTC time.Duration
	curFn         string
	globPattern   string
	mutex         sync.RWMutex
	outFh         *os.File
	pattern       *strftime.Strftime
	rotationTime  time.Duration
}

// New creates a new DrWriter object. A log filename pattern
// must be passed.
func New(pattern string) (*DrWriter, error) {
	globPattern := pattern
	strfobj, err := strftime.New(pattern)
	if err != nil {
		return nil, errors.Wrap(err, `invalid strftime pattern`)
	}

	var rl DrWriter
	rl.globPattern = globPattern
	rl.pattern = strfobj
	rl.rotationTime = 24 * time.Hour
	_, offset := time.Now().Zone()
	rl.timeDiffToUTC = time.Duration(offset) * time.Second
	return &rl, nil
}

func (rl *DrWriter) genFilename() string {
	now := time.Now()
	diff := time.Duration(now.UnixNano()) % rl.rotationTime
	t := now.Add(time.Duration(-1*diff) - rl.timeDiffToUTC)
	return rl.pattern.FormatString(t)
}

// Write satisfies the io.Writer interface. It writes to the
// appropriate file handle that is currently being used.
// If we have reached rotation time, the target file gets
// automatically rotated
func (rl *DrWriter) Write(p []byte) (n int, err error) {
	// Guard against concurrent writes
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	out, err := rl.getTargetWriter()
	if err != nil {
		return 0, errors.Wrap(err, `failed to acquite target io.Writer`)
	}

	return out.Write(p)
}

// must be locked during this operation
func (rl *DrWriter) getTargetWriter() (io.Writer, error) {
	// check if it's 00:00:00
	if rl.curFn != "" {
		t := time.Now()
		h := t.Hour()
		minute := t.Minute()
		s := t.Second()
		if h > 0 || minute > 0 || s > 0 {
			return rl.outFh, nil
		}
	}

	filename := rl.genFilename()

	if rl.curFn == filename {
		return rl.outFh, nil
	}

	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Errorf("failed to open file %s: %s", rl.pattern, err)
	}

	rl.outFh.Close()
	rl.outFh = fh
	rl.curFn = filename

	return fh, nil
}

// Close satisfies the io.Closer interface. You must
// call this method if you performed any writes to
// the object.
func (rl *DrWriter) Close() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if rl.outFh == nil {
		return nil
	}

	rl.outFh.Close()
	rl.outFh = nil
	return nil
}

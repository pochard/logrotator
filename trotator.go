package logrotator

import (
	_"fmt"
	strftime "github.com/lestrrat/go-strftime"
	"github.com/pkg/errors"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

type TimeBasedRotator struct {
	timeDiffToUTC int64
	lastTime      int64
	period        int64
	dirname       string
	filename      string
	mutex         sync.RWMutex
	outFile       *os.File
	pattern       *strftime.Strftime
}

func (tw *TimeBasedRotator) Write(p []byte) (n int, err error) {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	fh, err := tw.getFileHandler()
	if err != nil {
		return 0, errors.Wrap(err, `failed to acquite target io.Writer`)
	}

	if fh == nil {
		return 0, errors.Wrap(err, `target io.Writer is closed`)
	}

	return fh.Write(p)
}

func (tw *TimeBasedRotator) getFileHandler() (io.Writer, error) {
	nowUnixNano := time.Now().UnixNano()
	current := (nowUnixNano - ((nowUnixNano + tw.timeDiffToUTC) % tw.period))
	if (current - tw.lastTime) < tw.period {
		return tw.outFile, nil
	}
	filename := tw.pattern.FormatString(time.Unix(0, current))
	//fmt.Printf("FormatString filename=%s\n", filename)
	if tw.filename == filename {
		return tw.outFile, nil
	}
	//fmt.Printf("Rotate filename=%s %s\n", filename, path.Dir(filename))
	dirname := path.Dir(filename)
	if tw.dirname != dirname {
		os.MkdirAll(path.Dir(filename),os.ModePerm)
		tw.dirname = dirname
	}

	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Errorf("failed to open file %s: %s", tw.pattern, err)
	}

	tw.outFile.Close()
	tw.outFile = fh
	tw.filename = filename
	tw.lastTime = current

	return fh, nil
}

func (tw *TimeBasedRotator) Close() error {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	if tw.outFile != nil {
		tw.outFile.Close()
		tw.outFile = nil
	}
	return nil
}

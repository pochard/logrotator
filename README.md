# logrotator

This is a golang log rotator implementing the io.Writer interface and it's goroutine safe.

Import it in your program as:
```go
      import "github.com/pochard/logrotator"
```

## Write Example
Rotate a log at a specified time.Duration
```go
package main

import (
	"fmt"
	"github.com/pochard/logrotator"
	"io"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	s := "120024,1583893679,183.199.195.151,C03FD5AAB4CA,158379896859008,b_ntms_1,335,0,-,0,68050,102599,0,0,ebe60030-633f-11ea-8bbb-b943b30ce79e,-,36781,pc,192_168_1_23,-,-,-,-,-,1,1505784723f8f21,-,-,-,b_token,-,-,ebe60030-633f-11ea-8bbb-b943b30ce79e-010001,-,-,-,-,183.199.195.151,-,10000001,1,www.pengpengzhou.com,-,183.199.195.151,-,-,-,-,-,-,-,\"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729)\"\n"
	writer, err := logrotator.NewTimeBasedRotator("/data/web_log/click-%Y%m%d-%H%M.log", 1*time.Hour)
	if err != nil {
		fmt.Printf("config local file system logger error. %v\n", err)
	}
	defer writer.Close()
	test(writer, s)
}

func test(writer io.Writer, s string) {
	tSaved := time.Now()
	for i := 0; i != 200000; i++ {
		wg.Add(1)
		go func() {
			_, err := writer.Write([]byte(s))
			if err != nil {
				fmt.Printf("Failed to write to log, %v\n", err)
			}
			wg.Add(-1)
		}()
	}
	wg.Wait()
	fmt.Println(time.Now().Sub(tSaved))
}

```
Time elapse:
1.105725753s

Output files:
click-20200107-0800.log
click-20200107-0900.log
click-20200107-1000.log


## Clean Example: 
Clean all old log files that have not being modified for at the least 7 days. And the job is scheduled at 1:05 in every morning.
```go
package main

import (
	"fmt"
	"github.com/pochard/logrotator"
	"github.com/robfig/cron/v3"
	"net/http"
	"time"
)

func main() {
	cleaner, err := logrotator.NewTimeBasedCleaner("/data/web_log/*.log", 7 * 24 * time.Hour)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	c := cron.New()
	c.AddFunc("5 1 * * *", func() {
		deleted, err := cleaner.Clean()
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		for _, d := range deleted {
			fmt.Printf("%s deleted\n", d)
		}
	})
	c.Start()

	http.ListenAndServe(":8080", nil)
}

```

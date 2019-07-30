package util

import (
	"fmt"
	"time"
)

type customTimer struct {
	name  string
	start time.Time
}

func Start(name string, args ...interface{}) *customTimer {
	return &customTimer{
		name:  fmt.Sprintf(name, args...),
		start: time.Now(),
	}
}

func (t *customTimer) Stop() {
	took := time.Since(t.start)
	fmt.Printf("It took %s to run %s!\n", took, t.name)
}

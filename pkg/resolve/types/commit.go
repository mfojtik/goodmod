package types

import (
	"fmt"
	"time"
)

type Commit struct {
	SHA       string
	Timestamp time.Time
}

func (c Commit) String() string {
	t := c.Timestamp.UTC()
	timestamp := fmt.Sprintf("%d%.2d%.2d%.2d%.2d%.2d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return fmt.Sprintf("v0.0.0-%s-%s", timestamp, c.SHA[0:12])
}

package util

import (
	"time"
)

// TimePtr returns a pointer to a variable holding t.
func TimePtr(t time.Time) *time.Time {
	return &t
}

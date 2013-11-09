package sclib

import (
	"testing"
	"time"
)

func Test_IsExpiredTrue(t *testing.T) {
	oldTime := time.Date(1992, time.November, 24, 5, 0, 0, 0, time.UTC)
	expired := IsExpired(oldTime)
	if !expired {
		t.Error("Expired time reported as not expired")
	}
}

func Test_IsExpiredFalse(t *testing.T) {
	nowTime := time.Now()
	futureTime := nowTime.AddDate(1, 1, 1)
	expired := IsExpired(futureTime)
	if expired {
		t.Error("Unexpired time reported as expired")
	}
}

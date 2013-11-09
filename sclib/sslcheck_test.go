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

func Test_GetExpiredDays(t *testing.T) {
	nowTime := time.Now()
	futureTime := nowTime.AddDate(0, 0, 1)
	days := GetExpireDays(futureTime)
	if days != 1 {
		t.Error("Expired day count: ", days, " != 1")
	}
	otherTime := nowTime.AddDate(0, 0, 10)
	otherDays := GetExpireDays(otherTime)
	if otherDays != 10 {
		t.Error("Expired day count: ", otherDays, " != 10")
	}
}

func Test_GetExpiredDaysNeg(t *testing.T) {
	nowTime := time.Now()
	futureTime := nowTime.AddDate(0, 0, -1)
	days := GetExpireDays(futureTime)
	if days != -1 {
		t.Error("Expired day count: ", days, " != -1")
	}
	otherTime := nowTime.AddDate(0, 0, -10)
	otherDays := GetExpireDays(otherTime)
	if otherDays != -10 {
		t.Error("Expired day count: ", otherDays, " != -10")
	}
}

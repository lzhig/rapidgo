package base

import (
	"time"
)

// GetTodayZeroClockTime function 获取当地时间第二天 0点
func GetTodayZeroClockTime(t *time.Time) time.Time {
	h, m, s := t.Clock()
	return t.Add(-time.Hour*time.Duration(h) - time.Second*time.Duration(m*60+s))
}

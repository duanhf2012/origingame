package util

import (
	"time"
)

// 新的cooldown，多少秒内允许多少次操作
type CoolDown struct {
	BeginTime time.Time
	Count     int
}

// timeCycle 内可以进行 countLimit 次操作
// 比如 TestCoolDown(60*nano, 5) 就是 60秒内允许5次操作
// TestCoolDown(20*nano, 1) 就退化成之前的20秒一次操作那种cd了

func (c *CoolDown) TestCoolDown(timeCycle time.Duration, countLimit int) bool {
	nowTime := time.Now()
	if nowTime.Sub(c.BeginTime) > timeCycle {
		c.BeginTime = nowTime
		c.Count = 0
		return true
	} else {
		if c.Count >= countLimit {
			return false
		}

		c.Count++
		return true
	}
}

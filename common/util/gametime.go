package util

import (
	"github.com/duanhf2012/origin/v2/log"
	"time"
)

const DayOfSecond = int64(86400)
const HourOfSecond = int64(3600)
const MinuteOfSecond = int64(60)
const HourMs = int64(3600 * 1000)
const MinuteMs = int64(60 * 1000)
const DayOfMills = int64(86400000)
const WeekOfMills = int64(86400000 * 7)

const (
	GameLogicTimeRatio    = 3 //游戏逻辑时间倍率，必须是可以被24整除
	GameLogicZeroTimeHour = 4 //游戏逻辑起始点时间(24小时制 0-23)

	GameLogicDailyTime = (24 / GameLogicTimeRatio) * HourOfSecond //游戏逻辑时间每天毫秒
)

// 检测小时是否是有效值
func IsValidHour(hour int) bool {
	return hour >= 0 && hour < 24
}

// 检测分钟是否是有效值
func IsValidMinute(minute int) bool {
	return minute >= 0 && minute < 60
}

// 检测分钟是否是有效值
func IsValidSecond(second int) bool {
	return second >= 0 && second < 60
}

// 游戏逻辑时间结构体
type GameTime struct {
	hour   int //游戏逻辑：小时 0-23
	minute int //游戏逻辑：分钟 0-59
	second int //游戏逻辑：秒  0-59
}

// 获取当前游戏逻辑时间
func NowGameTime() *GameTime {
	unixTime := time.Now().Unix()
	systemElapseTime := int64(unixTime - GetServerFixedTimeSec(unixTime, GameLogicZeroTimeHour, 0, 0))
	if systemElapseTime < 0 {
		systemElapseTime += DayOfSecond
	}
	//游戏逻辑当天已过时间
	logicElapseTime := (systemElapseTime % GameLogicDailyTime) * GameLogicTimeRatio
	return &GameTime{
		hour:   int(logicElapseTime / HourOfSecond),
		minute: int((logicElapseTime % HourOfSecond) / MinuteOfSecond),
		second: int(logicElapseTime % MinuteOfSecond),
	}
}

func NewGameTimeBySecond(gameTimeSecond int64) *GameTime {
	gameTime := NewGameTime(0, 0, 0)
	gameTime.SetGameTimeSec(gameTimeSecond)
	return gameTime
}

func NewGameTime(hour, minute, second int) *GameTime {
	if IsValidHour(hour) == false || IsValidMinute(minute) == false || IsValidSecond(second) == false {
		log.SError("NewGameTime invalid param: ", hour, ", ", minute, ", ", second)
		return nil
	}
	return &GameTime{hour: hour, minute: minute, second: second}
}

func (gt *GameTime) Hour() int {
	return gt.hour
}

func (gt *GameTime) Minute() int {
	return gt.minute
}

func (gt *GameTime) Second() int {
	return gt.second
}

func (gt *GameTime) SetHour(hour int) {
	if IsValidHour(hour) == false {
		log.SError("SetHour error, hour: ", hour)
		return
	}
	gt.hour = hour
}

func (gt *GameTime) SetMinute(minute int) {
	if IsValidMinute(minute) == false {
		log.SError("SetMinute error, minute: ", minute)
		return
	}
	gt.minute = minute
}

func (gt *GameTime) SetSecond(second int) {
	if IsValidSecond(second) == false {
		log.SError("SetSecond error, hour: ", second)
		return
	}
	gt.second = second
}

func (gt *GameTime) SetGameTime(hour, minute, second int) {
	if IsValidHour(hour) == false || IsValidMinute(minute) == false || IsValidSecond(second) == false {
		log.SError("SetGameTime invalid param: ", hour, ", ", minute, ", ", second)
		return
	}
	gt.hour = hour
	gt.minute = minute
	gt.second = second
}

// 获取游戏逻辑时间 当天已过秒数
func (gt *GameTime) GetGameTimeSec() int64 {
	return int64(gt.hour)*HourOfSecond + int64(gt.minute)*MinuteOfSecond + int64(gt.second)
}

// 设置游戏逻辑时间戳 0-24小时的秒杀
func (gt *GameTime) SetGameTimeSec(gameTimeSec int64) {
	if gameTimeSec < 0 || gameTimeSec > DayOfSecond {
		log.SError("SetGameTimeSec error, gameTimeSec: ", gameTimeSec)
		return
	}
	gt.hour = int(gameTimeSec / 24)
	gt.minute = int((gameTimeSec % 24) / 60)
	gt.second = int(gameTimeSec % 60)
}

// 比较时间前后, 返回true:gt小
func (gt *GameTime) Before(b *GameTime) bool {
	return gt.GetGameTimeSec() < b.GetGameTimeSec()
}

// 判断gt是否在[start,end]时间范围
func (gt *GameTime) InRange(start *GameTime, end *GameTime) bool {
	currentGameTimeSec := gt.GetGameTimeSec()
	startGameTimeSec := start.GetGameTimeSec()
	endGameTimeSec := end.GetGameTimeSec()
	if start.Before(end) {
		// start<end 没跨天
		return startGameTimeSec <= currentGameTimeSec && currentGameTimeSec <= endGameTimeSec
	} else {
		//跨天
		if currentGameTimeSec >= startGameTimeSec {
			return true
		} else {
			return currentGameTimeSec <= endGameTimeSec
		}
	}
}

// 获取当天服务器真实系统时间戳
func (gt *GameTime) GetUnixTime() int64 {
	unixTime := time.Now().Unix()
	gameLogicUnixTime := GetServerFixedTimeSec(unixTime, GameLogicZeroTimeHour, 0, 0)
	if gameLogicUnixTime > unixTime {
		gameLogicUnixTime -= int64(DayOfSecond)
	}

	realUnixTime := gameLogicUnixTime
	logicElapseTime := int64(gt.GetGameTimeSec() / GameLogicTimeRatio)
	for i := 0; i < GameLogicTimeRatio; i++ {
		checkUnixTime := gameLogicUnixTime + logicElapseTime + int64(i)*GameLogicDailyTime
		if checkUnixTime <= unixTime {
			realUnixTime = checkUnixTime
		} else {
			break
		}
	}
	return realUnixTime
}

package player

import "time"

func (p *Player) SafeAfterTimer(timerId *uint64, d time.Duration, additionData interface{}, cb func(uint64, interface{})) {
	p.GetPlayerTimer().SafeTimerAfter(timerId, p.GetId(), d, additionData, cb)
}

func (p *Player) SafeTickerTimer(tickerId *uint64, d time.Duration, additionData interface{}, cb func(uint64, interface{})) {
	p.GetPlayerTimer().SafeTimerTicker(tickerId, p.GetId(), d, additionData, cb)
}

func (p *Player) SafeCancelTimer(timerId *uint64) {
	p.GetPlayerTimer().SafeCancelTimer(timerId)
}

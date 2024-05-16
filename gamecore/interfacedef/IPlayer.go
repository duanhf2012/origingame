package interfacedef

import (
	"origingame/common/collect"
	"time"
)

type IPlayerDBCallBack interface {
	OnLoadDBEnd(suc bool)
	OnLoadMultiDBEnd(collectType collect.MultiCollectionType)
}

type IPlayerTimer interface {
	SafeTimerAfter(timerId *uint64, objectId string, d time.Duration, AdditionData interface{}, cb func(uint64, interface{}))
	SafeTimerTicker(tickerId *uint64, objectId string, d time.Duration, AdditionData interface{}, cb func(uint64, interface{}))
	SafeCancelTimer(timerId *uint64)
}

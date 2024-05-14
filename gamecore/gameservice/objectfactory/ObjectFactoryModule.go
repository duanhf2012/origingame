package factory

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/sync"
	global "origingame/common/keyword"
	"origingame/common/performance"
	"origingame/gamecore/gameservice/player"
	"origingame/gamecore/interfacedef"
	"reflect"
	"runtime"

	"time"
)

var playerPool = sync.NewPoolEx(make(chan sync.IPoolData, 1000), func() sync.IPoolData {
	return &player.Player{}
})

type ObjectFactoryModule struct {
	service.Module

	mapPlayer map[string]*player.Player
	Analyzer  *performance.PerformanceAnalyzer
}

func NewGameObjectFactoryModule() *ObjectFactoryModule {
	return &ObjectFactoryModule{}
}

func (m *ObjectFactoryModule) OnInit() error {
	m.mapPlayer = make(map[string]*player.Player, 2048)
	return nil
}

func (m *ObjectFactoryModule) NewPlayer(id string, sender interfacedef.IMsgSender, gsService interfacedef.IGSService) *player.Player {
	player := playerPool.Get().(*player.Player)
	player.Init(id, sender, gsService)
	m.mapPlayer[id] = player

	m.Analyzer.Inc(performance.ObjectNumAnalyzer, int(global.ObjectTypePlayer), performance.NewObjectTotalNumColumn)
	m.Analyzer.Set(performance.ObjectNumAnalyzer, int(global.ObjectTypePlayer), performance.ObjectPoolNum, int64(len(playerPool.C)))
	m.Analyzer.Set(performance.ObjectNumAnalyzer, 0, performance.MapObjectTotalLen, int64(len(m.mapPlayer)))

	return player
}

func (m *ObjectFactoryModule) ReleasePlayer(object *player.Player) {
	log.SInfo("player[", object.GetId(), "] release")

	object.SaveToDB(true)
	delete(m.mapPlayer, object.GetId())
	playerPool.Put(object)
	m.Analyzer.Inc(performance.ObjectNumAnalyzer, int(global.ObjectTypePlayer), performance.ReleaseObjectTotalNumColumn)
	m.Analyzer.Set(performance.ObjectNumAnalyzer, 0, performance.MapObjectTotalLen, int64(len(m.mapPlayer))) //0表示所有类型
}

func (m *ObjectFactoryModule) SafeTimerAfter(timerId *uint64, objectId string, d time.Duration, AdditionData interface{}, cb func(uint64, interface{})) {
	id := *timerId
	m.SafeAfterFunc(timerId, d, AdditionData, func(uint64, interface{}) {
		_, ok := m.mapPlayer[objectId]
		if ok == true {
			cb(id, AdditionData)
		} else {
			funName := runtime.FuncForPC(reflect.ValueOf(cb).Pointer()).Name()
			log.SError("GameObjectFactoryModule TimerAfter obj[", objectId, "] not be found", ",funName", funName)
		}
	})
}

func (m *ObjectFactoryModule) SafeTimerTicker(tickerId *uint64, objectId string, d time.Duration, AdditionData interface{}, cb func(uint64, interface{})) {
	m.SafeNewTicker(tickerId, d, AdditionData, func(uint64, interface{}) {
		_, ok := m.mapPlayer[objectId]
		if ok == true {
			cb(*tickerId, AdditionData)
		} else {
			funName := runtime.FuncForPC(reflect.ValueOf(cb).Pointer()).Name()
			log.SError("GameObjectFactoryModule TimerTicker obj[", objectId, "] not be found", ",funName", funName)
			if m.CancelTimerId(tickerId) == false {
				log.Stack(fmt.Sprint("GameObjectFactoryModule SafeTimerTicker -> CancelTimerId[", tickerId, "] failed"))
			}
		}
	})
}

func (m *ObjectFactoryModule) SafeCancelTimer(timerId *uint64) {
	if m.CancelTimerId(timerId) == false {
		log.Stack(fmt.Sprint("GameObjectFactoryModule SafeCancelTimer[", timerId, "] failed"))
	}
}

func (m *ObjectFactoryModule) GetPlayer(id string) *player.Player {
	return m.mapPlayer[id]
}

func (m *ObjectFactoryModule) GetPlayerNum() int {
	return len(m.mapPlayer)
}

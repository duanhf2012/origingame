package interfacedef

import "origingame/common/collect"

type IPlayerDBCallBack interface {
	OnLoadDBEnd(suc bool)
	OnDelayLoadMCDBEnd(collectType collect.MultiCollectionType)
}

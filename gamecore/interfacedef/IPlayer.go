package interfacedef

import "origingame/common/collect"

type IPlayerDBCallBack interface {
	OnLoadDBEnd(suc bool)
	OnLoadMultiDBEnd(collectType collect.MultiCollectionType)
}

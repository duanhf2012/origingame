package collect

// CollectionType 新增表需要新增类型
type CollectionType = int32

const (
	//定义CollectionType类型
	CTUserInfo CollectionType = iota //userInfo表

	CTMax
)

// 新增玩家多行表
type MultiCollectionType = int32

const (
	MCTUserMail MultiCollectionType = iota //邮件表
	MCTMax
)

// 新增公共多行表
const (
	MPCTMailRecord MultiCollectionType = iota

	MCTShowArea
	MTCRealArea

	MPCTMax
)

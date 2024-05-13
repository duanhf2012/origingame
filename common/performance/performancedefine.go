package performance

type AnalyzerType = int

const (
	MsgAnalyzer          AnalyzerType = 0 //消息统计
	ObjectNumAnalyzer    AnalyzerType = 1 //消息对象数
	ServiceStateAnalyzer AnalyzerType = 2 //服务状态
	GameServerAnalyzer   AnalyzerType = 3 //game server连接数统计
	MaxAnalyzer
)

// Analyzer打印日志等级
const (
	AnalyzerLogLevel1 int = 1
	AnalyzerLogLevel2 int = 2
	AnalyzerLogLevel3 int = 3
	AnalyzerLogLevel4 int = 4
	AnalyzerLogLevel5 int = 5
	MaxAnalyzerLogLevel
)

const (
	GameServerPlayerStatic int = 0
)

// AnalyzerColumn定义
const (
	//MsgAnalyzer
	MsgCostTimeAnalyzer int = 0 //消息数量与耗时分析日志

	//ObjectNumAnalyzer
	NewObjectTotalNumColumn     int = 0 //对象数量分析
	ReleaseObjectTotalNumColumn int = 1
	ObjectPoolNum               int = 2 //对象池的数量
	MapObjectTotalLen           int = 3 //map中对象数量

	//SkillDirectorAnalyzer
	NewDirectorNumColumn     int = 0
	ReleaseDirectorNumColumn int = 1
	MapDirectorNumColumn     int = 2

	//SkillExecutorAnalyzer
	NewSkillExecutorColumn0     int = 0
	ReleaseSkillExecutorColumn1 int = 1

	//SkillExecutorCostAnalyzer
	SkillRealExecutorCostTime    int = 0
	SkillVirtualExecutorCostTime int = 1

	//GameServerAnalyzer
	GameServerPlayerNumColumn       int = 0
	GameServerClientPlayerNumColumn int = 1

	//EventCostAnalyzer
	EventCostTime int = 0

	//BuffAnalyzer
	BuffNewNumColumn     int = 0
	BuffNewReleaseColumn int = 1

	//PassiveAnalyzer
	PassiveNewNumColumn     int = 0
	PassiveNewReleaseColumn int = 1

	//HangUpAnalyzer
	HangUpCostTime int = 0

	//FunnyCaptureAnalyzer
	FunnyCaptureNewNumColumn     int = 0
	FunnyCaptureNewReleaseColumn int = 1

	//ItemAnalyzer
	ItemNewNumColumn  int = 0
	ItemReleaseColumn int = 1

	//MapTypeNumAnalyzer
	MapTypeNewNumColumn      int = 0
	MapTypeReleaseNumColumn  int = 1
	MapTypeLenColumn         int = 2
	MapTypeGlobalCountColumn int = 3

	//MatchCountAnalyzer
	MatchTypeNewNumColumn     int = 0 // 队伍的申请数量
	MatchTypeReleaseNumColumn int = 1 // 队伍释放数量
	MatchTypeLenColumn        int = 2 // 内存中队伍的数量 应该是 = 队伍的申请数量 -队伍释放数量 ，不然就有问题了

	//MatchPlayerCountAnalyzer
	MatchPlayerMatchColumn    int = 0 // 匹配的玩家数量
	MatchPlayerLeaveColumn    int = 1 // 取消匹配/匹配成功的玩家数量
	MatchNowPlayerMatchColumn int = 2 // 内存中正在匹配的数量  应该是 = 匹配的玩家数量 - 取消匹配/匹配成功的玩家数量

	//转服相关
	TransferGsCostTime int = 0 //玩家GS转移耗时
)

// 转移玩家事件定义
const (
	TransferPlayerAllCostTime       int = 0 //玩家转移总耗时
	TransferPlayerMarshalCostTime   int = 1 //玩家转移序列化耗时
	TransferPlayerUnMarshalCostTime int = 2 //玩家转移反序列化耗时
)

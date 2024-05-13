package keyword

import (
	"time"
)

const LogPointTimeLayout = string("20060102")
const EnterTimeCD = int64(3000)
const IntervalTimeMinTime = time.Minute * 5         //5分钟一次资源产出
const IntervalTimeMinInt64 = int64(time.Minute * 5) //5分钟一次资源产出
const SwitchPlateCD = int64(time.Minute * 5)
const RefreshLoginCDMap = 10 * time.Minute

// TODO 修改昵称CD为0，临时功能
const ModifyUserNickNameReqCD = int64(0) //const ModifyUserNickNameReqCD = int64(1000) //修改昵称请求CD

// 万分比
const PercentValue = 10000
const PercentValueM2 = 100000000
const PercentValueM3 = 1000000000000
const MaxCtDamageRatio = 500000000
const MinCtDamageRatio = 125000000
const BaseCtDamageRatio = 150000000

// 宠物资质洗练6品升7品需要的资质在爆区间的数量
const NeedFirePotentialCount = int32(2)

// 默认总权重
const DefaultTotalWeight = int32(10000)

// 默认重生宠物数量
const DefaultRebornPetCount = int(99)

const MaxMedalDataLen = 256    //勋章数据最大数量256*8=2048
const NoAfterMedalCount = 30   //没有后置勋章的数量,用于分配空间
const HaveAfterMedalCount = 30 //有后置勋章的数量,用于分配空间
const MaxMsgObjectiveDataLen = 3000
const ImportNickNameSeed = 1666150076479883600 //随机种子

const MaxRankAchieveDataLen = 256 //排行成就数据最大数量 256*8 = 2045
const MaxRankAchieveId = 256 * 8  //最大排行成就ID

const MaxCookbookDataLen = 25 //最大食谱ID

// 根据勋章ID,判断勋章是否有后置勋章
func CheckMedalHasAfterMedal(medalId uint32) bool {
	if medalId <= MaxMedalDataLen*BitLen {
		return false
	}
	return true
}

const BitLen = 8

const PreRandomPool = 5

// 回复血量时间间隔
const HpReplyTime = int64(5000 * time.Millisecond)

// 执行器最大执行次数
const ExecutorMaxDoCount = 10

// 技能误差CD
const SkillDeltaCD = int64(time.Second)

// 技能释放次数误差时间(ms)
const SkillOverlayDeltaTime = 500

// 技能持续时长误差时间
const SkillDurDeltaTime = int64(100 * time.Millisecond)

// 连携技公共CD
const LinkSkillPublicCD = int64(1000*time.Millisecond) - SkillDeltaCD

// 默认模型半径
const DefaultRadius = int32(1500)

// 预设阵容最大数量
const BattleArrayMaxNum = int32(8)

// 统计战力的最大阵容下标
const MaxBattleArrayIndex = int(2)

// 攻击方计算闪避时最低命中概率——万分比
const AttackDodgeMinHitPro = int64(2500)

// 命中调控参数——万分比
const AttackedHitRegulation = int64(5000)
const AttackedHitAddRegulation = int64(3000)

// 命中调控参数——万分比
const AttackedCtRegulation = int64(6000)
const AttackedCtAddRegulation = int64(500)

// 穿透调控参数——万分比
const AttackPierceRegulation = int64(6000)

// 最小穿透系数——万分比
const AttackMinPiercePro = int64(2500)

// 攻击方最大暴击概率
const AttackMaxCtPro = int64(9500)

// pvp伤害衰减最小值
const MinPvpDamageRatio = int64(100)

// 释放添加SP值，内置CD
const AddCastSpCd = int64(time.Millisecond * 600)

// 技能释放条件验证开关
const SkillCastCheckSwith = true

// 怪物返回出生点时间
const MonsterReturnBronTime = int64(2000 * time.Millisecond)

// 挂机产出周期限制,暂定为3天
const HangUpMaxTime = int64(259200000)

// 挂机随机产出循环最大次数限定
const HangUpBaseRandMaxLoopCount = int32(30)

// 最小伤害万分比
const MinDamageRatio = int64(500)

// 最小暴击伤害衰减
const MinCtDamageSubRatio = int64(500)

// 挂机周期内任务掉落次数限定
const HangUpDropTaskMaxNum = int32(24)

// 属性被动最多种类数
const MaxAttributeCount = 10

// 食物加成属性类型种类数量
const MaxFoodAdditionAttributeCount = 10

// 最大刷怪数
const MaxBrushMonsterCount = 50
const MaxBlackBrushMonsterCount = 1

// 地图最大对象数量
const MaxMapObjectCount = 55
const MaxMapMonsterCount = 30
const MaxMapTrapCount = 20

// 技能目标选择数量
const MaxContextDataLen = 12
const MaxTargetPosition = 12

// 天赋方案套数
const PreEquipmentItemSkillCount = 3
const PreEquipmentItemTalentCount = 10
const PreEquipmentItemPassiveCount = 10
const PreEquipmentItemAttrCount = 20

// 技能数量
const ObjMaxSkillCount = 20
const SkillMaxDamageCount = 5

// 技能最大被动数
const SkillMaxPassiveCount = 30

// 攻速最小值
const MinAttackSpeedTime = 200 * time.Millisecond

// buff附着最小概率
const MinBuffAddRatio = 100

// 地图中最大玩家数
const MapMaxPlayer = 3

// 解锁功能的功能ID
type EFuncUnLockIdType = int32

const (
	EFuncUnLockScheduleSpeedUp       EFuncUnLockIdType = 1  //1_派遣加速功能
	EFuncUnLockMedal                 EFuncUnLockIdType = 2  //2_勋章功能
	EFuncUnLockHangUp                EFuncUnLockIdType = 3  //3_挂机功能
	EFuncUnLockRankMainLevel         EFuncUnLockIdType = 4  //4_主线关卡进度排行榜功能
	EFuncUnLockRankBattleCombatScore EFuncUnLockIdType = 5  //5_阵容战力排行榜功能
	EFuncUnLockRankPlayerLevel       EFuncUnLockIdType = 6  //6_玩家等级排行榜功能
	EFuncUnLockRankTrialTowerCourage EFuncUnLockIdType = 7  //7_初始试炼塔排行榜功能 (勇者塔)
	EFuncUnLockGuild                 EFuncUnLockIdType = 8  //8_公会
	EFuncUnLockChat                  EFuncUnLockIdType = 9  //9_聊天系统
	EFuncUnlockHomeTask              EFuncUnLockIdType = 10 //10_神树目标功能
	EFuncUnlockOnlineAward           EFuncUnLockIdType = 11 //11_在线时长奖励
	EFuncUnlockCooking               EFuncUnLockIdType = 12 //12_烹饪
	EFuncUnlockBloodlineEquip        EFuncUnLockIdType = 13 //13_因子
	EFuncUnlockSpar                  EFuncUnLockIdType = 14 //14_能力系统
	EFuncUnlockPlayerSparPos0        EFuncUnLockIdType = 15 //15_玩家晶石槽位0
	EFuncUnlockPlayerSparPos1        EFuncUnLockIdType = 16 //16_玩家晶石槽位1
	EFuncUnlockPetPotentialRefine    EFuncUnLockIdType = 17 //17_宠物资质洗练 目前仅用于宠物幻能等级解锁
	EFuncUnlockPetPropertyRefine     EFuncUnLockIdType = 18 //18_宠物洗词条 目前仅用于宠物幻能等级解锁

	EFuncUnLockPlaceholder
	EFuncUnLockMax = EFuncUnLockPlaceholder + 1
)

// 挂机掉落是否是首次领取状态
type HangUpFirstAwardStateEnum int8

const (
	HangUpFirstInit         HangUpFirstAwardStateEnum = 0 //未领取首次奖励
	HangUpFirstInitComplete HangUpFirstAwardStateEnum = 1 //已领取首次奖励
)

// 挂机性能监控
type EHangUpAnalyzerType = int

const (
	MapLevelChangeType EHangUpAnalyzerType = 1 //主线关卡改变
	StuffAddType       EHangUpAnalyzerType = 2 //增加摆件
	GetHangUpAwardType EHangUpAnalyzerType = 3 //领取挂机奖励
)

// 抓宠相关
const CaptureMaxBattlePetCount = 10

// 派遣使用食物上限
const ScheduleMaxFoodCount = 5
const PreScheduleAdditionLen = 5

// 最大世界板块数量
const MaxWorldMapCount = 16
const PreItemCostCount = 10
const PreItemCfgCount = 100

// 地图控制器黑板特殊标记
const ThiefMaxLen = 5
const ShowAwardMaxLen = 12
const SubHpDropAwardLen = 20
const CollectAwardMaxLen = 20 //收集奖励
const RemoveItemMaxLen = 20   //删除道具长度
const AddScheduleMaxLen = 5   //添加的派遣事件个数限制

// 稀有宠物品质
const PetRareQuality = 6

// 目标相关
const PreConditionDataLen = 10
const PreHomeTaskCount = 5
const PreHomeTaskChapterCount = 5
const PreObjectiveCount = 20

// 舰船装备相关
const PreWarshipPropertyCount = 5
const PreWarshipPassiveCount = 5

// 生产加工相关
const SingleMachineMaxCount = 50   //单次生产最大值
const PreAdditionFormulaCount = 10 //预处理值——额外配方数量

type SpecialTagType = int64

/*
	一个地图实例上，上一个 MapDataContext map[int64]map[int64]int64 结构，双层map结构，可以用来存储数据
	形如 	[key1][key2]value
	为了避免名字空间的污染，大家的key冲突了，所以整理一下现在服务器中使用的规整
	1. key1 > 0	   ->   执行器级别，这是key1一定是executorId(key, value)，对应 Get/SetExecutorBlackBoard 这两个接口退化成 [executorId][key]value，只能在本executor中访问
	2. key1 = 0    ->   地图级别，对应 Get/SetMapBlackBoard(key, value) 这两个接口退化成 [key]value，是可以且一定会跨executor访问的，如果不需要跨executor访问，应该使用规则1
	3. key1 < 0    ->   地图级别map，对应 Get/SetContext(key1, key2, value) 这种情况下，是原生的双层map结构，是可以跨但应该不会跨executor访问的
*/

const MonsterAllFlag = int64(0)

const (
	// 规则1 编辑器级别
	// 这里是每个控制器自己的executorId

	// 规则2 地图级别
	MapExecutorBase SpecialTagType = 0

	RefreshMonsterGroupState SpecialTagType = -1000
	MapAdditionAward         SpecialTagType = -1001 //关卡额外奖励
	BlackBrushInfo           SpecialTagType = -1002 //刷怪数据
	ScheduleLogicBaseInfo    SpecialTagType = -1003 //触发关卡的事件数据
	DeathPetNpcState         SpecialTagType = -1004 //宠物NPC死亡计数
	ArriveTargetPetNpcState  SpecialTagType = -1005 //宠物NPC到达目标点计数
	CleanTreasureMapCount    SpecialTagType = -1006 //擦拭宝图次数
	RotateTreasureMap        SpecialTagType = -1007 //旋转宝图角度

	ExcludeGroupFlag          SpecialTagType = -1008
	LinkSkillUseCountFlag     SpecialTagType = -1010
	DamageFlag                SpecialTagType = -1011
	ExecutorStartTimer        SpecialTagType = -1012 //开始时间
	ShipBattleFormationIdx    SpecialTagType = -1013 //舰船战斗阵型计数
	DerivationTunTunRatInfo   SpecialTagType = -1014 //屯屯鼠衍生事件数据
	TunTunRatThrowNumCount    SpecialTagType = -1015 //屯屯鼠丢坚果数量
	TunTunRatAwardDropCount   SpecialTagType = -1016 //屯屯鼠奖励掉落次数
	TunTunRatAwardDropInfo    SpecialTagType = -1017 //屯屯鼠奖励掉落信息
	LevelMapTalkId            SpecialTagType = -1018 //关卡对话Id
	CapturedPetId             SpecialTagType = -1019 //抓宠宠物Id
	GradeLevel                SpecialTagType = -1020 //评级
	GradeLevelMonsterCount    SpecialTagType = -1021 //评级-击杀怪物数量
	HasBrushMonsterCount      SpecialTagType = -1022 //已经刷怪数
	MapShowAward              SpecialTagType = -1023 //用于结算展示的奖励
	HasBrushInteractiveCount  SpecialTagType = -1024 //已经刷的交互物数量
	HasDoInteractiveCount     SpecialTagType = -1025 //已经交互过的交互物数量
	MapCollectAward           SpecialTagType = -1026 //关卡收集的奖励
	HasBrushMonsterBatchCount SpecialTagType = -1027 //已经刷怪波数
	SceneObjectStatusInfo     SpecialTagType = -1028 //场景对象状态信息
	ForCountFreqFlag          SpecialTagType = -1029 //循环执行器执行频率信息
	WaringMonsterCD           SpecialTagType = -1030 //警告怪cd
	ExecutorSleepEndTime      SpecialTagType = -1031 //执行器延时结束时间
	ExecutorSleepLeftTime     SpecialTagType = -1032 //执行器延时剩余时间
	ExecutorPauseTime         SpecialTagType = -1033 //执行器暂停时间
	IntervalBrushData         SpecialTagType = -1034 //间隔时间刷怪数据
	PetSparUnlockSlot         SpecialTagType = -1035 //宠物能力闯关解锁槽位
	PetSparActive             SpecialTagType = -1036 //宠物能力晶石激活
	MapCollectRemoveItem      SpecialTagType = -1037 //关卡收集的删除道具(剧情对话)
	MapCollectAddSchedule     SpecialTagType = -1038 //关卡收集的添加派遣事件记录(剧情对话)
	CapturedPetInfo           SpecialTagType = -1039 //抓宠的宠物信息
	SpiritDojoMapInfo         SpecialTagType = -1040 //御灵道场关卡信息
	SpiritChallengeMapInfo    SpecialTagType = -1041 //御灵挑战关卡信息
)

// 定时执行器事件类型
type SleepEventType = int8

const (
	SleepEventTypeFinish SleepEventType = 0 //完成
	SleepEventTypePause  SleepEventType = 1 //暂停
	SleepEventTypeResume SleepEventType = 2 //恢复
)

// 任务池类型
type TaskPoolTypeEnum int8

const (
	TaskPoolTypeSingle     TaskPoolTypeEnum = 1 //一次性任务
	TaskPoolTypeRepetition TaskPoolTypeEnum = 2 //重复任务
)

// 地面类型
type EGroundType = int8

const (
	GroundTypeBanWalk EGroundType = 1 //1_禁止行走
	GroundBanSkill    EGroundType = 2 //2_禁止技能穿越
	GroundBanJump     EGroundType = 4 //4_禁止跳跃
)

// 动态阻挡-可通过对象类型
type EBlockFilterType = int8

const (
	BlockFilterTypeMonsterWalkable EBlockFilterType = 1 //1_怪物通过
	BlockFilterTypePlayerWalkable  EBlockFilterType = 2 //2_玩家通过
)

// 操作动态阻挡的类型
type ESetDynamicBlockType = int8

const (
	SetDynamicBlockTypeAdd ESetDynamicBlockType = 1 //添加动态阻挡
	SetDynamicBlockTypeDel ESetDynamicBlockType = 2 //删除动态阻挡
)

// 离开地图原因
type ELeaveAoiReason = int8

const (
	ELeaveNormal  ELeaveAoiReason = 0 //正常，不做特殊处理——表示移动到视野范围外
	ELeaveDestroy ELeaveAoiReason = 1 //销毁
	ELeaveDeath   ELeaveAoiReason = 2 //死亡离开地图
)

// 技能触发类型
type ESkillTriggerType = int32

const (
	TriggerTargetNone ESkillTriggerType = 0 //0_无目标选中
	BasePointRange    ESkillTriggerType = 1 //1_以基准点为中心的范围
	TargetPointRange  ESkillTriggerType = 2 //2_以目标点为中心的范围
	OwnerAura         ESkillTriggerType = 3 //3_自身光环
	TargetAura        ESkillTriggerType = 4 //4_目标光环
	StandFan          ESkillTriggerType = 5 //5_扇形持续伤害
	StandLine         ESkillTriggerType = 6 //6_直线持续伤害
)

// 技能目标类型
type ESkillTargetType = int32

const (
	TargetTypeNone      ESkillTargetType = 0   //0_无目标
	TargetTypeSelf      ESkillTargetType = 1   //1_自身目标
	TargetTypeFriend    ESkillTargetType = 2   //2_友军目标
	TargetTypeEnemy     ESkillTargetType = 4   //4_敌方目标
	TargetTypeMaster    ESkillTargetType = 8   //8_主人
	TargetTypePosition  ESkillTargetType = 16  //16_指定位置
	TargetTypeRandomPos ESkillTargetType = 32  //32_随机位置
	TargetTypePet       ESkillTargetType = 64  //64_宠物
	TargetTypePlayer    ESkillTargetType = 128 //128_玩家
)

// 技能目标优先级
type ESkillTargetPriority = int32

const (
	PriorityDefault    ESkillTargetPriority = 0  //0_默认优先级
	PriorityNearest    ESkillTargetPriority = 1  //1_距离最近
	PriorityFarthest   ESkillTargetPriority = 2  //2_距离最远
	PriorityMoreHP     ESkillTargetPriority = 3  //3_血量百分比最高
	PriorityLessHP     ESkillTargetPriority = 4  //4_血量百分比最低
	PriorityMoreEnmity ESkillTargetPriority = 5  //5_仇恨最高
	PriorityRandom     ESkillTargetPriority = 6  //6_随机选择
	PriorityAll        ESkillTargetPriority = 7  //7_全体
	PriorityTank       ESkillTargetPriority = 8  //8_技能优先坦克
	PriorityWarrior    ESkillTargetPriority = 9  //9_技能优先战士
	PriorityAssassin   ESkillTargetPriority = 10 //10_技能优先刺客
	PriorityADC        ESkillTargetPriority = 11 //11_技能优先射手
	PrioritySUP        ESkillTargetPriority = 12 //12_技能优先辅助
)

// 是否是优先级职业
func IsPriorityCareer(priorityType ESkillTargetPriority, careerType ECareerType) bool {
	switch priorityType {
	case PriorityTank:
		return careerType == ECareerTypeTank
	case PriorityWarrior:
		return careerType == ECareerTypeWarrior
	case PriorityAssassin:
		return careerType == ECareerTypeAssassin
	case PriorityADC:
		return careerType == ECareerTypeADC
	case PrioritySUP:
		return careerType == ECareerTypeSUP
	}
	return false
}

// 技能流程类型,用于控制技能执行点的顺序方法
type ESkillProcessType = int32

const (
	SkillProcessLOOP ESkillProcessType = 0 //0_循环
	SkillProcessLINE ESkillProcessType = 1 //1_线型
)

// 地图中对象类型 0-15
type EObjectType = int32
type ECampType = int32

// 目前对象类型降营只有
const (
	CampTypeMiddle  ECampType = 0
	CampTypePlayer  ECampType = 1 << ObjectTypePlayer
	CampTypeMonster ECampType = 1 << ObjectTypeMonster
	CampTypePeace             = CampTypePlayer + CampTypeMonster
	CampTypeMax
)

const (
	ObjectTypeInvalid EObjectType = 0 //无效类型
	//以下为可视对象服务器需要关注
	ObjectTypeVisibilityBegin EObjectType = 1 //可视开始
	ObjectTypePlayer          EObjectType = 1 //玩家对象
	ObjectTypeMonster         EObjectType = 2 //怪物对象
	ObjectTypePet             EObjectType = 3 //宠物对象
	ObjectTypeBlock           EObjectType = 4 //动态阻档对象
	ObjectTypeTrap            EObjectType = 5 //陷阱对象
	ObjectTypeFightPlayer     EObjectType = 6 //离线玩家对象
	ObjectTypeFightPet        EObjectType = 7 //离线宠物对象

	ObjectTypePlaceholder
	ObjectTypeVisibilityEnd = ObjectTypePlaceholder + 1
	ObjectTypeSelf          = ObjectTypeVisibilityEnd + 1 //自身 实际不可视

	// 以下服务器不关注的对象
	ObjectTypePathNode              EObjectType = 11 //路径节点
	ObjectTypeTriggerArea           EObjectType = 12 //触发区域
	ObjectTypeBrushMonster          EObjectType = 13 //刷怪组点,目前刷pet npc点
	ObjectTypeRoom                  EObjectType = 14 //房间
	ObjectTypeInteractObj           EObjectType = 15 //可交互物件
	ObjectTypeBrushMonsterFormation EObjectType = 16 //刷怪按阵型
	ObjectTypeDecoAnimal            EObjectType = 17 //环境生物

	ObjectTypePlaceholder2
	ObjectTypeMax = ObjectTypePlaceholder2 + 1
	//不允许超过31,<=31
)

// 怪物类型
type EMonsterType = int

const (
	MonsterTypeNo        EMonsterType = 0  //不是怪物
	MonsterTypeNormal    EMonsterType = 1  //1_普通怪物
	MonsterTypeElite     EMonsterType = 2  //2_精英怪物
	MonsterTypeBoss      EMonsterType = 3  //3_Boss怪物
	MonsterTypePet       EMonsterType = 4  //4_宠物怪物
	MonsterTypeLord      EMonsterType = 5  //5_领主怪物
	MonsterTypeNpc       EMonsterType = 6  //6_NPC怪物
	MonsterTypeEnergy    EMonsterType = 7  //7_充能怪物
	MonsterTypeTTRat     EMonsterType = 8  //8_屯屯鼠怪物
	MonsterTypeObscured  EMonsterType = 9  //9_遮挡怪物
	MonsterTypeDandelion EMonsterType = 10 //10_蒲公英怪物
	MonsterTypeGuard     EMonsterType = 11 //11_守护怪物(帐篷)
	MonsterTypeFissure   EMonsterType = 12 //12_裂隙怪物
)

// 地图类型
type EMapType = int32

const (
	MapTypeInvalid         EMapType = 0  //无效地图类型
	MapTypeLevel           EMapType = 1  //关卡地图
	MapTypeWorld           EMapType = 2  //世界地图
	MapTypePagodaRelics    EMapType = 3  //遗迹地图
	MapTypePagodaChallenge EMapType = 4  //挑战地图
	MapTypeLevelTask       EMapType = 5  //关卡任务地图
	MapTypeTrialTower      EMapType = 7  //试练塔地图
	MapTypeThief           EMapType = 8  //盗贼地图
	MapTypeMainLine        EMapType = 9  //主线地图
	MapTypeInvadeDefend    EMapType = 10 //舰船入侵战斗地图
	MapTypeNihility        EMapType = 12 //虚无之地地图
	MapTypeGrade           EMapType = 13 //评级地图
	MapTypeDailyCopy       EMapType = 14 //日常副本地图
	MapTypePvp             EMapType = 15 //竞技场地图
	MapTypeMatch           EMapType = 16 //组队地图
	MapTypeSpar            EMapType = 17 //能力地图
	MapTypeSingleCopy      EMapType = 18 //单人副本地图
	MapTypeSpiritDojo      EMapType = 19 //御灵道场地图
	MapTypeSpiritChallenge EMapType = 20 //御灵挑战地图

	MapTypeHolder
	MaxMapType = MapTypeHolder + 1
)

// 技能sp释放类型
type ESkillCastSpType int8

const (
	SkillCastSpTypeNoConnection ESkillCastSpType = 0 //0_与sp无关
	SkillCastSpTypeFull         ESkillCastSpType = 1 //1_充满释放
	SkillCastSpTypeConsume      ESkillCastSpType = 2 //2_sp值释放 释放完成后，sp不清空
	SkillCastSpTypeStage        ESkillCastSpType = 3 //3_阶段释放
	SkillCastSpTypeBeyond       ESkillCastSpType = 4 //4_超出释放
)

// 技能减CD类型
type ESkillCDType int8

const (
	SkillCDTypeEvent   ESkillCDType = 0 //0_事件类型
	SkillCDTypeMap     ESkillCDType = 1 //1_地图
	SkillCDTypeEquip   ESkillCDType = 2 //2_装备
	SkillCDTypePassive ESkillCDType = 3 //3_被动
	SkillCDTypeMax     ESkillCDType = 4 //个数
)

// 道具类型
type ItemTypeEnum = int8

// 6_角色属性,客户端使用,服务器不用
const (
	ItemTypeSundryInvalid    ItemTypeEnum = 0  //0_无效
	ItemTypeSundry           ItemTypeEnum = 1  //1_杂货
	ItemTypeConsumable       ItemTypeEnum = 2  //2_消耗品
	ItemTypeEquipment        ItemTypeEnum = 3  //3_装备
	ItemTypeFragment         ItemTypeEnum = 4  //4_碎片
	ItemTypeArtifact         ItemTypeEnum = 5  //5_神器
	ItemTypePetGift          ItemTypeEnum = 7  //7_宠物礼物
	ItemTypeElementsChip     ItemTypeEnum = 8  //8_元素碎片
	ItemTypePotentialExp     ItemTypeEnum = 9  //9_宠物资质经验
	ItemTypePassiveExp       ItemTypeEnum = 10 //10_宠物被动经验
	ItemTypeBloodlineCrystal ItemTypeEnum = 11 //11_血统结晶,提升血统活力
	ItemTypeBloodlineEquip   ItemTypeEnum = 12 //12_血统因子
	ItemTypeFood             ItemTypeEnum = 13 //13_食物
	ItemTypePetMood          ItemTypeEnum = 15 //15_宠物心情
	ItemTypeHorn             ItemTypeEnum = 16 //16_号角
	ItemTypePetFragment      ItemTypeEnum = 17 //17_宠物碎片
	ItemTypeFoodMaterial     ItemTypeEnum = 18 //18_食材
	ItemTypeSpice            ItemTypeEnum = 19 //19_香料
	ItemTypeStoryLine        ItemTypeEnum = 20 //20_剧情道具
	ItemTypeSparSource       ItemTypeEnum = 21 //21_能力原晶

	ItemTypeWarshipStart    ItemTypeEnum = 41 //舰船装备开始
	ItemTypeBuildingAreaEq                    //41_建筑区装备
	ItemTypeTreeAreaEq      ItemTypeEnum = 42 //42_神树区装备
	ItemTypeControlAreaEq   ItemTypeEnum = 43 //43_主控区装备
	ItemTypeDynamicAreaEq   ItemTypeEnum = 44 //44_动力区装备
	ItemTypeCabinAreaEq     ItemTypeEnum = 45 //45_船舱区装备
	ItemTypeMastAreaEq      ItemTypeEnum = 46 //46_桅杆区装备
	ItemTypeSplintAreaEq    ItemTypeEnum = 47 //47_甲板区装备
	ItemTypeFireAreaFirEq   ItemTypeEnum = 48 //48_宝船火力装备1
	ItemTypeFireAreaSecEq   ItemTypeEnum = 49 //49_宝船火力装备2
	ItemTypeFireAreaThdEq   ItemTypeEnum = 50 //50_宝船火力装备3
	ItemTypeFireAreaFourEq  ItemTypeEnum = 51 //51_宝船火力装备4
	ItemTypeFastenAreaFirEq ItemTypeEnum = 52 //52_宝船加固装备1
	ItemTypeFastenAreaSecEq ItemTypeEnum = 53 //53_宝船加固装备2
	ItemTypeFastenAreaThdEq ItemTypeEnum = 54 //54_宝船加固装备3
	ItemTypeWarshipEnd
)

// 舰船装备道具数量
const WarshipEquipCount = 14

// 判断道具是不是宠物食物类
func CheckIsPetIntimateFoodByItemType(itemType ItemTypeEnum, foodType EFoodTypeEnum) bool {
	return itemType == ItemTypeFood && (foodType == EFoodTypeIntimate)
}

// 判断道具是不是舰船装备类
func CheckIsWarshipEquipItemType(itemType ItemTypeEnum) bool {
	if itemType >= ItemTypeWarshipStart && itemType <= ItemTypeWarshipEnd {
		return true
	}
	return false
}

// 食物类型
type EFoodTypeEnum = int8

const (
	EFoodTypeCommon   EFoodTypeEnum = 0 //0_通用食物类型
	EFoodTypeSchedule EFoodTypeEnum = 1 //1_派遣食物类型
	EFoodTypeIntimate EFoodTypeEnum = 2 //2_亲密度食物类型
	EFoodTypeMood     EFoodTypeEnum = 3 //3_心情食物类型
	EFoodTypeAttr     EFoodTypeEnum = 4 //4_属性食物类型
	EFoodTypeBrawn    EFoodTypeEnum = 5 //5_体力食物类型
	EFoodTypeFeast    EFoodTypeEnum = 6 //6_宴席食物类型
	EFoodTypeBuff     EFoodTypeEnum = 7 //7_BUFF食物类型

	EFoodTypeEnd
	EFoodTypeMax = EFoodTypeEnd + 1
)

// 道具功能类型
type ItemEffectType = int32

const (
	ItemEffectTypeInvalid                     ItemEffectType = 0  //0_无效
	ItemEffectTypeExp                         ItemEffectType = 1  //1_经验丹
	ItemEffectTypeLevel                       ItemEffectType = 2  //2_等级丹
	ItemEffectTypeDrop                        ItemEffectType = 3  //3_掉落
	ItemEffectTypeUnlockFormula               ItemEffectType = 4  //4_解锁配方
	ItemEffectTypePetExp                      ItemEffectType = 5  //5_宠物经验,不在背包使用
	ItemEffectTypePetIntimate                 ItemEffectType = 6  //6_宠物亲密度or抓宠概率or获得性格万分比,不在背包使用
	ItemEffectTypeUnlockStuff                 ItemEffectType = 7  //7_解锁摆件
	ItemEffectTypePotentialExp                ItemEffectType = 8  //8_宠物资质经验
	ItemEffectTypePassiveExp                  ItemEffectType = 9  //9_宠物被动经验
	ItemEffectTypeBloodlineExp                ItemEffectType = 10 //10_血统活力
	ItemEffectTypeCapTureQuality              ItemEffectType = 11 //11_提升抓宠品质,不在背包使用
	ItemEffectTypeSchedule                    ItemEffectType = 12 //12_派遣加成,不在背包使用
	ItemEffectTypeMood                        ItemEffectType = 13 //13_宠物心情
	ItemEffectTypeNihilityRestoreHP           ItemEffectType = 14 //14_虚无之地血量恢复
	ItemEffectTypePetIntimateOrCapTureQuality ItemEffectType = 15 //15_宠物亲密度or提升抓宠品质or获得性格万分比,不在背包使用
	ItemEffectTypeBrawn                       ItemEffectType = 16 //16_体力
)

// 装备位置
type EquipmentPos = int8

const (
	EquipmentPosFirst EquipmentPos = 0 //开始位置

	EquipmentPosArms EquipmentPos = 0 //0_武器

	EquipmentPosBegin EquipmentPos = 1 //常规槽位开始

	EquipmentPosClothes        EquipmentPos = 1  //1_衣服
	EquipmentPosCap            EquipmentPos = 2  //2_帽子
	EquipmentPosBelt           EquipmentPos = 3  //3_腰带
	EquipmentPosRing           EquipmentPos = 4  //4_戒指
	EquipmentPosNecklace       EquipmentPos = 5  //5_项链
	EquipmentPosPants          EquipmentPos = 6  //6_裤子
	EquipmentPosShoes          EquipmentPos = 7  //7_鞋子
	EquipmentPosHandProtection EquipmentPos = 8  //8_护手
	EquipmentPosEarrings       EquipmentPos = 9  //9_耳环
	EquipmentPosCloak          EquipmentPos = 10 //10_披风

	EquipmentPosEnd
	EquipmentPosMax = EquipmentPosEnd + 1
)

// 是否是合法的装备位置
func IsValidEquipmentPos(pos EquipmentPos) bool {
	return pos >= EquipmentPosBegin && pos <= EquipmentPosEnd
}

// 宠物上阵位置
type PetEquipmentPos = int8

const (
	PetEquipmentPosNull   PetEquipmentPos = -1 //未装备
	PetEquipmentPosFirst  PetEquipmentPos = 0  //0_位置1 先锋位,默认召唤出战
	PetEquipmentPosSecond PetEquipmentPos = 1  //1_位置2 先锋位,默认召唤出战

	PetEquipmentPosVanguardEnd
	PetEquipmentPosVanguardCount = PetEquipmentPosVanguardEnd + 1 //先锋个数

	PetEquipmentPosThird  PetEquipmentPos = 2 //2_位置3
	PetEquipmentPosFourth PetEquipmentPos = 3 //3_位置4
	PetEquipmentPosFifth  PetEquipmentPos = 4 //4_位置5

	PetEquipmentPosEnd
	PetEquipmentPosMaxNum = PetEquipmentPosEnd + 1 //全部出战位个数

	PetEquipmentMaxNum = PetEquipmentPosMaxNum - PetEquipmentPosVanguardCount // 协助出战个数
)

// 宠物召唤位置
type PetSummonPos = int8

const (
	PetSummonPosNull   PetSummonPos = -1 //未召唤
	PetSummonPosFirst  PetSummonPos = 0  //0_召唤位置1
	PetSummonPosSecond PetSummonPos = 1  //1_召唤位置2

	PetSummonPosMaxNum PetSummonPos = 2 //个数
)

// 条件关系
type RelationType = int8

const (
	RelationTypeNull RelationType = 0 //0_无条件关系
	RelationTypeOr   RelationType = 1 //1_关系或
	RelationTypeAnd  RelationType = 2 //2_关系且
)

// 条件类型
type ConditionType = int32

const PreConditionTypeCount = 20
const (
	ConditionTypeContinue         ConditionType = 0  //0_服务器不检测——直接跳过
	ConditionTypeProbability      ConditionType = 2  //2_概率判断
	ConditionTypeHasBuff          ConditionType = 4  //4_拥有BUFF判断
	ConditionTypeHasPet           ConditionType = 5  //5_拥有宠物判断
	ConditionTypeCarryPet         ConditionType = 6  //6_携带宠物判断
	ConditionTypeHasTask          ConditionType = 7  //7_拥有任务判断
	ConditionTypeHasItem          ConditionType = 8  //8_拥有物品判断
	ConditionTypeHasPassive       ConditionType = 9  //9_拥有被动判断
	ConditionTypeKillCount        ConditionType = 10 //10_击杀数量判断
	ConditionTypeSay              ConditionType = 11 //11_对话任务条件
	ConditionTypeObstructionRange ConditionType = 12 //12_周围是否有遮挡物
	//缺1
	ConditionTypeWeather      ConditionType = 14 //14_是否满足天气条件
	ConditionTypeLeisureStuff ConditionType = 15 //15_是否有挂机摆件
	ConditionTypePassDungeon  ConditionType = 16 //16_通关该关卡
	ConditionTypeBuildCount   ConditionType = 18 //18_指定ID的建筑数量
	ConditionTypeDispatch     ConditionType = 19 //19_家园派遣宠物信息 派遣宠物中有指定元素且x星级的宠物 或者 派遣宠物中有指定ID且x星级的宠物
	ConditionTypeHasWeapon    ConditionType = 20 //20_拥有武器判断
	//缺1
	ConditionTypeUnlockMap ConditionType = 22 //22_世界地图解锁判断
	//缺2
	ConditionTypePetIntimateStep       ConditionType = 25 //25_宠物亲密度等级阶段
	ConditionTypeObj                   ConditionType = 26 //26_对象判断
	ConditionTypePetLevel              ConditionType = 27 //27_宠物等级判断
	ConditionTypePetIntimateLevel      ConditionType = 28 //28_宠物亲密度等级判断
	ConditionTypePetRacial             ConditionType = 29 //29_宠物种族判断
	ConditionTypePetTotalIntimateLevel ConditionType = 30 //30_宠物总亲密度等级判断
	//ConditionTypePetElement            ConditionType = 31 //31_宠物元素判断
	ConditionTypePetId          ConditionType = 32 //32_宠物ID判断
	ConditionTypeOpenServerTime ConditionType = 33 //33_开服时间判断
	ConditionTypeCreateRoleTime ConditionType = 34 //34_创角时间判断
	ConditionTypeWeekTime       ConditionType = 35 //35_每周时间判断
	ConditionTypeDayTime        ConditionType = 36 //36_每天时间判断
	ConditionTypeFixedTime      ConditionType = 37 //37_固定时间判断
	ConditionTypeElement        ConditionType = 38 //38_元素判断
	ConditionTypeGuideFinish    ConditionType = 39 //39_引导完成判断

	ConditionTypeCanUseInObjective       ConditionType = 1000
	ConditionTypeLevel                   ConditionType = 1001 //1001_等级判断
	ConditionTypeVip                     ConditionType = 1002 //1002_VIP等级判断
	ConditionTypeAttribute               ConditionType = 1003 //1003_当前属性判断
	ConditionTypeGetPet                  ConditionType = 1004 //1004_获得宠物
	ConditionTypeGetItem                 ConditionType = 1005 //1005_物品获得
	ConditionTypeDoSchedule              ConditionType = 1006 //1006_事件开始或结束执行
	ConditionTypeFindPlace               ConditionType = 1008 //1008_地点发现——在指定地点进行交互
	ConditionTypePetMaxLevel             ConditionType = 1009 //1009_幻兽最高等级
	ConditionTypePetMaxQuality           ConditionType = 1010 //1010_幻兽最高稀有度
	ConditionTypePetMaxIntimateLevel     ConditionType = 1011 //1011_幻兽最高亲密度
	ConditionTypeBuildingLevel           ConditionType = 1012 //1012_建筑等级判断
	ConditionTypeTrialTowerLevel         ConditionType = 1013 //1013_试炼塔关卡进度判断
	ConditionTypeMainLevel               ConditionType = 1014 //1014_主线关卡进度
	ConditionArmsLevel                   ConditionType = 1018 //1018_武器等级判断
	ConditionArmsActiveTalent            ConditionType = 1019 //1019_武器激活的天赋判断
	ConditionTypeCapturePet              ConditionType = 1021 //1021_抓宠成功次数
	ConditionTypeStatsPetLevel           ConditionType = 1022 //1022_统计宠物等级
	ConditionTypeSpecifyPetLevel         ConditionType = 1023 //1023_指定宠物等级判断
	ConditionTypeStatsPetSkillLevel      ConditionType = 1024 //1024_统计宠物技能等级
	ConditionTypeStatsPetSkillPassive    ConditionType = 1025 //1025_统计宠物解锁能力数量判断
	ConditionTypePetFinishNihility       ConditionType = 1026 //1026_使用宠物完成一关虚无之地
	ConditionTypeFinishiNihilityCount    ConditionType = 1027 //1027_虚无之地通关层数次数
	ConditionTypeClearanceNihilityNoPet  ConditionType = 1028 //1028_不使用某些宠物通关虚无之地
	ConditionTypeNihilityReSelectPassive ConditionType = 1029 //1029_虚无之地重新选择被动次数——仅限困难难度
	ConditionTypeStatsPetIntimateLevel   ConditionType = 1030 //1030_统计宠物亲密度等级判断
	ConditionTypeSpecifyPetIntimateLevel ConditionType = 1031 //1031_指定宠物亲密度等级判断
	ConditionTypePetUnlockCharProperty   ConditionType = 1032 //1032_解锁对应的性格词条判断
	ConditionTypePetWashCharProperty     ConditionType = 1033 //1033_统计性格洗练次数判断
	ConditionTypePetTransferCharacter    ConditionType = 1034 //1034_统计性格转移次数判断
	ConditionTypeSpecifyPetBLLevel       ConditionType = 1035 //1035_指定宠物血统等级判断
	ConditionTypePetQualityBLLevel       ConditionType = 1036 //1036_统计指定稀有度宠物血统等级数量判断
	ConditionTypePetBLEquipSuitEffect    ConditionType = 1037 //1037_血统因子激活套装效果数量判断
	ConditionTypePetBLEquipLevel         ConditionType = 1038 //1038_血统因子等级判断
	ConditionTypeStatsPetBLEquipLevel    ConditionType = 1039 //1039_统计血统因子等级数量判断
	ConditionTypeHomeObjectiveStage      ConditionType = 1040 //1040_家园目标阶段
	//ConditionTypePetElementCount         ConditionType = 1041 //1041_指定属性宠物个数限制判断
	ConditionTypeEquipSlotLevel       ConditionType = 1042 //1042_角色身体强化等级判断
	ConditionTypeFormulaUseCount      ConditionType = 1043 //1043_工坊制造次数
	ConditionTypeArmActiveCount       ConditionType = 1044 //1044_武器激活数判断
	ConditionTypeTrialTowerFloorCount ConditionType = 1045 //1045_武器塔进度判断
	ConditionTypeShipBattleSucCount   ConditionType = 1046 //1046_舰船战斗胜利次数
	ConditionTypePetMaxSkillLevel     ConditionType = 1047 //1047_统计宠物最大技能等级判断

	//ConditionTypePetElementResist             ConditionType = 1049 //1049_宠物抗性值判断
	ConditionTypeMedalScore                   ConditionType = 1050 //1050_勋章积分判断
	ConditionTypeFormationCheck               ConditionType = 1051 //1051_上阵槽位判断
	ConditionTypeEquipWithLevelAndQuality     ConditionType = 1052 //1052_装备穿戴判断
	ConditionTypeCoreTalent                   ConditionType = 1053 //1053_核心天赋激活数量
	ConditionTypeDailyMapStar                 ConditionType = 1054 //1054_通关日常副本星级判断
	ConditionTypePvpFightCount                ConditionType = 1055 //1055_参与竞技场次数
	ConditionTypePvpFightRank                 ConditionType = 1056 //1056_竞技场达到的段位
	ConditionTypeJoinGuild                    ConditionType = 1057 //1057_加入公会
	ConditionTypeMaxCombatScore               ConditionType = 1058 //1058_阵容最高战力判断
	ConditionTypeFuncUnlock                   ConditionType = 1059 //1059_功能解锁判断
	ConditionTypeEquipEnhanceCount            ConditionType = 1060 //1060_装备强化次数判断
	ConditionTypePetLevelUpCount              ConditionType = 1061 //1061_宠物升级次数
	ConditionTypeBloodlineEquipCount          ConditionType = 1062 //1062_宠物因子吸收次数
	ConditionTypeHangUpSpeedUpCount           ConditionType = 1063 //1063_快速挂机次数
	ConditionTypeHangUpAwardCount             ConditionType = 1064 //1064_领取挂机奖励次数
	ConditionTypeMainLineCount                ConditionType = 1065 //1065_主线挑战次数
	ConditionTypeShareCount                   ConditionType = 1066 //1066_任意分享次数
	ConditionTypeHornOpenCount                ConditionType = 1067 //1067_号角开启次数
	ConditionTypeShopBuyCount                 ConditionType = 1068 //1068_商店购买次数
	ConditionTypePetSkillLevelUpCount         ConditionType = 1069 //1069_宠物技能升级次数
	ConditionTypeSchedulePlayCount            ConditionType = 1070 //1070_派遣玩法类型次数
	ConditionArmsActiveTalentClassic          ConditionType = 1071 //1071_武器激活的经典天赋判断
	ConditionEquipItemTotalLevel              ConditionType = 1072 //1072_穿戴装备总等级判断
	ConditionGetPetKindCount                  ConditionType = 1073 //1073_累积拥有幻兽种类数量判断
	ConditionGetPetQualityKindCount           ConditionType = 1074 //1074_累积拥有不同品质的幻兽种类数量判断
	ConditionTypePvpFinishCount               ConditionType = 1075 //1075_竞技场完成战斗次数判断
	ConditionTypeMaxNihilityId                ConditionType = 1076 //1076_虚无之地最高达到层数判断
	ConditionTypeNihilityGetEquipCount        ConditionType = 1077 //1077_虚无之地获得装备个数判断
	ConditionTypeHornOpenGetBLCount           ConditionType = 1078 //1078_在号角开启中获得因子个数判断
	ConditionTypeArmsActiveTalentCount        ConditionType = 1079 //1079_武器激活的天赋点数判断
	ConditionTypeNihilityInfiniteMaxLayer     ConditionType = 1080 //1080_虚无之地无限模式最高难度判断
	ConditionTypeArmsActiveClassicTalentCount ConditionType = 1081 //1081_武器激活的经典天赋点数判断
	ConditionTypeCumulativeLogonDayCount      ConditionType = 1082 //1082_累计登陆天数判断
	ConditionTypeTaskFinishCount              ConditionType = 1083 //1083_累计完成任务数量判断
	ConditionTypeEquipPosWear                 ConditionType = 1084 //1084_指定部位穿戴装备判断
	ConditonTypeEquipWearCount                ConditionType = 1085 //1085_装备穿戴件数判断
	ConditionTypeSkillTypeLevelPetCount       ConditionType = 1086 //1086_统计技能类型等级宠物数量判断
	ConditionTypeTrialTowerGetBLCount         ConditionType = 1087 //1087_在神树试炼中获取的因子个数判断
	ConditionTypeCookingCount                 ConditionType = 1088 //1088_累计烹饪个数判断
	ConditionTypeMainLineObjective            ConditionType = 1089 //1089_指定的主线任务是否完成判断
	ConditionTypeDrawPetCount                 ConditionType = 1090 //1090_香波抽卡完成次数判断
	//空1
	ConditionTypeLevelFinishCount ConditionType = 1092 //1092_副本完成次数判断
	//空1
	ConditionTypePetFeedFoodCount     ConditionType = 1094 //1094_幻兽喂食次数判断
	ConditionTypeUpgradeSparStarCount ConditionType = 1095 //1095_能力晶石升星次数判断
	ConditionTypePetInteractCount     ConditionType = 1096 //1096_幻兽互动次数判断
	ConditionTypeRankLikeCount        ConditionType = 1097 //1097_排行榜点赞次数判断
	ConditionTypeChatCount            ConditionType = 1098 //1098_聊天次数判断

	ConditionTypeClearMapTime      ConditionType = 1099 //1099_关卡通关时间判断
	ConditionTypeTotalPetHpRatio   ConditionType = 1100 //1100_阵容宠物总血量比例判断
	ConditionTypePetDieCount       ConditionType = 1101 //1101_死亡宠物个数判断
	ConditionTypeLinkSkillPayCount ConditionType = 1102 //1102_连携技释放次数判断
	ConditionTypeSwitchPetCount    ConditionType = 1103 //1103_更换宠物次数判断
	ConditionTypeReputationLevel   ConditionType = 1104 //1104_声望等级判断
	ConditionTypeTamerTalentCount  ConditionType = 1105 //1105_职业天赋等级判断

	ConditionTypeServerEnd
	ConditionTypeServerMax = ConditionTypeServerEnd + 1

	//客户端条件,服务器不参与逻辑
	ConditionTypeClientStart          ConditionType = 2001
	ConditionTypeTreeObjectiveReceive               //2001_神树目标是否领奖

	ConditionTypeClientEnd
	ConditionTypeClientMax = ConditionTypeClientEnd + 1
)

// 条件上下文相关
type ConditionContextKey = int32

const (
	CCKObjectiveProgress    ConditionContextKey = int32(1)
	CCKObjectiveMaxProgress ConditionContextKey = int32(2)
)

// 条件操作符
type ConditionOperatorType = int32

const (
	COTEqual              ConditionOperatorType = 1 //1_等于
	COTGreaterThan        ConditionOperatorType = 2 //2_大于
	COTGreaterThanOrEqual ConditionOperatorType = 3 //3_大于等于
	COTLessThan           ConditionOperatorType = 4 //4_小于
	COTLessThanOrEqual    ConditionOperatorType = 5 //5_小于等于
	COTNotEqual           ConditionOperatorType = 6 //6_不等于
)

// 条件操作符字符串
var ConditionOperatorTypeStr = [7]string{"?", "=", ">", ">=", "<", "<=", "!="}

// 属性——只能添加
type AttributeType = int32

const AttributeTypeSaveMaxLen AttributeType = 100
const AttributeTypeAllMaxLen AttributeType = 200
const PetInitAttributeLen = 35

const AttributeTypeAdditionPreLen = 10
const (
	//特殊属性
	AttributeTypeExp             AttributeType = 0  //0_经验
	AttributeTypeLevel           AttributeType = 1  //1_等级
	AttributeTypeCoin            AttributeType = 2  //2_货币（金币）
	AttributeTypeDiamonds        AttributeType = 3  //3_钻石（绑定梦之晶）
	AttributeTypePetExp          AttributeType = 4  //4_宠物经验
	AttributeTypeIngredients     AttributeType = 5  //5_食材
	AttributeTypeMineral         AttributeType = 6  //6_矿物
	AttributeTypeUpgradeMaterial AttributeType = 7  //7_升级材料
	AttributeTypeEnergy          AttributeType = 8  //8_能量(麻吉橡果)
	AttributeTypePostureEnergy   AttributeType = 9  //9_姿态能量
	AttributeTypeSpecialEnergy   AttributeType = 10 //10_特殊能量
	AttributeTypeArenaCoin       AttributeType = 11 //11_竞技币
	AttributeTypeHornScore       AttributeType = 12 //12_号角积分
	AttributeTypeBrawn           AttributeType = 13 //13_体力
	AttributeTypeSavvy           AttributeType = 14 //14_悟性
	AttributeTypeDreamSpar       AttributeType = 15 //15_梦之晶（只能充值获得）
	AttributeTypeAdvancedCoin    AttributeType = 16 //16_丹堤
	AttributeTypeBreakLevel      AttributeType = 18 //18_突破等级
	AttributeTypeReputationLevel AttributeType = 19 //19_声望等级
	AttributeTypeReputationExp   AttributeType = 20 //20_声望经验
	AttributeTypeReputation      AttributeType = 21 //21_声望

	//一级属性
	PlayerAttributeTypeNoSaveStart AttributeType = 50
	PlayerAttributeTypePhysique    AttributeType = 51 //51_体质	实数
	PlayerAttributeTypeStrength    AttributeType = 52 //52_力量	实数
	PlayerAttributeTypeIntellect   AttributeType = 53 //53_智慧	实数
	PlayerAttributeTypeDexterous   AttributeType = 54 //54_灵巧	实数
	PlayerAttributeTypeStamina     AttributeType = 55 //55_耐力	实数
	PlayerAttributeTypePrimaryEnd                     //一级属性结束

	//一般属性
	AttributeTypeMaxHp              AttributeType = 101 //101_最大生命
	AttributeTypeAttack             AttributeType = 102 //102_攻击力
	AttributeTypeMagicAttack        AttributeType = 103 //103_幻力攻击力
	AttributeTypeDefense            AttributeType = 104 //104_防御力
	AttributeTypeMagicDefense       AttributeType = 105 //105_幻力防御力
	AttributeTypeMaxMana            AttributeType = 106 //106_最大能量
	AttributeTypePierce             AttributeType = 107 //107_穿透
	AttributeTypeArmor              AttributeType = 108 //108_护甲
	AttributeTypeSpEfficiency       AttributeType = 109 //109_Sp值获取率，万分比
	AttributeTypeSpReply            AttributeType = 110 //110_Sp自动获取值，实数值
	AttributeTypeEnmityRatio        AttributeType = 111 //111_仇恨系数，万分比
	AttributeTypeHpReply            AttributeType = 112 //112_生命回复，实数值
	AttributeTypeAttackHpReply      AttributeType = 113 //113_攻击回复生命，实数值
	AttributeTypeAttackHpReplyRatio AttributeType = 114 //114_生命吸取，万分比
	AttributeTypeAttackSpeed        AttributeType = 115 //115_攻击速度，万分比
	AttributeTypeManaReply          AttributeType = 116 //116_Mana自动获取值，实数值
	AttributeTypePetReplyNowHpRatio AttributeType = 117 //117_每x秒恢复最大生命比例，万分比 宠物独有，暂定x=2
	AttributeTypeMoveNowHpRatio     AttributeType = 118 //118_移动每x秒恢复最大生命比例，万分比 移动每x秒恢复最大生命比例，暂定x=2
	AttributeTypePetReplyNowHpValue AttributeType = 119 //119_每x秒恢复生命值，实数值 宠物独有，暂定x=2
	AttributeTypeSkillCooldown      AttributeType = 120 //120_技能冷却加成，万分比
	AttributeTypeMaxShieldHp        AttributeType = 121 //121_最大血量护盾，实数值

	//战斗轮盘相关属性
	AttributeTypeHit                 AttributeType = 201 //201_命中，万分比
	AttributeTypeHitValue            AttributeType = 202 //202_命中值，实数值
	AttributeTypeDodge               AttributeType = 203 //203_闪避，万分比
	AttributeTypeDodgeValue          AttributeType = 204 //204_闪避值，实数值
	AttributeTypeCritical            AttributeType = 205 //205_暴击几率，万分比
	AttributeTypeCtValue             AttributeType = 206 //206_暴击值，实数值
	AttributeTypeCtDamage            AttributeType = 207 //207_暴击伤害，万分比
	AttributeTypeCtDamageValue       AttributeType = 208 //208_暴击额外伤害，实数值
	AttributeTypeCtResist            AttributeType = 209 //209_暴击抵抗几率，万分比
	AttributeTypeCtResistValue       AttributeType = 210 //210_暴击抵抗值，实数值
	AttributeTypeCtDamageResist      AttributeType = 211 //211_暴击伤害抵抗，万分比
	AttributeTypeCtDamageAttenuation AttributeType = 212 //212_暴击伤害总衰减，万分比
	AttributeHitParam                AttributeType = 213 //213_命中参数，实数值
	AttributeCtParam                 AttributeType = 214 //214_暴击参数，实数值
	AttributePierceParam             AttributeType = 215 //215_穿透参数，实数值

	//增伤相关属性
	AttributeTypeAllDamage            AttributeType = 301 //301_总伤害修正，万分比
	AttributeTypeAllDamageResist      AttributeType = 302 //302_总受伤修正，万分比
	AttributeTypeLordDamage           AttributeType = 303 //303_对领主伤害，万分比
	AttributeTypeBossDamage           AttributeType = 304 //304_对首领伤害，万分比
	AttributeTypeEliteDamage          AttributeType = 305 //305_对精英伤害，万分比
	AttributeTypeNormalDamage         AttributeType = 306 //306_对普通伤害，万分比
	AttributeTypeThrowSkill           AttributeType = 307 //307_投射技能伤害，万分比
	AttributeTypeMeleeSkill           AttributeType = 308 //308_近战技能伤害，万分比
	AttributeTypeMagicSkill           AttributeType = 309 //309_法术技能伤害，万分比
	AttributeTypeRangeSkill           AttributeType = 310 //310_范围技能伤害，万分比
	AttributeTypeNormalSkill          AttributeType = 311 //311_普通技能伤害，万分比
	AttributeTypeLinkSkill            AttributeType = 312 //312_连携技能伤害，万分比
	AttributeTypeCommonlySkill        AttributeType = 313 //313_一般伤害修正，万分比
	AttributeTypeRemoteSkill          AttributeType = 314 //314_远程技能伤害，万分比
	AttributeTypeEvadeSkill           AttributeType = 315 //315_位移技能伤害，万分比
	AttributeTypeContinuedSkill       AttributeType = 316 //316_持续技能伤害，万分比
	AttributeTypeThrowSkillResist     AttributeType = 317 //317_投射受伤修正，万分比
	AttributeTypeMeleeSkillResist     AttributeType = 318 //318_近战受伤修正，万分比
	AttributeTypeMagicSkillResist     AttributeType = 319 //319_法术受伤修正，万分比
	AttributeTypeRemoteSkillResist    AttributeType = 320 //320_远程受伤修正，万分比
	AttributeTypeLinkSummon           AttributeType = 321 //321_出战幻兽连携伤害修正，万分比
	AttributeTypeLinkPrepare          AttributeType = 322 //322_备战幻兽连携伤害修正，万分比
	AttributeTypeLinkSKillDamageValue AttributeType = 323 //323_连携技伤害固定值，实数值
	AttributeTypePlayerDamage         AttributeType = 324 //324_对玩家伤害，万分比
	AttributeTypePetDamage            AttributeType = 325 //325_对幻兽伤害，万分比
	AttributeTypeSufferPlayerDamage   AttributeType = 326 //326_受玩家伤害，万分比
	AttributeTypeSufferPetDamage      AttributeType = 327 //327_受幻兽伤害，万分比
	AttributeTypeSufferMonsterDamage  AttributeType = 328 //328_受怪物伤害，万分比
	AttributeTypeBuffStateTime        AttributeType = 329 //329_增益状态时间，万分比
	AttributeTypeDeBuffStateTime      AttributeType = 330 //330_减益状态时间，万分比
	AttributeTypeAttackUpDizzyDamage  AttributeType = 331 //331_对控制状态造成伤害提升，攻击方，万分比
	AttributeTypeDefenseUpDizzyDamage AttributeType = 332 //332_控制状态受到伤害提升，防御方，万分比
	AttributeTypeAbnormalHit          AttributeType = 333 //333_异常状态命中，万分比
	AttributeTypeAbnormalDodge        AttributeType = 334 //334_异常状态闪避，万分比

	AttributeTypeAdditionalAttackElement AttributeType = 401 //401_附加攻击力元素伤害	万分比
	AttributeTypeAdditionalAttackChaotic AttributeType = 402 //402_附加攻击力混乱伤害	万分比
	AttributeTypeAdditionalAttackVoid    AttributeType = 403 //403_附加攻击力虚无伤害	万分比
	AttributeTypeAttackElement           AttributeType = 404 //404_元素攻击	实数
	AttributeTypeDefenseElement          AttributeType = 405 //405_元素防御	实数
	AttributeTypeAttackChaotic           AttributeType = 406 //406_混乱攻击	实数
	AttributeTypeDefenseChaotic          AttributeType = 407 //407_混乱防御	实数
	AttributeTypeAttackVoid              AttributeType = 408 //408_虚无攻击	实数
	AttributeTypeDefenseVoid             AttributeType = 409 //409_虚无防御	实数

	//技能及执行器相关
	AttributeTypeEvadeDistance              AttributeType = 551 //551_位移距离，实数值——非防御技能且为位移技能生效
	AttributeTypeBulletSize                 AttributeType = 552 //552_子弹大小，实数值
	AttributeTypeBulletCount                AttributeType = 553 //553_额外子弹数量，实数值 todo数量暂时不做
	AttributeTypeLinkSkillRange             AttributeType = 554 //554_连接技能范围，万分比
	AttributeTypeLinkSkillTargetCount       AttributeType = 555 //555_连接技能目标数量，实数值
	AttributeTypeSummonLiveTime             AttributeType = 556 //556_召唤时间，实数值 单位ms
	AttributeTypeSummonInheritAttack        AttributeType = 557 //557_召唤属性继承额外攻击比例，万分比
	AttributeTypeSummonInheritHp            AttributeType = 558 //558_召唤属性继承额外生命比例，万分比
	AttributeTypeSummonInheritDefense       AttributeType = 559 //559_召唤属性继承额外防御比例，万分比
	AttributeTypeSummonLimitCount           AttributeType = 560 //560_召唤总数，实数值
	AttributeTypeDefenseSkillUseCount       AttributeType = 561 //561_防御次数上限，实数值
	AttributeTypeDefenseSkillEvadeDistance  AttributeType = 562 //562_防御技能距离，实数值——防御技能且为位移技能生效
	AttributeTypeDefenseSkillInvincibleTime AttributeType = 563 //563_防御无敌时间，实数值
	AttributeTypeRangeSkillRange            AttributeType = 564 //564_技能范围修改，万分比

	//治疗相关
	AttributeTypeSwitchRecoveryEffect AttributeType = 601 //601_治疗效果，万分比
	AttributeTypeReplyHpShieldEffect  AttributeType = 602 //602_回盾效果，万分比

	//PVP相关
	AttributeTypePvpAllReduceTreated   AttributeType = 701 //701_PVP总受治疗降低,万分比
	AttributeTypePvpAllReduceDamage    AttributeType = 702 //702_PVP总伤害降低,万分比
	AttributeTypePvpAllReduceToughness AttributeType = 703 //703_PVP韧性,万分比

	//陷阱相关
	AttributeTypeTrapAttackExt      AttributeType = 740 // 740_陷阱伤害加,万分比成
	AttributeTypeTrapRangeExt       AttributeType = 741 // 741_陷阱效果范围,万分比围
	AttributeTypeTrapTimeExt        AttributeType = 742 // 742_陷阱生存和触发时间,毫秒
	AttributeTypeTrapAttackInherit  AttributeType = 743 // 743_陷阱属性继承额外攻击比例属,万分比性
	AttributeTypeTrapHpInherit      AttributeType = 744 // 744_陷阱属性继承额外生命比例属,万分比性
	AttributeTypeTrapDefenceInherit AttributeType = 745 // 745_陷阱属性继承额外防御比例属,万分比性
	AttributeTypeTrapMaxCount       AttributeType = 746 // 746_陷阱总数量

	AttributeTypeSaveMaxNum //最大存档类型
	//不需要存档的属性
	AttributeTypeNoSaveStart    AttributeType = 1000
	AttributeTypeNowHp          AttributeType = 1001 //1001_当前生命
	AttributeTypeNowMana        AttributeType = 1002 //1002_当前能量
	AttributeTypeSpeed          AttributeType = 1003 //1003_移动速度
	AttributeTypeModeRadius     AttributeType = 1004 //1004_模型半径
	AttributeTypeModeZoomChange AttributeType = 1005 //1005_模型缩放
	AttributeTypeShieldTimes    AttributeType = 1006 //1006_次数护盾
	AttributeTypeShieldHp       AttributeType = 1007 //1007_血量护盾

	AttributeTypeMaxNum
)

type ESkillFromType = int8

const (
	ESkillFromSelf   ESkillFromType = 0 //0_自身
	ESkillFromMaster ESkillFromType = 1 //1_主人
	ESkillFromEnemy  ESkillFromType = 2 //2_敌人
)

type PotentialType = int32

const (
	PTypeHp             PotentialType = 0 //0_生命资质
	PTypePhysicsAttack  PotentialType = 1 //1_物理攻击资质
	PTypeMagicAttack    PotentialType = 2 //2_幻法攻击资质
	PTypePhysicsDefense PotentialType = 3 //3_物理防御资质
	PTypeMagicDefense   PotentialType = 4 //4_幻法防御资质
	PTypeEnd

	PTypeMax = PTypeEnd + 1
)

func CheckIsValidPotentialType(pType PotentialType) bool {
	if pType >= PTypeHp && pType < PTypeMax {
		return true
	}
	return false
}

// 事件类型
type EventType = int32

const (
	EventTypeUserAttribute  EventType = 0  //0_玩家属性修改
	EventTypeDamage         EventType = 1  //1_技能伤害
	EventTypeSkillSwitch    EventType = 2  //2_技能开关
	EventTypeTauntEd        EventType = 3  //3_强制目标选择
	EventTypeAddTrap        EventType = 4  //4_添加陷阱
	EventTypeSummonMonster  EventType = 5  //5_召唤小怪事件
	EventTypeSkillCd        EventType = 7  //7_技能CD修改
	EventZoomChange         EventType = 8  //8_模型缩放事件
	EventTypeDispelBuff     EventType = 9  //9_驱散&消除BUFF
	EventTypePassive        EventType = 10 //10_添加&移除被动事件
	EventTypeEnd            EventType = 11 //11_终结事件
	EventClearEnmity        EventType = 12 //12_清理仇恨
	EventBrushSummonMonster EventType = 13 //13_召唤小怪组事件
	EventSkillAddBlood      EventType = 14 //14_技能回血
	EventAiNodeEnable       EventType = 15 //15_AI节点开关
	EventCanBeAttack        EventType = 16 //16_可被攻击状态
	EventTypeRemoveBuff     EventType = 17 //17_移除指定BUFF
	EventTypeBuff           EventType = 18 //18_添加&移除BUFF，通过BuffCfgId
	EventTypeGoods          EventType = 19 //19_添加&移除指定物品
	EventTypeBrushMonster   EventType = 20 //20_怪物组刷怪事件
	//EventTypeFinishTaskSayStep       EventType = 21 //21_完成任务对话步骤
	EventTypeReplySp                 EventType = 22 //22_回复Sp
	EventTypeSkillNowCd              EventType = 23 //23_修改技能当前CD
	EventTypeSummonPet               EventType = 24 //24_召唤宠物事件
	EventTypeChangeWeather           EventType = 25 //25_改变天气事件
	EventTypeAddSchedule             EventType = 26 //26_添加派遣事件
	EventTypeBrushInteractiveMonster EventType = 27 //27_添加可交互对象
	EventTypeOperateTreasureMap      EventType = 28 //28_操作宝图
	EventTypeDropEvent               EventType = 29 //29_掉落事件
	AddDandelionEvent                EventType = 30 //30_掉落蒲公英事件
	BuffTransmitDamageEvent          EventType = 31 //31_Buff传导伤害事件
	BuffCommunicateEvent             EventType = 32 //32_Buff传染事件
	BuffDamageEvent                  EventType = 33 //33_Buff伤害事件
	AttrDamageEvent                  EventType = 34 //34_属性伤害事件
	SkillAddHpShieldEvent            EventType = 35 //35_技能回复血量护盾
	DieEvent                         EventType = 36 //36_死亡事件
	KillSkillFormObjEvent            EventType = 37 //37_击杀技能来源对象
	EventTypeTalk                    EventType = 38 //38_对话事件
	BuffBreakSelfEvent               EventType = 39 //39_buff打断自身事件

	EventTypeTail
	EventTypeMaxNum = EventTypeTail + 1 //数量
)

// 事件值类型
type EventValueType = int32

const (
	EVTypeAbsolute    EventValueType = 1 //1_绝对值
	EVTypePercentage  EventValueType = 2 //2_万分比
	EVTypeSetAbsolute EventValueType = 3 //3_设置绝对值
	EVTypeSetRatio    EventValueType = 4 //4_设置万分比
)

// 掉落循环最大次数
const DropCycleCount = int8(4)
const MaxSubHpDropCount = 20 //扣血掉落最多不能超过20次

// 掉落方法
type DropFunc = int8

const (
	DropFuncIndependent DropFunc = 1 //1_掉落独立
	DropFuncRepeat      DropFunc = 2 //2_掉落重复
	DropFuncNoRepeat    DropFunc = 3 //3_掉落不重复
)

const (
	MaxDropIndependent = int(20)  // 独立掉落最大配置数
	MaxDropRepeat      = int(100) // 重复掉落最大
	MaxDropNoRepeat    = int(100) // 不重复掉落最大
)

// 奖励类型
type GoodsTypeEnum = int8

const (
	GoodsTypeItem            GoodsTypeEnum = 0  //0_道具
	GoodsTypeAttribute       GoodsTypeEnum = 1  //1_属性
	GoodsTypePet             GoodsTypeEnum = 2  //2_宠物
	GoodsTypeTalent          GoodsTypeEnum = 3  //3_天赋点
	GoodsTypeTask            GoodsTypeEnum = 4  //4_任务
	GoodsTypeGem             GoodsTypeEnum = 5  //5_宝石
	GoodsTypeDropId          GoodsTypeEnum = 6  //6_掉落ID
	GoodsTypeResources       GoodsTypeEnum = 7  //7_资源——入家园仓库
	GoodsTypePetIntimate     GoodsTypeEnum = 8  //8_宠物亲密度
	GoodsTypePetMood         GoodsTypeEnum = 9  //9_宠物心情
	GoodsTypeAction          GoodsTypeEnum = 11 //11_营地动作
	GoodsTypeGuildContribute GoodsTypeEnum = 12 //12_公会贡献
	GoodsTypeTalent2         GoodsTypeEnum = 13 //13_2型天赋点
	GoodsTypeTalentClassic   GoodsTypeEnum = 14 //14_经典天赋点
	GoodsTypeCookbook        GoodsTypeEnum = 15 //15_食谱
	GoodsTypeTamerTalent     GoodsTypeEnum = 16 //16_职业天赋点

	GoodsTypeEnd
	GoodsTypeMaxNum = GoodsTypeEnd + 1
)

// Buff类型
type EBuffType = int32

const (
	GainBuff    EBuffType = 0 //0_增益型Buff
	DeBuff      EBuffType = 1 //1_减益型Buff
	PassiveBuff EBuffType = 2 //2_被动型Buff(暂时不用)
	ControlBuff EBuffType = 3 //3_控制型Buff

	BuffPlaceholder
	MaxBuffType = BuffPlaceholder + 1
)

// Buff驱散/净化类型
type EBuffDispelType = int32

const (
	PurifyType EBuffDispelType = 1 //1_净化
	DispelType EBuffDispelType = 2 //2_驱散
)

type EBuffTableStatus = int32

const (
	BuffTableHit      EBuffTableStatus = 0 //buff命中
	BuffTableImmunity EBuffTableStatus = 1 //buff免疫
	BuffTableResist   EBuffTableStatus = 2 //buff抵抗
)

// 行为限制类型
type EBehaviorRestrict = int

const (
	NoneRestrict EBehaviorRestrict = 0 //0_无
	Dizzy        EBehaviorRestrict = 1 //1_眩晕
	Frozen       EBehaviorRestrict = 2 //2_冰冻
	Afraid       EBehaviorRestrict = 3 //3_恐惧
	Silent       EBehaviorRestrict = 4 //4_沉默
	ForbidMove   EBehaviorRestrict = 5 //5_禁止移动
	BeRidiculed  EBehaviorRestrict = 6 //6_被嘲讽
	HitDown      EBehaviorRestrict = 7 //7_击倒

	HitBack EBehaviorRestrict = 11 //11_击退 --特殊，不属于buff行为

	MaxRestrict EBehaviorRestrict = 20 //最大限制
)

// 特殊能力类型
type ESpecialAbility = int

const (
	NoneSpecialAbility ESpecialAbility = 0 //0_无
	InvincibleMode     ESpecialAbility = 1 //1_无敌模式   会受到攻击，但不费HP
	SafeMode           ESpecialAbility = 2 //2_安全模式   不会受到攻击
	Invisible          ESpecialAbility = 3 //3_隐身  不会被敌人主动发现
	AntiInvisible      ESpecialAbility = 4 //4_反隐身  可以发现隐身对象
	ChangeSkin         ESpecialAbility = 5 //5_换皮肤

	EndSpecialAbility
	MaxSpecialAbility = EndSpecialAbility + 1 //最大子限制
)

const InvalidPosDirector = uint32(360)

// 免疫效果标识
type EImmuneTag = int

const (
	NoneImmuneTag     EImmuneTag = 0  //0_无
	ImmuneDizzy       EImmuneTag = 1  //1_免疫眩晕
	ImmuneFrozen      EImmuneTag = 2  //2_免疫冰冻
	ImmuneAfraid      EImmuneTag = 3  //3_免疫恐惧
	ImmuneSilent      EImmuneTag = 4  //4_免疫沉默
	ImmuneForbidMove  EImmuneTag = 5  //5_免疫禁止移动
	ImmuneHitBack     EImmuneTag = 6  //6_免疫击退
	ImmuneHitFly      EImmuneTag = 7  //7_免疫击飞
	ImmuneRidiculed   EImmuneTag = 8  //8_免疫嘲讽
	ImmuneHitDown     EImmuneTag = 9  //9_免疫击倒
	ImmuneMaxHpChange EImmuneTag = 10 //10_免疫最大血量修改

	ImmuneEnd
	ImmuneMax EImmuneTag = ImmuneEnd + 1
)

// Buff打断类型
type EBreakType = int

const (
	NoneBreak       EBreakType = 0  //0_无
	MoveBreak       EBreakType = 1  //1_移动打断
	AttackBreak     EBreakType = 2  //2_攻击打断
	HurtBreak       EBreakType = 3  //3_受伤打断
	LeaveMapBreak   EBreakType = 4  //4_离开场景打断
	DieBreak        EBreakType = 5  //5_死亡打断
	LeaveWarBreak   EBreakType = 6  //6_脱战打断
	DodgeBreak      EBreakType = 7  //7_闪避打断
	HitBreak        EBreakType = 8  //8_命中打断
	CtBreak         EBreakType = 13 //13_暴击打断
	NotCtBreak      EBreakType = 14 //14_未暴击打断
	SummonBreak     EBreakType = 15 //15_进入出战区打断
	ExitSummonBreak EBreakType = 16 //16_离开出战区打断

	EndBreak
	MaxBreak = EndBreak + 1 //最大类型

)

// 技能被影响效果类型
type ESkillEffectType = int

const (
	NoneSkillEffectType           ESkillEffectType = 0  //0_无
	MoveBreakSkillEffectType      ESkillEffectType = 1  //1_移动时打断技能
	NoHitFlySkillEffectType       ESkillEffectType = 2  //2_禁止被击飞
	NoHitDownSkillEffectType      ESkillEffectType = 4  //4_禁止被击倒
	NoRotateSkillEffectType       ESkillEffectType = 8  //8_禁止旋转
	SkillBreakMoveSkillEffectType ESkillEffectType = 16 //16_技能打断移动
	InvincibleEffectType          ESkillEffectType = 32 //32_技能释放时无敌
)

// 冲突方法
type EConflictMethod = uint8

const (
	Stack        EConflictMethod = 0 //0_叠加
	OneSelf      EConflictMethod = 1 //1_独占 先到先得
	Replace      EConflictMethod = 2 //2_替换 等级高的优先
	Accumulation EConflictMethod = 3 //3_累加
)

// 动作状态
type EAFSMType = int32

const (
	Die        = 3
	Born       = 10
	DieAndBorn = 23
)

// 对象移除类型
type EObjRemoveType = int32

const (
	EORTDie      = 0 //0_死亡
	EORTRemove   = 1 //1_直接删除
	EORTCaptured = 2 //2_被捕捉
	EORTEscape   = 3 //3_逃跑
)

// 闪避配置参数
type EHitDodgeType = int8

const (
	EHDTNormalHit EHitDodgeType = 0 // 0_正常计算闪避命中
	EHDTMustHit   EHitDodgeType = 1 // 1_必定命中
)

// 闪避配置参数
type ECriticalType = int8

const (
	ECTNormal       ECriticalType = 0 // 0_正常计算暴击率
	ECTMustCritical ECriticalType = 1 // 1_必定暴击
)

// 技能类型
type ESkillType = uint64

const (
	SkillTypeNone           ESkillType = 0
	SkillTypeCommonly       ESkillType = 1      // 1_一般技能
	SkillTypeNormal         ESkillType = 2      // 2_普通技能
	SkillTypeEvade          ESkillType = 4      // 4_位移技能
	SkillTypeLink           ESkillType = 8      // 8_连携技能
	SkillTypeMelee          ESkillType = 16     // 16_近战技能
	SkillTypeThrow          ESkillType = 32     // 32_投射技能
	SkillTypeRange          ESkillType = 64     // 64_范围技能
	SkillTypeTreat          ESkillType = 128    // 128_治疗技能
	SkillTypeMagic          ESkillType = 256    // 256_法术技能
	SkillTypeContinued      ESkillType = 512    // 512_持续技能
	SkillTypeDefense        ESkillType = 1024   // 1024_防御技能
	SkillTypeControl        ESkillType = 2048   // 2048_控制技能
	SkillTypeTrap           ESkillType = 4096   // 4096_陷阱技能
	SkillTypeSummon         ESkillType = 8129   // 8129_召唤技能
	SkillTypeStatus         ESkillType = 16384  // 16384_状态技能
	SkillTypeBullet         ESkillType = 32768  // 32768_子弹技能
	SkillTypeRemote         ESkillType = 65536  // 65536_远程技能
	SkillTypePassiveTrigger ESkillType = 131072 // 131072_被动触发技能
)

const NotFeaturesSkillTypeFlag = SkillTypeCommonly | SkillTypeNormal | SkillTypeLink | SkillTypeDefense | SkillTypePassiveTrigger //非功能的技能类型——在技能中,该标签有且只能有一个 及：一个技能不能是一般技能又是连携技能
const AllSkillTypeFlag = SkillTypeCommonly | SkillTypeNormal | SkillTypeEvade | SkillTypeLink | SkillTypeMelee | SkillTypeThrow | SkillTypeRange | SkillTypeTreat | SkillTypeMagic | SkillTypeContinued | SkillTypeDefense | SkillTypeControl | SkillTypeTrap | SkillTypeSummon | SkillTypeStatus | SkillTypeBullet | SkillTypeRemote
const SkillTypeCount = 17

var SkillTypeList = [SkillTypeCount]uint64{SkillTypeCommonly, SkillTypeNormal, SkillTypeEvade, SkillTypeLink, SkillTypeMelee, SkillTypeThrow, SkillTypeRange, SkillTypeTreat, SkillTypeMagic, SkillTypeContinued, SkillTypeDefense, SkillTypeControl, SkillTypeTrap, SkillTypeSummon, SkillTypeStatus, SkillTypeBullet, SkillTypeRemote}
var NotFeaturesSkillTypeList = [5]uint64{SkillTypeCommonly, SkillTypeNormal, SkillTypeLink, SkillTypeDefense, SkillTypePassiveTrigger} //非功能的技能类型——在技能中,该标签有且只能有一个 及：一个技能不能是一般技能又是连携技能

// 获取宠物的技能类型,返回只有3种技能
func GetPetSkillType(skillTypeFlag uint64) ESkillType {
	if skillTypeFlag&SkillTypeCommonly == SkillTypeCommonly { //1_一般技能
		return SkillTypeCommonly
	} else if skillTypeFlag&SkillTypeNormal == SkillTypeNormal { //2_普通技能 普通攻击
		return SkillTypeNormal
	} else if skillTypeFlag&SkillTypeLink == SkillTypeLink { //8_连携技能
		return SkillTypeLink
	}
	return SkillTypeNone
}

// 技能范围
type ESkillRangeType = uint8

const (
	ESkillRangeNone         ESkillRangeType = 0 //0_无技能范围
	ESkillRangeCircle       ESkillRangeType = 1 //1_圆形
	ESkillRangeRect         ESkillRangeType = 2 //2_矩形
	ESkillRangeFan          ESkillRangeType = 3 //3_扇形
	ESkillRangePoints       ESkillRangeType = 4 //4_点阵
	ESkillRangeVariableRect ESkillRangeType = 5 //5_可变矩形
)

// 执行器范围选择基点类型
type ExecutorSelectSrcType = uint8

const (
	UseParentTarget     ExecutorSelectSrcType = 1 //1_使用父节点目标dd
	UseParentEndPoint   ExecutorSelectSrcType = 2 //2_使用父节点终点
	UseParentPosition   ExecutorSelectSrcType = 3 //3_使用父执行器轨迹——子弹特殊含义
	UseParentSrc        ExecutorSelectSrcType = 4 //4_使用父节点起点
	UseParentSrcPoint   ExecutorSelectSrcType = 5 //5_使用父节点起点位置
	UseParentOwner      ExecutorSelectSrcType = 6 //6_使用技能所有者
	UseParentOwnerPoint ExecutorSelectSrcType = 7 //7_使用技能所有者位置
)

// 执行器效果
type ESkillHitEffect = uint8

const (
	EHitNone = 0 //0_无打击效果
	EHItDown = 1 //1_击倒
	EHItFly  = 2 //2_击飞
)

type ExecutorNameType = uint8

const (
	NoneExecutor          ExecutorNameType = 0
	CommonExecutor        ExecutorNameType = 1 //通用执行器
	IntervalExecutor      ExecutorNameType = 2 //间隔时间执行器
	EvadeExecutor         ExecutorNameType = 3 //位移执行器
	RepelExecutor         ExecutorNameType = 4 //击退执行器
	BulletExecutor        ExecutorNameType = 5 //子弹执行器
	VirtualExecutor       ExecutorNameType = 6 //虚拟执行器
	CycleChainExecutor    ExecutorNameType = 7 //闪电链执行器
	PassiveExecutor       ExecutorNameType = 8 //被动触发执行器
	ConditionDecoExecutor ExecutorNameType = 9 //条件触发执行器

	EndExecutor
	MaxExecutor = EndExecutor + 1
)

type ExecutorRealNameType = uint8

const (
	RealTypeNoneExecutor                     ExecutorRealNameType = 0
	RealTypeNormalExecutor                   ExecutorRealNameType = 1
	RealTypeEvadeExecutor                    ExecutorRealNameType = 2
	RealTypeRepelExecutor                    ExecutorRealNameType = 3
	RealNameTypeBulletExecutor               ExecutorRealNameType = 4
	RealTypeTrajactoryExecutor               ExecutorRealNameType = 5
	RealTypeCycleChain                       ExecutorRealNameType = 6
	RealTypeBulletFalseConstantSpeedExecutor ExecutorRealNameType = 7
	RealTypeBulletFalseConstantTimeExecutor  ExecutorRealNameType = 8
	RealTypeVirtualExecutor                  ExecutorRealNameType = 9
	RealTypeCycleChainExecutor               ExecutorRealNameType = 10
	RealTypePassiveExecutor                  ExecutorRealNameType = 11

	RealTypeEndExecutor
	RealTypeMaxExecutor = RealTypeEndExecutor + 1
)

type DamageFunType = int32

const (
	DamageFunMin              DamageFunType = 0 //最小
	DamageFunNormal           DamageFunType = 0 //0_普通伤害算法
	DamageFunTargetDecreasing DamageFunType = 1 //1_目标递减算法
	DamageFunTargetHitCount   DamageFunType = 2 //2_多次命中同一目标算法

	DFEnd
	DFMax = DFEnd + 1
)

type DamageFromType = int8

const (
	DamageFromMin     DamageFromType = 0 //最小
	DamageFromSkill   DamageFromType = 0 //技能
	DamageFromTrap    DamageFromType = 1 //陷阱
	DamageFromBuff    DamageFromType = 2 //buff
	DamageFromPassive DamageFromType = 3 //被动
	DamageFromBorn    DamageFromType = 4 //出生
	DamageFromDie     DamageFromType = 5 //死亡
)

type ESkillDamageTableStatus = int32

const (
	TableStatusNone  ESkillDamageTableStatus = 0 //0_技能正常命中
	TableStatusDodge ESkillDamageTableStatus = 1 //1_技能被闪避
	//TableStatusParry    ESkillDamageTableStatus = 2 //2_技能被招架
	//TableStatusBlock    ESkillDamageTableStatus = 3 //3_技能被格挡
	TableStatusCritical ESkillDamageTableStatus = 4 //4_技能暴击
)

var mapCheckRangeSwitch map[ExecutorNameType]bool

var mapExecuteName map[string]ExecutorNameType

var mapExecuteRealName map[string]ExecutorRealNameType

var mapMoveBehaviorRestrict map[EBehaviorRestrict]bool

var mapSkillBehaviorRestrict map[EBehaviorRestrict]bool

// 限制切换幻兽的限制
var mapSwitchSummonPetBehaviorRestrict map[EBehaviorRestrict]bool

var mapNoCostHomeResoures map[EHomeResources]struct{}

// 镜像玩家、宠物需要同步的属性
var mapFightNotifyAttribute map[AttributeType]struct{}

// 资质转换目标属性
var mapPotentialTypeToAttribute map[PotentialType]AttributeType

func init() {
	mapExecuteName = make(map[string]ExecutorNameType)
	mapMoveBehaviorRestrict = make(map[EBehaviorRestrict]bool)
	mapSkillBehaviorRestrict = make(map[EBehaviorRestrict]bool)
	mapSwitchSummonPetBehaviorRestrict = make(map[EBehaviorRestrict]bool)
	mapExecuteRealName = make(map[string]ExecutorRealNameType)
	mapClientCarePassive = make(map[PassiveType]struct{})
	mapNoCostHomeResoures = make(map[EHomeResources]struct{})
	mapFightNotifyAttribute = make(map[AttributeType]struct{})
	mapPotentialTypeToAttribute = make(map[PotentialType]AttributeType, PTypeMax)

	mapExecuteName["commonexecutor"] = CommonExecutor
	mapExecuteName["intervalexecutor"] = IntervalExecutor
	mapExecuteName["NormalExecutor"] = CommonExecutor
	mapExecuteName["EvadeExecutor"] = EvadeExecutor
	mapExecuteName["RepelExecutor"] = RepelExecutor
	mapExecuteName["BulletExecutor"] = BulletExecutor
	mapExecuteName["TrajactoryExecutor"] = CommonExecutor
	mapExecuteName["BulletFalseConstantSpeedExecutor"] = CommonExecutor
	mapExecuteName["BulletFalseConstantTimeExecutor"] = CommonExecutor
	mapExecuteName["VirtualExecutor"] = VirtualExecutor
	mapExecuteName["CycleChain"] = CycleChainExecutor
	mapExecuteName["HelicalBulletExecutor"] = BulletExecutor
	mapExecuteName["PassiveExecutor"] = PassiveExecutor
	mapExecuteName["ConditionDecoExecutor"] = ConditionDecoExecutor

	mapExecuteRealName["NormalExecutor"] = RealTypeNormalExecutor
	mapExecuteRealName["EvadeExecutor"] = RealTypeEvadeExecutor
	mapExecuteRealName["RepelExecutor"] = RealTypeRepelExecutor
	mapExecuteRealName["BulletExecutor"] = RealNameTypeBulletExecutor
	mapExecuteRealName["HelicalBulletExecutor"] = RealNameTypeBulletExecutor
	mapExecuteRealName["TrajactoryExecutor"] = RealTypeTrajactoryExecutor
	mapExecuteRealName["CycleChain"] = RealTypeCycleChain
	mapExecuteRealName["BulletFalseConstantSpeedExecutor"] = RealTypeBulletFalseConstantSpeedExecutor
	mapExecuteRealName["BulletFalseConstantTimeExecutor"] = RealTypeBulletFalseConstantTimeExecutor
	mapExecuteRealName["VirtualExecutor"] = RealTypeVirtualExecutor
	mapExecuteRealName["CycleChain"] = RealTypeCycleChainExecutor
	mapExecuteRealName["PassiveExecutor"] = RealTypePassiveExecutor
	mapExecuteRealName["ConditionDecoExecutor"] = ConditionDecoExecutor

	mapCheckRangeSwitch = make(map[ExecutorNameType]bool)
	mapCheckRangeSwitch[CommonExecutor] = false
	mapCheckRangeSwitch[IntervalExecutor] = false
	mapCheckRangeSwitch[EvadeExecutor] = false //执行器内部再进行验算,这里不打开
	mapCheckRangeSwitch[RepelExecutor] = false //执行器内部再进行验算,这里不打开
	mapCheckRangeSwitch[BulletExecutor] = false
	mapCheckRangeSwitch[VirtualExecutor] = false //虚拟执行器不需要验证——一定不要填ture
	mapCheckRangeSwitch[CycleChainExecutor] = false
	mapCheckRangeSwitch[PassiveExecutor] = false
	mapCheckRangeSwitch[ConditionDecoExecutor] = false

	mapMoveBehaviorRestrict[Dizzy] = true
	mapMoveBehaviorRestrict[Frozen] = true
	mapMoveBehaviorRestrict[Afraid] = false
	mapMoveBehaviorRestrict[Silent] = false
	mapMoveBehaviorRestrict[ForbidMove] = true
	mapMoveBehaviorRestrict[BeRidiculed] = false
	mapMoveBehaviorRestrict[HitDown] = true

	mapSkillBehaviorRestrict[Dizzy] = true
	mapSkillBehaviorRestrict[Frozen] = true
	mapSkillBehaviorRestrict[Afraid] = false
	mapSkillBehaviorRestrict[Silent] = false
	mapSkillBehaviorRestrict[ForbidMove] = false
	mapSkillBehaviorRestrict[BeRidiculed] = false
	mapSkillBehaviorRestrict[HitDown] = true

	mapSwitchSummonPetBehaviorRestrict[Dizzy] = true
	mapSwitchSummonPetBehaviorRestrict[HitDown] = true

	mapClientCarePassive[PTChangeAttribute] = struct{}{}
	mapClientCarePassive[PTSkill] = struct{}{}
	mapClientCarePassive[PTSkillEffectAttr] = struct{}{}
	mapClientCarePassive[PTSkillForcedRelease] = struct{}{}
	mapClientCarePassive[PTAttrTransfer] = struct{}{}
	mapClientCarePassive[PTAttributeFixed] = struct{}{}
	mapClientCarePassive[PTHalo] = struct{}{}

	mapNoCostHomeResoures[EHRUpgradeMaterial] = struct{}{}
	mapNoCostHomeResoures[EHRCoin] = struct{}{}

	mapFightNotifyAttribute[AttributeTypeHit] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeHitValue] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeDodge] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeDodgeValue] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeMaxHp] = struct{}{}
	mapFightNotifyAttribute[AttributeHitParam] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeEvadeDistance] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeBulletSize] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeBulletCount] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeLinkSkillRange] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeLinkSkillTargetCount] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeDefenseSkillUseCount] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeDefenseSkillEvadeDistance] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeDefenseSkillInvincibleTime] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeRangeSkillRange] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeShieldTimes] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeShieldHp] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeLevel] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeMaxShieldHp] = struct{}{}
	mapFightNotifyAttribute[PlayerAttributeTypePhysique] = struct{}{}
	mapFightNotifyAttribute[PlayerAttributeTypeStrength] = struct{}{}
	mapFightNotifyAttribute[PlayerAttributeTypeIntellect] = struct{}{}
	mapFightNotifyAttribute[PlayerAttributeTypeDexterous] = struct{}{}
	mapFightNotifyAttribute[PlayerAttributeTypeStamina] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeSpeed] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeNowMana] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeMaxMana] = struct{}{}
	mapFightNotifyAttribute[AttributeTypeManaReply] = struct{}{}

	// 资质转换目标属性
	mapPotentialTypeToAttribute[PTypeHp] = AttributeTypeMaxHp
	mapPotentialTypeToAttribute[PTypePhysicsAttack] = AttributeTypeAttack
	mapPotentialTypeToAttribute[PTypeMagicAttack] = AttributeTypeMagicAttack
	mapPotentialTypeToAttribute[PTypePhysicsDefense] = AttributeTypeDefense
	mapPotentialTypeToAttribute[PTypeMagicDefense] = AttributeTypeMagicDefense
}

func GetExecutorNeedCheckRange(typeValue ExecutorNameType) bool {
	checkSwitch, ok := mapCheckRangeSwitch[typeValue]
	if ok {
		return checkSwitch
	}

	return true
}

func GetExecutorNameType(name string) ExecutorNameType {
	typ, ok := mapExecuteName[name]
	if ok == false {
		return NoneExecutor
	}

	return typ
}

func GetExecutorRealNameType(name string) ExecutorRealNameType {
	typ, ok := mapExecuteRealName[name]
	if ok == false {
		return RealTypeNoneExecutor
	}

	return typ
}

// CheckMoveBehaviorRestrict 验证该行为是否限制移动
func CheckMoveBehaviorRestrict(er EBehaviorRestrict) bool {
	isRestrict, ok := mapMoveBehaviorRestrict[er]
	if ok == false {
		return false
	}

	return isRestrict
}

// CheckSkillBehaviorRestrict 验证该行为是否限制释放技能
func CheckSkillBehaviorRestrict(er EBehaviorRestrict) bool {
	isRestrict, ok := mapSkillBehaviorRestrict[er]
	if ok == false {
		return false
	}

	return isRestrict
}

// CheckSwitchSummonPetBehaviorRestrict 验证该行为是否限制切换幻兽
func CheckSwitchSummonPetBehaviorRestrict(er EBehaviorRestrict) bool {
	isRestrict, ok := mapSwitchSummonPetBehaviorRestrict[er]
	if ok == false {
		return false
	}

	return isRestrict
}

func GetMapFightNotifyAttribute() map[AttributeType]struct{} {
	return mapFightNotifyAttribute
}

func GetAttributeNeedNotify(attrType AttributeType) bool {
	_, ok := mapFightNotifyAttribute[attrType]
	return ok
}

// 根据资质类型获取转换的目标属性
func GetPotentialTypeToAttribute(pType PotentialType) (bool, AttributeType) {
	attType, ok := mapPotentialTypeToAttribute[pType]

	return ok, attType
}

type MapExecutorType int

const MinStateSkillMonsterGroupId = 1000
const MinStateSkillMonsterBrushId = 2000000
const (
	Root      MapExecutorType = 0 //0_Root
	TimeSleep MapExecutorType = 1 //1_时间延迟执行器
	//缺少1
	DieSettlement         MapExecutorType = 3 //3_死亡结算执行器_已删除
	FinishSettlement      MapExecutorType = 4 //4_完成结算执行器_已删除
	KillMonsterCount      MapExecutorType = 5 //5_杀怪计数执行器
	ChangeWeatherExecutor MapExecutorType = 6 //6_改变天气
	//ChangeElementExecutor   MapExecutorType = 7  //7_改变元素
	PlayStoryExecutor       MapExecutorType = 8  //8_动画演绎事件
	AiWaitExecutor          MapExecutorType = 9  //9_AI等待
	RefreshNpcExecutor      MapExecutorType = 10 //10_刷NPC执行器
	RemoveNpcExecutor       MapExecutorType = 11 //11_删除NPC执行器
	PlayTalkExecutor        MapExecutorType = 12 //12_开始对话执行器
	WaitTalkEndExecutor     MapExecutorType = 13 //13_等待对话结束执行器
	StorySettlementExecutor MapExecutorType = 14 //14_剧情结算执行器_已删除
	WaitStoryEndExecutor    MapExecutorType = 15 //15_等待动画演绎事件结束
	SettlementExecutor      MapExecutorType = 16 //16_结算执行器
	AndExecutor             MapExecutorType = 17 //17_逻辑与执行器
	OrExecutor              MapExecutorType = 18 //18_逻辑或执行器
	ForCountExecutor        MapExecutorType = 19 //19_计数执行器
	BrushMonstersExecutor   MapExecutorType = 20 //20_刷怪组执行器

	SetDoorExecutor         MapExecutorType = 21 //21_设置门开关执行器
	GotoRoomExecutor        MapExecutorType = 22 //22_走去房间执行器
	HotAreaExecutor         MapExecutorType = 23 //23_热点范围触发执行器
	RemoveMonsterExecutor   MapExecutorType = 24 //24_删除怪物执行器
	TransferToRoom          MapExecutorType = 25 //25_传送到指定房间
	JudgeNotPassMapExecutor MapExecutorType = 26 //26_判断是否通达关卡
	//JudgeTaskStatus         MapExecutorType = 27 //27_判断任务状态

	ControlMonsterAIExecutor       MapExecutorType = 28 //28_控制怪物AI节点执行器
	CompareMonsterCount            MapExecutorType = 29 //29_怪物数比较执行器
	MusicExecutor                  MapExecutorType = 30 //30_音乐设置执行器
	HpCompare                      MapExecutorType = 31 //31_生命比较执行器
	DynamicBlockExecutor           MapExecutorType = 32 //32_动态阻挡执行器
	BrushMonsterFormationsExecutor MapExecutorType = 33 //33_刷怪阵型执行器
	AttrPlayerCompareExecutor      MapExecutorType = 34 //34_属性判断执行器
	//缺少1
	RandomRatioExecutor       MapExecutorType = 36 //36_概率判断执行器
	LinkSkillUseCountExecutor MapExecutorType = 37 //37_连携技使用执行器
	//缺少1
	DamageExecutor                 MapExecutorType = 39 //39_伤害量统计执行器
	RefreshPetNpcExecutor          MapExecutorType = 40 //40_刷新宠物Npc执行器
	PetNpcAIControlExecutor        MapExecutorType = 41 //41_PetNpc的AI控制
	PetNpcHpCompareExecutor        MapExecutorType = 42 //42_PetNpc生命比较执行器
	SuperSleepExecutor             MapExecutorType = 43 //43_Super定时器(可被取消)
	CancelSuperSleepExecutor       MapExecutorType = 44 //44_取消Super定时器
	ShowMessageIdExecutor          MapExecutorType = 45 //45_显示消息文本
	BlackBrushMonsterExecutor      MapExecutorType = 46 //46_按黑板数据刷怪
	CapturedMonsterExecutor        MapExecutorType = 47 //47_抓宠执行器
	ReplaceSummonAssistPetExecutor MapExecutorType = 48 //48_临时替换出战或连携宠物执行器
	BrushInteractiveExecutor       MapExecutorType = 49 //49_刷交互对象执行器

	PetNpcDeathExecutor             MapExecutorType = 50 //50_NPC宠物死亡
	PetNpcArriveTargetPointExecutor MapExecutorType = 51 //51_NPC宠物是否到达目标点
	ChargeEnergyRatioExecutor       MapExecutorType = 52 //52_对象充能比例执行器

	EnergyMonsterDeathExecutor      MapExecutorType = 53 //53_怪物死亡充能执行器
	EnergyMonsterBeAttackedExecutor MapExecutorType = 54 //54_充能怪物被攻击充能
	MutableSleepExecutor            MapExecutorType = 55 //55_可变睡眠时间执行器
	ChangeMutableSleepExecutor      MapExecutorType = 56 //56_修改可变睡眠时间执行器
	CancelMutableSleepExecutor      MapExecutorType = 57 //57_取消可变睡眠时间执行器
	MonitorTreasureMapExecutor      MapExecutorType = 58 //58_监听宝图信息执行器
	MonitorInteractiveExecutor      MapExecutorType = 59 //59_监听交互执行器

	ShipBattlePlayerFormationExecutor   MapExecutorType = 60  // 60_玩家上阵下一个舰船阵型执行器
	ShipBattleAllDieExecutor            MapExecutorType = 61  // 61_一个阵型全死判断执行器
	ShipBattleHasFormationExecutor      MapExecutorType = 62  // 62_是否还有舰船阵型判断执行器
	ShipBattleHasEquipPetExecutor       MapExecutorType = 63  // 63_舰船阵型中是否有连携宠物判断执行器
	TunTunRatThrowNutCountExecutor      MapExecutorType = 64  // 64_屯屯鼠丢坚果次数判断执行器
	TunTunRatPickUpFinishExecutor       MapExecutorType = 65  // 65_屯屯鼠捡坚果完成判断执行器
	BrushMonstersPlayerAroundExecutor   MapExecutorType = 66  // 66_玩家周围刷怪组执行器
	BrushMonstersRandPosExecutor        MapExecutorType = 67  // 67_指定位置列表随机刷怪组执行器
	DandelionDeathExecutor              MapExecutorType = 68  // 68_蒲公英全部死亡执行器
	SwitchExecutor                      MapExecutorType = 69  // 69_分支执行器
	ModifyNickNameExecutor              MapExecutorType = 70  // 70_修改昵称执行器
	CameraFocusObjectExecutor           MapExecutorType = 71  // 71_摄像机聚焦目标执行器
	RecoverCameraFocusObjectExecutor    MapExecutorType = 72  // 72_恢复摄像机目标执行器
	StopClientAiExecutor                MapExecutorType = 73  // 73_暂停客户端AI执行器
	RecoverClientAiExecutor             MapExecutorType = 74  // 74_恢复客户端AI执行器
	WaitCapturedFinishExecutor          MapExecutorType = 75  // 75_等待抓宠完成执行器
	ReplaceBlackAssistPetExecutor       MapExecutorType = 76  // 76_临时替换黑板宠物连携执行器
	RefreshCapturedPetExecutor          MapExecutorType = 77  // 77_刷新抓宠宠物执行器
	NotifyClientLogicExecutor           MapExecutorType = 78  // 78_通知客户端自定义逻辑执行器
	MonsterCountExecutor                MapExecutorType = 79  // 79_怪物总数执行器
	MonsterBrushCountExecutor           MapExecutorType = 80  // 80_判断已刷怪数量执行器
	TimeGradeExecutor                   MapExecutorType = 81  // 81_时间评级执行器
	TargetSubHpGradeExecutor            MapExecutorType = 82  // 82_目标血量评级执行器
	KillMonsterNumGradeExecutor         MapExecutorType = 83  // 83_杀怪数量评级执行器
	MonsterKillCountExecutor            MapExecutorType = 84  // 84_刷怪组杀怪通知计数器
	InteractiveCountExecutor            MapExecutorType = 85  // 85_交互对象数量执行器
	BrushMirrorPlayerExecutor           MapExecutorType = 86  // 86_刷镜像玩家阵容执行器
	MirrorPlayerDeathExecutor           MapExecutorType = 87  // 87_监听镜像状态执行器
	MapTriggerEventExecutor             MapExecutorType = 88  // 88_地图事件执行器
	MonsterBatchCountExecutor           MapExecutorType = 89  // 89_怪物总波数执行器
	BrushMonsterGeneralExecutor         MapExecutorType = 90  // 90_通用阵型刷怪执行器
	SceneObjectStatusChangeExecutor     MapExecutorType = 91  // 91_场景对象状态切换执行器
	SceneObjectStatusSyncExecutor       MapExecutorType = 92  // 92_场景对象状态同步执行器
	MonitorTargetHpExecutor             MapExecutorType = 93  // 93_监控目标血量执行器
	StopAllAiExecutor                   MapExecutorType = 94  // 94_暂停全部AI执行器
	IntervalBrushMonsterGeneralExecutor MapExecutorType = 95  // 95_通用阵型间隔时间刷怪执行器
	MainLineObjectiveFinishExecutor     MapExecutorType = 96  // 96_主线目标完成执行器
	BuffControlExecutor                 MapExecutorType = 97  // 97_BUFF控制执行器(目前只控制怪物)
	MonitorTargetBuffExecutor           MapExecutorType = 98  // 98_监听目标BUFF执行器
	MonitorObjectiveDieExecutor         MapExecutorType = 99  // 99_监听目标死亡执行器
	BrushPetBattleMirrorPlayerExecutor  MapExecutorType = 100 // 100_刷新幻兽对决对手玩家执行器

	//_EXECUTOR_CODE_STUB
	//这行别删，也别在这行下面加新的 Executor代码

	ExecutorTypeHolder
	MaxMapExecutorType = ExecutorTypeHolder + 1 //max值
)

type TrapTriggerMethod int32

const (
	TrouchDisappear TrapTriggerMethod = 1 //触碰即消失
	TickTrigger     TrapTriggerMethod = 2 //脉冲触发
	TimeoutTrigger  TrapTriggerMethod = 4 //到期触发后，然后消失
)

type WeatherType = uint32

const (
	WTNone    WeatherType = 0  //0_无
	WTSunny   WeatherType = 1  //1_晴天
	WTRainy   WeatherType = 2  //2_雨天
	WTThunder WeatherType = 4  //4_打雷
	WTGale    WeatherType = 8  //8_大风
	WTDrought WeatherType = 16 //16_干旱
)

var AllWeatherTypeList = []WeatherType{WTNone, WTSunny, WTRainy, WTThunder, WTGale, WTDrought}

// 初始类型
type PassiveType = int

const (
	PTNone            PassiveType = 0
	PTChangeAttribute PassiveType = 1 //修改属性类被动

	PTCastSkill   PassiveType = 2 //释放技能触发
	PTBeAttack    PassiveType = 3 //被攻击时触发
	PTHitTarget   PassiveType = 4 //命中触发
	PTNoHitTarget PassiveType = 5 //未命中触发
	PTCritical    PassiveType = 6 //暴击触发
	PTNoCritical  PassiveType = 7 //未暴击触发
	PTDodge       PassiveType = 8 //闪避触发

	PTAttackDistance PassiveType = 11 //攻击距离类被动:距离不同伤害属性有所变化
	PTSkill          PassiveType = 12 //修改技能类被动:修改技能CD，技能SP获取量，技能sp消耗量，技有分辨率或者值

	PTChangeBuff    PassiveType = 13 //修改Buff，对持续时间，分辨率修改，生效几率增加
	PTSuckBlood     PassiveType = 14 //吸血被动:攻击命中/暴击命中时，吸血值为伤害的万分比 ——废弃
	PTReboundDamage PassiveType = 15 //反弹伤害，受伤害会按比例反弹一定伤害
	PTDieDelay      PassiveType = 16 //死亡延时
	PTKillTarget    PassiveType = 17 //击杀目标时产生被动效果
	PTShareDamage   PassiveType = 18 //分担伤害
	PTRevive        PassiveType = 19 //复活被动

	PTAttributeCondition  PassiveType = 20 //属性条件被动
	PTBuffStatusCondition PassiveType = 21 //状态条件判断
	PTSkillStatus         PassiveType = 22 //技能状态判断被动
	PTTargetCount         PassiveType = 23 //目标数量判断被动
	PTBuffCountCondition  PassiveType = 24 //状态数量判断被动
	PTCounteredDamage     PassiveType = 25 //反击被动
	//PTTableStatusExtend        PassiveType = 26 //招格闪扩展被动
	PTSkillEffectAttr          PassiveType = 27 //受技能攻击影响招格闪被动
	PTSkillForcedRelease       PassiveType = 28 //客户端强制释放技能被动
	PTHpShieldChange           PassiveType = 29 //血量护盾增多被动
	PTAttrTransfer             PassiveType = 30 //属性转移被动
	PTTiming                   PassiveType = 31 //定时被动
	PTHalo                     PassiveType = 32 //光环被动
	PTEnterMapRecoverySp       PassiveType = 33 //进入地图恢复Sp被动
	PTAttributeFixed           PassiveType = 34 //属性固定被动
	PTChangeSummonPetAttribute PassiveType = 35 //修改召唤宠物属性被动——仅玩家生效
	//PTStatisticalChangeAttribute PassiveType = 36 //统计元素修改属性类被动
	PTDamageCorrect           PassiveType = 37 //伤害修正被动
	PTBuffDispel              PassiveType = 38 //buff净化驱散被动——驱散or净化buff时触发
	PTSwitchBattle            PassiveType = 39 //出战备战被动
	PTHurt                    PassiveType = 40 //受到伤害被动
	PTChangeAttributeNihility PassiveType = 41 //虚无之地效果层数修改属性被动
	PTFixedSubHp              PassiveType = 42 //固定扣血被动
	PTAfterCastSkill          PassiveType = 43 //释放技能后触发
	PTSwitchBattleLinkCost    PassiveType = 44 //上场宠物连携技消耗被动 ps:该被动只能存在于宠物身上
	PTCastSkillCountAttr      PassiveType = 45 //释放技能次数提升属性
	PTUpSkillLevel            PassiveType = 46 //修改技能等级被动
	PTEnterMap                PassiveType = 47 //进入地图触发被动
	PTAttackCount             PassiveType = 48 //攻击(普通执行器执行buff与event)次数触发被动
	PTJudgeDamageValue        PassiveType = 49 //伤害值判断被动
	PTMoveSpeedDown           PassiveType = 50 //移动受损被动

	PTMax
)

// 技能被动来源类型
type SkillPassiveFromType = int8

const (
	SPFTSelf  SkillPassiveFromType = 0 // 技能被动来源自身
	SPFTOther SkillPassiveFromType = 1 // 技能被动来源其他（天赋、装备、血统、因子）

	SPFTEnd
	SPFTMax = SPFTEnd + 1
)

var mapClientCarePassive map[PassiveType]struct{}

func CheckPassiveTypeClientCare(typeId PassiveType) bool {
	_, ok := mapClientCarePassive[typeId]
	return ok
}

type PassiveTargetType = int

const (
	PassiveTargetTypeSelf               PassiveTargetType = 1     //1_被动拥有者
	PassiveTargetTypeEnemy              PassiveTargetType = 2     //2_被动拥有者的敌方阵营
	PassiveTargetTypeFriends            PassiveTargetType = 4     //4_被动拥有者的友方阵营
	PassiveTargetTypeAttack             PassiveTargetType = 8     //8_被动拥有者的攻击方 打我的那个对象
	PassiveTargetTypeAttackTarget       PassiveTargetType = 16    //16_被动拥有者的被攻击方 我打的那个对象
	PassiveTargetTypeMaster             PassiveTargetType = 32    //32_被动拥有者的主人
	PassiveTargetTypeOtherPet           PassiveTargetType = 64    //64_被动拥有者的其他宠物——出战+备战
	PassiveTargetTypeAllPet             PassiveTargetType = 128   //128_被动拥有者的宠物——出战+备战
	PassiveTargetTypeOtherBattleZonePet PassiveTargetType = 256   //256_被动拥有者的其他出战宠物
	PassiveTargetTypeAllBattleZonePet   PassiveTargetType = 512   //512_被动拥有者的出战宠物
	PassiveTargetTypeOtherBackgroundPet PassiveTargetType = 1024  //1024_被动拥有者的其他备战宠物
	PassiveTargetTypeAllBackgroundPet   PassiveTargetType = 2048  //2048_被动拥有者的备战宠物
	PassiveTargetTypeRandomOneFriend    PassiveTargetType = 4096  //4096_被动拥有者的随机一个友方单位
	PassiveTargetTypeRandomTwoFriend    PassiveTargetType = 8192  //8192_被动拥有者的随机两个友方单位
	PassiveTargetTypeRandomThreeFriend  PassiveTargetType = 16384 //16384_被动拥有者的随机三个友方单位
)

const CanUseAllTargetType = PassiveTargetTypeFriends | PassiveTargetTypeSelf | PassiveTargetTypeOtherPet | PassiveTargetTypeAllPet | PassiveTargetTypeOtherBattleZonePet |
	PassiveTargetTypeAllBattleZonePet | PassiveTargetTypeOtherBackgroundPet | PassiveTargetTypeAllBackgroundPet | PassiveTargetTypeRandomOneFriend |
	PassiveTargetTypeRandomTwoFriend | PassiveTargetTypeRandomThreeFriend

type EPetBattleStatus = int8

const (
	EPetBattleAll        EPetBattleStatus = 0 //所有使用方式
	EPetBattleIn         EPetBattleStatus = 1 //出战区
	EPetBattleBackground EPetBattleStatus = 2 //备战区
)

// 出战备战被动触发类型
type ESwitchBattleTriggerType = int

const (
	ESwitchNone         ESwitchBattleTriggerType = 0 //无操作标记
	EBattleTriggerType  ESwitchBattleTriggerType = 1 //1_出战触发
	EPrepareTriggerType ESwitchBattleTriggerType = 2 //2_备战触发
	EDieTriggerType     ESwitchBattleTriggerType = 4 //4_死亡触发
)

// 结算类型
type SettlementType = int32

const (
	STNone    SettlementType = 0 //无
	STDie     SettlementType = 1 //死亡结算
	STFinish  SettlementType = 2 //完成结算
	STStory   SettlementType = 3 //剧情结算
	STTimeOut SettlementType = 4 //超时结算
	STFailed  SettlementType = 5 //失败结算
)

// 击杀数量类型
type KillCountType = int

const (
	KCTNone         KillCountType = 0 //无
	KCTMonsterGroup KillCountType = 1 //按怪物组
	KCTMonsterId    KillCountType = 2 //按怪物ID
	KCTMonsterBrush KillCountType = 3 //按刷怪组ID
	KCTMonsterType  KillCountType = 4 //按刷怪类型
	KCTMonsterCamp  KillCountType = 5 //按阵营类型
)

// 技能被动影响属性类型
type SkillPassiveAttrType = int

const (
	SPATCd          SkillPassiveAttrType = 0 //0_技能CD
	SPATSpGet       SkillPassiveAttrType = 1 //1_技能SP获取
	SPATRatioDamage SkillPassiveAttrType = 2 //2_技能倍率
	SPATFixDamage   SkillPassiveAttrType = 3 //3_技能固定伤害
	SPATSetMustHit  SkillPassiveAttrType = 4 //4_设置必定命中
	SPATSetMustCt   SkillPassiveAttrType = 5 //5_设置必定暴击
	SPATManaCost    SkillPassiveAttrType = 6 //6_技能mana消耗修正
	SPATRandomCd    SkillPassiveAttrType = 7 //7_修改技能随机冷却时间
	SPATHpCost      SkillPassiveAttrType = 8 //8_技能hp消耗修正
)

// 技能条件被动-状态类型
type SkillPassiveStatusType = int

const (
	SPSTNone     SkillPassiveStatusType = 0 //0_无
	SPSTCDIng    SkillPassiveStatusType = 1 //1_技能在CD
	SPSTNoCDing  SkillPassiveStatusType = 2 //2_技能未在CD
	SPSTSPFull   SkillPassiveStatusType = 3 //3_技能Sp满
	SPSTSPNoFull SkillPassiveStatusType = 4 //4_技能Sp未满
)

// 天赋流派类型
type TalentSectType = int32

const (
	TCNone    TalentSectType = 0 //0_无
	TCSectFir TalentSectType = 1 //1_流派1
	TCSectSec TalentSectType = 2 //2_流派2
)

// 天赋类型
type TalentType int32

const (
	TTNone         TalentType = 0 //0_无
	TTSectTalent   TalentType = 1 //1_流派天赋
	TTCoreTalent   TalentType = 2 //2_核心天赋
	TTMainTalent   TalentType = 3 //3_主要天赋
	TTNormalTalent TalentType = 4 //4_普通天赋
)

// 天赋流派类型
type TalentClassicSectType int32

const (
	TCCNone    TalentClassicSectType = 0 //0_无
	TCCPublic  TalentClassicSectType = 1 //1_公共
	TCCSectFir TalentClassicSectType = 2 //2_流派1
	TCCSectSec TalentClassicSectType = 3 //3_流派2
)

// 经典天赋类型
type TalentClassicType int32

const (
	TCTNone         TalentClassicType = 0 //0_无
	TCTSectTalent   TalentClassicType = 1 //1_流派天赋
	TCTCoreTalent   TalentClassicType = 2 //2_核心天赋
	TCTMainTalent   TalentClassicType = 3 //3_主要天赋
	TCTNormalTalent TalentClassicType = 3 //4_普通天赋
)

// 职业天赋类型
type TamerTalentType int32

const (
	TTTNone         TamerTalentType = 0 //0_无
	TTTSecTalent    TamerTalentType = 1 //1_流派天赋
	TTTCoreTalent   TamerTalentType = 2 //2_核心天赋
	TTTMainTalent   TamerTalentType = 3 //3_主要天赋
	TTTNormalTalent TamerTalentType = 4 //4_普通天赋
)

// 宠物进化消耗类型
type EvolutionCostPetType int32

const (
	EPTNone         EvolutionCostPetType = 0 //0_无
	EPTEqualID      EvolutionCostPetType = 1 //1_同ID宠物
	EPTEqualElement EvolutionCostPetType = 2 //1_同元素宠物
)

// 技能组类型
type ESkillGroupType int

const (
	ESGTNormal  ESkillGroupType = 0 //0_普通技能组
	ESGTSpecial ESkillGroupType = 1 //1_特殊技能组
)

// 距离判断类型
type EDistanceTargetType int32

const (
	DTTMaster      EDistanceTargetType = 0 //0_主人
	DTTTargetEnemy EDistanceTargetType = 1 //1_当前目标敌人
	DTTAlertEnemy  EDistanceTargetType = 2 //2_警戒范围内敌人
)

// AI节点开启状态类型
type EAiNodeEnableType int32

const (
	ANENone    EAiNodeEnableType = 0 //0_无
	ANEEnable  EAiNodeEnableType = 1 //1_激活
	ANEDisable EAiNodeEnableType = 2 //2_关闭
)

type ETaskType = int32

const (
	TaskTypeDaily     ETaskType = 1 //1_每日任务
	TaskTypeWeek      ETaskType = 2 //2_每周任务
	TaskTypeEternal   ETaskType = 3 //3_成就任务
	TaskTypePetGet    ETaskType = 4 //4_幻兽获取任务
	TaskTypePetFoster ETaskType = 5 //5_幻兽养成任务
)

const MaxScheduleDecodeIdArray = 100                               //派遣识别ID数组大小 识别id最大为100*8
const MaxScheduleGetCountMaskArray = 1250                          //派遣事件次数控制掩码数组；最大可控制数量的派遣事件个数=MaxScheduleCountArray*2
const MaxSchedulePoolMaskArray = 125                               //派遣事件池数量控制掩码数; 最大数量125*8
const MaxSchedulePoolMaskId = MaxSchedulePoolMaskArray * 8         //最大一次性事件池掩码Id
const MaxScheduleGetCountMaskId = MaxScheduleGetCountMaskArray * 2 //最大派遣事件次数控制掩码ID
const MaxDispatchPetCount = 5                                      //派遣最大宠物个数
const MaxSchedulePetTime = int64(3600000 * 8)                      //最大宠物派遣奖励时间8小时
const MaxSchedulePool = 25                                         //派遣事件池最大数量限制
const MaxOfflineScheduleDay = 3                                    //派遣离线事件最多刷新天数
const MaxOfflineScheduleWeek = 3                                   //派遣离线事件最多刷新周数
const MaxOfflineScheduleExecutorCount = 8                          //离线刷新事件执行次数
const MinOfflineTimeOffset = 10000                                 //离线刷新时间偏差
const MaxOfflineTimeOffset = 30000
const MaxTriggerRelationScheduleCount = 5      //最大触发关联事件数量
const MaxTriggerRelationScheduleLoopCont = 100 //触发关联事件最大循环次数限制
const MaxScheduleGroupLoopCount = 50           //触发事件组最大循环次数限制
const MaxRefreshScheduleLoopCount = 50         //刷新派遣事件最大循环次数限制
const MaxDailyScheduleCountLimit = 512         //每日派遣类型完成次数上限
const MaxScheduleTime = 7 * 86400000           //最大派遣时间7天
const MaxScheduleResidentAwardCount = 8        //常驻事件最大领取奖励次数
const MaxScheduleManualRefreshCount = 1000     //派遣事件手动刷新次数限制

// 稀有度类型
type ERarityType = int32

const (
	ERarityUnknown  ERarityType = 0 //0_未知稀有度
	ERarityUbiquity ERarityType = 1 //1_无处不在
	ERarityCommon   ERarityType = 2 //2_常见
	ERarityRarely   ERarityType = 3 //3_罕见
	ERarityScarce   ERarityType = 4 //4_稀有
	ERaritySScarce  ERarityType = 5 //5_超稀有
	ERarityEpic     ERarityType = 6 //6_史诗
	ERarityMyth     ERarityType = 7 //7_神话

	ERarityEnd
	ERarityMax = ERarityEnd + 1
)

// 检测宠物品质是否正确
func CheckIsRarityType(rarityType ERarityType) bool {
	if rarityType <= ERarityUnknown || rarityType > ERarityEnd {
		return false
	}
	return true
}

// 奇遇宠物随机类型
type EAdventurePetRandType = uint8

const (
	EAPTNone     EAdventurePetRandType = 0 //0_未知类型
	EAPTRandom   EAdventurePetRandType = 1 //1_父事件宠物纯随机
	EAPTPriority EAdventurePetRandType = 2 //2_父事件宠物优先级
)

// 事件触发关联事件类型
type ETriggerRelationScheduleType = int8

const (
	TriggerMultiRelationUnknown       ETriggerRelationScheduleType = 0 //0_未知
	TriggerMultiRelationSchedule      ETriggerRelationScheduleType = 1 //1_多个事件
	TriggerContinuousRelationSchedule ETriggerRelationScheduleType = 2 //2_连续事件

	TriggerScheduleEnd
	TriggerScheduleMax = TriggerScheduleEnd + 1
)

// 检测是否是事件触发关联事件类型
func CheckIsETriggerRelationScheduleType(curType ETriggerRelationScheduleType) bool {
	if curType <= TriggerMultiRelationUnknown || curType > TriggerScheduleEnd {
		return false
	}

	return true
}

// 派遣事件消失类型
type EScheduleDisappearType uint8

const (
	EScheduleFinishDisappear EScheduleDisappearType = 0 //0_完成消失
	EScheduleNoneDisappear   EScheduleDisappearType = 1 //1_完成不消失
)

// 派遣事件触发类型
type EScheduleTriggerType uint8

const (
	//注:新增触发类型时注意 CheckCanTriggerRelationScheduleByTriggerType 函数
	//0_默认获得 1_系统触发 2_探索触发 可触发关联事件
	EScheduleTriggerNone    EScheduleTriggerType = 0 //0_默认获得
	EScheduleTriggerSystem  EScheduleTriggerType = 1 //1_系统触发
	EScheduleTriggerExplore EScheduleTriggerType = 2 //2_识别触发
	//只有 3_事件触发 才能触发关联事件的连续事件
	EScheduleTriggerEvent EScheduleTriggerType = 3 //3_事件触发
)

// 根据事件触发类型判断是否可以触发关联事件
func CheckCanTriggerRelationScheduleByTriggerType(triggerType EScheduleTriggerType) bool {
	if triggerType < EScheduleTriggerNone || triggerType > EScheduleTriggerExplore {
		return false
	}

	return true
}

// 派遣事件执行类型
type EScheduleExecutorType uint8

const (
	EScheduleExecutorDispatch EScheduleExecutorType = 0 //0_派遣执行
	EScheduleExecutorPick     EScheduleExecutorType = 1 //1_点击执行
	EScheduleExecutorFight    EScheduleTriggerType  = 2 //2_进入副本
)

// 派遣事件完成条件类型
type ESCFinishType uint8

const (
	ESCFinishNone            ESCFinishType = 0 //0_无条件
	ESCFinishLevelFinish     ESCFinishType = 1 //1_关卡通关（成功）
	ESCFinishLevelSettlement ESCFinishType = 2 //2_关卡结算（成功、失败）
	ESCFinishLevelEnd        ESCFinishType = 3 //3_关卡结束（成功、失败、退出）
	ESCFinishEnterCapture    ESCFinishType = 4 //4_开始抓宠
	ESCFinishLevelEnter      ESCFinishType = 5 //5_关卡进入

	ESCFinishHolder
	ESCFinishMax = ESCFinishHolder + 1
)

// 派遣事件主类型
type EScheduleMainType = uint8

const (
	EScheduleMainExplore  EScheduleMainType = 0 //0_探索大类
	EScheduleMainTask     EScheduleMainType = 1 //1_任务大类
	EScheduleMainFight    EScheduleMainType = 2 //2_战斗大类
	EScheduleMainCapture  EScheduleMainType = 3 //3_抓宠大类
	EScheduleMainSearch   EScheduleMainType = 4 //4_搜索大类
	EScheduleMainResource EScheduleMainType = 5 //5_资源大类
	EScheduleMainSystem   EScheduleMainType = 6 //6_系统大类
	EScheduleMainBuilding EScheduleMainType = 7 //7_建筑大类

	EScheduleMainTypeEnd
	EScheduleMainTypeMax = EScheduleMainTypeEnd + 1
)

// 派遣事件类型
type EScheduleType = int16

const (
	ESchedulePry               EScheduleType = 0   //0_打探消息
	EScheduleExplore           EScheduleType = 1   //1_内容探索
	EScheduleAwardExplore      EScheduleType = 2   //2_奖励探索
	EScheduleDecodeExplore     EScheduleType = 3   //3_识别探索
	EScheduleMainLineTask      EScheduleType = 100 //100_主线任务
	EScheduleSubLineTask       EScheduleType = 101 //101_支线任务
	EScheduleAdventure         EScheduleType = 102 //102_奇遇任务
	EScheduleStoryPick         EScheduleType = 103 //103_剧本点击
	EScheduleStoryTalk         EScheduleType = 104 //104_剧本对话
	EScheduleCoinTalk          EScheduleType = 105 //105_金币对话
	EScheduleEnergyTalk        EScheduleType = 106 //106_能量对话
	EScheduleBoxPick           EScheduleType = 107 //107_宝箱点击
	EScheduleCampTalk          EScheduleType = 108 //108_营地对话
	EScheduleEliminateTrouble  EScheduleType = 200 //200_消除隐患
	EScheduleFunnyFight        EScheduleType = 201 //201_趣味战斗
	ESchedulePressureFight     EScheduleType = 202 //202_压力战斗
	EScheduleThief             EScheduleType = 203 //203_盗贼副本
	EScheduleBattleDragon      EScheduleType = 204 //204_好斗龙
	EScheduleChallenge         EScheduleType = 205 //205_挑战副本
	EScheduleEquipTreasure     EScheduleType = 211 //211_装备藏宝区
	EScheduleCoinTreasure      EScheduleType = 212 //212_金币藏宝区
	EScheduleBloodlineTreasure EScheduleType = 213 //213_因子藏宝区
	EScheduleMedalTreasure     EScheduleType = 214 //214_勋章藏宝区
	EScheduleRelicTreasure     EScheduleType = 215 //215_遗物藏宝区
	EScheduleStoryLevel        EScheduleType = 221 //221_剧情副本
	EScheduleExpLevel          EScheduleType = 222 //222_经验副本
	EScheduleEnergyLevel       EScheduleType = 223 //223_能量副本
	EScheduleSkillBookLevel    EScheduleType = 224 //224_技能书副本
	EScheduleBloodlineLevel    EScheduleType = 225 //225_因子副本
	EScheduleEquipLevel        EScheduleType = 226 //226_装备副本
	EScheduleBoxLevel          EScheduleType = 227 //227_宝箱副本
	EScheduleCoinLevel         EScheduleType = 228 //228_金币副本
	EScheduleCaptureFight      EScheduleType = 300 //300_战斗抓宠
	EScheduleCaptureSoul       EScheduleType = 301 //301_灵魂回响
	EScheduleCaptureDraw       EScheduleType = 302 //302_抽卡抓宠
	EScheduleCaptureNaughty    EScheduleType = 303 //303_抓住那个调皮的家伙
	EScheduleFreelySearch      EScheduleType = 400 //400_自由搜索
	EScheduleTreasureSearch    EScheduleType = 401 //401_寻宝搜索
	ESchedulePickupSearch      EScheduleType = 402 //402_拾遗搜索
	EScheduleDeriveDandelion   EScheduleType = 403 //403_蒲蒲苗
	EScheduleBoxSearch         EScheduleType = 404 //404_宝箱搜索
	EScheduleStorySearch       EScheduleType = 405 //405_剧情搜索
	EScheduleGather            EScheduleType = 501 //501_采集
	EScheduleMining            EScheduleType = 502 //502_挖矿
	ESchedulePickup            EScheduleType = 503 //503_拾遗
	EScheduleEnergy            EScheduleType = 504 //504_能量
	EScheduleJellyTree         EScheduleType = 505 //505_果冻树
	EScheduleStreamTrain       EScheduleType = 506 //506_溪流训练
	EScheduleStory             EScheduleType = 507 //507_剧情
	EScheduleFortifiedStone    EScheduleType = 508 //508_强化石
	EScheduleSkillBook         EScheduleType = 509 //509_技能书
	EScheduleCoin              EScheduleType = 510 //510_金币
	EScheduleBloodline         EScheduleType = 511 //511_因子
	EScheduleEquip             EScheduleType = 512 //512_装备
	EScheduleDeriveTunTunRat   EScheduleType = 599 //599_屯屯鼠
	EScheduleReal              EScheduleType = 601 //601_真实
	EScheduleLogon             EScheduleType = 611 //611_登入
	ESchedulePetWorld          EScheduleType = 621 //621_幻兽世界
	EScheduleCircle            EScheduleType = 631 //631_跑环
	ESchedulePersistent        EScheduleType = 641 //641_长期
	EScheduleReserve           EScheduleType = 651 //651_储备
	EScheduleBusyPeople        EScheduleType = 661 //661_大忙人
	EScheduleExpeditions       EScheduleType = 671 //671_探险团
	EScheduleTickets           EScheduleType = 681 //681_门票

	//TODO 优化时拆分建筑
	EScheduleHomeStart EScheduleType = 701 //家园派遣开始标记
	EScheduleSpring                        //701_温泉
	ESchedulePlay      EScheduleType = 702 //702_乐园
	EScheduleTraining  EScheduleType = 703 //703_训练室
	EScheduleHomeEnd                       //家园派遣结束标记
	EScheduleHomeMax   = EScheduleHomeEnd + 1
)

// 派遣玩法类型
type ESchedulePlayType = int8

const (
	ESchedulePlayPry            ESchedulePlayType = 1  //1_打探
	ESchedulePlayExplore        ESchedulePlayType = 2  //2_探索
	ESchedulePlayAward          ESchedulePlayType = 3  //3_奖励
	ESchedulePlayDispatchPet    ESchedulePlayType = 4  //4_派遣宠物
	ESchedulePlayThief          ESchedulePlayType = 5  //5_盗贼
	ESchedulePlayBuilding       ESchedulePlayType = 7  //7_建筑派遣
	ESchedulePlayCaptureFight   ESchedulePlayType = 8  //8_战斗抓宠
	ESchedulePlayCaptureSoul    ESchedulePlayType = 9  //9_灵魂回响
	ESchedulePlayCaptureNaughty ESchedulePlayType = 10 //10_抓住那个调皮的家伙
	ESchedulePlayFight          ESchedulePlayType = 12 //12_关卡战斗
	ESchedulePlaySearch         ESchedulePlayType = 13 //13_关卡搜索
	ESchedulePlayResident       ESchedulePlayType = 14 //14_常驻
)

// 衍生事件类型
type EDerivationScheduleType = int32

const (
	EDerivationScheduleNone            EDerivationScheduleType = -1 //无效类型
	EDerivationScheduleFlyFish         EDerivationScheduleType = 0  //0_探索大类_小飞鱼
	EDerivationScheduleFourLeaf        EDerivationScheduleType = 1  //1_任务大类_四叶灵
	EDerivationScheduleBattleDragon    EDerivationScheduleType = 2  //2_战斗大类_好斗龙
	EDerivationScheduleGluttonousSnake EDerivationScheduleType = 3  //3_抓宠大类_贪吃蛇
	EDerivationScheduleDandelion       EDerivationScheduleType = 4  //4_搜索大类_蒲蒲苗
	EDerivationScheduleTunTunRat       EDerivationScheduleType = 5  //5_资源大类_屯屯鼠

	EDerivationScheduleTypeEnd
	EDerivationScheduleTypeMax = EDerivationScheduleTypeEnd + 1
)

// 蒲公英类型
type EDandelionType = int32

const (
	EDandelionTypeNone   EDandelionType = 0 //无效类型
	EDandelionTypeNormal EDandelionType = 1 //普通蒲公英
	EDandelionTypeRare   EDandelionType = 2 //稀有蒲公英
)

// 事件随机贪吃蛇状态
type EScheduleGluttonousSnakeStateType = int8

const (
	EScheduleGluttonousSnakeStateTypeNone EScheduleGluttonousSnakeStateType = 0 //没随机过
	EScheduleGluttonousSnakeStateTypeGet  EScheduleGluttonousSnakeStateType = 1 //随机到贪吃蛇
	EScheduleGluttonousSnakeStateTypeMiss EScheduleGluttonousSnakeStateType = 2 //未随机到贪吃蛇
)

// 时间类型
type ETimeType = int8

const (
	ETimeTypeNormal     ETimeType = 0 //0_默认时间
	ETimeTypeOpenServer ETimeType = 1 //1_开服时间
	ETimeTypeCreateRole ETimeType = 2 //2_创角时间
	ETimeTypeEveryWeek  ETimeType = 3 //3_每周几时间
	ETimeTypeEveryDay   ETimeType = 4 //4_每天几点时间
	ETimeTypeFixed      ETimeType = 5 //5_固定日期时间
	ETimeTypeDayCount   ETimeType = 6 //6_天数
)

type ETaskConditionType = int32

const (
	TaskConditionTypeLevel ETaskConditionType = 0 //0_玩家等级判断
	TaskConditionFinishMap ETaskConditionType = 1 //1_通过关卡条件判断

	TaskConditionHolder
	MaxTaskConditionHolder = TaskConditionHolder + 1
)

type EMoveInterruptReason = int32

const (
	EMIRPathNodeCount EMoveInterruptReason = 0 //0_路径单必须大于2小于6
	EMIRSpeed         EMoveInterruptReason = 1 //1_速度不一致
	EMIRPathNodeLine  EMoveInterruptReason = 2 //2_路径点不连通
	EMIRStartPos      EMoveInterruptReason = 3 //3_起点与服务器当前点偏差3个格子以上
	EMIRStartBlock    EMoveInterruptReason = 4 //4_起点为阻挡点
	EMIRTimeDiff      EMoveInterruptReason = 5 //5_时间偏差值为负
	EMIRMoveToBlock   EMoveInterruptReason = 6 //6_行动到阻挡点
	EMIRTransfer      EMoveInterruptReason = 7 //7_传送到目标点
)

// 邮件状态
type EMailStatus = int32

const MailMaxCount = 50
const MailMaxCountDb = 100
const MaxRowNum int32 = 100 //单集合支持最大行数
const PreWaitMailCount = 10
const OnceLoadMailRecordCount = 100
const MaxMailRecordCount = 1000
const MaxFriendCountDb = 150

const (
	EMailStatusNull        EMailStatus = 0 //空
	EMailStatusNotRead     EMailStatus = 1 //未读
	EMailStatusNotReceived EMailStatus = 2 //已读但未领取
	EMailStatusReceived    EMailStatus = 3 //已领取
	EMailStatusDel         EMailStatus = 4 //删除
)

// 邮件类型
type EMailType = uint8

const (
	EMailTypeSystem EMailType = 0 //系统邮件
	EMailTypeGM     EMailType = 1 //GM邮件
	EMailTypeSurvey EMailType = 2 //问卷邮件
)

// 性别类型
type ESexType = int32

const (
	ESexMale   ESexType = 1
	ESexFeMale ESexType = 1 << 1
	ESexAll    ESexType = ESexMale | ESexFeMale
)

// 资源来源类型,奖励和消耗用同一个枚举
type ResourceSourceTypeEnum int32

// ///注意:添加来源需要同步到keyword
const (
	//奖励来源
	SourceAwardTypeNone                ResourceSourceTypeEnum = 0  //未知
	SourceAwardTypeInit                ResourceSourceTypeEnum = 1  //角色初始化
	SourceAwardTypeGM                  ResourceSourceTypeEnum = 2  //gm添加
	SourceAwardTypeLevelUp             ResourceSourceTypeEnum = 3  //升级奖励
	SourceAwardTypeTask                ResourceSourceTypeEnum = 4  //任务产出
	SourceAwardTypeTaskBox             ResourceSourceTypeEnum = 5  //任务成就宝箱产出
	SourceAwardTypeMainLevel           ResourceSourceTypeEnum = 6  //主线关卡
	SourceAwardTypePagodaRelics        ResourceSourceTypeEnum = 7  //遗迹关卡
	SourceAwardTypePagodaChallenge     ResourceSourceTypeEnum = 8  //挑战关卡
	SourceAwardTypeBranchLevel         ResourceSourceTypeEnum = 9  //支线关卡
	SourceAwardTypeAutoTask            ResourceSourceTypeEnum = 10 //自动领取任务,包括主线、支线、爬塔、派遣任务
	SourceAwardTypePetFree             ResourceSourceTypeEnum = 11 //宠物放生
	SourceAwardTypeMonsterDeath        ResourceSourceTypeEnum = 12 //怪物死亡掉落
	SourceAwardTypeUseItem             ResourceSourceTypeEnum = 13 //使用道具
	SourceAwardTypeItemEvent           ResourceSourceTypeEnum = 14 //道具事件获取
	SourceAwardTypeUnLock              ResourceSourceTypeEnum = 15 //解锁功能奖励
	SourceAwardTypeMail                ResourceSourceTypeEnum = 16 //邮件奖励
	SourceAwardInteractTalk            ResourceSourceTypeEnum = 17 //宠物互动剧情奖励
	SourceAwardTypeCapturedMonster     ResourceSourceTypeEnum = 18 //抓宠奖励
	SourceAwardTypeLevelCaptured       ResourceSourceTypeEnum = 19 //抓宠奖励-主线关卡战斗抓宠
	SourceAwardTypeSchedule            ResourceSourceTypeEnum = 20 //派遣事件
	SourceAwardTypeBuildingUnlock      ResourceSourceTypeEnum = 21 //建筑解锁奖励
	SourceAwardTypePetFeed             ResourceSourceTypeEnum = 22 //宠物喂养
	SourceAwardTypePetGiveGift         ResourceSourceTypeEnum = 23 //宠物赠送礼物
	SourceAwardTypeBuildingMake        ResourceSourceTypeEnum = 24 //建筑生产
	SourceAwardTypeBuildingMachine     ResourceSourceTypeEnum = 25 //建筑加工
	SourceAwardTypeThief               ResourceSourceTypeEnum = 26 //盗贼副本奖励
	SourceAwardTypeRecharge            ResourceSourceTypeEnum = 27 //充值获得
	SourceAwardTypeShop                ResourceSourceTypeEnum = 28 //商店购买获得
	SourceAwardTypeActiveMedal         ResourceSourceTypeEnum = 29 //激活勋章奖励
	SourceAwardTypeTaskObjective       ResourceSourceTypeEnum = 30 //任务目标奖励
	SourceAwardTypeInteractive         ResourceSourceTypeEnum = 31 //交互奖励
	SourceAwardTypeFourLeaf            ResourceSourceTypeEnum = 32 //衍生事件-四叶灵回礼获得
	SourceAwardGluttonousSnake         ResourceSourceTypeEnum = 33 //趣味抓宠贪吃蛇抓宠
	SourceAwardTypeHangUp              ResourceSourceTypeEnum = 34 //挂机离线奖励
	SourceAwardTypeHomeStage           ResourceSourceTypeEnum = 35 //家园目标阶段奖励
	SourceAwardTypeHomeObjective       ResourceSourceTypeEnum = 36 //家园任务目标奖励
	SourceAwardTypeAtlasEntry          ResourceSourceTypeEnum = 37 //图鉴词条解锁奖励
	SourceAwardTypeInvadeDefend        ResourceSourceTypeEnum = 38 //舰船战斗副本奖励
	SourceAwardTypeNihility            ResourceSourceTypeEnum = 39 //虚无之地副本奖励
	SourceAwardTypeChapter             ResourceSourceTypeEnum = 40 //章节奖励
	SourceAwardTypeHangUpSpeedUp       ResourceSourceTypeEnum = 41 //挂机加速
	SourceAwardInteractEvent           ResourceSourceTypeEnum = 42 //宠物互动事件奖励
	SourceAwardInteractCaress          ResourceSourceTypeEnum = 43 //宠物抚摸奖励
	SourceAwardTypeDrawPet             ResourceSourceTypeEnum = 44 //抽卡
	SourceAwardDailyCopyHangUp         ResourceSourceTypeEnum = 45 //日常副本挂机
	SourceAwardUnloadBloodlineEquip    ResourceSourceTypeEnum = 46 //卸载血统因子返还到背包
	SourceAwardLevelUpBloodlineBack    ResourceSourceTypeEnum = 47 //升级宠物血统等级消耗宠物返还血统因子
	SourceAwardTransferCharacterBack   ResourceSourceTypeEnum = 48 //宠物性格转移返还血统因子
	SourceAwardCreateGuildFailBack     ResourceSourceTypeEnum = 49 //创建公会失败返还
	SourceAwardModifyGuildNameFailBack ResourceSourceTypeEnum = 50 //修改公会名称失败返还
	SourceAwardArenaPetTiredReset      ResourceSourceTypeEnum = 51 //竞技场宠物疲劳度修复
	SourceAwardArenaWinReward          ResourceSourceTypeEnum = 52 //竞技场结算奖励
	SourceAwardTypeMapTypeLevel        ResourceSourceTypeEnum = 53 //关卡地图
	SourceAwardTypeMapTypeWorld        ResourceSourceTypeEnum = 54 //世界地图
	SourceAwardTypeMapTypeTrialTower   ResourceSourceTypeEnum = 55 //试练塔地图
	SourceAwardTypeMapTypeGrade        ResourceSourceTypeEnum = 57 //评级地图
	SourceAwardTypeMapTypeDailyCopy    ResourceSourceTypeEnum = 58 //日常副本地图
	SourceAwardTypeMapTypePvp          ResourceSourceTypeEnum = 59 //竞技场地图
	SourceAwardModifyNameFailBack      ResourceSourceTypeEnum = 60 //修改昵称失败返还
	SourceAwardOpenHorn                ResourceSourceTypeEnum = 61 //开号角
	SourceAwardGetHornReward           ResourceSourceTypeEnum = 62 //消耗积分得到号角
	SourceAwardRankAchieve             ResourceSourceTypeEnum = 63 //排行成就奖励
	SourceAwardRankListExtol           ResourceSourceTypeEnum = 64 //排行榜赞颂奖励
	SourceAwardDailyTask               ResourceSourceTypeEnum = 65 //每日任务领奖奖励
	SourceAwardWeekTask                ResourceSourceTypeEnum = 66 //每周任务领奖奖励
	SourceAwardEternalTask             ResourceSourceTypeEnum = 67 //成就任务领奖奖励
	SourceAwardDailyTaskBox            ResourceSourceTypeEnum = 68 //每日任务积分宝箱领奖奖励
	SourceAwardWeekTaskBox             ResourceSourceTypeEnum = 69 //每周任务积分宝箱领奖奖励
	SourceAwardPetRebornBack           ResourceSourceTypeEnum = 70 //宠物重生返还奖励
	SourceAwardSharePet                ResourceSourceTypeEnum = 71 //分享宠物奖励
	SourceAwardTypeOnlineAward         ResourceSourceTypeEnum = 72 //在线时长奖励
	SourceAwardTypeCooking             ResourceSourceTypeEnum = 73 //烹饪奖励
	SourceAwardTypePetTransfer         ResourceSourceTypeEnum = 74 //宠物继承（转移）返还奖励
	SourceAwardTypeMainLineObjective   ResourceSourceTypeEnum = 75 //主线目标奖励
	SourceAwardTypeSparMapUnLockSlot   ResourceSourceTypeEnum = 76 //宠物能力槽位解锁地图奖励
	SourceAwardTypeMachineReward       ResourceSourceTypeEnum = 77 //配方加工奖励
	SourceAwardTypeSingleCopy          ResourceSourceTypeEnum = 78 //单人副本地图奖励
	SourceAwardTypeFormulaUnlock       ResourceSourceTypeEnum = 79 //配方解锁奖励
	SourceAwardTypeMatchCopy           ResourceSourceTypeEnum = 80 //多人副本地图奖励
	SourceAwardTypeMapRevive           ResourceSourceTypeEnum = 81 //地图复活消耗物品
	SourceAwardTypeMainLineTask        ResourceSourceTypeEnum = 82 //主线任务奖励
	SourceAwardTypeDropEvent           ResourceSourceTypeEnum = 83 //掉落事件奖励
	SourceAwardTypeNpcTalkOptionAward  ResourceSourceTypeEnum = 84 //剧情对话选项奖励
	SourceHomeTaskDayReward            ResourceSourceTypeEnum = 85 //每日任务奖励
	SourceTrialTowerDayReward          ResourceSourceTypeEnum = 86 //每日爬塔奖励
	SourceAwardTypeTrialTowerStage     ResourceSourceTypeEnum = 87 //爬塔阶段奖励
	SourceAwardTypeExtendPetBag        ResourceSourceTypeEnum = 88 //扩展宠物背包
	SourceAwardTypeSpiritDojo          ResourceSourceTypeEnum = 89 //御灵道场副本奖励
	SourceAwardTypeDojoDraw            ResourceSourceTypeEnum = 90 //御灵道场抽奖奖励
	SourcePublicShop                   ResourceSourceTypeEnum = 91 //公共商店
	SourceItemSale                     ResourceSourceTypeEnum = 92 //出售物品
	SourcePrivateShop                  ResourceSourceTypeEnum = 93 //私有商店
	SourceAwardDojoRefreshBack         ResourceSourceTypeEnum = 94 //刷新御灵道场关卡返还
	SourceAwardTypeMatchTwistedLand    ResourceSourceTypeEnum = 95 //扭曲之地地图奖励
	SourceAwardTypeMax

	//消耗来源
	SourceCostTypeNone                      ResourceSourceTypeEnum = 10000 //未知
	SourceCostTypePetLevelUp                ResourceSourceTypeEnum = 10001 //宠物升级消耗
	SourceCostTypePetPassiveRefine          ResourceSourceTypeEnum = 10002 //宠物被动洗练消耗
	SourceCostTypePetRestore                ResourceSourceTypeEnum = 10003 //宠物等级重塑消耗
	SourceCostTypePetArtifice               ResourceSourceTypeEnum = 10004 //宠物资质炼化消耗
	SourceCostTypeItemEvent                 ResourceSourceTypeEnum = 10005 //移除道具事件消耗
	SourceCostTypeBagCapacity               ResourceSourceTypeEnum = 10006 //背包扩容消耗
	SourceCostTypeRestTalent                ResourceSourceTypeEnum = 10007 //重置or切换天赋消耗
	SourceCostTypeUnlockPlot                ResourceSourceTypeEnum = 10008 //解锁建筑用地
	SourceCostTypeBuildBuilding             ResourceSourceTypeEnum = 10009 //修建建筑
	SourceCostTypeBuildingLevelUp           ResourceSourceTypeEnum = 10010 //升级建筑
	SourceCostTypePetBreakthroughLevel      ResourceSourceTypeEnum = 10011 //宠物突破等级消耗
	SourceCostTypePetFeed                   ResourceSourceTypeEnum = 10012 //宠物喂养
	SourceCostTypePetGiveGift               ResourceSourceTypeEnum = 10013 //宠物赠送礼物
	SourceCostTypePetTransferChar           ResourceSourceTypeEnum = 10014 //宠物性格转移消耗
	SourceCostTypePetWashChar               ResourceSourceTypeEnum = 10015 //宠物性格洗礼消耗
	SourceCostTypePetResetChar              ResourceSourceTypeEnum = 10016 //宠物重置性格消耗
	SourceCostTypePetAddElementResist       ResourceSourceTypeEnum = 10017 //宠物元素抗性加点消耗
	SourceCostTypeAddPotentialExp           ResourceSourceTypeEnum = 10018 //增加宠物资质经验消耗
	SourceCostTypePetSkillLevelUp           ResourceSourceTypeEnum = 10019 //宠物升级技能消耗
	SourceCostTypePetLevelUpPassive         ResourceSourceTypeEnum = 10020 //升级宠物被动消耗
	SourceCostTypePetLevelUpBloodline       ResourceSourceTypeEnum = 10021 //提升宠物血统等级消耗
	SourceCostTypePetLoadBloodlineEquip     ResourceSourceTypeEnum = 10022 //安装血统因子消耗
	SourceCostTypePetLevelUpBloodlineEquip  ResourceSourceTypeEnum = 10023 //升级血统因子消耗
	SourceCostFunnyCapture                  ResourceSourceTypeEnum = 10025 //趣味抓宠消耗
	SourceCostInteract                      ResourceSourceTypeEnum = 10026 //互动消耗
	SourceCostSchedule                      ResourceSourceTypeEnum = 10027 //派遣消耗
	SourceCostScheduleSpeedUp               ResourceSourceTypeEnum = 10028 //派遣加速消耗
	SourceCostUseItem                       ResourceSourceTypeEnum = 10029 //使用道具消耗
	SourceCostGluttonousSnake               ResourceSourceTypeEnum = 10031 //贪吃蛇使用食物消耗
	SourceCostTypeBuildingSpeedUp           ResourceSourceTypeEnum = 10032 //修建/升级建筑加速消耗
	SourceCostTypeBanishInvade              ResourceSourceTypeEnum = 10033 //驱逐入侵事件消耗
	SourceCostSwitchWarshipPosture          ResourceSourceTypeEnum = 10034 //舰船姿态切换消耗
	SourceCostUpgradeQuality                ResourceSourceTypeEnum = 10035 //舰船装备升级品质消耗
	SourceCostSellItem                      ResourceSourceTypeEnum = 10036 //道具出售消耗
	SourceCostEquipSlotLevelUp              ResourceSourceTypeEnum = 10037 //装备槽位升级消耗
	SourceCostShopBuyGoods                  ResourceSourceTypeEnum = 10038 //商店购买消耗
	SourceCostTypeArmsLevelUp               ResourceSourceTypeEnum = 10039 //武器升级消耗
	SourceCostTypeArmsLevelSpeedUp          ResourceSourceTypeEnum = 10040 //武器升级加速消耗
	SourceCostTypeNilihityHardPassiveSelect ResourceSourceTypeEnum = 10041 //虚无之地困难模式被动重选
	SourceCostTypeMachineCost               ResourceSourceTypeEnum = 10042 //配方加工消耗
	SourceCostTypeResetElementResist        ResourceSourceTypeEnum = 10043 //重置元素加点消耗
	SourceCostTypePetTransfer               ResourceSourceTypeEnum = 10044 //宠物养成转移消耗
	SourceCostHangUpSpeedUp                 ResourceSourceTypeEnum = 10045 //挂机加速消耗
	SourceCostShopRefresh                   ResourceSourceTypeEnum = 10046 //商店刷新消耗
	SourceCostCreateGuild                   ResourceSourceTypeEnum = 10047 //创建公会消耗
	SourceCostModifyGuildName               ResourceSourceTypeEnum = 10048 //修改帮会名称消耗
	SourceCostGuildLevelUp                  ResourceSourceTypeEnum = 10049 //公会升级消耗
	SourceCostModifyName                    ResourceSourceTypeEnum = 10050 //修改昵称消耗
	SourceCostTypePetReborn                 ResourceSourceTypeEnum = 10051 //宠物重生消耗
	SourceCostScheduleRefresh               ResourceSourceTypeEnum = 10052 //派遣手动刷新消耗
	SourceCostTypeCooking                   ResourceSourceTypeEnum = 10053 //烹饪消耗
	SourceCostScheduleBuyTimes              ResourceSourceTypeEnum = 10054 //派遣次数购买消耗
	SourceCostCampTentsLevelUp              ResourceSourceTypeEnum = 10055 //营地帐篷升级消耗
	SourceCostCampTentsAdditionActive       ResourceSourceTypeEnum = 10056 //营地帐篷增益激活消耗
	SourceCostUnlockFormula                 ResourceSourceTypeEnum = 10057 //解锁配方消耗
	SourceCostActiveSpar                    ResourceSourceTypeEnum = 10058 //激活能力晶石消耗
	SourceCostUpgradeSparStar               ResourceSourceTypeEnum = 10059 //升级能力晶石星级消耗
	SourceCostTamerTalentChange             ResourceSourceTypeEnum = 10060 //职业天赋升级消耗
	SourceCostSpiritDojoRefresh             ResourceSourceTypeEnum = 10061 //御灵道场关卡刷新消耗
	SourceCostUsePotentialPill              ResourceSourceTypeEnum = 10062 //使用资质丹消耗
	SourceCostUseInnerPill                  ResourceSourceTypeEnum = 10063 // 使用内丹消耗
	SourceCostWashPetPotential              ResourceSourceTypeEnum = 10064 //宠物资质洗练消耗

	SourceCostTypeMax
)

// 日志类型
type ESPReplyMode = int32

const (
	ReplyModeAll             ESPReplyMode = 0 //0_全部回复
	ReplyModeCastSkill       ESPReplyMode = 1 //1_仅触发技能回复
	ReplyModeExceptCastSkill ESPReplyMode = 2 //2_除外触发技能回复

	ReplyModeMax
)

// 召唤宠物属性继承
type EInheritType = int32

const (
	EInheritNone   EInheritType = 0 //0_不继承
	EInheritPlayer EInheritType = 1 //1_继承玩家属性
	EInheritPet    EInheritType = 2 //2_继承宠物
)

// 日志类型
type ELogTypeEnum int32

const (
	ELogTypeCreateRole        ELogTypeEnum = 1  //创建角色
	ELogTypeLogin             ELogTypeEnum = 2  //登录
	ELogTypeLogout            ELogTypeEnum = 3  //登出
	ELogTypePlayerExp         ELogTypeEnum = 4  //玩家经验变化
	ELogTypeResource          ELogTypeEnum = 5  //资源获取和消耗,包括货币和道具等
	ELogTypePetGet            ELogTypeEnum = 6  //宠物获得
	ELogTypePetExp            ELogTypeEnum = 7  //宠物经验变化
	ELogTypePetPotentialExp   ELogTypeEnum = 8  //添加宠物资质经验
	ELogTypePetSkillLevelUp   ELogTypeEnum = 9  //宠物技能升级
	ELogTypePetPassiveLevelUp ELogTypeEnum = 10 //宠物强化被动
	//ELogTypePetAddElementResist     ELogTypeEnum = 11 //宠物元素抗性加点
	ELogTypePetIntimateExp          ELogTypeEnum = 12 //宠物亲密度变化
	ELogTypePetUnlockCharSite       ELogTypeEnum = 13 //宠物解锁性格槽位
	ELogTypePetAddCharacter         ELogTypeEnum = 14 //宠物获得性格
	ELogTypePetResetCharacter       ELogTypeEnum = 15 //宠物重置性格
	ELogTypePetTransferCharacter    ELogTypeEnum = 16 //宠物性格转移
	ELogTypePetUnlockProperty       ELogTypeEnum = 17 //宠物解锁词条
	ELogTypePetWashCharacter        ELogTypeEnum = 18 //宠物性格洗礼
	ELogTypePetBloodlineLevelUp     ELogTypeEnum = 19 //提升宠物血统活力和血统等级
	ELogTypePetOptionBloodlineEquip ELogTypeEnum = 20 //宠物安装或卸载血统因子
	ELogTypePetAddBloodlineEquipExp ELogTypeEnum = 21 //宠物血统因子增加经验
	ELogTypePetRelease              ELogTypeEnum = 22 //宠物放生
	ELogTypeTaskEnd                 ELogTypeEnum = 23 //领取任务奖励
	ELogTypeTaskBox                 ELogTypeEnum = 24 //领取任务宝箱奖励
	ELogTypeTalent                  ELogTypeEnum = 25 //天赋激活或升级
	ELogTypeEnterLevel              ELogTypeEnum = 26 //进入关卡
	ELogTypeLeaveLevel              ELogTypeEnum = 27 //离开关卡
	ELogTypeInteractTalk            ELogTypeEnum = 28 //宠物互动剧情
	ELogTypeOnline                  ELogTypeEnum = 29 //在线玩家数量
	ELogTypeStartSchedule           ELogTypeEnum = 30 //开始派遣
	ELogTypeAwardSchedule           ELogTypeEnum = 31 //派遣领取奖励
	ELogTypeTriggerSchedule         ELogTypeEnum = 32 //触发新派遣
	ELogTypeSpeedUpSchedule         ELogTypeEnum = 33 //派遣加速
	ELogTypeBuildingLevelUp         ELogTypeEnum = 34 //建筑升级请求
	ELogTypeBuildingLevelUpResult   ELogTypeEnum = 35 //建筑升级结果
	ELogTypeBuildingMachine         ELogTypeEnum = 36 //建筑生产加工
	//ELogTypeBuildingGetAward          ELogTypeEnum = 36 //建筑领奖
	//ELogTypeBuildingSpeedUp           ELogTypeEnum = 37 //加速
	ELogTypeUnlockFormula             ELogTypeEnum = 38 //解锁配方
	ELogTypeCapturedStart             ELogTypeEnum = 39 //趣味抓宠开启
	ELogTypeCapturedEnd               ELogTypeEnum = 40 //趣味抓宠结束
	ELogTypePetBreakthroughLevel      ELogTypeEnum = 41 //宠物突破等级
	ELogTypeTriggerDeriveSchedule     ELogTypeEnum = 42 //触发衍生事件
	ELogTypeExeFlyFishDeriveSchedule  ELogTypeEnum = 43 //执行小飞鱼衍生事件
	ELogTypeExeFourLeafDeriveSchedule ELogTypeEnum = 44 //执行四叶灵衍生事件
	ELogTypeExeSnakeDeriveSchedule    ELogTypeEnum = 45 //执行贪吃蛇衍生事件
	ELogTypeExeDragonDeriveSchedule   ELogTypeEnum = 46 //执行好斗龙衍生事件
	ELogTypeHangUp                    ELogTypeEnum = 47 //挂机领奖
	ELogTypeMail                      ELogTypeEnum = 48 //邮件
	ELogTypeDailyCopyHangUp           ELogTypeEnum = 49 //日常副本挂机奖励
	ELogTypeDrawPet                   ELogTypeEnum = 50 //抽卡
	ELogTypeChangeEquipment           ELogTypeEnum = 51 //切换装备
	ELogTypeLevelUpEquipSlot          ELogTypeEnum = 52 //强化装备槽位
	ELogTypeMedalActive               ELogTypeEnum = 53 //激活勋章
	ELogTypeHomeObjective             ELogTypeEnum = 54 //神树目标
	ELogTypeCreateGuild               ELogTypeEnum = 55 //创建公会
	ELogTypeGuildLevelUp              ELogTypeEnum = 56 //公会升级
	ELogTypeGuildApplyVerify          ELogTypeEnum = 57 //公会加入审核
	ELogTypeGuildModifyName           ELogTypeEnum = 58 //公会修改名称
	ELogTypeGuildModifySundry         ELogTypeEnum = 59 //公会修改杂项
	ELogTypeGuildPresidentChange      ELogTypeEnum = 60 //公会会长变更
	ELogTypeExitGuild                 ELogTypeEnum = 61 //玩家退出公会
	ELogTypeGuildContribution         ELogTypeEnum = 62 //获得公会贡献值
	ELogTypeShopRefresh               ELogTypeEnum = 63 //商店刷新日志
	ELogTypeShopBuy                   ELogTypeEnum = 64 //商店购买日志
	ELogTypeResetTired                ELogTypeEnum = 65 //重置宠物疲劳
	ELogTypeTalentClassic             ELogTypeEnum = 66 //经典天赋激活或升级
	ELogTypeReconnect                 ELogTypeEnum = 67 //重连
	ELogTypePurchase                  ELogTypeEnum = 68 //充值支付回调日志
	ELogTypeOpenHorn                  ELogTypeEnum = 69 //大宝库开奖
	ELogTypeHornScoreExchange         ELogTypeEnum = 70 //大宝库积分兑换
	ELogTypePetReborn                 ELogTypeEnum = 71 //宠物重生
	ELogTypePetTransfer               ELogTypeEnum = 72 //宠物继承（转移）
	ELogTypeRankListLikeExtol         ELogTypeEnum = 73 //排行榜点赞、赞颂日志
	ELogTypeRankListAchieveAward      ELogTypeEnum = 74 //领取排行榜成就奖励日志
	ELogTypePlayerLevel               ELogTypeEnum = 75 //玩家等级
	ELogTypeCookingStart              ELogTypeEnum = 76 //开始烹饪日志
	ELogTypeCookingAward              ELogTypeEnum = 77 //领取烹饪食物日志
	ELogTypeResetTalent               ELogTypeEnum = 78 //天赋重置
	ELogTypeResetTalentClassic        ELogTypeEnum = 79 //经典天赋重置
	ELogTypeExchangeCode              ELogTypeEnum = 80 //兑换码
	ELogTypeTamerTalent               ELogTypeEnum = 81 //职业天赋激活或升级
	ELogTypeResetTamerTalent          ELogTypeEnum = 82 //职业天赋重置

	ELogTypeMax
)

// 派遣事件触发来源类型
type ELogScheduleTriggerType int32

const (
	ELogTriggerTypeSchedule  ELogScheduleTriggerType = 1 //其他派遣事件触发
	ELogTriggerTypeEvent     ELogScheduleTriggerType = 2 //事件触发
	ELogTriggerTypeUnlock    ELogScheduleTriggerType = 3 //功能解锁触发
	ELogTriggerTypeUnlockMap ELogScheduleTriggerType = 4 //解锁世界地图
	ELogTriggerTypeBuilding  ELogScheduleTriggerType = 5 //建筑升级解锁触发
	ELogTriggerTypeRefresh   ELogScheduleTriggerType = 6 //每日刷新触发
	ELogTriggerTypeGM        ELogScheduleTriggerType = 7 //GM触发
	ELogTriggerTypeHomeTask  ELogScheduleTriggerType = 8 //神树目标触发
)

// 家园相关
const PreBuildingNum = 10
const PreBuildingPetNum = 30
const PreSurpriseCount = 5

type EPlotType = uint32

const (
	EPlotTypeFir     EPlotType = 0
	EPlotTypeSec     EPlotType = 1
	EPlotTypeThird   EPlotType = 2
	EPlotTypeFourth  EPlotType = 3
	EPlotTypeFifth   EPlotType = 4
	EPlotTypeSixth   EPlotType = 5
	EPlotTypeSeventh EPlotType = 6
	EPlotTypeEighth  EPlotType = 7
	EPlotTypeNinth   EPlotType = 8

	EPlotTypeEnd
	EPlotTypeMax = EPlotTypeEnd + 1
)

type EBuildingType = int32

const (
	EBTStart     EBuildingType = 0 //开始标记
	EBTDispatch                    //派遣中心
	EBTWorkshop  EBuildingType = 1 //工坊
	EBTParadise  EBuildingType = 2 //乐园
	EBTHotSpring EBuildingType = 3 //温泉
	EBTLounge    EBuildingType = 4 //休息室
	EBTWarehouse EBuildingType = 5 //仓库
	EBTRune      EBuildingType = 6 //符文室
	EBTTraining  EBuildingType = 7 //训练室

	EBTEnd
	EBTMax = EBTEnd + 1

	//战斗建筑——不管生产/加工
	EBTFightFirst   EBuildingType = 101 //舰船战斗1号建筑
	EBTFightSecond  EBuildingType = 102 //舰船战斗2号建筑
	EBTFightThird   EBuildingType = 103 //舰船战斗3号建筑
	EBTFightFourth  EBuildingType = 104 //舰船战斗4号建筑
	EBTFightFifth   EBuildingType = 105 //舰船战斗5号建筑
	EBTFightSixth   EBuildingType = 106 //舰船战斗6号建筑
	EBTFightSeventh EBuildingType = 107 //舰船战斗7号建筑

	EBTFightEnd
	EBTFightMax = EBTFightEnd + 1

	//特殊建筑，仅展示修建过程
	EBTSpecialMap  EBuildingType = 201 //地图建筑
	EBTSpecialBook EBuildingType = 202 //书籍建筑
	EBTSpecialTree EBuildingType = 203 //神树建筑
)

// 通过 建筑的配置ID 判断是否是舰船战斗建筑
func IsShipBattleBuilding(BuildingCnfId EBuildingType) bool {
	if BuildingCnfId >= EBTFightFirst && BuildingCnfId <= EBTFightSeventh {
		return true
	}
	return false
}

type BrushMonsterGroupType = int32

const (
	BCommon    BrushMonsterGroupType = 0 //普通刷怪
	BFormation BrushMonsterGroupType = 1 //阵型编队刷怪
)

// 刷怪方式
type BrushMonsterType = int32

const (
	BrushTypeRandom BrushMonsterType = 0 //0_随机刷怪
	BrushTypeOrder  BrushMonsterType = 1 //1_顺序刷怪
)

// 刷怪点方式类型
type BrushMonsterModeType = int32

const (
	BrushrModeTypePoint BrushMonsterModeType = 0 //0_刷怪点刷怪
	BrushModeTypePlayer BrushMonsterModeType = 1 //1_玩家点刷怪
)

// 组合刷怪类型
type CombBrushMonsterType = uint64

const (
	BrushModePointRandom  CombBrushMonsterType = uint64(BrushrModeTypePoint)<<32 | uint64(BrushTypeRandom) //刷怪点-随机刷怪
	BrushModePointOrder   CombBrushMonsterType = uint64(BrushrModeTypePoint)<<32 | uint64(BrushTypeOrder)  //刷怪点-顺序刷怪
	BrushModePlayerRandom CombBrushMonsterType = uint64(BrushModeTypePlayer)<<32 | uint64(BrushTypeRandom) //玩家点-随机刷怪
	BrushModePlayerOrder  CombBrushMonsterType = uint64(BrushModeTypePlayer)<<32 | uint64(BrushTypeOrder)  //玩家点-顺序刷怪
)

// 获取组合刷怪类型
func GetBrushModeCombType(brushModeType BrushMonsterModeType, brushMonsterType BrushMonsterType) CombBrushMonsterType {
	return uint64(brushModeType)<<32 | uint64(brushMonsterType)
}

// 建筑升级提升上限类型
type EBuildingUpLimitType = int32

const (
	MaxLimitTypeNone            EBuildingUpLimitType = 0 //0_空上限
	MaxPetLimitType             EBuildingUpLimitType = 1 //1_宠物背包存储上限
	MaxItemLimitType            EBuildingUpLimitType = 2 //2_道具背包存储上限
	MaxDefendFormationLimitType EBuildingUpLimitType = 3 //3_舰船战斗阵防守型数量值
	MaxIngredientsLimitType     EBuildingUpLimitType = 4 //4_食材存储上限
	MaxMineralLimitType         EBuildingUpLimitType = 5 //5_矿物存储上限
	MaxUpgradeMaterialLimitType EBuildingUpLimitType = 6 //6_升级材料存储上限
	MaxEnergyLimitType          EBuildingUpLimitType = 7 //7_能量存储上限
	MaxPostureEnergyLimitType   EBuildingUpLimitType = 8 //8_姿态能量存储上限

	MaxLimitTypeHolder
	MaxUpLimitType = MaxLimitTypeHolder + 1
)

type EHomeResources = int32

const (
	EHRStart           EHomeResources = 0
	EHRIngredients     EHomeResources = 0 //0_食材
	EHRMineral         EHomeResources = 1 //1_矿物
	EHRUpgradeMaterial EHomeResources = 2 //2_升级材料
	EHRCoin            EHomeResources = 3 //3_金币
	EHREnergy          EHomeResources = 4 //4_能量

	EHREnd
	EHRMax = EHREnd + 1
)

func GetNoCostHomeResoures() map[EHomeResources]struct{} {
	return mapNoCostHomeResoures
}

func CheckHomeResourceIsNoCost(resourceId EHomeResources) bool {
	_, ok := mapNoCostHomeResoures[resourceId]
	return ok
}

func GetHomeResourceAttrType(resourceType EHomeResources) AttributeType {
	switch resourceType {
	case EHRIngredients:
		return AttributeTypeIngredients
	case EHRMineral:
		return AttributeTypeMineral
	case EHRUpgradeMaterial:
		return AttributeTypeUpgradeMaterial
	case EHRCoin:
		return AttributeTypeCoin
		//case EHREnergy:
		//	return AttributeTypeEnergy
	}

	return 0
}

type EFormulaType = int32

const (
	EFTCatchPetRatioFood   EFormulaType = 0 //0_抓宠提升几率食物配方
	EFTCatchPetQualityFood EFormulaType = 1 //1_抓宠提升品质食物配方
	EFTWashNegativeFood    EFormulaType = 2 //2_洗负性格食物配方
	EFTDispatchFood        EFormulaType = 3 //3_派遣效果加成食物配方
	EFTStuff               EFormulaType = 4 //4_摆件配方
	EFTGift                EFormulaType = 5 //5_礼物配方
	EFTWashFood            EFormulaType = 6 //6_洗性格食物配方

	EFTEnd
	EFTMax = EFTEnd + 1
)

type EPropertyType = int32

const (
	//派遣相关
	EPPATScheduleBegin              EPropertyType = 301 //派遣相关开始标记
	EPPATHotSpringMood                                  //提升温泉单次心情固定数值-绝对值-温泉-单个宠物
	EPPATHotSpringMoodRatio         EPropertyType = 302 //提升温泉单次心情万分比增益-万分比-温泉-单个宠物 总基础值 * 总万分比
	EPPATHotSpringMoodMultiple      EPropertyType = 303 //提升温泉单次心情倍数比例-万分比-温泉-单个宠物
	EPPATHotSpringMoodMultipleRatio EPropertyType = 304 //提升温泉单次心情倍数几率-万分比-温泉-单个宠物
	EPPATPatrolExpRatio             EPropertyType = 308 //提升巡逻单次经验万分比增益-万分比-探索
	EPPATThreadExpRatio             EPropertyType = 310 //提升主线单次经验万分比增益-万分比-主线
	EPPATBranchExpRatio             EPropertyType = 311 //提升支线单次经验万分比增益-万分比-支线
	EPPATAdventureExpRatio          EPropertyType = 312 //提升奇遇单次经验万分比增益-万分比-奇遇
	EPPATScheduleTime               EPropertyType = 313 //减少任意派遣事件的耗时-万分比-任意派遣
	EPPATScheduleAddGatherRatio     EPropertyType = 319 //提升采集产出万分比收益
	EPPATScheduleAddMiningRatio     EPropertyType = 320 //提升挖矿产出万分比收益
	EPPATScheduleCatchPetRatio      EPropertyType = 322 //提升探索后抓宠事件的概率-万分比-探索
	EPPATScheduleGetTaskRatio       EPropertyType = 323 //提升探索后奇遇事件的概率-万分比-探索
	EPPATPlayIntimate               EPropertyType = 324 //提升乐园单次好感度固定数值
	EPPATPlayIntimateRatio          EPropertyType = 325 //提升乐园单次好感度万分比增益
	EPPATPlayIntimateMultiple       EPropertyType = 326 //提升乐园单次好感度倍数比例
	EPPATPlayIntimateMultipleRatio  EPropertyType = 327 //提升乐园单次好感度倍数几率
	EPPATSunnyPlayIntimateRatio     EPropertyType = 328 //提升晴天乐园好感度加成比例
	EPPATRainyHotSpringExpRatio     EPropertyType = 329 //提升雨天温泉经验加成比例

	EPPATAddWaterRatio       EPropertyType = 330 //水元素宠物概率增加
	EPPATAddRareWaterRatio   EPropertyType = 331 //水元素稀有宠物概率增加
	EPPATAddFireRatio        EPropertyType = 332 //火元素宠物概率增加
	EPPATAddRareFireRatio    EPropertyType = 333 //火元素稀有宠物概率增加
	EPPATAddWindRatio        EPropertyType = 334 //风元素宠物概率增加
	EPPATAddRareWindRatio    EPropertyType = 335 //风元素稀有宠物概率增加
	EPPATAddThunderRatio     EPropertyType = 336 //雷元素宠物概率增加
	EPPATAddRareThunderRatio EPropertyType = 337 //雷元素稀有宠物概率增加
	EPPATAddSoilRatio        EPropertyType = 338 //土元素宠物概率增加
	EPPATAddRareSoilRatio    EPropertyType = 339 //土元素稀有宠物概率增加

	EPPATThiefTriggerRatio    EPropertyType = 350 //影响盗贼副本触发几率
	EPPATThiefRobRatio        EPropertyType = 351 //影响盗贼副本被抢量
	EPPATThiefAwardRatio      EPropertyType = 352 //增加盗贼副本奖励
	EPPATThiefExtraAward      EPropertyType = 353 //盗贼副本会获得额外奖励掉落ID
	EPPATSunnyThiefRobRatio   EPropertyType = 355 //提升晴天强盗事件损失比例
	EPPATRainyThiefRobRatio   EPropertyType = 356 //提升雨天强盗事件损失比例
	EPPATThunderThiefRobRatio EPropertyType = 357 //提升打雷天强盗事件损失比例
	EPPATGaleThiefRobRatio    EPropertyType = 358 //提升大风天强盗事件损失比例
	EPPATDroughtThiefRobRatio EPropertyType = 359 //提升干旱天强盗事件损失比例

	EPPATScheduleAddSunnyGatherRatio   EPropertyType = 360 //提升晴天采集比例
	EPPATScheduleAddRainyGatherRatio   EPropertyType = 361 //提升雨天采集比例
	EPPATScheduleAddGaleMiningRatio    EPropertyType = 362 //提升大风天挖矿比例
	EPPATScheduleAddDroughtMiningRatio EPropertyType = 363 //提升沙霾天挖矿比例
	EPPATScheduleAddThunderGatherRatio EPropertyType = 364 //提升打雷天采集比例
	EPPATScheduleAddThunderMiningRatio EPropertyType = 365 //提升打雷天挖矿比例

	EPPScheduleGatherTimeRatio       EPropertyType = 366 //采集派遣事件耗时降低
	EPPScheduleMiningTimeRatio       EPropertyType = 367 //挖矿派遣事件耗时降低
	EPPScheduleEnergyTimeRatio       EPropertyType = 368 //能量派遣事件耗时降低
	EPPScheduleAddEnergyRatio        EPropertyType = 369 //提升能量产出万分比收益
	EPPScheduleBoxCoinRatio          EPropertyType = 370 //提高宝箱派遣事件金币产出万分比收益
	EPPScheduleExploreTimeRatio      EPropertyType = 372 //内容探索事件耗时降低
	EPPScheduleExploreResourceRatio  EPropertyType = 373 //内容探索出现资源类型概率增加
	EPPScheduleExploreFightRatio     EPropertyType = 374 //内容探索出现战斗类型概率增加
	EPPScheduleExploreSearchRatio    EPropertyType = 375 //内容探索出现搜索类型概率增加
	EPPScheduleExploreRareRatio      EPropertyType = 376 //内容探索出现稀有事件概率增加
	EPPScheduleExplorePetRatio       EPropertyType = 377 //内容探索出现宠物类型概率增加
	EPPScheduleGatherTimeExtraCoin   EPropertyType = 378 //根据采集事件时长获得额外金币
	EPPScheduleMiningTimeExtraCoin   EPropertyType = 379 //根据挖矿事件时长获得额外金币
	EPPScheduleEnergyTimeExtraCoin   EPropertyType = 380 //根据能量事件时长获得额外金币
	EPPScheduleBoxSkillItemRatio     EPropertyType = 381 //派遣宝箱事件时，有几率获得额外技能书+1
	EPPScheduleBoxGiftRatio          EPropertyType = 382 //派遣宝箱事件时，有几率获得额外礼物+1
	EPPSchedulePetCount              EPropertyType = 383 //派遣宠物数量
	EPPScheduleDecodeTimeRatio       EPropertyType = 384 //识别派遣事件耗时
	EPPScheduleBoxTimeRatio          EPropertyType = 385 //宝箱派遣事件耗时
	EPPScheduleDecodeRarityEpicRatio EPropertyType = 386 //识别史诗事件权重比例
	EPPATTrainingExp                 EPropertyType = 387 //提升训练所单次经验固定数值-绝对值-训练所-单个宠物
	EPPATTrainingExpRatio            EPropertyType = 388 //提升训练所单次经验万分比增益-万分比-训练所-单个宠物 总基础值 * 总万分比
	EPPScheduleAwardGroup            EPropertyType = 389 //修改派遣奖励组ID

	EPPATScheduleEnd
)

// 宠物性格属性加成方式
type EPropertyAddType = int8

const (
	EPropertyAddTypeUnknown EPropertyAddType = 0 //0_未知加成方式
	EPropertyAddTypeRatio   EPropertyAddType = 1 //1_万分比
	EPropertyAddTypeValue   EPropertyAddType = 2 //2_数值

	EPropertyAddTypeEnd
	EPropertyAddTypeMax = EPropertyAddTypeEnd + 1
)

// 词条正负
type EPositiveDefectType = int32

const (
	EPositiveType EPositiveDefectType = 1  // 1,正反馈
	EDefectType   EPositiveDefectType = -1 //-1,负反馈
)

const RefreshDailyDataHour = 4             //每日刷新时间
const RefreshMonthDay = 1                  //每月刷新日期
const MaxPetNickNameLen = 16               //宠物昵称最长字节限制
const MaxCharacterCount = 3                //每个宠物性格数量最大限制
const MaxPropertyGroupCount = 5            //宠物每个性格词条组最大限制
const MaxCharacterPropertyCount = 5        //宠物每个性格词条组词条数量最大限制
const MaxAllPetCharacterPropertyCount = 20 //宠物所有性格属性条数
const MaxBloodlineEquipSiteCount = 6       //宠物血统因子曹位数
const MaxPetGiveGiftOrFeedUseCount = 20    //宠物赠送礼物或宠物喂养每次使用不同道具的数量最大限制
const MaxPetAddPotentialExpUseCount = 20   //请求添加宠物资质经验使用不同道具的数量最大限制
const InitBloodlinePassiveCount = 10       //初始化血统因子和血统因子套装被动数量

func IsValidBloodlinePos(pos int32) bool {
	return pos >= 0 && pos < MaxBloodlineEquipSiteCount
}

// 宠物种族类型
type EPetRacialType = int8

const (
	EPetRacialNone       EPetRacialType = 0 //0_未知
	EPetRacialWildAnimal EPetRacialType = 1 //1_野兽
	EPetRacialDragon     EPetRacialType = 2 //2_龙
	EPetRacialAquatic    EPetRacialType = 3 //3_水栖
	EPetRacialFlight     EPetRacialType = 4 //4_飞行
	EPetRacialNatural    EPetRacialType = 5 //5_自然

	EPetRacialEnd
	EPetRacialMax = EPetRacialEnd + 1
)

// 宠物性格槽位状态
type EPetCharacterSiteStateType int8

const (
	EPetCharacterSiteStateUnOpen   EPetCharacterSiteStateType = 0 //0_未解锁
	EPetCharacterSiteStateInit     EPetCharacterSiteStateType = 1 //1_初始性格
	EPetCharacterSiteStateOpen     EPetCharacterSiteStateType = 2 //2_已解锁
	EPetCharacterSiteStateTransfer EPetCharacterSiteStateType = 3 //3_已转移

	EPetCharacterSiteStateEnd
	EPetCharacterSiteStateMax = EPetCharacterSiteStateEnd + 1
)

// 宠物子类型
type PetSubType = int

const (
	PPetSubTypeCommon   PetSubType = 0 //0_通用宠物类型
	PPetSubTypeNpc      PetSubType = 1 //1_Npc宠物类型
	PPetSubTypeSummon   PetSubType = 2 //2_召唤宠物类型
	PPetSubTypeCaptured PetSubType = 3 //3_抓宠宠物类型
)

/*
// 宠物定位
type EPetType = int32

const (
	EPetTypeMeat      EPetType = 1 //1_肉盾
	EPetTypeOutput    EPetType = 2 //2_输出
	EPetTypeAuxiliary EPetType = 3 //3_辅助
)
const AllPetType = 15
*/

// 职业定位(玩家和宠物)
type ECareerType = int32

const (
	ECareerTypeNone     ECareerType = 0 // 0_无职业
	ECareerTypeWarrior  ECareerType = 1 // 1_近战输出
	ECareerTypeADC      ECareerType = 2 // 2_远程输出
	ECareerTypeSUP      ECareerType = 3 // 3_辅助
	ECareerTypeTank     ECareerType = 4 // 4_坦克
	ECareerTypeAssassin ECareerType = 5 // 5_近战刺客

	ECareerTypeEnd
	ECareerTypeMax = ECareerTypeEnd + 1
)
const AllCareerType = 0xFFFF

// 判断是否包含职业定位
func CheckHaveCareerType(careerTypeFlag uint32, careerType ECareerType) bool {
	careerTypeTag := uint32(1 << careerType)
	return careerTypeFlag&careerTypeTag == careerTypeTag
}

// 趣味抓宠类型
type ECatchPetLevelType = int32

const (
	ECPNone    ECatchPetLevelType = -1 //无效信息
	ECPStart   ECatchPetLevelType = 0  //开始标记
	ECPSoul                            //0_灵魂回响
	ECPNaughty ECatchPetLevelType = 1  //1_抓住那个调皮的家伙
	ECPFight   ECatchPetLevelType = 2  //2_战斗抓宠

	ECPEnd
	ECPMax = ECPEnd + 1
)

// 抓宠大分类(目标主要是用在宠物随机品质的地方，做区分)
type ECatchPetType = int32

const (
	ECPTDefault ECatchPetType = 0 // 默认抓宠
	ECPTFunny   ECatchPetType = 1 // 趣味抓宠
	ECPTDraw    ECatchPetType = 2 // 抽卡
)

// 对象体型大小
type ESceneObjSizeType = int32

const (
	EPSStart  ESceneObjSizeType = 0 //开始标记
	EPSLarge                        //0_大体型
	EPSMedium ESceneObjSizeType = 1 //1_中体型
	EPSSmall  ESceneObjSizeType = 2 //2_小体型

	EPSEnd
	EPSMax = EPSEnd + 1
)

// 宠物被动解锁条件类型
type EPetPassiveUnlockType int8

const (
	EPetPassiveUnlockTypeNone       EPetPassiveUnlockType = 0 //0_无解锁条件
	EPetPassiveUnlockTypeSkillLevel EPetPassiveUnlockType = 1 //1_技能等级
	EPetPassiveUnlockTypeBloodLevel EPetPassiveUnlockType = 2 //2_血统等级

	EPetPassiveUnlockTypeEnd
	EPetPassiveUnlockTypeMax = EPetPassiveUnlockTypeEnd + 1
)

type ECaptureResultType int8

const (
	ECRTSucc   ECaptureResultType = 1 //1_成功捕捉
	ECRTDie    ECaptureResultType = 2 //2_抓宠对象死亡
	ECRTEscape ECaptureResultType = 4 //4_抓宠对象逃跑
)

// 操作血统因子类型
type EOptionBloodlineEquipType int8

const (
	EOptionBloodlineEquipTypeInstall   EOptionBloodlineEquipType = 1 //安装血统因子
	EOptionBloodlineEquipTypeUninstall EOptionBloodlineEquipType = 2 //卸载血统因子

	EOptionBloodlineEquipTypeEnd
	EOptionBloodlineEquipTypeMax = EOptionBloodlineEquipTypeEnd + 1
)

// 宠物互动
type EPetInteractType = int32

const (
	PetInteractTypeNone     EPetInteractType = -1 //错误类型
	PetInteractTypeCaress   EPetInteractType = 0  //抚摸
	PetInteractTypeInteract EPetInteractType = 1  //互动
	PetInteractTypeEvent    EPetInteractType = 2  //互动事件
	PetInteractTypeEnd

	PetInteractTypeMax = PetInteractTypeEnd + 1
)

// 影响可变时间条件类型
type EMutableTimeConditionType = int32

const (
	EMTCMonsterDeath EMutableTimeConditionType = 0 //0_怪物死亡
	EMTCPetNpcDeath  EMutableTimeConditionType = 1 //1_NPC宠物死亡
	EMTCPetNpcDamage EMutableTimeConditionType = 2 //2_NPC宠物受伤
)

// 惊喜意外效果类型
type ESurpriseAccidentType = int32

const (
	ESATPetElementIncreaseProduction ESurpriseAccidentType = 1 //1_宠物元素属性提升产量
	ESATIncreaseProduction           ESurpriseAccidentType = 2 //2_产量提升
	ESATPetIdIncreaseProduction      ESurpriseAccidentType = 3 //3_宠物ID个数提升产量
	ESATPetCharacterAddition         ESurpriseAccidentType = 4 //4_提升宠物性格加成
)

// 舰船姿态
type EShipAttitudeType = int8

const (
	EShipNone    EShipAttitudeType = 0 //无姿态
	EShipMake    EShipAttitudeType = 1 //生产
	EShipExplore EShipAttitudeType = 2 //探索
	EShipFight   EShipAttitudeType = 3 //战斗

	EShipEnd
	EShipMax = EShipEnd + 1
)

// 惊喜意外有效时间模式
type ESurpriseTimeModeType = int8

const (
	ESTMTDurationTime EShipAttitudeType = 0 //0_持续毫秒时间
	ESTMTDurationDay  EShipAttitudeType = 1 //1_持续天数
	ESTMTUseCount     EShipAttitudeType = 2 //2_使用的次数

	ESTMTEnd
	ESTMTMax = ESTMTEnd + 1
)

// 舰船装备词条属性类型
type EWarshipEntryType = int8

const (
	WETProperty     EWarshipEntryType = 1 //1_性格词条类
	WETFightPassive EWarshipEntryType = 2 //2_战斗被动类
)

const ShopFieldCount = 16      //商品栏数量
const ShopGoodsCount = 64      //商店商品数量
const ShopFieldGoodsCount = 16 //商品栏商品数量
const ShopCount = 20           //商店数量

// 商店类型
type EShopType = int8

// 商店系统刷新周期类型
type EShopRefreshCycleType = int8

const (
	EShopRefreshCycleNone    EShopRefreshCycleType = 0 //0_不刷新
	EShopRefreshCycleDays    EShopRefreshCycleType = 1 //1_间隔天数刷新, 凌晨4点
	EShopRefreshCycleMonthly EShopRefreshCycleType = 2 //2_每月几号刷新
	EShopRefreshCycleWeekly  EShopRefreshCycleType = 3 //3_每周几刷新
	EShopRefreshCycleHours   EShopRefreshCycleType = 4 //4_间隔小时刷新, 开服时间为起点
	EShopRefreshCycleMinutes EShopRefreshCycleType = 5 //5_间隔分钟刷新, 开服时间为起点
)

// 商店栏位重置类型
type EShopFieldRefreshType = int8

const (
	EShopFieldRefreshSystem EShopFieldRefreshType = 0 // 0_系统刷新重置
	EShopFieldRefreshAll    EShopFieldRefreshType = 1 // 1_每次刷新重置
)

// 商品库存刷新类型
type EGoodsCountRefreshType = int8

const (
	EGoodsCountRefreshNone   EGoodsCountRefreshType = 0 //0_不刷新库存
	EGoodsCountRefreshSystem EGoodsCountRefreshType = 1 //1_仅系统刷新库存
	EGoodsCountRefreshAll    EGoodsCountRefreshType = 2 //2_都刷新库存
)

// 首充标记
const FirstPurchaseArray = 20

type EPetMoodType = int32

const (
	EPetMoodUnhappy      EPetMoodType = 1 // 1_不愉快
	EPetMoodDispleasure  EPetMoodType = 2 // 2_较不愉快
	EPetMoodMorePleasant EPetMoodType = 3 // 3_较为愉快
	EPetMoodHappy        EPetMoodType = 4 // 4_愉快
	EPetMoodVeryHappy    EPetMoodType = 5 // 5_非常愉快

	EPetMoodTypeEnd
	EPetMoodTypeMax = EPetMoodTypeEnd + 1
)

// 心情变化原因
type EMoodChangeWayType = int32

const (
	EMoodChangeWayUseItem = iota
	EMoodChangeWayBattle
	EMoodChangeWayTime
	EMoodChangeWayCaress
	EMoodChangeWayBuild
	EMoodChangeWaySchedule
	EMoodChangeWayGM
)

// 任务目标来源
type ObjectiveFromType = int8

const (
	OFTNone              ObjectiveFromType = 0
	OFTMainTask          ObjectiveFromType = 1  //来源——主线任务
	OFTPlate             ObjectiveFromType = 2  //来源——地图区域
	OFTHomeTask          ObjectiveFromType = 3  //来源——家园神树任务
	OFTBadgeTask         ObjectiveFromType = 4  //来源——徽章任务
	OFTAtlasTask         ObjectiveFromType = 5  //来源——图鉴任务
	OFTRankAchieveTask   ObjectiveFromType = 6  //来源——排行成就
	OFTGmTask            ObjectiveFromType = 7  //来源——GM
	OFTDailyTask         ObjectiveFromType = 8  //来源——每日任务
	OFTWeekTask          ObjectiveFromType = 9  //来源——每周任务
	OFTEternalTask       ObjectiveFromType = 10 //来源——成就任务
	OFTMainLineObjective ObjectiveFromType = 11 //来源--主线目标
	OFTPetTask           ObjectiveFromType = 12 //来源——宠物任务
)

// 目标进度计数类型
type EObjectiveCountType = int8

const (
	EOCTCreateRole   EObjectiveCountType = 0 //0_创角后计数
	EOCTAddObjective EObjectiveCountType = 1 //1_任务开启后计数
)

type NihilityModeType = int32

const (
	NihilityModeNormal   NihilityModeType = 1 // 简单模式
	NihilityModeHard     NihilityModeType = 2 // 困难模式
	NihilityModeInfinite NihilityModeType = 3 // 无尽模式

	NihilityModeEnd
	NihilityModeMax = NihilityModeEnd + 1
)

// 是否是合法的虚无之地模式
func IsValidNihilityMode(mode NihilityModeType) bool {
	return mode >= NihilityModeNormal && mode <= NihilityModeEnd
}

type NihilityPassiveConfigType = int32

const (
	NihilityPassiveConfigSelect NihilityPassiveConfigType = 1 // 1_3选1被动
	NihilityDecPassiveConfig    NihilityPassiveConfigType = 2 // 2_困难模式被动
)

type NihilityPetBattleStatusType = int32

const (
	NihilityPetBattleStatusNone          NihilityPetBattleStatusType = 0
	NihilityPetBattleStatusSummon        NihilityPetBattleStatusType = 1 << 0                                                       // 1_召唤过
	NihilityPetBattleStatusEquip         NihilityPetBattleStatusType = 1 << 1                                                       // 2_连携过
	NihilityPetBattleStatusSummonOrEquip NihilityPetBattleStatusType = NihilityPetBattleStatusSummon | NihilityPetBattleStatusEquip // 3_召唤或者连携过
)

// 虚无之地常量
const (
	NIHILITY_AREAN_PASSIVE_MAX     = 3  // 简单模式下最多3个可选被动
	NIHILITY_AREAN_PLAYER_RANK_MAX = 3  // 每个等级只需要存3个排名最靠前的
	NIHILITY_AREAN_GEN_CONFIG      = 2  // 配置个数，2个应该就够了
	NIHILITY_AREAN_LEVEL_MAKE_NUM  = 20 // 虚无之地关卡数，注意这个只是make用，具体多少关是策划配的，不能用这个字段

	NIHILITY_AREAN_PLAYER_RANK_SEND_TO_CLIENT = 5 // 每个等级给客户端发几个信息
)

type NihilityPassiveReSelect struct {
	Mode       NihilityModeType
	IsReSelect bool
}

// 勋章类型
type EMedalType = int8

const (
	EMedalTypeNone     EMedalType = 0 //未知勋章
	EMedalTypePet      EMedalType = 1 //1_宠物勋章
	EMedalTypeArms     EMedalType = 2 //2_武器勋章
	EMedalTypeBuilding EMedalType = 3 //3_建筑勋章
	EMedalTypeRisk     EMedalType = 4 //4_冒险勋章
	EMedalTypeExplore  EMedalType = 5 //5_探索勋章
	EMedalTypeSocial   EMedalType = 6 //6_社交勋章

	EMedalTypeEnd
	EMedalTypeMax = EMedalTypeEnd + 1
)

// 判断是否是勋章类型
func CheckIsMedalType(medalType EMedalType) bool {
	if medalType > EMedalTypeNone && medalType < EMedalTypeMax {
		return true
	}
	return false
}

const (
	MAX_QUERY_PLAYER_BRIEF_INFO_COUNT int32 = 30 // 查询玩家简要信息最大数量
)

// 杂项唯一标记类型
type EMiscUniqueType = byte

const (
	EMiscUniqueTypeNone    EMiscUniqueType = 0 //0_位置杂项标记
	EMiscUniqueTypeQQGroup EMiscUniqueType = 1 //1_添加QQ群
)

// 营地背包动作类型
type ELeisureActionType = int32

const (
	ELeisureActionTent     ELeisureActionType = 1 // 1_支起帐篷
	ELeisureActionWin      ELeisureActionType = 2 // 2_营地胜利
	ELeisureActionRest     ELeisureActionType = 3 // 3_营地休息
	ELeisureActionWarning  ELeisureActionType = 4 // 4_营地警戒
	ELeisureActionTouchPet ELeisureActionType = 5 // 5_抚摸幻兽
)

const MaxLeisureAction = 200

const (
	MaxDrawPetCount   = 10  // 一次抽卡最多出宠物数量
	DefaultSingleDraw = 100 // 默认单抽配置
	DefaultMultiDraw  = 200 // 默认多抽配置
)

// 日常副本类型
type EDailyCopyType = int8

const (
	EDailyCopyStart      EDailyCopyType = 1 //开始标记
	EDailyCopyCoin       EDailyCopyType = 1 //1_日常金币副本
	EDailyCopyPetExpBook EDailyCopyType = 2 //2_日常宠物经验书副本
	EDailyCopyEquipment  EDailyCopyType = 3 //3_日常装备副本

	EDailyCopyEnd
	EDailyCopyMax = EDailyCopyEnd + 1
)

const MaxPetCount = 300       //最大宠物数限制
const AllocSharePetCount = 32 //分享宠物分配数量

// 试练塔类型
type ETrialTowerType = int8

const (
	ETrialTowerTypeCourage ETrialTowerType = 1 // 1_勇气塔
	ETrialTowerTypeBows    ETrialTowerType = 2 // 2_弓箭塔
	ETrialTowerTypeSword   ETrialTowerType = 3 // 3_大剑塔
	ETrialTowerTypeStaff   ETrialTowerType = 4 // 4_法杖塔
	ETrialTowerTypeShield  ETrialTowerType = 4 // 5_剑盾塔

	ETrialTowerTypeEnd
	ETrialTowerTypeMax = ETrialTowerTypeEnd + 1
)

// 排行榜类型
type ERankType = uint64

const (
	ERankTypeMainLevel         ERankType = 1 // 1_主线关卡进度
	ERankTypeBattleCombatScore ERankType = 2 // 2_当前阵容战力排行榜
	ERankTypePlayerLevel       ERankType = 3 // 3_玩家等级排行榜
	ERankTypeTrialTowerCourage ERankType = 4 // 4_初始试炼塔(勇者塔)
	ERankTypeGuild             ERankType = 5 // 5_公会排行榜,包含所有的公会
	ERankTypeGuildNotFull      ERankType = 6 // 6_公会排行榜,只包含未满员的公会
	ERankTypeLike              ERankType = 7 // 7_点赞排行榜

	ERankTypeEnd
	ERankTypeMax = ERankTypeEnd + 1
)

// 判断排行榜类型是否合法
func CheckRankTypeIsValid(rankType ERankType) bool {
	if rankType < ERankTypeMainLevel || rankType > ERankTypeEnd {
		return false
	}

	return true
}

// 挂机收菜的收菜类型
type EHangUpRewardOpType = int8

const (
	EHungUpNormal  EHangUpRewardOpType = 1 // 正常挂机收菜
	EHungUpSpeedUp EHangUpRewardOpType = 2 // 加速挂机收菜
)

// 竞技场段位
type EArenaRange = int32

const (
	EArenaRangeLegend     EArenaRange = 1 //传奇
	EArenaRangeChief      EArenaRange = 2 //首席
	EArenaRangeMaster     EArenaRange = 3 //大师
	EArenaRangeExpert     EArenaRange = 4 //专家
	EArenaRangeElite      EArenaRange = 5 //精英
	EArenaRangeRookie     EArenaRange = 6 //新秀
	EArenaRangeInternship EArenaRange = 7 //见习

	EArenaRangeEnd
	EArenaRangeMax = EArenaRangeEnd + 1
)

// 是否是合法的竞技场段位
func IsValidArenaRange(arenaRange EArenaRange) bool {
	return arenaRange >= EArenaRangeLegend && arenaRange <= EArenaRangeEnd
}

// 是否是合法的竞技场小段编号
func IsValidArenaNode(nodeId int32) bool {
	return nodeId >= 1 && nodeId <= MAX_ARENA_NODE_ID
}

// 根据玩家Uid判断是否是机器人
func ArenaIsRobot(playerUid uint64) bool {
	if playerUid < 10000 {
		return true
	}
	return false
}

func CheckIsRobot(objId uint64) bool {
	return objId <= 1000000
}

const MAX_ARENA_BATTLE_DETAIL_COUNT = 15 // 最大战报数量
const MAX_ARENA_NODE_ID = 10             // 最小段数
const MAX_RANK_ITEM_COUNT = 50           // 排行榜发送的最大数据，只发送前 50 名
const NOT_IN_ARENA_PLAYER_POS = -1       // 没上榜的玩家的pos

const GuildNoticeMaxLen = int(300) //公会公告长度限制
const GuildLogAllocLen = int(30)   //公会日志预分配长度

// 公会职位枚举
type EGuildPositionType = int8

const (
	EGuildPositionMember    EGuildPositionType = 0 // 0_公会普通成员
	EGuildPositionManager   EGuildPositionType = 1 // 1_公会管理
	EGuildPositionPresident EGuildPositionType = 2 // 2_公会会长
)

// 公会权限类型枚举
type EGuildPermissionType = int8

const (
	EGuildPermissionModifyInfo        EGuildPermissionType = 1 // 1_公会个性化 修改公会个性化信息,包括公会名称、icon、入会限制、公告等
	EGuildPermissionJoinGuild         EGuildPermissionType = 2 // 2_入会请求处理 入会请求处理
	EGuildPermissionKickManager       EGuildPermissionType = 3 // 3_踢管理 把管理人员踢出公会
	EGuildPermissionKickMember        EGuildPermissionType = 4 // 4_踢会员 把普通人员踢出公会
	EGuildPermissionManagePosition    EGuildPermissionType = 5 // 5_职位管理 任命/收回管理
	EGuildPermissionTransferPresident EGuildPermissionType = 6 // 6_转让会长
	EGuildPermissionLevelUp           EGuildPermissionType = 7 // 7_公会升级
)

// 公会个性化修改类型
type EGuildModifyInfoType = int32

const (
	EGuildModifyName   EGuildModifyInfoType = 1 // 1_修改公会名
	EGuildModifyNotify EGuildModifyInfoType = 2 // 2_修改公会公告
	EGuildModifySundry EGuildModifyInfoType = 3 // 3_修改公会杂项 包括入会限制,Icon等
)

// 公会日志类型
type EGuildLogType = int8

const (
	EGuildLogCreate            EGuildLogType = 1 //1_创建公会日志
	EGuildLogJoin              EGuildLogType = 2 //2_加入公会日志
	EGuildLogExit              EGuildLogType = 3 //3_退出公会日志
	EGuildLogKick              EGuildLogType = 4 //4_踢出公会日志
	EGuildLogAppoint           EGuildLogType = 5 //5_任命管理日志
	EGuildLogDepose            EGuildLogType = 6 //6_回收管理日志
	EGuildLogTransferPresident EGuildLogType = 7 //7_转让会长日志
)

// 公会会长变更类型
type EGuildPresidentChangeType = int8

const (
	EPresidentChangeExit     EGuildPresidentChangeType = 1 //退出公会
	EPresidentChangeTransfer EGuildPresidentChangeType = 2 //转移会长
	EPresidentChangeTimeOut  EGuildPresidentChangeType = 3 //超时不登录
)

// 异步消息类型
type EPlayerAsyncMsgType = int32

const (
	EPlayerAsyncMsgTypeArenaCoin EPlayerAsyncMsgType = 2 // 竞技币
)

// 系统消息类型
type ESysMsgType = int32

const (
	ESysMsgTypeDrawPetRarity  ESysMsgType = 1 //1_抽卡稀有度
	ESysMsgTypeBloodLineLevel ESysMsgType = 2 //2_宠物血统等级
	ESysMsgTypeIntimateLevel  ESysMsgType = 3 //3_宠物亲密度等级
	ESysMsgTypeJoinGuild      ESysMsgType = 4 //4_加入公会
	ESysMsgTypeExitGuild      ESysMsgType = 5 //5_退出公会
	ESysMsgTypeTransferGuild  ESysMsgType = 6 //6_转让公会
	ESysMsgTypeBecomeFriends  ESysMsgType = 7 //7_成为好友

	ESysMsgTypeHolder
	ESysMsgTypeMax = ESysMsgTypeHolder + 1
)

const ChatMsgMaxLen = 300          //聊天消息最大长度
const PrivateChatMaxNum = 50       // 私聊最大存储数量
const DefaultGuildChatSaveNum = 50 //默认的公会聊天存储条数

// 聊天频道类型
type EChatChannelType = int32

const (
	EChatChannelSys     EChatChannelType = 1 //1_系统频道
	EChatChannelWorld   EChatChannelType = 2 //2_世界频道
	EChatChannelGuild   EChatChannelType = 3 //3_公会频道
	EChatChannelPrivate EChatChannelType = 4 //4_私聊

	EChatChannelEnd
	EChatChannelMax = EChatChannelEnd + 1
)

// 判断是否是有效的聊天频道
func CheckChatChannelIsValid(chatChannel EChatChannelType) bool {
	if chatChannel > EChatChannelSys && chatChannel < EChatChannelMax {
		return true
	}
	return false
}

// 天赋方案套数
const TalentClassPlanCount = 2

func GetPassiveTypeIdAndLevel(passiveId int) (int, int) {
	return passiveId / 1000, passiveId % 1000
}

type EBuffExternalEventType = int8

const (
	NoneTriggerEvent          EBuffExternalEventType = 0 //没有
	BeSkillDamageTriggerEvent EBuffExternalEventType = 1 //1_被技能击中触发的事件
	BeBuffDamageTriggerEvent  EBuffExternalEventType = 2 //2_被Buff伤害击中触发的事件
	KillTriggerEvent          EBuffExternalEventType = 3 //3_被杀死触发的事件
	IntervalTriggerEvent      EBuffExternalEventType = 4 //4_Buff脉冲触发的事件
	BeCtDamageTriggerEvent    EBuffExternalEventType = 5 //5_受到暴击伤害触发事件

	BuffExternalEventEnd
	BuffExternalEventMax = BuffExternalEventEnd + 1
)

// 虚无之地效果池子类型
type ENihilityEffectPoolType = int32

const (
	ENihilityEffectPoolPetId   ENihilityEffectPoolType = 1 // 1_宠物Id
	ENihilityEffectPoolElement ENihilityEffectPoolType = 2 // 2_元素类型
	ENihilityEffectPoolCommon  ENihilityEffectPoolType = 3 // 3_通用
)

// 一次随机的个数
const NihilityEffectPoolCount int = 3
const NihilityEffectEffectInvalidCount int32 = 999999

// 血统因子默认等级
const BloodlineEquipInitLevel int32 = 0

// 血统因子类型
type EBloodlineEquipType = int32

// 注:因子类型对服务器来说没有实际意义，只有同一宠物不允许穿相同类型的因子(且不会用到枚举)，但因子种类太多，所以简化枚举命名
const (
	BEquipTypeInit EBloodlineEquipType = 0  //0_天生因子
	BEquipType1    EBloodlineEquipType = 1  //1_因子类型1  生命值因子
	BEquipType2    EBloodlineEquipType = 2  //2_因子类型2 生命万分比因子
	BEquipType3    EBloodlineEquipType = 3  //3_因子类型3 攻击值因子
	BEquipType4    EBloodlineEquipType = 4  //4_因子类型4 攻击万分比因子
	BEquipType5    EBloodlineEquipType = 5  //5_因子类型5 防御值因子
	BEquipType6    EBloodlineEquipType = 6  //6_因子类型6 防御万分比因子
	BEquipType7    EBloodlineEquipType = 7  //7_因子类型7 命中值因子
	BEquipType8    EBloodlineEquipType = 8  //8_因子类型8 命中万分比因子
	BEquipType9    EBloodlineEquipType = 9  //9_因子类型9 闪避值因子
	BEquipType10   EBloodlineEquipType = 10 //10_因子类型10 闪避万分比因子
	BEquipType11   EBloodlineEquipType = 11 //11_因子类型11 穿透值因子
	BEquipType12   EBloodlineEquipType = 12 //12_因子类型12 穿透万分比因子
	BEquipType13   EBloodlineEquipType = 13 //13_因子类型13 护甲值因子
	BEquipType14   EBloodlineEquipType = 14 //14_因子类型14 护甲万分比因子
	BEquipType15   EBloodlineEquipType = 15 //15_因子类型15 暴击值因子
	BEquipType16   EBloodlineEquipType = 16 //16_因子类型16 暴击万分比因子
	BEquipType17   EBloodlineEquipType = 17 //17_因子类型17 抗暴值因子
	BEquipType18   EBloodlineEquipType = 18 //18_因子类型18 抗暴万分比因子
	BEquipType19   EBloodlineEquipType = 19 //19_因子类型19 命中率值因子
	BEquipType20   EBloodlineEquipType = 20 //20_因子类型20 闪避率值因子
	BEquipType21   EBloodlineEquipType = 21 //21_因子类型21 穿透率值因子
	BEquipType22   EBloodlineEquipType = 22 //22_因子类型22 减伤率值因子
	BEquipType23   EBloodlineEquipType = 23 //23_因子类型23 暴击率值因子
	BEquipType24   EBloodlineEquipType = 24 //24_因子类型24 暴伤率值因子
	BEquipType25   EBloodlineEquipType = 25 //25_因子类型25 暴伤抵抗率值因子
	BEquipType26   EBloodlineEquipType = 26 //26_因子类型26 攻速值因子
	BEquipType27   EBloodlineEquipType = 27 //27_因子类型27 抗暴率值因子
	BEquipType28   EBloodlineEquipType = 28 //28_因子类型28
	BEquipType29   EBloodlineEquipType = 29 //29_因子类型29
	BEquipType30   EBloodlineEquipType = 30 //30_因子类型30
	BEquipType31   EBloodlineEquipType = 31 //31_因子类型31
	BEquipType32   EBloodlineEquipType = 32 //32_因子类型32
	BEquipType33   EBloodlineEquipType = 33 //33_因子类型33
	BEquipType34   EBloodlineEquipType = 34 //34_因子类型34
	BEquipType35   EBloodlineEquipType = 35 //35_因子类型35
	BEquipType36   EBloodlineEquipType = 36 //36_因子类型36
	BEquipType37   EBloodlineEquipType = 37 //37_因子类型37
	BEquipType38   EBloodlineEquipType = 38 //38_因子类型38
	BEquipType39   EBloodlineEquipType = 39 //39_因子类型39
	BEquipType40   EBloodlineEquipType = 40 //40_因子类型40

	BEquipTypeEnd
	BEquipTypeMax = BEquipTypeEnd + 1
)

// 检测血统因子类型是否合法
func CheckBloodlineEquipType(equipType EBloodlineEquipType) bool {
	if equipType < BEquipTypeInit || equipType > BEquipTypeEnd {
		return false
	}

	return true
}

// 获得奖励的时候，一些特殊的标记
type GoodsInfoType = int32

const (
	GoodsInfoTypeFromAutoReleasePet GoodsInfoType = 1 // 来自宠物自动转化
)

// 被动生效对象 对应keyword中的EBLUnlockPassiveUseObjType
type EPassiveUseObjType = int32

const (
	EPUseObjUnknown  EPassiveUseObjType = 0 //未知生效对象,不生效
	EPUseObjPet      EPassiveUseObjType = 1 //1_宠物生效
	ELUPUseObjPlayer EPassiveUseObjType = 5 //5_玩家生效

	EPUseObjEnd
	EPUseObjMax = EPUseObjEnd + 1
)

// 检查血统因子等级解锁的能力生效对象是否正确
func CheckBLUnlockPassiveUseObjType(eType EPassiveUseObjType) bool {
	if eType > EPUseObjUnknown && eType < EPUseObjMax {
		return true
	}

	return false
}

// 白名单类型
type WhiteListType = int32

const (
	WhiteListIp         WhiteListType = 1 // ip
	WhiteListChannelId  WhiteListType = 2 // 渠道号【一个渠道一个】
	WhiteListDevId      WhiteListType = 3 // 设备号
	WhiteListChannelUid WhiteListType = 4 // 渠道账号【玩家级别】
)

// 分享类型
type EShareType = int32

const (
	EShareTypeNormal EShareType = 0 //0_通用分享
	EShareTypePet    EShareType = 1 //1_宠物分享,游戏外分享
	EShareTypeRank   EShareType = 2 //2_排行榜分享
)

// 在线奖励类型
type EOnlineAwardType = int32

const (
	EOnlineAwardTypeCommon EShareType = 0 //0_普通奖励
	EOnlineAwardTypeGood   EShareType = 1 //1_好奖励
)

// 匹配类型
type EMatchType = int32

const (
	EMatchTypeNormal        EMatchType = 0 // 0_队列匹配
	EMatchTypePlayerLevel   EMatchType = 1 // 1_玩家等级匹配
	EMatchTypeHomeTaskLevel EMatchType = 2 // 2_神树信仰等级匹配

	MatchEnd
	MatchMax = MatchEnd + 1
)

const PreMatchPlayerCount = 100
const PreMatchArmyCount = 10

const (
	Day   = 24 * time.Hour
	Month = 30 * Day
)

const (
	StopServerAI          int32 = 0x01                        //关闭服务器AI
	StopClientAi          int32 = 0x02                        //关闭客户端AI
	StopServerAndClientAi       = StopServerAI | StopClientAi //关闭全部AI
)

// 关卡报警类型
type EAutoWarningFocusType = int

const (
	AutoWarningFocusTypeNone         EAutoWarningFocusType = 0 // 0_手动
	AutoWarningFocusTypeBossAndElite EAutoWarningFocusType = 1 // 1_boss精英警告
	AutoWarningFocusTypeOnlyBoss     EAutoWarningFocusType = 2 // 2_boss警告
)

// 营地帐篷增益类型
type ECampTentsAdditionType = int32

const (
	ECampTentsTypeNone             ECampTentsAdditionType = 0 // 0_帐篷无增益
	ECampTentsTypeScheduleCount    ECampTentsAdditionType = 1 // 1_帐篷增益派遣次数
	ECampTentsTypeCookExtAwardRate ECampTentsAdditionType = 2 // 2_帐篷增益烹饪额外奖励概率
	ECampTentsTypeFeedAttrRate     ECampTentsAdditionType = 3 // 3_帐篷增益喂食属性成功率

	ECampTentsTypeEnd
	ECampTentsTypeMax = ECampTentsTypeEnd + 1
)

// 挂机营地摆件类型
type ECampStuffType = int32

const (
	CampStuffTypeCamp ECampStuffType = 2 //2_营地 帐篷
)

// 多人匹配中玩家队伍的同步
type EMatchPlayerUpdateType = int32

const (
	EMatchPlayerAll EMatchPlayerUpdateType = 0 // 全更新
	EMatchPlayerAdd EMatchPlayerUpdateType = 1 // 增加
	EMatchPlayerDel EMatchPlayerUpdateType = 2 // 删除
)

// 地图规则限定类型
type EMapRuleRestrictType = uint32

const (
	EMapRestrictFoodBuff EMapRuleRestrictType = 1 //1_地图限制食物BUFF
	//例子：
	EMapRestrictTest1 = 1 << 1
	EMapRestrictTest2 = 1 << 2
)

// 日常任务组类型
type EDailyTaskGroupType = int32

const (
	EDailyTaskGroupNone    EDailyTaskGroupType = 0 //0_无效任务组类型
	EDailyTaskGroupCoexist EDailyTaskGroupType = 1 //1_并存型
	EDailyTaskGroupRandom  EDailyTaskGroupType = 2 //2_随机型
	EDailyTaskGroupCycle   EDailyTaskGroupType = 3 //3_每周轮循型

	EDailyTaskGroupTypeEnd
	EDailyTaskGroupTypeMax = EDailyTaskGroupTypeEnd + 1
)

// 日常任务组状态
type EDailyTaskGroupStateType = uint8

const (
	EDailyTaskGroupStateFinish EDailyTaskGroupStateType = 1 //1_任务组已完成
	EDailyTaskGroupStateAward  EDailyTaskGroupStateType = 2 //2_任务组已领奖
)

// 剧情对话选项执行操作类型
type ENpcTalkDoType = int8

const (
	ENTDoTypeEvent       ENpcTalkDoType = 1 // 1_事件
	ENTDoTypeDrop        ENpcTalkDoType = 2 // 2_掉落
	ENTDoTypeTimeLine    ENpcTalkDoType = 3 // 3_timeline
	ENTDoTypeJumpNpcTalk ENpcTalkDoType = 4 // 4_跳转NpcTalk

	ENTDoTypeEnd
	ENTDoTypeMax = ENTDoTypeEnd + 1
)

// 职业
type ERoleCareersType = int32

const (
	ERoleCareersNone ERoleCareersType = 0 //
	EHunter          ERoleCareersType = 1 // 猎手
	EMagician        ERoleCareersType = 2 // 术师
	EWarrior         ERoleCareersType = 3 // 斗士
	EBraver          ERoleCareersType = 4 // 勇气

	ERoleCareersTypeEnd
	ERoleCareersTypeMax = ERoleCareersTypeEnd + 1
)

func IsValidRoleCareersType(ct ERoleCareersType) bool {
	return ct >= EHunter && ct < ERoleCareersTypeMax
}

// 装备词条类型
type EEquipPropertyType = int8

const (
	EEquipProAttr    EEquipPropertyType = 1 // 1_属性词条
	EEquipProPassive EEquipPropertyType = 2 // 2_被动词条
	EEquipProSFX     EEquipPropertyType = 3 // 3_特技词条

	EEquipProEnd
	EEquipProMax = EEquipProEnd + 1
)

// 检查装备词条类型是否有效
func CheckEquipPropertyTypeIsValid(pType EEquipPropertyType) bool {
	if pType >= EEquipProAttr && pType < EEquipProMax {
		return true
	}
	return false
}

// 关卡通关点赞类型
type ELevelMapFinishLikeType = int32

const (
	FirstFinishLikeType  ELevelMapFinishLikeType = 1 //首通玩家点赞
	CareerFinishLikeType ELevelMapFinishLikeType = 2 //本职业最低战力玩家点赞
	FastFinishLikeType   ELevelMapFinishLikeType = 3 //最快通关玩家点赞
	LatestFinishLikeType ELevelMapFinishLikeType = 4 //最近通关玩家点赞
	LowestFightLikeType  ELevelMapFinishLikeType = 5 //最低战力通关玩家点赞
)

// 御灵道场关卡品质类型
type ESpiritDojoLevelQualityType = int32

const (
	ESpiritDojoLevelQualityNone   ESpiritDojoLevelQualityType = 0 //0_无效
	ESpiritDojoLevelQualityBlue   ESpiritDojoLevelQualityType = 1 //1_蓝色关卡
	ESpiritDojoLevelQualityPurple ESpiritDojoLevelQualityType = 2 //2_紫色关卡
	ESpiritDojoLevelQualityOrange ESpiritDojoLevelQualityType = 3 //3_橙色关卡

	ESpiritDojoLevelQualityMax
)

// 购买限制类型
type EShopBuyLimit = int32

const (
	EShopBuyLimitNoLimit EShopBuyLimit = 0 // 0_无限制
	EShopBuyLimitDay     EShopBuyLimit = 1 // 1_日限制
	EShopBuyLimitWeek    EShopBuyLimit = 2 // 2_周限制
	EShopBuyLimitMonth   EShopBuyLimit = 3 // 3_月限制
	EShopBuyLimitNever   EShopBuyLimit = 4 // 4_终生限制
)

// 查找匹配类型
type EFindMatchUpType = int32

const (
	EFindMatchUpTypeDown EFindMatchUpType = 1 //向下匹配查找
	EFindMatchUpTypeUp   EFindMatchUpType = 2 //向上匹配查找
	EFindMatchUpTypeMax
)

// 机器人玩法类型
type ERobotPlayType = int32

const (
	ENonePlayRobotType        ERobotPlayType = 0
	ESpiritDojoRobotType      ERobotPlayType = 1 //1_御灵道场机器人
	ESpiritChallengeRobotType ERobotPlayType = 2 //2_御灵挑战机器人

	EEndPlayRobotType
	EMaxPlayRobotType = EEndPlayRobotType + 1
)

// 幻兽分类
type EPetClassType = int8

const (
	EPetClassCommon   EPetClassType = 1 //1_普通幻兽
	EPetClassSpecial  EPetClassType = 2 //2_特殊珍兽
	EPetClassPrecious EPetClassType = 3 //3_珍兽
	EPetClassMythical EPetClassType = 4 //4_神兽

	EPetClassEnd
	EPetClassMax = EPetClassEnd + 1
)

type EMatchGameType = int32

const (
	EMatchGameTypeMatch       EMatchGameType = 1 // 1_组队匹配
	EMatchGameTypeTwistedLand EMatchGameType = 2 // 2_扭曲之地

	EMatchGameTypeEnd
	EMatchGameTypeMax = EMatchGameTypeEnd + 1
)

func IsValidMatchGameType(ct EMatchGameType) bool {
	return ct >= EMatchGameTypeMatch && ct < EMatchGameTypeMax
}

// 邮件内容显示类型
type EMailShowParamType = int32

const (
	EMailShowParamTypeNone                       EMailShowParamType = 0 // 0_无
	EMailShowParamTypeMatchTwistedLandFirstAward EMailShowParamType = 1 // 1_扭曲之地首通奖励
)

type EShopFlushMode = int32

const (
	EShopFlushModeOpenServerRunning EShopFlushMode = 1 // 按开服时间开始，每多少毫秒刷新
	EShopFlushModeHour              EShopFlushMode = 2 // 指定的时间（24小时）刷新
)

// 系统邮件来源类型
type ESysMailFromType = int8

const (
	ESMFTypeLevel        ESysMailFromType = 0 // 0_邮件来源玩家等级
	ESMFTypeExchangeCode ESysMailFromType = 1 // 1_邮件来源兑换码
	ESMFTypeQQ           ESysMailFromType = 2 // 2_邮件来源QQ群
	ESMFTypeSurvey       ESysMailFromType = 3 // 3_邮件来源问卷调查
	ESMFTypeGuild        ESysMailFromType = 4 // 4_邮件来源公会
	ESMFTypeTwisted      ESysMailFromType = 5 // 5_邮件来源扭曲之地
)

// 邮件来源问卷调查参数ID
type ESysMailFromSurveyType = int32

const (
	ESurveyMailFir ESysMailFromSurveyType = 0 // 0_问卷调查1
	ESurveyMailSec ESysMailFromSurveyType = 1 // 1_问卷调查2
)

// 邮件来源公会参数ID
type ESysMailFromGuildType = int32

const (
	EGuildMailBePresident      ESysMailFromGuildType = 0 // 0_成为会长
	EGuildMailRetiredPresident ESysMailFromGuildType = 1 // 1_卸任会长
)

// 邮件来源扭曲之地参数ID
type ESysMailFromTwistedType = int32

const (
	ETwistedMailHasNextDifficulty  ESysMailFromTwistedType = 0 // 0_扭曲之地首通有下一难度
	ETwistedMailNoneNextDifficulty ESysMailFromTwistedType = 1 // 1_扭曲之地首通没有下一难度
)

// 宠物因子穿戴类型
type EBloodlineEquipPosType = int32

const (
	EBloodlineEquipPosType0 EBloodlineEquipPosType = 0 // 0_宠物因子穿戴0类型
	EBloodlineEquipPosType1 EBloodlineEquipPosType = 1 // 1_宠物因子穿戴1类型
	EBloodlineEquipPosType2 EBloodlineEquipPosType = 2 // 2_宠物因子穿戴2类型

	EBloodlineEquipPosTypeEnd
	EBloodlineEquipPosTypeMax = EBloodlineEquipPosTypeEnd + 1
)

func IsBloodlineEquipPosType(ct EBloodlineEquipPosType) bool {
	return ct >= EBloodlineEquipPosType0 && ct < EBloodlineEquipPosTypeMax
}

type EQueryCrossMapRet = int32

const (
	EQueryCrossMapSuccess             EQueryCrossMapRet = 0
	EQueryCrossMapMapStatusError      EQueryCrossMapRet = 1 // 玩家在战斗服状态错误
	EQueryCrossMapMapPlayerNotInCross EQueryCrossMapRet = 2 // 玩家不在战斗服了
)

// 关卡行为限制类型
type EMapRestrictType = int

const (
	ERestrictTypeNone                EMapRestrictType = 0      //0_无
	ERestrictTypePlayerMove          EMapRestrictType = 1 << 0 //1_禁止玩家移动
	ERestrictTypeCastSkillExceptLink EMapRestrictType = 1 << 1 //2_禁止释放除连携外的技能
)

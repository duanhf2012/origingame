// Package rpcapi 定义 OriginGame 各 Service 共享的 Origin RPC 契约。
package rpcapi

import "context"

// DBService 是业务无关的 MongoDB 与 Redis 执行模板。
//
//origin:rpc
type DBService interface {
	// ExecuteMongo 在当前 DBService 固定绑定的数据库中执行 MongoDB 请求。
	ExecuteMongo(context.Context, MongoRequest) (MongoResult, error)
	// ExecuteRedis 在当前 DBService 固定绑定的 Redis 中执行请求。
	ExecuteRedis(context.Context, RedisRequest) (RedisResult, error)
}

// MongoExecuteMode 指定多操作请求的执行语义。
type MongoExecuteMode int32

const (
	MongoExecuteModeUnspecified MongoExecuteMode = iota
	MongoExecuteModeSequential
	MongoExecuteModeTransaction
)

// MongoOperationKind 指定一个结构化 MongoDB 操作。
type MongoOperationKind int32

const (
	MongoOperationKindUnspecified MongoOperationKind = iota
	MongoOperationKindInsertOne
	MongoOperationKindInsertMany
	MongoOperationKindFindOne
	MongoOperationKindFindMany
	MongoOperationKindCountDocuments
	MongoOperationKindEstimatedDocumentCount
	MongoOperationKindAggregate
	MongoOperationKindUpdateOne
	MongoOperationKindUpdateMany
	MongoOperationKindReplaceOne
	MongoOperationKindFindOneAndUpdate
	MongoOperationKindFindOneAndReplace
	MongoOperationKindFindOneAndDelete
	MongoOperationKindDeleteOne
	MongoOperationKindDeleteMany
	MongoOperationKindRawCommand
)

// MongoUpdateKind 区分普通更新文档和更新 Pipeline。
type MongoUpdateKind int32

const (
	MongoUpdateKindUnspecified MongoUpdateKind = iota
	MongoUpdateKindDocument
	MongoUpdateKindPipeline
)

// MongoReturnDocument 指定 FindOneAndXxx 返回修改前或修改后的文档。
type MongoReturnDocument int32

const (
	MongoReturnDocumentUnspecified MongoReturnDocument = iota
	MongoReturnDocumentBefore
	MongoReturnDocumentAfter
)

// MongoOperationStatus 是单个操作的最终状态。
type MongoOperationStatus int32

const (
	MongoOperationStatusUnspecified MongoOperationStatus = iota
	MongoOperationStatusSucceeded
	MongoOperationStatusFailed
	MongoOperationStatusNotExecuted
	MongoOperationStatusRolledBack
	MongoOperationStatusStateUnknown
)

// MongoFailureKind 是跨 RPC 稳定的 MongoDB 失败分类。
type MongoFailureKind int32

const (
	MongoFailureKindUnspecified MongoFailureKind = iota
	MongoFailureKindDuplicateKey
	MongoFailureKindWriteConflict
	MongoFailureKindDocumentValidation
	MongoFailureKindExpectationFailed
	MongoFailureKindUnsupported
	MongoFailureKindUnauthorized
	MongoFailureKindCommand
	MongoFailureKindTransactionAborted
	MongoFailureKindResultLimitExceeded
	MongoFailureKindTimeout
	MongoFailureKindCanceled
	MongoFailureKindNetwork
	MongoFailureKindUnknown
)

// MongoRequest 是一次同 Key 有序执行的 MongoDB 请求。
type MongoRequest struct {
	DispatchKey string           // 保证同 Key 有序的路由键。
	ExecuteMode MongoExecuteMode // 多操作执行语义。
	Operations  []MongoOperation // 待执行的 MongoDB 操作。
}

// MongoOperation 是 MongoDB 操作的判别联合；Kind 对应的参数必须恰好一个非空。
type MongoOperation struct {
	Kind        MongoOperationKind // 判别联合的操作类型。
	Collection  string             // 目标集合名称。
	Expectation *MongoExpectation  // 操作结果数量断言。

	InsertOne              *MongoInsertOne              // 单文档插入参数。
	InsertMany             *MongoInsertMany             // 多文档插入参数。
	FindOne                *MongoFindOne                // 单文档查询参数。
	FindMany               *MongoFindMany               // 多文档查询参数。
	CountDocuments         *MongoCountDocuments         // 文档计数参数。
	EstimatedDocumentCount *MongoEstimatedDocumentCount // 估算文档计数标记。
	Aggregate              *MongoAggregate              // 聚合管道参数。
	UpdateOne              *MongoUpdateOne              // 单文档更新参数。
	UpdateMany             *MongoUpdateMany             // 多文档更新参数。
	ReplaceOne             *MongoReplaceOne             // 单文档替换参数。
	FindOneAndUpdate       *MongoFindOneAndUpdate       // 查询并更新参数。
	FindOneAndReplace      *MongoFindOneAndReplace      // 查询并替换参数。
	FindOneAndDelete       *MongoFindOneAndDelete       // 查询并删除参数。
	DeleteOne              *MongoDeleteOne              // 单文档删除参数。
	DeleteMany             *MongoDeleteMany             // 多文档删除参数。
	RawCommand             *MongoRawCommand             // 受控原始命令参数。
}

type MongoInsertOne struct {
	Document []byte // BSON 文档。
}

type MongoInsertMany struct {
	Documents [][]byte // BSON 文档列表。
	Unordered bool     // 是否允许无序写入。
}

type MongoFindOne struct {
	Filter     []byte // BSON 过滤条件。
	Projection []byte // BSON 投影条件。
	Sort       []byte // BSON 排序条件。
}

type MongoFindMany struct {
	Filter     []byte // BSON 过滤条件。
	Projection []byte // BSON 投影条件。
	Sort       []byte // BSON 排序条件。
	Limit      int64  // 返回文档数量上限。
}

type MongoCountDocuments struct {
	Filter []byte // BSON 过滤条件。
}

// MongoEstimatedDocumentCount 是无参数操作的显式标记；值必须为 true。
// 使用具名 bool 是为了让静态 RPC Codec 保留判别联合中的存在性。
type MongoEstimatedDocumentCount bool

type MongoAggregate struct {
	Pipeline     [][]byte // BSON 聚合阶段列表。
	MaxDocuments int64    // 返回文档数量上限。
	AllowDiskUse bool     // 是否允许服务端磁盘暂存。
}

// MongoUpdate 保存普通更新文档或更新 Pipeline，二者只能选择一个。
type MongoUpdate struct {
	Kind     MongoUpdateKind // 更新参数编码类型。
	Document []byte          // BSON 更新文档。
	Pipeline [][]byte        // BSON 更新阶段列表。
}

type MongoUpdateOne struct {
	Filter       []byte      // BSON 过滤条件。
	Update       MongoUpdate // 更新内容。
	Upsert       bool        // 未匹配时是否插入。
	ArrayFilters [][]byte    // BSON 数组过滤条件。
}

type MongoUpdateMany struct {
	Filter       []byte      // BSON 过滤条件。
	Update       MongoUpdate // 更新内容。
	ArrayFilters [][]byte    // BSON 数组过滤条件。
}

type MongoReplaceOne struct {
	Filter      []byte // BSON 过滤条件。
	Replacement []byte // 替换用 BSON 文档。
	Upsert      bool   // 未匹配时是否插入。
}

type MongoFindOneAndUpdate struct {
	Filter         []byte              // BSON 过滤条件。
	Update         MongoUpdate         // 更新内容。
	Projection     []byte              // BSON 投影条件。
	Sort           []byte              // BSON 排序条件。
	Upsert         bool                // 未匹配时是否插入。
	ReturnDocument MongoReturnDocument // 返回修改前或后的文档。
	ArrayFilters   [][]byte            // BSON 数组过滤条件。
}

type MongoFindOneAndReplace struct {
	Filter         []byte              // BSON 过滤条件。
	Replacement    []byte              // 替换用 BSON 文档。
	Projection     []byte              // BSON 投影条件。
	Sort           []byte              // BSON 排序条件。
	Upsert         bool                // 未匹配时是否插入。
	ReturnDocument MongoReturnDocument // 返回修改前或后的文档。
}

type MongoFindOneAndDelete struct {
	Filter     []byte // BSON 过滤条件。
	Projection []byte // BSON 投影条件。
	Sort       []byte // BSON 排序条件。
}

type MongoDeleteOne struct {
	Filter []byte // BSON 过滤条件。
}

type MongoDeleteMany struct {
	Filter []byte // BSON 过滤条件。
}

// MongoRawCommand 调用 DBService 启动阶段登记的受控原始命令。
type MongoRawCommand struct {
	ID           string // 启动时登记的命令标识。
	Command      []byte // BSON 原始命令文档。
	MaxDocuments int64  // 返回文档数量上限。
}

// MongoExpectation 对单个操作的可观察数量做通用断言。
type MongoExpectation struct {
	Documents *MongoCountRange // 返回文档数量范围。
	Inserted  *MongoCountRange // 插入数量范围。
	Matched   *MongoCountRange // 匹配数量范围。
	Modified  *MongoCountRange // 修改数量范围。
	Deleted   *MongoCountRange // 删除数量范围。
	Upserted  *MongoCountRange // Upsert 数量范围。
}

type MongoCountRange struct {
	Min int64 // 允许的最小数量。
	Max int64 // 允许的最大数量。
}

// MongoResult 保留全部操作的逐项状态和首个整体失败。
type MongoResult struct {
	Results []MongoOperationResult // 各操作执行结果。
	Failure *MongoExecutionFailure // 首个整体执行失败。
}

type MongoOperationResult struct {
	Status          MongoOperationStatus // 最终执行状态。
	Documents       [][]byte             // 查询或聚合返回的 BSON 文档。
	Count           int64                // 文档计数结果。
	InsertedCount   int64                // 实际插入数量。
	MatchedCount    int64                // 更新匹配数量。
	ModifiedCount   int64                // 实际修改数量。
	DeletedCount    int64                // 实际删除数量。
	UpsertedCount   int64                // Upsert 数量。
	InsertedIDs     []MongoBSONValue     // 插入文档主键。
	UpsertedIDs     []MongoBSONValue     // Upsert 文档主键。
	WriteFailures   []MongoWriteFailure  // 逐文档写入失败。
	CommandResponse []byte               // 原始命令响应 BSON。
}

type MongoBSONValue struct {
	Type  byte   // BSON 值类型编号。
	Value []byte // BSON 值编码。
}

type MongoExecutionFailure struct {
	OperationIndex int32            // 发生失败的操作索引。
	Kind           MongoFailureKind // 跨 RPC 稳定失败分类。
	ServerCode     int32            // MongoDB 服务端错误码。
	StateUnknown   bool             // 是否无法确认写入状态。
}

type MongoWriteFailure struct {
	DocumentIndex int32            // 失败文档在批次中的索引。
	Kind          MongoFailureKind // 跨 RPC 稳定失败分类。
	ServerCode    int32            // MongoDB 服务端错误码。
}

// RedisExecuteMode 指定命令组合的执行语义。
type RedisExecuteMode int32

const (
	RedisExecuteModeUnspecified RedisExecuteMode = iota
	RedisExecuteModeCommand
	RedisExecuteModePipeline
	RedisExecuteModeTransaction
	RedisExecuteModeScript
)

// RedisValueKind 描述 Redis 返回值节点的跨 RPC 类型。
type RedisValueKind int32

const (
	RedisValueKindUnspecified RedisValueKind = iota
	RedisValueKindNull
	RedisValueKindBytes
	RedisValueKindInteger
	RedisValueKindBigNumber
	RedisValueKindDouble
	RedisValueKindBoolean
	RedisValueKindArray
	RedisValueKindMap
)

// RedisCommandStatus 是单条命令的最终状态。
type RedisCommandStatus int32

const (
	RedisCommandStatusUnspecified RedisCommandStatus = iota
	RedisCommandStatusSucceeded
	RedisCommandStatusFailed
	RedisCommandStatusNotExecuted
	RedisCommandStatusStateUnknown
)

// RedisFailureKind 是跨 RPC 稳定的 Redis 失败分类。
type RedisFailureKind int32

const (
	RedisFailureKindUnspecified RedisFailureKind = iota
	RedisFailureKindWrongType
	RedisFailureKindCommand
	RedisFailureKindNoScript
	RedisFailureKindTransactionAborted
	RedisFailureKindCrossSlot
	RedisFailureKindUnauthorized
	RedisFailureKindResultLimitExceeded
	RedisFailureKindTimeout
	RedisFailureKindCanceled
	RedisFailureKindNetwork
	RedisFailureKindUnknown
)

// RedisRequest 是一次同 Key 有序执行的 Redis 请求。
type RedisRequest struct {
	DispatchKey string           // 保证同 Key 有序的路由键。
	ExecuteMode RedisExecuteMode // 命令组合执行语义。
	Commands    []RedisCommand   // 命令或管道内容。
	Script      *RedisScriptCall // 受控 Script 调用参数。
}

type RedisCommand struct {
	Name string   // 大写 Redis 命令名称。
	Args [][]byte // 命令二进制参数。
}

type RedisScriptCall struct {
	ID   string   // 启动时登记的 Script 标识。
	Keys []string // Script 键参数。
	Args [][]byte // Script 二进制参数。
}

// RedisValue 使用节点表和索引表达非递归的嵌套结果。
type RedisValue struct {
	RootIndex uint32           // 根节点在 Nodes 中的索引。
	Nodes     []RedisValueNode // 扁平化结果节点表。
}

type RedisValueNode struct {
	Kind     RedisValueKind // 节点值类型。
	Bytes    []byte         // 字节串或大数文本。
	Integer  int64          // 整数值。
	Double   float64        // 双精度值。
	Boolean  bool           // 布尔值。
	Children []uint32       // 数组或映射的子节点索引。
}

type RedisResult struct {
	Results []RedisCommandResult   // 各命令或 Script 执行结果。
	Failure *RedisExecutionFailure // 首个整体执行失败。
}

type RedisCommandResult struct {
	Status  RedisCommandStatus   // 最终执行状态。
	Value   RedisValue           // 成功时的结构化返回值。
	Failure *RedisCommandFailure // 命令级失败信息。
}

type RedisCommandFailure struct {
	Kind RedisFailureKind // 跨 RPC 稳定失败分类。
}

type RedisExecutionFailure struct {
	Kind         RedisFailureKind // 跨 RPC 稳定失败分类。
	StateUnknown bool             // 是否无法确认写入状态。
}

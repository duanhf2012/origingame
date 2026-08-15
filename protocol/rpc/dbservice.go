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
	DispatchKey string
	ExecuteMode MongoExecuteMode
	Operations  []MongoOperation
}

// MongoOperation 是 MongoDB 操作的判别联合；Kind 对应的参数必须恰好一个非空。
type MongoOperation struct {
	Kind        MongoOperationKind
	Collection  string
	Expectation *MongoExpectation

	InsertOne              *MongoInsertOne
	InsertMany             *MongoInsertMany
	FindOne                *MongoFindOne
	FindMany               *MongoFindMany
	CountDocuments         *MongoCountDocuments
	EstimatedDocumentCount *MongoEstimatedDocumentCount
	Aggregate              *MongoAggregate
	UpdateOne              *MongoUpdateOne
	UpdateMany             *MongoUpdateMany
	ReplaceOne             *MongoReplaceOne
	FindOneAndUpdate       *MongoFindOneAndUpdate
	FindOneAndReplace      *MongoFindOneAndReplace
	FindOneAndDelete       *MongoFindOneAndDelete
	DeleteOne              *MongoDeleteOne
	DeleteMany             *MongoDeleteMany
	RawCommand             *MongoRawCommand
}

type MongoInsertOne struct {
	Document []byte
}

type MongoInsertMany struct {
	Documents [][]byte
	Unordered bool
}

type MongoFindOne struct {
	Filter     []byte
	Projection []byte
	Sort       []byte
}

type MongoFindMany struct {
	Filter     []byte
	Projection []byte
	Sort       []byte
	Limit      int64
}

type MongoCountDocuments struct {
	Filter []byte
}

// MongoEstimatedDocumentCount 是无参数操作的显式标记；值必须为 true。
// 使用具名 bool 是为了让静态 RPC Codec 保留判别联合中的存在性。
type MongoEstimatedDocumentCount bool

type MongoAggregate struct {
	Pipeline     [][]byte
	MaxDocuments int64
	AllowDiskUse bool
}

// MongoUpdate 保存普通更新文档或更新 Pipeline，二者只能选择一个。
type MongoUpdate struct {
	Kind     MongoUpdateKind
	Document []byte
	Pipeline [][]byte
}

type MongoUpdateOne struct {
	Filter       []byte
	Update       MongoUpdate
	Upsert       bool
	ArrayFilters [][]byte
}

type MongoUpdateMany struct {
	Filter       []byte
	Update       MongoUpdate
	ArrayFilters [][]byte
}

type MongoReplaceOne struct {
	Filter      []byte
	Replacement []byte
	Upsert      bool
}

type MongoFindOneAndUpdate struct {
	Filter         []byte
	Update         MongoUpdate
	Projection     []byte
	Sort           []byte
	Upsert         bool
	ReturnDocument MongoReturnDocument
	ArrayFilters   [][]byte
}

type MongoFindOneAndReplace struct {
	Filter         []byte
	Replacement    []byte
	Projection     []byte
	Sort           []byte
	Upsert         bool
	ReturnDocument MongoReturnDocument
}

type MongoFindOneAndDelete struct {
	Filter     []byte
	Projection []byte
	Sort       []byte
}

type MongoDeleteOne struct {
	Filter []byte
}

type MongoDeleteMany struct {
	Filter []byte
}

// MongoRawCommand 调用 DBService 启动阶段登记的受控原始命令。
type MongoRawCommand struct {
	ID           string
	Command      []byte
	MaxDocuments int64
}

// MongoExpectation 对单个操作的可观察数量做通用断言。
type MongoExpectation struct {
	Documents *MongoCountRange
	Inserted  *MongoCountRange
	Matched   *MongoCountRange
	Modified  *MongoCountRange
	Deleted   *MongoCountRange
	Upserted  *MongoCountRange
}

type MongoCountRange struct {
	Min int64
	Max int64
}

// MongoResult 保留全部操作的逐项状态和首个整体失败。
type MongoResult struct {
	Results []MongoOperationResult
	Failure *MongoExecutionFailure
}

type MongoOperationResult struct {
	Status          MongoOperationStatus
	Documents       [][]byte
	Count           int64
	InsertedCount   int64
	MatchedCount    int64
	ModifiedCount   int64
	DeletedCount    int64
	UpsertedCount   int64
	InsertedIDs     []MongoBSONValue
	UpsertedIDs     []MongoBSONValue
	WriteFailures   []MongoWriteFailure
	CommandResponse []byte
}

type MongoBSONValue struct {
	Type  byte
	Value []byte
}

type MongoExecutionFailure struct {
	OperationIndex int32
	Kind           MongoFailureKind
	ServerCode     int32
	StateUnknown   bool
}

type MongoWriteFailure struct {
	DocumentIndex int32
	Kind          MongoFailureKind
	ServerCode    int32
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
	DispatchKey string
	ExecuteMode RedisExecuteMode
	Commands    []RedisCommand
	Script      *RedisScriptCall
}

type RedisCommand struct {
	Name string
	Args [][]byte
}

type RedisScriptCall struct {
	ID   string
	Keys []string
	Args [][]byte
}

// RedisValue 使用节点表和索引表达非递归的嵌套结果。
type RedisValue struct {
	RootIndex uint32
	Nodes     []RedisValueNode
}

type RedisValueNode struct {
	Kind     RedisValueKind
	Bytes    []byte
	Integer  int64
	Double   float64
	Boolean  bool
	Children []uint32
}

type RedisResult struct {
	Results []RedisCommandResult
	Failure *RedisExecutionFailure
}

type RedisCommandResult struct {
	Status  RedisCommandStatus
	Value   RedisValue
	Failure *RedisCommandFailure
}

type RedisCommandFailure struct {
	Kind RedisFailureKind
}

type RedisExecutionFailure struct {
	Kind         RedisFailureKind
	StateUnknown bool
}

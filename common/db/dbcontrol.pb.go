// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.0
// source: dbcontrol.proto

package db

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// 如果包含Insert+Update，如果存在由更新
type OptType int32

const (
	OptType_None             OptType = 0
	OptType_Find             OptType = 1
	OptType_Insert           OptType = 2
	OptType_Update           OptType = 4
	OptType_DelOne           OptType = 8
	OptType_Upset            OptType = 16
	OptType_FindOneAndUpdate OptType = 32
	OptType_Count            OptType = 64
	// InsertLog           = 128; //插入日志(日志插入专用)
	OptType_DelMany        OptType = 256   //删除
	OptType_UpdateMany     OptType = 512   //批量更新
	OptType_FindManyKey    OptType = 1024  //通过多个key查找
	OptType_UpdateFieldOpt OptType = 2048  //修改单个字段，支持修改缓存
	OptType_Redis_SetKey   OptType = 4096  //redis Set命令
	OptType_Redis_GetKey   OptType = 8192  //redis Get命令
	OptType_Redis_DelKey   OptType = 16384 //
)

// Enum value maps for OptType.
var (
	OptType_name = map[int32]string{
		0:     "None",
		1:     "Find",
		2:     "Insert",
		4:     "Update",
		8:     "DelOne",
		16:    "Upset",
		32:    "FindOneAndUpdate",
		64:    "Count",
		256:   "DelMany",
		512:   "UpdateMany",
		1024:  "FindManyKey",
		2048:  "UpdateFieldOpt",
		4096:  "Redis_SetKey",
		8192:  "Redis_GetKey",
		16384: "Redis_DelKey",
	}
	OptType_value = map[string]int32{
		"None":             0,
		"Find":             1,
		"Insert":           2,
		"Update":           4,
		"DelOne":           8,
		"Upset":            16,
		"FindOneAndUpdate": 32,
		"Count":            64,
		"DelMany":          256,
		"UpdateMany":       512,
		"FindManyKey":      1024,
		"UpdateFieldOpt":   2048,
		"Redis_SetKey":     4096,
		"Redis_GetKey":     8192,
		"Redis_DelKey":     16384,
	}
)

func (x OptType) Enum() *OptType {
	p := new(OptType)
	*p = x
	return p
}

func (x OptType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OptType) Descriptor() protoreflect.EnumDescriptor {
	return file_dbcontrol_proto_enumTypes[0].Descriptor()
}

func (OptType) Type() protoreflect.EnumType {
	return &file_dbcontrol_proto_enumTypes[0]
}

func (x OptType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OptType.Descriptor instead.
func (OptType) EnumDescriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{0}
}

type Sort struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SortField string `protobuf:"bytes,1,opt,name=SortField,proto3" json:"SortField,omitempty"`
	Asc       bool   `protobuf:"varint,2,opt,name=Asc,proto3" json:"Asc,omitempty"`
}

func (x *Sort) Reset() {
	*x = Sort{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbcontrol_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Sort) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Sort) ProtoMessage() {}

func (x *Sort) ProtoReflect() protoreflect.Message {
	mi := &file_dbcontrol_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Sort.ProtoReflect.Descriptor instead.
func (*Sort) Descriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{0}
}

func (x *Sort) GetSortField() string {
	if x != nil {
		return x.SortField
	}
	return ""
}

func (x *Sort) GetAsc() bool {
	if x != nil {
		return x.Asc
	}
	return false
}

type Placeholder struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Placeholder) Reset() {
	*x = Placeholder{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbcontrol_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Placeholder) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Placeholder) ProtoMessage() {}

func (x *Placeholder) ProtoReflect() protoreflect.Message {
	mi := &file_dbcontrol_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Placeholder.ProtoReflect.Descriptor instead.
func (*Placeholder) Descriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{1}
}

type FindManyKeyOption struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ConditionField string                  `protobuf:"bytes,1,opt,name=ConditionField,proto3" json:"ConditionField,omitempty"`                                                                                   //条件字段名，目前只支持单个字段
	Key            []string                `protobuf:"bytes,2,rep,name=key,proto3" json:"key,omitempty"`                                                                                                         //多key
	SelectField    map[string]*Placeholder `protobuf:"bytes,3,rep,name=SelectField,proto3" json:"SelectField,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` //选择的字段名称，注意如果字段名找不到，则不会报错，数据默认为空
}

func (x *FindManyKeyOption) Reset() {
	*x = FindManyKeyOption{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbcontrol_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindManyKeyOption) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindManyKeyOption) ProtoMessage() {}

func (x *FindManyKeyOption) ProtoReflect() protoreflect.Message {
	mi := &file_dbcontrol_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindManyKeyOption.ProtoReflect.Descriptor instead.
func (*FindManyKeyOption) Descriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{2}
}

func (x *FindManyKeyOption) GetConditionField() string {
	if x != nil {
		return x.ConditionField
	}
	return ""
}

func (x *FindManyKeyOption) GetKey() []string {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *FindManyKeyOption) GetSelectField() map[string]*Placeholder {
	if x != nil {
		return x.SelectField
	}
	return nil
}

type RedisKeyValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReidsKey string `protobuf:"bytes,1,opt,name=reidsKey,proto3" json:"reidsKey,omitempty"`
	Data     []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *RedisKeyValue) Reset() {
	*x = RedisKeyValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbcontrol_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RedisKeyValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RedisKeyValue) ProtoMessage() {}

func (x *RedisKeyValue) ProtoReflect() protoreflect.Message {
	mi := &file_dbcontrol_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RedisKeyValue.ProtoReflect.Descriptor instead.
func (*RedisKeyValue) Descriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{3}
}

func (x *RedisKeyValue) GetReidsKey() string {
	if x != nil {
		return x.ReidsKey
	}
	return ""
}

func (x *RedisKeyValue) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type DBControllerRedisReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type     OptType          `protobuf:"varint,1,opt,name=type,proto3,enum=db.OptType" json:"type,omitempty"`
	ModKey   uint64           `protobuf:"varint,2,opt,name=modKey,proto3" json:"modKey,omitempty"`
	KeyValue []*RedisKeyValue `protobuf:"bytes,3,rep,name=keyValue,proto3" json:"keyValue,omitempty"`
}

func (x *DBControllerRedisReq) Reset() {
	*x = DBControllerRedisReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbcontrol_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DBControllerRedisReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DBControllerRedisReq) ProtoMessage() {}

func (x *DBControllerRedisReq) ProtoReflect() protoreflect.Message {
	mi := &file_dbcontrol_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DBControllerRedisReq.ProtoReflect.Descriptor instead.
func (*DBControllerRedisReq) Descriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{4}
}

func (x *DBControllerRedisReq) GetType() OptType {
	if x != nil {
		return x.Type
	}
	return OptType_None
}

func (x *DBControllerRedisReq) GetModKey() uint64 {
	if x != nil {
		return x.ModKey
	}
	return 0
}

func (x *DBControllerRedisReq) GetKeyValue() []*RedisKeyValue {
	if x != nil {
		return x.KeyValue
	}
	return nil
}

type DBControllerRedisRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ret      int32            `protobuf:"varint,1,opt,name=Ret,proto3" json:"Ret,omitempty"`
	KeyValue []*RedisKeyValue `protobuf:"bytes,2,rep,name=keyValue,proto3" json:"keyValue,omitempty"`
}

func (x *DBControllerRedisRes) Reset() {
	*x = DBControllerRedisRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbcontrol_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DBControllerRedisRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DBControllerRedisRes) ProtoMessage() {}

func (x *DBControllerRedisRes) ProtoReflect() protoreflect.Message {
	mi := &file_dbcontrol_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DBControllerRedisRes.ProtoReflect.Descriptor instead.
func (*DBControllerRedisRes) Descriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{5}
}

func (x *DBControllerRedisRes) GetRet() int32 {
	if x != nil {
		return x.Ret
	}
	return 0
}

func (x *DBControllerRedisRes) GetKeyValue() []*RedisKeyValue {
	if x != nil {
		return x.KeyValue
	}
	return nil
}

type DBControllerReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type               OptType            `protobuf:"varint,1,opt,name=type,proto3,enum=db.OptType" json:"type,omitempty"`
	Key                string             `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	CollectName        string             `protobuf:"bytes,3,opt,name=collectName,proto3" json:"collectName,omitempty"`
	Condition          []byte             `protobuf:"bytes,4,opt,name=condition,proto3" json:"condition,omitempty"`
	SelectField        []byte             `protobuf:"bytes,5,opt,name=selectField,proto3" json:"selectField,omitempty"`
	MaxRow             int32              `protobuf:"varint,6,opt,name=maxRow,proto3" json:"maxRow,omitempty"`
	Sort               []*Sort            `protobuf:"bytes,7,rep,name=sort,proto3" json:"sort,omitempty"`
	Data               [][]byte           `protobuf:"bytes,8,rep,name=data,proto3" json:"data,omitempty"` //如果是OptType_FindOneAndUpdate，则data参数更新数据 rawData为完整插入数据
	RawData            [][]byte           `protobuf:"bytes,9,rep,name=rawData,proto3" json:"rawData,omitempty"`
	TimeoutMs          int32              `protobuf:"varint,10,opt,name=timeoutMs,proto3" json:"timeoutMs,omitempty"`
	AdditionalData     int32              `protobuf:"varint,11,opt,name=additionalData,proto3" json:"additionalData,omitempty"`
	Upsert             bool               `protobuf:"varint,12,opt,name=Upsert,proto3" json:"Upsert,omitempty"`  //是否插入
	CacheId            string             `protobuf:"bytes,13,opt,name=cacheId,proto3" json:"cacheId,omitempty"` //cacheId
	Version            int32              `protobuf:"varint,14,opt,name=version,proto3" json:"version,omitempty"`
	Ordered            bool               `protobuf:"varint,15,opt,name=Ordered,proto3" json:"Ordered,omitempty"` //是否有序, 插入
	Skip               int32              `protobuf:"varint,16,opt,name=Skip,proto3" json:"Skip,omitempty"`       //跳过
	ManyKeyCondition   *FindManyKeyOption `protobuf:"bytes,17,opt,name=ManyKeyCondition,proto3" json:"ManyKeyCondition,omitempty"`
	QueryDocumentCount bool               `protobuf:"varint,18,opt,name=QueryDocumentCount,proto3" json:"QueryDocumentCount,omitempty"` //是否查询文档总数量
	NotUseCache        bool               `protobuf:"varint,19,opt,name=NotUseCache,proto3" json:"NotUseCache,omitempty"`               //不使用cache
}

func (x *DBControllerReq) Reset() {
	*x = DBControllerReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbcontrol_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DBControllerReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DBControllerReq) ProtoMessage() {}

func (x *DBControllerReq) ProtoReflect() protoreflect.Message {
	mi := &file_dbcontrol_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DBControllerReq.ProtoReflect.Descriptor instead.
func (*DBControllerReq) Descriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{6}
}

func (x *DBControllerReq) GetType() OptType {
	if x != nil {
		return x.Type
	}
	return OptType_None
}

func (x *DBControllerReq) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *DBControllerReq) GetCollectName() string {
	if x != nil {
		return x.CollectName
	}
	return ""
}

func (x *DBControllerReq) GetCondition() []byte {
	if x != nil {
		return x.Condition
	}
	return nil
}

func (x *DBControllerReq) GetSelectField() []byte {
	if x != nil {
		return x.SelectField
	}
	return nil
}

func (x *DBControllerReq) GetMaxRow() int32 {
	if x != nil {
		return x.MaxRow
	}
	return 0
}

func (x *DBControllerReq) GetSort() []*Sort {
	if x != nil {
		return x.Sort
	}
	return nil
}

func (x *DBControllerReq) GetData() [][]byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *DBControllerReq) GetRawData() [][]byte {
	if x != nil {
		return x.RawData
	}
	return nil
}

func (x *DBControllerReq) GetTimeoutMs() int32 {
	if x != nil {
		return x.TimeoutMs
	}
	return 0
}

func (x *DBControllerReq) GetAdditionalData() int32 {
	if x != nil {
		return x.AdditionalData
	}
	return 0
}

func (x *DBControllerReq) GetUpsert() bool {
	if x != nil {
		return x.Upsert
	}
	return false
}

func (x *DBControllerReq) GetCacheId() string {
	if x != nil {
		return x.CacheId
	}
	return ""
}

func (x *DBControllerReq) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *DBControllerReq) GetOrdered() bool {
	if x != nil {
		return x.Ordered
	}
	return false
}

func (x *DBControllerReq) GetSkip() int32 {
	if x != nil {
		return x.Skip
	}
	return 0
}

func (x *DBControllerReq) GetManyKeyCondition() *FindManyKeyOption {
	if x != nil {
		return x.ManyKeyCondition
	}
	return nil
}

func (x *DBControllerReq) GetQueryDocumentCount() bool {
	if x != nil {
		return x.QueryDocumentCount
	}
	return false
}

func (x *DBControllerReq) GetNotUseCache() bool {
	if x != nil {
		return x.NotUseCache
	}
	return false
}

type FindManyKeyData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key  string   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`   //查找回来的key
	Data [][]byte `protobuf:"bytes,2,rep,name=data,proto3" json:"data,omitempty"` //数据
}

func (x *FindManyKeyData) Reset() {
	*x = FindManyKeyData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbcontrol_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindManyKeyData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindManyKeyData) ProtoMessage() {}

func (x *FindManyKeyData) ProtoReflect() protoreflect.Message {
	mi := &file_dbcontrol_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindManyKeyData.ProtoReflect.Descriptor instead.
func (*FindManyKeyData) Descriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{7}
}

func (x *FindManyKeyData) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *FindManyKeyData) GetData() [][]byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type DBControllerRet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type          OptType            `protobuf:"varint,1,opt,name=type,proto3,enum=db.OptType" json:"type,omitempty"`
	Res           [][]byte           `protobuf:"bytes,2,rep,name=res,proto3" json:"res,omitempty"`
	MatchedCount  int32              `protobuf:"varint,3,opt,name=MatchedCount,proto3" json:"MatchedCount,omitempty"`
	ModifiedCount int32              `protobuf:"varint,4,opt,name=ModifiedCount,proto3" json:"ModifiedCount,omitempty"`
	UpsertedCount int32              `protobuf:"varint,5,opt,name=UpsertedCount,proto3" json:"UpsertedCount,omitempty"`
	DeletedCount  int32              `protobuf:"varint,6,opt,name=DeletedCount,proto3" json:"DeletedCount,omitempty"`
	ManyKeyData   []*FindManyKeyData `protobuf:"bytes,7,rep,name=ManyKeyData,proto3" json:"ManyKeyData,omitempty"`      //返回的许多key的数据
	DocumentCount int64              `protobuf:"varint,8,opt,name=DocumentCount,proto3" json:"DocumentCount,omitempty"` //文档总数量
}

func (x *DBControllerRet) Reset() {
	*x = DBControllerRet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbcontrol_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DBControllerRet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DBControllerRet) ProtoMessage() {}

func (x *DBControllerRet) ProtoReflect() protoreflect.Message {
	mi := &file_dbcontrol_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DBControllerRet.ProtoReflect.Descriptor instead.
func (*DBControllerRet) Descriptor() ([]byte, []int) {
	return file_dbcontrol_proto_rawDescGZIP(), []int{8}
}

func (x *DBControllerRet) GetType() OptType {
	if x != nil {
		return x.Type
	}
	return OptType_None
}

func (x *DBControllerRet) GetRes() [][]byte {
	if x != nil {
		return x.Res
	}
	return nil
}

func (x *DBControllerRet) GetMatchedCount() int32 {
	if x != nil {
		return x.MatchedCount
	}
	return 0
}

func (x *DBControllerRet) GetModifiedCount() int32 {
	if x != nil {
		return x.ModifiedCount
	}
	return 0
}

func (x *DBControllerRet) GetUpsertedCount() int32 {
	if x != nil {
		return x.UpsertedCount
	}
	return 0
}

func (x *DBControllerRet) GetDeletedCount() int32 {
	if x != nil {
		return x.DeletedCount
	}
	return 0
}

func (x *DBControllerRet) GetManyKeyData() []*FindManyKeyData {
	if x != nil {
		return x.ManyKeyData
	}
	return nil
}

func (x *DBControllerRet) GetDocumentCount() int64 {
	if x != nil {
		return x.DocumentCount
	}
	return 0
}

var File_dbcontrol_proto protoreflect.FileDescriptor

var file_dbcontrol_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x64, 0x62, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x02, 0x64, 0x62, 0x22, 0x36, 0x0a, 0x04, 0x53, 0x6f, 0x72, 0x74, 0x12, 0x1c, 0x0a,
	0x09, 0x53, 0x6f, 0x72, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x53, 0x6f, 0x72, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x41,
	0x73, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x03, 0x41, 0x73, 0x63, 0x22, 0x0d, 0x0a,
	0x0b, 0x50, 0x6c, 0x61, 0x63, 0x65, 0x68, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x22, 0xe8, 0x01, 0x0a,
	0x11, 0x46, 0x69, 0x6e, 0x64, 0x4d, 0x61, 0x6e, 0x79, 0x4b, 0x65, 0x79, 0x4f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x26, 0x0a, 0x0e, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x46,
	0x69, 0x65, 0x6c, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x43, 0x6f, 0x6e, 0x64,
	0x69, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x48, 0x0a, 0x0b,
	0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x26, 0x2e, 0x64, 0x62, 0x2e, 0x46, 0x69, 0x6e, 0x64, 0x4d, 0x61, 0x6e, 0x79, 0x4b,
	0x65, 0x79, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x46,
	0x69, 0x65, 0x6c, 0x64, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0b, 0x53, 0x65, 0x6c, 0x65, 0x63,
	0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x1a, 0x4f, 0x0a, 0x10, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74,
	0x46, 0x69, 0x65, 0x6c, 0x64, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x25, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x64, 0x62,
	0x2e, 0x50, 0x6c, 0x61, 0x63, 0x65, 0x68, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x3f, 0x0a, 0x0d, 0x52, 0x65, 0x64, 0x69, 0x73,
	0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x69, 0x64,
	0x73, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x69, 0x64,
	0x73, 0x4b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x7e, 0x0a, 0x14, 0x44, 0x42, 0x43, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x52, 0x65, 0x64, 0x69, 0x73, 0x52, 0x65, 0x71,
	0x12, 0x1f, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0b,
	0x2e, 0x64, 0x62, 0x2e, 0x4f, 0x70, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x4b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x4b, 0x65, 0x79, 0x12, 0x2d, 0x0a, 0x08, 0x6b, 0x65, 0x79,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x64, 0x62,
	0x2e, 0x52, 0x65, 0x64, 0x69, 0x73, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x08,
	0x6b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x57, 0x0a, 0x14, 0x44, 0x42, 0x43, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x52, 0x65, 0x64, 0x69, 0x73, 0x52, 0x65, 0x73,
	0x12, 0x10, 0x0a, 0x03, 0x52, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x52,
	0x65, 0x74, 0x12, 0x2d, 0x0a, 0x08, 0x6b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x64, 0x62, 0x2e, 0x52, 0x65, 0x64, 0x69, 0x73, 0x4b,
	0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x08, 0x6b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x22, 0xdf, 0x04, 0x0a, 0x0f, 0x44, 0x42, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c,
	0x65, 0x72, 0x52, 0x65, 0x71, 0x12, 0x1f, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x0b, 0x2e, 0x64, 0x62, 0x2e, 0x4f, 0x70, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x6f, 0x6c, 0x6c,
	0x65, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63,
	0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x6f,
	0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x63,
	0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x65, 0x6c, 0x65,
	0x63, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x73,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x61,
	0x78, 0x52, 0x6f, 0x77, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6d, 0x61, 0x78, 0x52,
	0x6f, 0x77, 0x12, 0x1c, 0x0a, 0x04, 0x73, 0x6f, 0x72, 0x74, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x08, 0x2e, 0x64, 0x62, 0x2e, 0x53, 0x6f, 0x72, 0x74, 0x52, 0x04, 0x73, 0x6f, 0x72, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x61, 0x77, 0x44, 0x61, 0x74, 0x61, 0x18,
	0x09, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x72, 0x61, 0x77, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1c,
	0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x4d, 0x73, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x4d, 0x73, 0x12, 0x26, 0x0a, 0x0e,
	0x61, 0x64, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x44, 0x61, 0x74, 0x61, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x61, 0x64, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c,
	0x44, 0x61, 0x74, 0x61, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x18, 0x0c,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x12, 0x18, 0x0a, 0x07,
	0x63, 0x61, 0x63, 0x68, 0x65, 0x49, 0x64, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63,
	0x61, 0x63, 0x68, 0x65, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x18, 0x0a, 0x07, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x18, 0x0f, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x07, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x53, 0x6b,
	0x69, 0x70, 0x18, 0x10, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x53, 0x6b, 0x69, 0x70, 0x12, 0x41,
	0x0a, 0x10, 0x4d, 0x61, 0x6e, 0x79, 0x4b, 0x65, 0x79, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x11, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x64, 0x62, 0x2e, 0x46, 0x69,
	0x6e, 0x64, 0x4d, 0x61, 0x6e, 0x79, 0x4b, 0x65, 0x79, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x10, 0x4d, 0x61, 0x6e, 0x79, 0x4b, 0x65, 0x79, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x2e, 0x0a, 0x12, 0x51, 0x75, 0x65, 0x72, 0x79, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65,
	0x6e, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x12, 0x20, 0x01, 0x28, 0x08, 0x52, 0x12, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x75, 0x6e,
	0x74, 0x12, 0x20, 0x0a, 0x0b, 0x4e, 0x6f, 0x74, 0x55, 0x73, 0x65, 0x43, 0x61, 0x63, 0x68, 0x65,
	0x18, 0x13, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x4e, 0x6f, 0x74, 0x55, 0x73, 0x65, 0x43, 0x61,
	0x63, 0x68, 0x65, 0x22, 0x37, 0x0a, 0x0f, 0x46, 0x69, 0x6e, 0x64, 0x4d, 0x61, 0x6e, 0x79, 0x4b,
	0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x12, 0x10, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x4b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0xb5, 0x02, 0x0a,
	0x0f, 0x44, 0x42, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x52, 0x65, 0x74,
	0x12, 0x1f, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0b,
	0x2e, 0x64, 0x62, 0x2e, 0x4f, 0x70, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x03,
	0x72, 0x65, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x64, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x4d, 0x61, 0x74, 0x63, 0x68,
	0x65, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x4d, 0x6f, 0x64, 0x69, 0x66,
	0x69, 0x65, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d,
	0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x24, 0x0a,
	0x0d, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x65, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x65, 0x64, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x35, 0x0a, 0x0b, 0x4d, 0x61, 0x6e, 0x79, 0x4b,
	0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x64,
	0x62, 0x2e, 0x46, 0x69, 0x6e, 0x64, 0x4d, 0x61, 0x6e, 0x79, 0x4b, 0x65, 0x79, 0x44, 0x61, 0x74,
	0x61, 0x52, 0x0b, 0x4d, 0x61, 0x6e, 0x79, 0x4b, 0x65, 0x79, 0x44, 0x61, 0x74, 0x61, 0x12, 0x24,
	0x0a, 0x0d, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x2a, 0xed, 0x01, 0x0a, 0x07, 0x4f, 0x70, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x08, 0x0a, 0x04, 0x4e, 0x6f, 0x6e, 0x65, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x46, 0x69,
	0x6e, 0x64, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x10, 0x02,
	0x12, 0x0a, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x10, 0x04, 0x12, 0x0a, 0x0a, 0x06,
	0x44, 0x65, 0x6c, 0x4f, 0x6e, 0x65, 0x10, 0x08, 0x12, 0x09, 0x0a, 0x05, 0x55, 0x70, 0x73, 0x65,
	0x74, 0x10, 0x10, 0x12, 0x14, 0x0a, 0x10, 0x46, 0x69, 0x6e, 0x64, 0x4f, 0x6e, 0x65, 0x41, 0x6e,
	0x64, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x10, 0x20, 0x12, 0x09, 0x0a, 0x05, 0x43, 0x6f, 0x75,
	0x6e, 0x74, 0x10, 0x40, 0x12, 0x0c, 0x0a, 0x07, 0x44, 0x65, 0x6c, 0x4d, 0x61, 0x6e, 0x79, 0x10,
	0x80, 0x02, 0x12, 0x0f, 0x0a, 0x0a, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4d, 0x61, 0x6e, 0x79,
	0x10, 0x80, 0x04, 0x12, 0x10, 0x0a, 0x0b, 0x46, 0x69, 0x6e, 0x64, 0x4d, 0x61, 0x6e, 0x79, 0x4b,
	0x65, 0x79, 0x10, 0x80, 0x08, 0x12, 0x13, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x46,
	0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x10, 0x80, 0x10, 0x12, 0x11, 0x0a, 0x0c, 0x52, 0x65,
	0x64, 0x69, 0x73, 0x5f, 0x53, 0x65, 0x74, 0x4b, 0x65, 0x79, 0x10, 0x80, 0x20, 0x12, 0x11, 0x0a,
	0x0c, 0x52, 0x65, 0x64, 0x69, 0x73, 0x5f, 0x47, 0x65, 0x74, 0x4b, 0x65, 0x79, 0x10, 0x80, 0x40,
	0x12, 0x12, 0x0a, 0x0c, 0x52, 0x65, 0x64, 0x69, 0x73, 0x5f, 0x44, 0x65, 0x6c, 0x4b, 0x65, 0x79,
	0x10, 0x80, 0x80, 0x01, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x3b, 0x64, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_dbcontrol_proto_rawDescOnce sync.Once
	file_dbcontrol_proto_rawDescData = file_dbcontrol_proto_rawDesc
)

func file_dbcontrol_proto_rawDescGZIP() []byte {
	file_dbcontrol_proto_rawDescOnce.Do(func() {
		file_dbcontrol_proto_rawDescData = protoimpl.X.CompressGZIP(file_dbcontrol_proto_rawDescData)
	})
	return file_dbcontrol_proto_rawDescData
}

var file_dbcontrol_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_dbcontrol_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_dbcontrol_proto_goTypes = []interface{}{
	(OptType)(0),                 // 0: db.OptType
	(*Sort)(nil),                 // 1: db.Sort
	(*Placeholder)(nil),          // 2: db.Placeholder
	(*FindManyKeyOption)(nil),    // 3: db.FindManyKeyOption
	(*RedisKeyValue)(nil),        // 4: db.RedisKeyValue
	(*DBControllerRedisReq)(nil), // 5: db.DBControllerRedisReq
	(*DBControllerRedisRes)(nil), // 6: db.DBControllerRedisRes
	(*DBControllerReq)(nil),      // 7: db.DBControllerReq
	(*FindManyKeyData)(nil),      // 8: db.FindManyKeyData
	(*DBControllerRet)(nil),      // 9: db.DBControllerRet
	nil,                          // 10: db.FindManyKeyOption.SelectFieldEntry
}
var file_dbcontrol_proto_depIdxs = []int32{
	10, // 0: db.FindManyKeyOption.SelectField:type_name -> db.FindManyKeyOption.SelectFieldEntry
	0,  // 1: db.DBControllerRedisReq.type:type_name -> db.OptType
	4,  // 2: db.DBControllerRedisReq.keyValue:type_name -> db.RedisKeyValue
	4,  // 3: db.DBControllerRedisRes.keyValue:type_name -> db.RedisKeyValue
	0,  // 4: db.DBControllerReq.type:type_name -> db.OptType
	1,  // 5: db.DBControllerReq.sort:type_name -> db.Sort
	3,  // 6: db.DBControllerReq.ManyKeyCondition:type_name -> db.FindManyKeyOption
	0,  // 7: db.DBControllerRet.type:type_name -> db.OptType
	8,  // 8: db.DBControllerRet.ManyKeyData:type_name -> db.FindManyKeyData
	2,  // 9: db.FindManyKeyOption.SelectFieldEntry.value:type_name -> db.Placeholder
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_dbcontrol_proto_init() }
func file_dbcontrol_proto_init() {
	if File_dbcontrol_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_dbcontrol_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Sort); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_dbcontrol_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Placeholder); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_dbcontrol_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindManyKeyOption); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_dbcontrol_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RedisKeyValue); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_dbcontrol_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DBControllerRedisReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_dbcontrol_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DBControllerRedisRes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_dbcontrol_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DBControllerReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_dbcontrol_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindManyKeyData); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_dbcontrol_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DBControllerRet); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_dbcontrol_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_dbcontrol_proto_goTypes,
		DependencyIndexes: file_dbcontrol_proto_depIdxs,
		EnumInfos:         file_dbcontrol_proto_enumTypes,
		MessageInfos:      file_dbcontrol_proto_msgTypes,
	}.Build()
	File_dbcontrol_proto = out.File
	file_dbcontrol_proto_rawDesc = nil
	file_dbcontrol_proto_goTypes = nil
	file_dbcontrol_proto_depIdxs = nil
}

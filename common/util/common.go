package util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/util/algorithms"
)

const WeekDays = int(7) //一周的天数

const GuildNameMinLen = int(2)  // 公会昵称最小长度
const GuildNameMaxLen = int(10) // 公会昵称长度限制
const UserNameMinLen = int(2)   // 用户昵称最小长度
const UserNameMaxLen = int(7)   // 用户昵称最大长度

func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func IsSameDayByMilliTimestamp(t1, t2 int64) bool {
	time1 := time.UnixMilli(t1)
	time2 := time.UnixMilli(t2)
	return IsSameDay(time1, time2)
}

// 获取指定时间的当天的指定小时的时间
func GetDayDateByHour(t int64, hour int) time.Time {
	time1 := time.UnixMilli(t)
	return time.Date(time1.Year(), time1.Month(), time1.Day(), hour, 0, 0, 0, time1.Location())
}

// 获取指定时间的下个指定小时的时间
func GetNextDayDate(t time.Time, hour int) time.Time {
	t1 := time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	if t1.After(t) {
		return t1
	}
	return t1.AddDate(0, 0, 1)
}

// 获取指定时间的下一个指定周几的指定小时的时间(中国习惯)
// 默认是Sunday(周日)开始到Saturday(星期六)算 0,1,2,3,4,5,6 （欧美日历）
// 所以只有Monday(周一)减去Sunday(周日)的时候是正数，特殊处理下就好了
func GetNextWeekDay(t time.Time, weekday time.Weekday, hour int) time.Time {
	nowWeekDay := t.Weekday()
	offset := int(weekday - nowWeekDay) //周一减去指定时间的星期几
	//特殊处理周日(周日是0)
	if nowWeekDay == time.Sunday {
		offset = -6
	}

	//获取本周的周几的指定时间
	t1 := time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location()).AddDate(0, 0, offset)
	if t1.After(t) {
		return t1
	}

	return t1.AddDate(0, 0, 7)
}

// 获取指定时间的刷新开始时间
func GetRefreshDayDate(t time.Time, hour int) time.Time {
	t1 := time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	if t.After(t1) {
		return t1
	}
	return t1.AddDate(0, 0, -1)
}

// 获取下次按时间刷新的时间
func GetRefreshNextDayDate(t time.Time, hour int) time.Time {
	t1 := time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	if t.Before(t1) {
		return t1
	}
	return t1.AddDate(0, 0, 1)
}

// 检查是否可以刷新 t1 必须比 t2小 ms级时间戳
func CheckCanRefreshByHours(t1, t2 int64, hour int) bool {
	if t1 >= t2 {
		return false
	}
	time1 := time.UnixMilli(t1)
	nextDateTime := GetNextDayDate(time1, hour)
	if t2 < nextDateTime.UnixMilli() {
		return false
	}

	return true
}

// 检查是否可以刷新周任务
func CheckCanDoWeekRefresh(t1, t2 int64, weekday time.Weekday, hour int) bool {
	if t1 >= t2 {
		return false
	}
	time1 := time.UnixMilli(t1)
	//获取time1的周一时间
	nextWeekDayTime := GetNextWeekDay(time1, weekday, hour)
	if t2 < nextWeekDayTime.UnixMilli() {
		return false
	}

	return true
}

// 获取指定时间的下个指定周几和小时的时间
func GetNextWeekDate(t time.Time, weekday time.Weekday, hour int) time.Time {
	offset := int(weekday - t.Weekday())
	//offset大于0时直接就是需要加的天数
	if offset == 0 {
		if t.Hour() >= hour {
			offset = WeekDays
		}
	} else if offset < 0 {
		offset = WeekDays - int(t.Weekday()-weekday)
	}

	return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location()).AddDate(0, 0, offset)
}

// 获得本周刷新开始时间
func GetCurrentWeekDate(t time.Time, weekday time.Weekday, hour int) time.Time {
	return GetNextWeekDate(t, weekday, hour).AddDate(0, 0, -WeekDays)
}

// 检查是否可以刷新-每周刷新 t1 必须比 t2小 ms级时间戳 weekday:星期几刷新 hour:哪个小时刷新
func CheckCanRefreshByHoursForWeek(t1, t2 int64, weekday time.Weekday, hour int) bool {
	if t1 >= t2 {
		return false
	}
	time1 := time.UnixMilli(t1)

	nextDateTime := GetNextWeekDate(time1, weekday, hour)
	if t2 < nextDateTime.UnixMilli() {
		return false
	}

	return true
}

// 获取指定时间的下个指定日期和小时的时间
func GetNextMonthDate(t time.Time, day, hour int) time.Time {
	t1 := time.Date(t.Year(), t.Month(), day, hour, 0, 0, 0, t.Location())
	if t1.After(t) {
		return t1
	}
	return t1.AddDate(0, 1, 0)
}

// 检查是否可以刷新-每周刷新 t1 必须比 t2小 ms级时间戳 day:日 hour:哪个小时刷新
func CheckCanRefreshByHoursForMonth(t1, t2 int64, day, hour int) bool {
	if t1 >= t2 {
		return false
	}
	time1 := time.UnixMilli(t1)

	nextDateTime := GetNextMonthDate(time1, day, hour)
	if t2 < nextDateTime.UnixMilli() {
		return false
	}

	return true
}

// 组合数字 n1 << 32 | n2
func CombinationNumber(n1, n2 uint64) uint64 {
	return n1<<32 | n2
}

// 解析组合数字
func AnalysisCombinationNumber(value uint64) (uint64, uint64) {
	num1 := value >> 32
	num2 := value & 0xFFFFFFFF
	return num1, num2
}

// 组合数字 n1 << 48 | n2 << 32 | n3
func CombinationThreeNumber(n1, n2, n3 uint64) uint64 {
	maxN1 := uint64(1<<(64-48) - 1)
	maxN2 := uint64(1<<(48-32) - 1)
	maxN3 := uint64(1<<32 - 1)

	if n1 > maxN1 {
		log.Stack("n1 is overflow ", maxN1)
	}
	if n2 > maxN2 {
		log.Stack("n1 is overflow ", maxN2)
	}
	if n3 > maxN3 {
		log.Stack("n1 is overflow ", maxN3)
	}

	return (n1 << 48) | (n2 << 32) | n3
}

// 解析组合数字
func AnalysisCombinationThreeNumber(value uint64) (uint64, uint64, uint64) {
	num1 := value >> 48
	num2 := (value >> 32) & 0xFFFF
	num3 := value & 0xFFFFFFFF
	return num1, num2, num3
}

// 组合日期 y<<16 | m<<8 | d<<8
func CombinationTime(n1, n2, n3 int32) int32 {
	maxN1 := int32(1<<(32-16) - 1)
	maxN2 := int32(1<<(16-8) - 1)
	maxN3 := int32(1<<8 - 1)

	if n1 > maxN1 {
		log.Stack("n1 is overflow ", maxN1)
	}
	if n2 > maxN2 {
		log.Stack("n2 is overflow ", maxN2)
	}
	if n3 > maxN3 {
		log.Stack("n3 is overflow ", maxN3)
	}

	return (n1 << 32) | (n2 << 8) | n3
}

// &^ 与非(位清空操作)
// 表达式z = x &^ y
// 如果对应y中bit位为1的话,z的bit位为0,否则对应的bit位等于x相应的bit位的值
func RemoveBitValue(value int64, index int) int64 {
	bit := 1 << index
	value = value &^ int64(bit)
	return value
}

// 检查bit位是否为1
func CheckBitValueIsTrue(value int32, index int32) bool {
	if value>>index&1 == 1 {
		return true
	}
	return false
}

// 获取timeNow到下durationDay个hour点间隔时间
func GetNowToDaySubTime(timeNow time.Time, durationDay, hour int) time.Duration {
	if timeNow.Hour() < hour {
		durationDay--
	}

	nextTime := timeNow.AddDate(0, 0, durationDay)
	newTime := time.Date(nextTime.Year(), nextTime.Month(), nextTime.Day(), hour, 0, 0, 0, nextTime.Location())

	return newTime.Sub(timeNow)
}

// 获取timeNow到下durationDay个hour点的时间戳
func GetNowToDayTimestampNano(timeNow time.Time, durationDay, hour int) int64 {
	return GetNowToDayTime(timeNow, durationDay, hour).UnixNano()
}

// 获取timeNow到下durationDay个hour点的时间
func GetNowToDayTime(timeNow time.Time, durationDay, hour int) time.Time {
	if timeNow.Hour() < hour {
		durationDay--
	}

	nextTime := timeNow.AddDate(0, 0, durationDay)
	return time.Date(nextTime.Year(), nextTime.Month(), nextTime.Day(), hour, 0, 0, 0, nextTime.Location())
}

// 获取服务器时间（单位：秒）
func GetServerFixedTimeSec(unixTime int64, hour, min, sec int) int64 {
	tm := time.Unix(unixTime, 0)
	return time.Date(tm.Year(), tm.Month(), tm.Day(), hour, min, sec, 0, tm.Location()).Unix()
}

//权重圆桌概率相关代码
/*
	1. 经典权重随机方法
		1. 必须实现IWeight接口，圆桌权重必须是叫Weight，int32，权重之和必须在int32范围之内就行
		   eg:
		   type WeightTest struct {
				util.RandomBase 或者是 util.RandomBaseMutli 前者是单随机，后者是多重随机
				OtherData int32 // 其他数据
			}

		2. 加载配置的时候需要调用，注意这里不用再存储sum
			util.PrepareWeight(WeightTestList)

		   随机的时候使用
			tempValue := util.RandomByWeight[int32, WeightTest](WeightTestList)

		注： 预处理了权重，查找用二分法查找

	2. 扩展权重随机方法
		1. 必须实现IWeight接口，同上

        2. 随机的时候使用
			outList := make([]***, 0, 2)
			outList,ok := util.RandomByWeightAdv(WeightTestList, 2, true, outList)
			参数说明：(带权重的列表，权重和，要随机的个数，是否可重复，返回的列表)

			eg：outList,ok := util.RandomByWeightAdv(WeightTestList, 2, true, outList)
				随机2个，可以重复
				outList,ok := util.RandomByWeightAdv(WeightTestList, 3, false, outList)
				随机3个，不可以重复

*/

type WeightType interface {
	~int32
}

type IWeight[ValueType WeightType] interface {
	comparable
	GetWeight() ValueType
	SetWeight(w ValueType)
	GetValue() int32
}

type RandomBase struct {
	Weight int32
}

func (ff *RandomBase) GetWeight() int32 {
	return ff.Weight
}
func (ff *RandomBase) SetWeight(w int32) {
	ff.Weight = w
}
func (ff *RandomBase) GetValue() int32 {
	return ff.Weight
}

// 单个的权重随机场景
func PrepareWeight[ValueType WeightType, T IWeight[ValueType]](weightList []T) ValueType {
	var sum ValueType
	arrLen := len(weightList)
	for i := 0; i < arrLen; i++ {
		sum += weightList[i].GetWeight()
		weightList[i].SetWeight(sum)
	}
	return sum
}

func RandomByWeight[ValueType WeightType, T IWeight[ValueType]](weightList []T) T {
	randValue := RandBetweenInt32(1, int32(weightList[len(weightList)-1].GetWeight()))
	index := algorithms.BiSearch[int32, T](weightList, randValue, 1)
	return weightList[index]
}

// 多个权重随机场景

type IWeightMutli[ValueType WeightType] interface {
	comparable
	GetWeight() ValueType
	SetWeight(w ValueType)
	Dummy() // 这里故意的，用类型系统帮助，防止PrepareWeight被误调
}

type RandomBaseMutli struct {
	Weight int32 `json:"Weight,omitempty"`
}

func (ff *RandomBaseMutli) GetWeight() int32 {
	return ff.Weight
}
func (ff *RandomBaseMutli) SetWeight(w int32) {
	ff.Weight = w
}
func (ff *RandomBaseMutli) Dummy() {
}

type weightKey[ValueType WeightType] struct {
	weightValue ValueType
	listIndex   int
}

// (带权重的列表，权重和，要随机的个数，是否可重复，返回的列表)
func RandomByWeightAdv[ValueType WeightType, T IWeightMutli[ValueType]](weightList []T, count int, canDup bool, outElement []T) ([]T, bool) {
	if !canDup {
		// 不能重复的话，这里必须进行个数判断，因为可能元素不够用
		orgLen := len(weightList)
		if orgLen < count {
			log.Stack("RandomByWeightAdv len error, in canot dup mode", log.Int("listSize:", orgLen), log.Int("willRandomCount", count))
			return nil, false
		}

		if orgLen == count {
			outElement = append(outElement, weightList...)
			return outElement, true
		}
	}

	var sum ValueType
	mapArr := make(map[weightKey[ValueType]]struct{}, 20)
	for index, ele := range weightList {
		weight := ele.GetWeight()
		key := weightKey[ValueType]{
			weightValue: weight,
			listIndex:   index,
		}

		mapArr[key] = struct{}{}
		sum += weight
	}

	for i := 0; i < count; i++ {
		randValue := ValueType(RandBetweenInt32(1, int32(sum)))

		var sumWeight ValueType = 0
		for weightInfo, _ := range mapArr {

			sumWeight += weightInfo.weightValue

			if sumWeight >= randValue {
				outElement = append(outElement, weightList[weightInfo.listIndex])
				if !canDup {
					sum -= weightInfo.weightValue
					delete(mapArr, weightInfo)
				}
				break
			}

		}
	}

	return outElement, true
}

//end

// 多个权重带调整权重随机场景

type IWeightMutliWithAdjust[ValueType WeightType] interface {
	comparable
	GetWeight() ValueType
	SetWeight(w ValueType)
	GetAdjustId() int32
	Dummy() // 这里故意的，用类型系统帮助，防止PrepareWeight被误调
}

// (带权重的列表，权重和，要随机的个数，是否可重复，返回的列表，权重调整Map)
func RandomByWeightAdvWithAdjust[ValueType WeightType, T IWeightMutliWithAdjust[ValueType]](weightList []T, count int,
	canDup bool, outElement []T, adjustMap map[int32]ValueType) ([]T, bool) {
	if !canDup {
		// 不能重复的话，这里必须进行个数判断，因为可能元素不够用
		orgLen := len(weightList)
		if orgLen < count {
			log.Stack("RandomByWeightAdv len error, in canot dup mode, listSize:", orgLen, ", willRandomCount:", count)
			return nil, false
		}

		if orgLen == count {
			outElement = append(outElement, weightList...)
			return outElement, true
		}
	}

	var sum ValueType
	mapArr := make(map[weightKey[ValueType]]struct{}, 20)
	for index, ele := range weightList {
		weight := ele.GetWeight()

		adjustValue, ok := adjustMap[ele.GetAdjustId()]
		if ok {
			weight += adjustValue
		}

		key := weightKey[ValueType]{
			weightValue: weight,
			listIndex:   index,
		}

		mapArr[key] = struct{}{}
		sum += weight
	}

	for i := 0; i < count; i++ {
		randValue := ValueType(RandBetweenInt32(1, int32(sum)))

		var sumWeight ValueType = 0
		for weightInfo, _ := range mapArr {

			sumWeight += weightInfo.weightValue

			if sumWeight >= randValue {
				outElement = append(outElement, weightList[weightInfo.listIndex])
				if !canDup {
					sum -= weightInfo.weightValue
					delete(mapArr, weightInfo)
				}
				break
			}

		}
	}

	return outElement, true
}

//end

// 配置中的Struct可以继承
type ICnfCheck interface {
	Check() error
}

// 判断一个元素是否在一个slice中
type IMyComparable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~string | ~float32 | ~float64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func IsInSlice[T IMyComparable](s []T, ele T) bool {
	for _, v := range s {
		if v == ele {
			return true
		}
	}
	return false
}

// 看两个数组是否有交集
func IsSubIntersection[T IMyComparable](sub1 []T, sub2 []T) bool {
	eleMap := make(map[T]struct{}, 30)
	for _, v := range sub1 {
		eleMap[v] = struct{}{}
	}

	for _, v := range sub2 {
		if _, ok := eleMap[v]; ok {
			return true
		}
	}

	return false
}

// sub是否是total的子集，可以排除元素，通常是0
func IsSubSlice[T IMyComparable](total []T, sub []T, exclude T) bool {
	eleMap := make(map[T]struct{}, 30)
	for _, v := range total {
		if v == exclude {
			continue
		}
		eleMap[v] = struct{}{}
	}

	for _, v := range sub {
		if v == exclude {
			continue
		}
		if _, ok := eleMap[v]; !ok {
			return false
		}
	}
	return true
}

// 判断slice是否有重复元素，可以排除一个元素，常常这个要排除的是0
func HasRepInSliceExclude[T IMyComparable](s []T, exclude T) bool {
	eleMap := make(map[T]struct{}, 30)
	effectCount := 0
	for _, v := range s {
		if v == exclude {
			continue
		}
		eleMap[v] = struct{}{}
		effectCount++
	}
	return len(eleMap) != effectCount
}

// 判断slice是否有重复元素
func HasRepInSlice[T IMyComparable](s []T) bool {
	eleMap := make(map[T]struct{}, 30)
	for _, v := range s {
		eleMap[v] = struct{}{}
	}
	return len(eleMap) != len(s)
}

// 删除指定元素
func RemoveElementInSlice[T IMyComparable](s []T, exclude T) ([]T, bool) {
	if len(s) == 0 {
		return s, false
	}

	removeIndex := -1
	for index, v := range s {
		if v == exclude {
			removeIndex = index
			break
		}
	}

	if removeIndex < 0 {
		return s, false
	}

	if len(s) == 1 {
		return s[:0], true
	} else {
		return append(s[:removeIndex], s[removeIndex+1:]...), true
	}
}

func SaveToBsonFile(bsonFileName string, data interface{}) error {
	file, err := os.OpenFile(bsonFileName, os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create bson file error, file name is %s, error: %s", bsonFileName, err.Error())
	}

	defer file.Close()

	buf, err := bson.Marshal(data)
	if err != nil {
		return fmt.Errorf("data unmarshal error, file name is %s, error: %s", bsonFileName, err.Error())
	}

	_, err = file.Write(buf)
	if err != nil {
		return fmt.Errorf("write file error, file name is %s, error: %s", bsonFileName, err.Error())
	}

	return nil
}

// 如果是service里面读取存盘失败保存的数据，isDeleteFile 一定要传true
func ReadFromBsonFile(bsonFileName string, data interface{}, isDeleteFile bool) error {
	file, err := os.OpenFile(bsonFileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open bson file error, file name is %s, error %s", bsonFileName, err.Error())
	}

	defer file.Close()

	// 读出全部bson数据
	var allBuf []byte
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("read buf fail", err)
			return fmt.Errorf("read bson data error, file name is %s, error %s", bsonFileName, err.Error())
		}
		//说明读取结束
		if n == 0 {
			break
		}
		//读取到最终的缓冲区中
		allBuf = append(allBuf, buf[:n]...)
	}

	// 反序列化
	err = bson.Unmarshal(allBuf, data)
	if err != nil {
		return fmt.Errorf("bson data unmarshal error, file name is %s, error %s", bsonFileName, err.Error())
	}

	// 删除文件
	if isDeleteFile {
		// 删除bson文件
		file.Close()
		err = os.Remove(bsonFileName)
		if err != nil {
			return fmt.Errorf("remove bson file error, file name is %s, error %s", bsonFileName, err.Error())
		}
	}

	return nil
}

type DataType interface {
	int | uint | int64 | uint64 | float32 | float64 | int32 | uint32 | int16 | uint16
}

func ConvertToNumber[DType DataType](val interface{}) (error, DType) {
	switch val.(type) {
	case int64:
		return nil, DType(val.(int64))
	case int:
		return nil, DType(val.(int))
	case uint:
		return nil, DType(val.(uint))
	case uint64:
		return nil, DType(val.(uint64))
	case float32:
		return nil, DType(val.(float32))
	case float64:
		return nil, DType(val.(float64))
	case int32:
		return nil, DType(val.(int32))
	case uint32:
		return nil, DType(val.(uint32))
	case int16:
		return nil, DType(val.(int16))
	case uint16:
		return nil, DType(val.(uint16))
	}

	return errors.New("unsupported type"), 0
}

type PickElementType interface {
}

type PickFunction func(pickElement any) bool

func PickSlice[PickElement PickElementType](pickSlick []PickElement, pickFun PickFunction) []PickElement {
	//leftPos := 0
	rightPos := len(pickSlick) - 1
	leftPos := 0
	for ; leftPos < rightPos+1; leftPos++ {
		//pickFun(pickSlick[leftPos])
		leftSelOk := pickFun(pickSlick[leftPos]) //mapSelectField[pickSlick[leftPos].Key]
		if leftSelOk == true {
			continue
		}

		for ; rightPos >= leftPos; rightPos-- {
			rightSelOk := pickFun(pickSlick[rightPos]) //mapSelectField[pickSlick[leftPos].Key]
			//找不到，右pos往走继续找
			if rightSelOk == false {
				continue
			}

			pickSlick[leftPos] = pickSlick[rightPos]
			rightPos -= 1
			break
		}
	}

	return pickSlick[:rightPos+1]
}

// 把字符串转换为数字
func StringToUint64(s string) (uint64, bool) {
	v, err := strconv.ParseUint(s, 10, 64)
	return v, err == nil
}

// 检测是否是有效的公会名称，长度和空格检测
func CheckIsValidGuildName(name string) error {
	runeName := []rune(name)
	if len(runeName) < GuildNameMinLen || len(runeName) > GuildNameMaxLen {
		return fmt.Errorf("GuildName[%s] len error, range[%d,%d] but input len: %d", name, GuildNameMinLen, GuildNameMaxLen, len(runeName))
	}
	for idx := 0; idx < len(runeName); idx++ {
		if unicode.IsSpace(runeName[idx]) {
			return fmt.Errorf("GuildName[%s] include space", name)
		}
	}
	return nil
}

// 检测是否是有效的用户名称，长度和空格检测
func CheckIsValidUserName(name string) error {
	runeName := []rune(name)
	if len(runeName) < UserNameMinLen || len(runeName) > UserNameMaxLen {
		return fmt.Errorf("UserName[%s] len error, range[%d,%d] but input len: %d", name, UserNameMinLen, UserNameMaxLen, len(runeName))
	}
	for idx := 0; idx < len(runeName); idx++ {
		if unicode.IsSpace(runeName[idx]) {
			return fmt.Errorf("UserName[%s] include space", name)
		}
	}
	return nil
}

// 把 "112,33,123" 转成对应的 slice
func ParseNumberList[T IMyComparable](value string, retList []T) []T {
	vKeyList := strings.Split(value, ",")
	if len(vKeyList) <= 0 {
		return nil
	}

	for _, val := range vKeyList {
		keyValue, errValue := strconv.Atoi(val)
		if errValue != nil {
			continue
		}

		retList = append(retList, T(keyValue))
	}

	return retList
}

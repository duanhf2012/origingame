package util

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
)

func SubAbs(a int32, b int32) int32 {
	sub := a - b
	if sub < 0 {
		return -1 * sub
	}

	return sub
}

func Abs(a int32) int32 {
	if a < 0 {
		return -1 * a
	}

	return a
}

func Round(x float64) float64 {
	if math.Abs(x) < 1e-6 {
		return 0
	}

	if x > 0 {
		return math.Floor(x + 0.5)
	}
	return math.Ceil(x - 0.5)
}

func Max(a int64, b int64) int64 {
	if a >= b {
		return a
	}
	return b
}

const RadicalTwo = float64(1.414)

func Pythagorean(a int32, b int32) int {
	aa := a * a
	bb := b * b
	return int(math.Sqrt(float64(aa + bb)))
}

// RandNum 获取一个随机整数
func RandNum(max int32) int32 {
	if max <= 0 {
		return 0
	}
	r := rand.Int31n(max)
	return r
}

// RandNumRange 获取min~max之间的随机数
func RandNumRange(min, max int64) int64 {
	if min == max {
		return min
	}
	if min > max {
		return max
	}

	r := rand.Int63n(max - min)
	return r + min
}

// RandNumRange 获取min~max之间的随机数 [min,max]
func RandBetweenNum(min, max int64) int64 {
	if min == max {
		return min
	}
	if min > max {
		return max
	}

	r := rand.Int63n(max - min + 1)
	return r + min
}

// RandNumRange 获取min~max之间的随机数 [min,max]
func RandBetweenInt32(min, max int32) int32 {
	if min == max {
		return min
	}
	if min > max {
		return max
	}

	r := rand.Int31n(max - min + 1)
	return r + min
}

func RandBetweenInt(min, max int) int {
	if min == max {
		return min
	}
	if min > max {
		return max
	}

	r := rand.Intn(max - min + 1)
	return r + min
}

// RandArrayBySumWeight 根据权重List,随机获取下标
func RandArrayBySumWeight(array []int32) int32 {
	arraylen := len(array)
	if arraylen <= 0 {
		return -1
	}
	weight := array[arraylen-1]
	r := RandNum(weight)
	for i, v := range array {
		if r < v {
			return int32(i)
		}
	}

	return -1
}

// 根据固定权重随机 格式为[[值,权重],[值,权重]]
// mapRepeat用于排重,记录已随机到的
// 返回值：随机到的值、总权重-随机到的值所占权重、错误
func RandomValueByFixedWeight(totalWeight int32, mapRepeat map[int32]struct{}, cfgSlice [][]int32) (int32, int32, error) {
	//1.参数判断
	if totalWeight <= 0 {
		return 0, totalWeight, fmt.Errorf("RandomValueByFixedWeight param totalWeight:%d must > 0", totalWeight)
	}
	lens := len(cfgSlice)
	if lens <= 0 {
		return 0, totalWeight, nil //数据为空表示不随机
	}

	//2.随机
	randValue := RandNum(totalWeight)

	//3.获取随机到的值
	tempWeight := int32(0)
	for i := 0; i < lens; i++ {
		if len(cfgSlice[i]) != 2 {
			return 0, totalWeight, fmt.Errorf("RandomValueByFixedWeight param cfgSlice index:%d len:%d must = 2", i, len(cfgSlice[i]))
		}
		//排重判断
		if _, ok := mapRepeat[cfgSlice[i][0]]; ok {
			continue
		}

		tempWeight += cfgSlice[i][1]
		if randValue < tempWeight {
			//记录排重信息
			if mapRepeat != nil {
				mapRepeat[cfgSlice[i][0]] = struct{}{}
			}

			return cfgSlice[i][0], totalWeight - cfgSlice[i][1], nil
		}
	}

	//4.如果没有随机到,则表示传入的总权重totalWeight比参数cfgSlice里面的总权重大,因此随机不到 有这种情况,属于正常的
	return 0, totalWeight, nil
}

// 随机多个值
// pCount:随机个数
// totalWeight:总权重
// cfgSlice:随机池 格式必须是[[值,权重],[值,权重]]
// bDuplicate:随机是否可重复 true可以重复 false不可重复
// pSlice:记录随机到的值
func RandomMultiValueByFixedWeight(pCount int32, totalWeight int32, cfgSlice [][]int32, bDuplicate bool, pSlice []int32) []int32 {
	//1.参数判断
	if totalWeight <= 0 || pCount <= 0 || len(cfgSlice) <= 0 {
		return pSlice
	}

	//2.随机
	//用于记录已随机到的值,用于排重
	mapValue := make(map[int32]struct{})
	//记录总权重
	tWeight := totalWeight
	//循环随机
	randId := int32(0)
	var err error
	for i := int32(0); i < pCount; i++ {
		//去随机
		if bDuplicate == true {
			randId, tWeight, err = RandomValueByFixedWeight(tWeight, nil, cfgSlice)
		} else {
			randId, tWeight, err = RandomValueByFixedWeight(tWeight, mapValue, cfgSlice)
		}

		if err != nil {
			continue
		}
		if randId > 0 {
			pSlice = append(pSlice, randId)
		}
	}

	return pSlice
}

// HashString2Number 计算字符串的hash值
func HashString2Number(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// 0到90度
const BasePosEnlarge = 1000
const SinAndCosEnlarge = 1000
const PosEnlarge = BasePosEnlarge * BasePosEnlarge
const Base = 180 / math.Pi
const MaxAngle = 90
const MaxGridWith = 200

var SinValue [MaxAngle + 1]int32
var CosValue [MaxAngle + 1]int32
var ATan2Value [MaxGridWith*2 + 2][MaxGridWith*2 + 2]int32
var SqrtValue [MaxGridWith + 1][MaxGridWith + 1]int32

func init() {
	for i := 0; i <= 90; i++ {
		SinValue[i] = int32(math.Sin(float64(i)*math.Pi/180) * SinAndCosEnlarge)
		CosValue[i] = int32(math.Cos(float64(i)*math.Pi/180) * SinAndCosEnlarge)
	}

	for i := -MaxGridWith; i <= MaxGridWith; i++ {
		for j := -MaxGridWith; j <= MaxGridWith; j++ {
			ATan2Value[i+MaxGridWith][j+MaxGridWith] = int32(Base * math.Atan2(float64(i*BasePosEnlarge), float64(j*BasePosEnlarge)))
		}
	}

	for i := 0; i <= MaxGridWith; i++ {
		for j := 0; j <= MaxGridWith; j++ {
			SqrtValue[i][j] = int32(math.Sqrt(float64(i*BasePosEnlarge*i*BasePosEnlarge) + float64(j*BasePosEnlarge*j*BasePosEnlarge)))
		}
	}
}

// GetTwoPointDistance 返回距离 <0表示俩点相差>100
func GetTwoPointDistance(startX, startY, endX, endY int32) int32 {
	deltaX := SubAbs(startX, endX)
	deltaY := SubAbs(startY, endY)

	if deltaX > MaxGridWith || deltaY > MaxGridWith {
		return 0xfffffff
	}

	return SqrtValue[deltaX][deltaY]
}

func GetSinValue(angle int32) int32 {
	if angle > MaxAngle {
		return 0
	}

	return SinValue[angle]
}

func GetCosValue(angle int32) int32 {
	if angle > MaxAngle {
		return 0
	}

	return CosValue[angle]
}

func GetSinCosRange(angle int32) (int32, int32) {
	angle = int32(angle) % 360
	if angle < 0 {
		angle += 360
	}

	var sin, cos int32
	if angle <= 90 {
		sin = GetSinValue(angle)
		cos = GetCosValue(angle)
	} else if angle <= 180 {
		ag := 180 - angle
		sin = GetSinValue(ag)
		cos = -1 * GetCosValue(ag)
	} else if angle <= 270 {
		//使用计算器验证了下 sin(260)=-1*cos(10)
		ag := 270 - angle
		sin = -1 * GetCosValue(ag)
		cos = -1 * GetSinValue(ag)
	} else {
		ag := 360 - angle
		sin = -1 * GetSinValue(ag)
		cos = 1 * GetCosValue(ag)
	}

	return sin, cos
}

func GetATan2Value(startX, startY, endX, endY int32) int32 {
	subX := endX - startX
	subY := endY - startY
	indexX := subX + MaxGridWith
	indexY := subY + MaxGridWith

	if indexX < 0 || indexX >= MaxGridWith*2+2 || indexY < 0 || indexY >= MaxGridWith*2+2 {
		return 0
	}

	return ATan2Value[indexX][indexY]
}

var longLetters = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ=_")

func RandString(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	arc := uint8(0)
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	for i, x := range b {
		arc = x & 63
		b[i] = longLetters[arc]
	}
	return string(b)
}

type SortData struct {
	Id        uint64
	SortValue int32
}

// 升序
type SortAsc []SortData

func (a SortAsc) Len() int           { return len(a) }
func (a SortAsc) Less(i, j int) bool { return a[i].SortValue < a[j].SortValue }
func (a SortAsc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// 降序
type SortDesc []SortData

func (a SortDesc) Len() int           { return len(a) }
func (a SortDesc) Less(i, j int) bool { return a[i].SortValue > a[j].SortValue }
func (a SortDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// 获取伤害系数
func GetDamageCoefficients(casterAttack, targetDefense int64) int64 {
	ret := casterAttack - targetDefense
	minRet := casterAttack * 500 / 10000
	if ret < minRet {
		ret = minRet
	}

	return ret
}

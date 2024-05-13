package util

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
)

// 签名类型
type SignType = int8

const (
	SignLogin    SignType = 1 //1登录
	SignPurchase SignType = 2 //2充值
	SignGM       SignType = 3 //3GM
)

type UrlParamData struct {
	Key   string
	Value string
}

type ParamSortAsc []UrlParamData

func (p ParamSortAsc) Len() int           { return len(p) }
func (p ParamSortAsc) Less(i, j int) bool { return p[i].Key < p[j].Key }
func (p ParamSortAsc) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

/*
//聚合平台签名
1. 排序：将【所有请求数据（不包括 sign 本身）】按key升序排列
2. 拼接：将第1步骤的结果组合成value1+value2+value3+...格式
3. 追加：【将第2步骤的结果md5后-*实例代码没有此操作*-】追加"KEY（KEY由平台分配）
4. 加密：将第3步骤的结果进行md5加密(32位16进制)，并将结果转换为小写字母，便得到sign参数
*/
func GetFusionAppSign(params *[]UrlParamData, appKey string, signType SignType) string {
	sort.Sort(ParamSortAsc(*params))
	rs := ""
	switch signType {
	case SignLogin:
		for idx := 0; idx < len(*params); idx++ {
			rs += (*params)[idx].Value
		}
	case SignPurchase:
		for idx := 0; idx < len(*params); idx++ {
			rs += (*params)[idx].Key + "=" + (*params)[idx].Value
		}
	case SignGM:
		lenParam := len(*params)
		for idx := 0; idx < lenParam; idx++ {
			rs += (*params)[idx].Key + "=" + (*params)[idx].Value
			if idx == lenParam-1 {
				continue
			}
			rs += "&"
		}
	}

	rs += appKey
	return strings.ToLower(fmt.Sprintf("%x", md5.Sum([]byte(rs))))
}

package util

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const RemarksRowsCount = 5
const ServerSkipCount = 3

var Comma = ','
var Comment = '#'

type FileItem struct {
	KindStr   string
	IndexOf   int
	IndexList []int
}

type loadRowsFinishCallBack func(rowsData interface{}) error

func generateFileItemMap(nameLine, typeLine []string, mapFileItem map[string]*FileItem) {
	if len(nameLine) != len(typeLine) {
		return
	}

	for i := 0; i < len(nameLine); i++ {
		if nameLine[i] == "" || typeLine[i] == "" {
			continue
		}

		name := nameLine[i]
		typeName := typeLine[i]
		if typeName != "slice" {
			mapFileItem[name] = &FileItem{
				KindStr: typeName,
				IndexOf: i,
			}
		} else {
			if _, ok := mapFileItem[name]; ok {
				mapFileItem[name].IndexList = append(mapFileItem[name].IndexList, i)
			} else {
				mapFileItem[name] = &FileItem{
					KindStr:   typeName,
					IndexOf:   0,
					IndexList: []int{i},
				}
			}
		}
	}

	return
}

func analysisLine(typeLine, dataLine []string) error {
	if len(dataLine) != len(typeLine) {
		return fmt.Errorf("len err")
	}

	for i := 0; i < len(dataLine); i++ {
		if typeLine[i] == "string" {
			continue
		}

		dataTemp := dataLine[i]
		dataTemp = strings.Replace(dataTemp, " ", "", -1)
		dataTemp = strings.Replace(dataTemp, "\n", "", -1)
		dataTemp = strings.Replace(dataTemp, "\r", "", -1)
		if !IsUtf8([]byte(dataTemp)) {
			dataTemp = ConvertToUTF8(dataTemp)
		}

		dataLine[i] = AnalysisString(dataTemp)
	}

	return nil
}

func AnalysisString(strData string) string {
	retStr := strData
	byteData := []byte(strData)
	for i := 0; i < len(byteData); i++ {
		if byteData[i] != '_' {
			continue
		}

		needAnalysis := true
		startIndex := i - 1
		endIndex := i + 1
		for ; startIndex >= 0; startIndex-- {
			if byteData[startIndex] == '"' {
				needAnalysis = false
				break
			}
			if byteData[startIndex] == ':' || byteData[startIndex] == '{' || byteData[startIndex] == '[' || byteData[startIndex] == ',' {
				startIndex++
				break
			}
		}
		if !needAnalysis {
			continue
		}
		if startIndex < 0 {
			startIndex = 0
		}

		for ; endIndex < len(byteData); endIndex++ {
			if byteData[endIndex] == '"' {
				needAnalysis = false
				break
			}
			if byteData[endIndex] == ',' || byteData[endIndex] == '}' || byteData[endIndex] == ']' {
				break
			}
		}
		if !needAnalysis {
			continue
		}
		if endIndex > len(byteData) {
			endIndex = len(byteData)
		}

		replaceStr := strData[startIndex:endIndex]
		v := strings.Split(replaceStr, "_")
		if len(v) > 0 {
			retStr = strData[:startIndex] + v[0] + strData[endIndex:]
			return AnalysisString(retStr)
		}
	}

	return retStr
}

type Index map[interface{}]interface{}
type RecordFile struct {
	Comma      rune
	Comment    rune
	typeRecord reflect.Type
	records    []interface{}
	indexes    []Index
}

// NewRecordFile 解析配置结构
func NewRecordFile(st interface{}) (*RecordFile, error) {
	typeRecord := reflect.TypeOf(st)
	if typeRecord == nil || typeRecord.Kind() != reflect.Struct {
		return nil, errors.New("st must be a struct")
	}

	for i := 0; i < typeRecord.NumField(); i++ {
		f := typeRecord.Field(i)
		tag := f.Tag
		if tag == "omitempty" {
			continue
		}

		kind := f.Type.Kind()
		switch kind {
		case reflect.Bool:
		case reflect.Int:
		case reflect.Int8:
		case reflect.Int16:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Uint:
		case reflect.Uint8:
		case reflect.Uint16:
		case reflect.Uint32:
		case reflect.Uint64:
		case reflect.Float32:
		case reflect.Float64:
		case reflect.String:
		case reflect.Struct:
		case reflect.Array:
		case reflect.Slice:
		case reflect.Map:
		default:
			return nil, fmt.Errorf("invalid type: %v %s", f.Name, kind)
		}

		if tag == "index" {
			switch kind {
			case reflect.Slice, reflect.Map, reflect.Float32, reflect.Float64:
				return nil, fmt.Errorf("could not index %s field %v %v", kind, i, f.Name)
			}
		}
	}

	rf := new(RecordFile)
	rf.typeRecord = typeRecord

	return rf, nil
}

// 读取csv文件
func (rf *RecordFile) Read(name string, callBack loadRowsFinishCallBack) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	//读取文件所有行信息
	if rf.Comma == 0 {
		rf.Comma = Comma
	}
	if rf.Comment == 0 {
		rf.Comment = Comment
	}
	reader := csv.NewReader(file)
	reader.Comma = rf.Comma
	reader.Comment = rf.Comment
	lines, err := reader.ReadAll()
	if err != nil {
		return err
	}

	typeRecord := rf.typeRecord
	// make records
	records := make([]interface{}, len(lines)-RemarksRowsCount)

	// make indexes
	indexes := []Index{}
	for i := 0; i < typeRecord.NumField(); i++ {
		tag := typeRecord.Field(i).Tag
		if tag == "index" || tag == "indexCombination" {
			indexes = append(indexes, make(Index))
		}
	}

	n := ServerSkipCount
	//第四行开始为服务器标记位
	//第四行为数据名、第五行为数据类型
	nameLine := lines[n]
	typeLine := lines[n+1]
	mapFileItem := make(map[string]*FileItem, typeRecord.NumField())
	generateFileItemMap(nameLine, typeLine, mapFileItem)

	n += 2
	recordsIndex := 0
	//遍历每行数据
	for ; n < len(lines); n++ {
		value := reflect.New(typeRecord)
		records[recordsIndex] = value.Interface()
		record := value.Elem()

		line := lines[n]
		if analysisLine(typeLine, line) != nil {
			return fmt.Errorf("line %v, field count mismatch: %v (file) %v (st)", n, len(line), typeRecord.NumField())
		}

		iIndex := 0
		//遍历配置结构每个参数，整合
		for i := 0; i < typeRecord.NumField(); i++ {
			f := typeRecord.Field(i)
			if f.Tag == "omitempty" {
				continue
			}

			if f.Tag == "indexCombination" {
				field := record.Field(i)
				for n := 0; n < f.Type.NumField(); n++ {
					fileIndexCombinationItem, ok := mapFileItem[f.Type.Field(n).Name]
					if !ok {
						return fmt.Errorf("parse field (row=%v, col=%v) error: %s has not in config file", n+1, i+1, f.Name)
					}
					fIndexCombination := field.Field(n)

					strField := line[fileIndexCombinationItem.IndexOf]
					switch f.Type.Field(n).Type.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						strField = rf.checkKeyWords(strField)
						var v int64
						v, err = strconv.ParseInt(strField, 0, fIndexCombination.Type().Bits())
						if err == nil {
							fIndexCombination.SetInt(v)
						}
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						strField = rf.checkKeyWords(strField)
						var v uint64
						v, err = strconv.ParseUint(strField, 0, fIndexCombination.Type().Bits())
						if err == nil {
							fIndexCombination.SetUint(v)
						}
					case reflect.String:
						fIndexCombination.SetString(strField)
					default:
						err = fmt.Errorf("in IndexCombination itemName=%s kind err: %+v", f.Name, f.Type.Field(n).Type.Kind())
					}
				}

				index := indexes[iIndex]
				iIndex++
				if _, ok := index[field.Interface()]; ok {
					return fmt.Errorf("indexCombination error: duplicate at (itemName=%s row=%v, col=%v)", f.Name, n, i)
				}
				index[field.Interface()] = records[recordsIndex]
				continue
			}

			itemKind := f.Type.Kind()
			itemTypeStr := f.Type.Kind().String()
			fileItem, ok := mapFileItem[f.Name]
			if !ok {
				if f.Tag == "index" {
					return fmt.Errorf("parse field (row=%v, col=%v) error: %s has not in config file", n+1, i+1, f.Name)
				}
				continue
			}
			//特殊类型解析
			if fileItem.KindStr == "duration" {
				field := record.Field(i)
				strField := line[fileItem.IndexOf]
				strField = rf.checkKeyWords(strField)
				if strField == "" {
					field.SetInt(0)
					continue
				}
				var v int64
				v, err = strconv.ParseInt(strField, 0, f.Type.Bits())
				if err != nil {
					return fmt.Errorf("parse field (itemName=%s row=%v, col=%v) error: %+v", f.Name, n, i, err)
				}
				v = v * int64(time.Millisecond)
				field.SetInt(v)
				continue
			}

			if itemTypeStr != fileItem.KindStr || fileItem.IndexOf >= len(line) {
				return fmt.Errorf("parse field (itemName=%s row=%v, col=%v) error: %s type is not equal", f.Name, n, i, fileItem.KindStr)
			}
			// records
			if itemKind != reflect.Slice {
				strField := line[fileItem.IndexOf]
				field := record.Field(i)
				if !field.CanSet() || strField == "" {
					continue
				}

				switch itemKind {
				case reflect.Bool:
					var v bool
					v, err = strconv.ParseBool(strField)
					if err == nil {
						field.SetBool(v)
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					strField = rf.checkKeyWords(strField)
					var v int64
					v, err = strconv.ParseInt(strField, 0, f.Type.Bits())
					if err == nil {
						field.SetInt(v)
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					strField = rf.checkKeyWords(strField)
					var v uint64
					v, err = strconv.ParseUint(strField, 0, f.Type.Bits())
					if err == nil {
						field.SetUint(v)
					}
				case reflect.Float32, reflect.Float64:
					var v float64
					v, err = strconv.ParseFloat(strField, f.Type.Bits())
					if err == nil {
						field.SetFloat(v)
					}
				case reflect.String:
					field.SetString(strField)
				case reflect.Struct, reflect.Array, reflect.Map:
					err = json.Unmarshal([]byte(strField), field.Addr().Interface())
				default:
					err = fmt.Errorf("itemName=%s kind err: %+v", f.Name, itemKind)
				}

				if err != nil {
					return fmt.Errorf("data[%s] parse field (itemName=%s row=%v, col=%v) error: %+v", f.Name, strField, n, i, err)
				}

				if f.Tag == "index" {
					index := indexes[iIndex]
					iIndex++
					if _, ok := index[field.Interface()]; ok {
						return fmt.Errorf("index error: duplicate at (itemName=%s row=%v, col=%v, value:%+v)", f.Name, n, i, field.Interface())
					}
					index[field.Interface()] = records[recordsIndex]
				}
			} else {
				if f.Tag == "index" {
					return fmt.Errorf("slice can not be index itemName=%s", f.Name)
				}

				for _, indexItem := range fileItem.IndexList {
					strField := line[indexItem]
					field := record.Field(i)
					if !field.CanSet() || strField == "" {
						continue
					}

					tempSlice := reflect.New(field.Type())
					err = json.Unmarshal([]byte(strField), tempSlice.Interface())
					if err != nil {
						return fmt.Errorf("data[%s] parse field (itemName=%s, row=%v, col=%v) error: %+v", f.Name, strField, n, i, err)
					}
					for n := 0; n < tempSlice.Elem().Len(); n++ {
						e := tempSlice.Elem().Index(n)
						field.Set(reflect.Append(field, e))
					}
				}
			}
		}

		checkRowsErr := callBack(value.Interface())
		if checkRowsErr != nil {
			return checkRowsErr
		}

		recordsIndex++
	}

	rf.records = records
	rf.indexes = indexes

	return nil
}

// 获取获取第i个对象指针
func (rf *RecordFile) Record(i int) interface{} {
	return rf.records[i]
}

// 获取个数
func (rf *RecordFile) NumRecord() int {
	return len(rf.records)
}

// 获取第i个key组成的map数据
func (rf *RecordFile) Indexes(i int) Index {
	if i >= len(rf.indexes) {
		return nil
	}
	return rf.indexes[i]
}

// 获取第0个key,值为i对应的数据指针
func (rf *RecordFile) Index(i interface{}) interface{} {
	index := rf.Indexes(0)
	if index == nil {
		return nil
	}
	return index[i]
}

func (rf *RecordFile) checkKeyWords(value string) string {
	vStr := strings.Split(value, "_")
	if len(vStr) >= 2 {
		return vStr[0]
	}

	return value
}

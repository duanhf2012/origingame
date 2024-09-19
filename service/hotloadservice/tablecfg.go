package hotloadservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/service"
	jsoniter "github.com/json-iterator/go"
	"origingame/common/module/tablecfg"
	tabledef "origingame/table/TableDef"
	"os"
	"path/filepath"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type TableProcessor func(tables *tablecfg.Tables) error
type TableCfgModule struct {
	service.Module
	jsonPath string

	mapIgnoreFileName map[string]struct{}
	processorFunList  []TableProcessor
}

// SetJsonPath 设置Json路径
func (tc *TableCfgModule) SetJsonPath(path string) error {
	if !pathExists(path) {
		return fmt.Errorf("path %s does not exist", path)
	}

	tc.jsonPath = path
	return nil
}

func (tc *TableCfgModule) OnInit() error {
	if tc.jsonPath == "" {
		return fmt.Errorf("json path is empty,please call SetJsonPath first")
	}

	return tc.LoadCfg()
}

func (tc *TableCfgModule) LoadCfg() error {
	tables, err := tabledef.NewTables(tc.loaderJson)
	if err != nil {
		return err
	}
	var t tablecfg.Tables
	t.Tables = tables
	t.ProcessedTables = &tablecfg.ProcessedTables{}

	for _, processor := range tc.processorFunList {
		err = processor(&t)
		if err != nil {
			return err
		}
	}

	tablecfg.SetTables(t)

	return nil
}

// AppendTableProcessor 追加表处理器，当表加载完成时，可以追加函数对数据进行预处理。配置加载完成后将按该顺序执行
func (tc *TableCfgModule) AppendTableProcessor(processorFun TableProcessor) {
	tc.processorFunList = append(tc.processorFunList, processorFun)
}

// SetCustomConfigFile 设置自定义加载的配置文件，不设置加载全部
func (tc *TableCfgModule) SetCustomConfigFile(mapIgnoreFileName map[string]struct{}) {
	tc.mapIgnoreFileName = mapIgnoreFileName
}

func (tc *TableCfgModule) loaderJson(file string) ([]map[string]interface{}, error) {
	if _, ok := tc.mapIgnoreFileName[file]; ok {
		return []map[string]interface{}{}, nil
	}

	if bytes, err := os.ReadFile(filepath.Join(tc.jsonPath, file+".json")); err != nil {
		return nil, err
	} else {
		jsonData := make([]map[string]interface{}, 0)
		if err = json.Unmarshal(bytes, &jsonData); err != nil {
			return nil, err
		}
		return jsonData, nil
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	// 其他错误，如权限问题，可能会阻止我们访问文件
	return false
}

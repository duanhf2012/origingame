package area

import (
	"errors"
	"sync/atomic"
)

// catalogSnapshot 使用原子指针发布不可变区服列表，HTTP 热路径无需锁。
type catalogSnapshot struct {
	areas []Info // 完整且不可变的显示区服列表。
}

// Catalog 保存最后一份完整有效的区服快照。
type Catalog struct {
	current atomic.Pointer[catalogSnapshot] // 原子发布的当前区服快照。
}

// Replace 深复制并原子发布完整快照；空快照不能覆盖线上数据。
func (catalog *Catalog) Replace(areas []Info) error {
	if len(areas) == 0 {
		return errors.New("区服快照不能为空")
	}
	catalog.current.Store(&catalogSnapshot{areas: clone(areas)})
	return nil
}

// Snapshot 返回调用方独占副本，JSON 编码和测试不能修改在线快照。
func (catalog *Catalog) Snapshot() []Info {
	if catalog == nil {
		return nil
	}
	current := catalog.current.Load()
	if current == nil {
		return nil
	}
	return clone(current.areas)
}

func clone(source []Info) []Info {
	result := make([]Info, len(source))
	for index := range source {
		result[index] = source[index]
		result[index].GateList = append([]GateInfo(nil), source[index].GateList...)
	}
	return result
}

// Package area 负责 Gateway 的显示区服到真实区服映射快照。
package area

import (
	"errors"
	"sync/atomic"
)

type snapshot struct{ realByShow map[int64]int64 }

// Catalog 原子发布完整且不可变的区服映射。
type Catalog struct{ current atomic.Pointer[snapshot] }

// Replace 校验并替换整份映射；失败不会覆盖上一份有效数据。
func (catalog *Catalog) Replace(mapping map[int64]int64) error {
	// 空映射不能覆盖当前可用区服快照。
	if len(mapping) == 0 {
		return errors.New("Gateway 区服映射不能为空")
	}
	// 校验所有标识并复制输入，防止调用方后续修改快照。
	copyMap := make(map[int64]int64, len(mapping))
	for showAreaID, realAreaID := range mapping {
		if showAreaID <= 0 || realAreaID <= 0 {
			return errors.New("Gateway 区服映射包含无效 ID")
		}
		copyMap[showAreaID] = realAreaID
	}
	// 一次性发布完整不可变映射。
	catalog.current.Store(&snapshot{realByShow: copyMap})
	return nil
}

// Resolve 返回显示区服当前对应的真实区服。
func (catalog *Catalog) Resolve(showAreaID int64) (int64, bool) {
	// 未初始化目录视为没有可路由区服。
	if catalog == nil {
		return 0, false
	}
	current := catalog.current.Load()
	if current == nil {
		return 0, false
	}
	// 从当前快照读取显示区服的真实归属。
	realAreaID, ok := current.realByShow[showAreaID]
	return realAreaID, ok
}

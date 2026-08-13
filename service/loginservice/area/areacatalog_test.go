package area

import "testing"

// TestCatalogCopiesSnapshots 验证调用方无法修改已经发布的区服与 Gateway 列表。
func TestCatalogCopiesSnapshots(t *testing.T) {
	catalog := &Catalog{}
	input := []Info{{ShowAreaID: 1, AreaName: "体验服", GateList: []GateInfo{{Protocol: "tcp", Address: "a"}}}}
	if err := catalog.Replace(input); err != nil {
		t.Fatalf("Replace() error = %v", err)
	}
	input[0].AreaName = "changed"
	input[0].GateList[0].Address = "changed"
	first := catalog.Snapshot()
	if first[0].AreaName != "体验服" || first[0].GateList[0].Address != "a" {
		t.Fatalf("published snapshot was mutated: %+v", first)
	}
	first[0].GateList[0].Address = "again"
	second := catalog.Snapshot()
	if second[0].GateList[0].Address != "a" {
		t.Fatalf("returned snapshot aliases catalog: %+v", second)
	}
}

// TestCatalogRejectsEmpty 保证运行期空查询不能清空最后有效区服数据。
func TestCatalogRejectsEmpty(t *testing.T) {
	catalog := &Catalog{}
	if err := catalog.Replace(nil); err == nil {
		t.Fatal("Replace(nil) succeeded")
	}
}

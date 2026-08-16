package area

import "testing"

func TestCatalogRejectsInvalidReplacementWithoutLosingSnapshot(t *testing.T) {
	var catalog Catalog
	if err := catalog.Replace(map[int64]int64{10: 1}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Replace(map[int64]int64{11: 0}); err == nil {
		t.Fatal("无效映射不应发布")
	}
	if got, ok := catalog.Resolve(10); !ok || got != 1 {
		t.Fatalf("Resolve(10) = %d, %v", got, ok)
	}
}


//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package tabledef;


import "errors"

type ItemData struct {
    ItemID int32
    Num int64
}

const TypeId_ItemData = 1241678205

func (*ItemData) GetTypeId() int32 {
    return 1241678205
}

func NewItemData(_buf map[string]interface{}) (_v *ItemData, err error) {
    _v = &ItemData{}
    { var _ok_ bool; var __json_ItemID__ interface{}; if __json_ItemID__, _ok_ = _buf["ItemID"]; !_ok_ || __json_ItemID__ == nil { err = errors.New("ItemID error"); return } else { var __x__ int32;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_ItemID__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = int32(_x_) }; _v.ItemID = __x__ }}
    { var _ok_ bool; var __json_Num__ interface{}; if __json_Num__, _ok_ = _buf["Num"]; !_ok_ || __json_Num__ == nil { err = errors.New("Num error"); return } else { var __x__ int64;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_Num__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = int64(_x_) }; _v.Num = __x__ }}
    return
}



//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package tabledef;


import "errors"

type AttrData struct {
    AttrType int32
    AttrValue int64
}

const TypeId_AttrData = 618152283

func (*AttrData) GetTypeId() int32 {
    return 618152283
}

func NewAttrData(_buf map[string]interface{}) (_v *AttrData, err error) {
    _v = &AttrData{}
    { var _ok_ bool; var __json_AttrType__ interface{}; if __json_AttrType__, _ok_ = _buf["AttrType"]; !_ok_ || __json_AttrType__ == nil { err = errors.New("AttrType error"); return } else { var __x__ int32;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_AttrType__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = int32(_x_) }; _v.AttrType = __x__ }}
    { var _ok_ bool; var __json_AttrValue__ interface{}; if __json_AttrValue__, _ok_ = _buf["AttrValue"]; !_ok_ || __json_AttrValue__ == nil { err = errors.New("AttrValue error"); return } else { var __x__ int64;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_AttrValue__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = int64(_x_) }; _v.AttrValue = __x__ }}
    return
}


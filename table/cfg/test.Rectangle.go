
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type TestRectangle struct {
    Width float32
    Height float32
}

const TypeId_TestRectangle = -31893773

func (*TestRectangle) GetTypeId() int32 {
    return -31893773
}

func NewTestRectangle(_buf map[string]interface{}) (_v *TestRectangle, err error) {
    _v = &TestRectangle{}
    { var _ok_ bool; var __json_width__ interface{}; if __json_width__, _ok_ = _buf["width"]; !_ok_ || __json_width__ == nil { err = errors.New("width error"); return } else { var __x__ float32;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_width__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = float32(_x_) }; _v.Width = __x__ }}
    { var _ok_ bool; var __json_height__ interface{}; if __json_height__, _ok_ = _buf["height"]; !_ok_ || __json_height__ == nil { err = errors.New("height error"); return } else { var __x__ float32;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_height__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = float32(_x_) }; _v.Height = __x__ }}
    return
}


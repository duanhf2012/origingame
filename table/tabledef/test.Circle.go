
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package tabledef;


import "errors"

type TestCircle struct {
    Radius float32
}

const TypeId_TestCircle = 2131829196

func (*TestCircle) GetTypeId() int32 {
    return 2131829196
}

func NewTestCircle(_buf map[string]interface{}) (_v *TestCircle, err error) {
    _v = &TestCircle{}
    { var _ok_ bool; var __json_radius__ interface{}; if __json_radius__, _ok_ = _buf["radius"]; !_ok_ || __json_radius__ == nil { err = errors.New("radius error"); return } else { var __x__ float32;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_radius__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = float32(_x_) }; _v.Radius = __x__ }}
    return
}


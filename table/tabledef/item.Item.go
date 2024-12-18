
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package tabledef;


import "errors"

type ItemItem struct {
    Id int32
    Name string
    Price int32
    UpgradeToItemId int32
    ExpireTime *int64
    Quality int32
    ExchangeList []*ItemItemExchange
    ExchangeColumn *ItemItemExchange
    AddAttr map[int32]int64
    AddItem []*ItemData
}

const TypeId_ItemItem = 2107285806

func (*ItemItem) GetTypeId() int32 {
    return 2107285806
}

func NewItemItem(_buf map[string]interface{}) (_v *ItemItem, err error) {
    _v = &ItemItem{}
    { var _ok_ bool; var __json_id__ interface{}; if __json_id__, _ok_ = _buf["id"]; !_ok_ || __json_id__ == nil { err = errors.New("id error"); return } else { var __x__ int32;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_id__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = int32(_x_) }; _v.Id = __x__ }}
    { var _ok_ bool; var __json_name__ interface{}; if __json_name__, _ok_ = _buf["name"]; !_ok_ || __json_name__ == nil { err = errors.New("name error"); return } else { var __x__ string;  {  if __x__, _ok_ = __json_name__.(string); !_ok_ { err = errors.New("__x__ error"); return } }; _v.Name = __x__ }}
    { var _ok_ bool; var __json_price__ interface{}; if __json_price__, _ok_ = _buf["price"]; !_ok_ || __json_price__ == nil { err = errors.New("price error"); return } else { var __x__ int32;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_price__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = int32(_x_) }; _v.Price = __x__ }}
    { var _ok_ bool; var __json_upgrade_to_item_id__ interface{}; if __json_upgrade_to_item_id__, _ok_ = _buf["upgrade_to_item_id"]; !_ok_ || __json_upgrade_to_item_id__ == nil { err = errors.New("upgrade_to_item_id error"); return } else { var __x__ int32;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_upgrade_to_item_id__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = int32(_x_) }; _v.UpgradeToItemId = __x__ }}
    { var _ok_ bool; var __json_expire_time__ interface{}; if __json_expire_time__, _ok_ = _buf["expire_time"]; !_ok_ || __json_expire_time__ == nil { _v.ExpireTime = nil } else { var __x__ int64;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_expire_time__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = int64(_x_) }; _v.ExpireTime = &__x__ }}
    { var _ok_ bool; var __json_quality__ interface{}; if __json_quality__, _ok_ = _buf["quality"]; !_ok_ || __json_quality__ == nil { err = errors.New("quality error"); return } else { var __x__ int32;  { var _ok_ bool; var _x_ float64; if _x_, _ok_ = __json_quality__.(float64); !_ok_ { err = errors.New("__x__ error"); return }; __x__ = int32(_x_) }; _v.Quality = __x__ }}
    { var _ok_ bool; var __json_exchange_list__ interface{}; if __json_exchange_list__, _ok_ = _buf["exchange_list"]; !_ok_ || __json_exchange_list__ == nil { err = errors.New("exchange_list error"); return } else { var __x__ []*ItemItemExchange;  {
                    var _arr0_ []interface{}
                    var _ok0_ bool
                    if _arr0_, _ok0_ = (__json_exchange_list__).([]interface{}); !_ok0_ { err = errors.New("__x__ error"); return }
    
                    __x__ = make([]*ItemItemExchange, 0, len(_arr0_))
                    
                    for _, _e0_ := range _arr0_ {
                        var _list_v0_ *ItemItemExchange
                        { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _e0_.(map[string]interface{}); !_ok_ { err = errors.New("_list_v0_ error"); return }; if _list_v0_, err = NewItemItemExchange(_x_); err != nil { return } }
                        __x__ = append(__x__, _list_v0_)
                    }
                }
    ; _v.ExchangeList = __x__ }}
    { var _ok_ bool; var __json_exchange_column__ interface{}; if __json_exchange_column__, _ok_ = _buf["exchange_column"]; !_ok_ || __json_exchange_column__ == nil { err = errors.New("exchange_column error"); return } else { var __x__ *ItemItemExchange;  { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = __json_exchange_column__.(map[string]interface{}); !_ok_ { err = errors.New("__x__ error"); return }; if __x__, err = NewItemItemExchange(_x_); err != nil { return } }; _v.ExchangeColumn = __x__ }}
    { var _ok_ bool; var __json_AddAttr__ interface{}; if __json_AddAttr__, _ok_ = _buf["AddAttr"]; !_ok_ || __json_AddAttr__ == nil { err = errors.New("AddAttr error"); return } else { var __x__ map[int32]int64;  {
                    var _arr0_ []interface{}
                    var _ok0_ bool
                    if _arr0_, _ok_ = (__json_AddAttr__).([]interface{}); !_ok_ { err = errors.New("__x__ error"); return }
    
                    __x__ = make(map[int32]int64)
                    
                    for _, _e0_ := range _arr0_ {
                        var _kv0_ []interface{}
                        if _kv0_, _ok0_ = _e0_.([]interface{}); !_ok0_ || len(_kv0_) != 2 { err = errors.New("__x__ error"); return }
                        var _key0_ int32
                        { var _ok_ bool; var _x_ float64; if _x_, _ok_ = _kv0_[0].(float64); !_ok_ { err = errors.New("_key0_ error"); return }; _key0_ = int32(_x_) }
                        var _value0_ int64
                        { var _ok_ bool; var _x_ float64; if _x_, _ok_ = _kv0_[1].(float64); !_ok_ { err = errors.New("_value0_ error"); return }; _value0_ = int64(_x_) }
                        __x__[_key0_] = _value0_
                    }
                    }; _v.AddAttr = __x__ }}
    { var _ok_ bool; var __json_AddItem__ interface{}; if __json_AddItem__, _ok_ = _buf["AddItem"]; !_ok_ || __json_AddItem__ == nil { err = errors.New("AddItem error"); return } else { var __x__ []*ItemData;  {
                    var _arr0_ []interface{}
                    var _ok0_ bool
                    if _arr0_, _ok0_ = (__json_AddItem__).([]interface{}); !_ok0_ { err = errors.New("__x__ error"); return }
    
                    __x__ = make([]*ItemData, 0, len(_arr0_))
                    
                    for _, _e0_ := range _arr0_ {
                        var _list_v0_ *ItemData
                        { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _e0_.(map[string]interface{}); !_ok_ { err = errors.New("_list_v0_ error"); return }; if _list_v0_, err = NewItemData(_x_); err != nil { return } }
                        __x__ = append(__x__, _list_v0_)
                    }
                }
    ; _v.AddItem = __x__ }}
    return
}


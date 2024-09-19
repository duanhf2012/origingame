
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package tabledef;

type JsonLoader func(string) ([]map[string]interface{}, error)

type Tables struct {
    TbItem *ItemTbItem
}

func NewTables(loader JsonLoader) (*Tables, error) {
    var err error
    var buf []map[string]interface{}

    tables := &Tables{}
    if buf, err = loader("item_tbitem") ; err != nil {
        return nil, err
    }
    if tables.TbItem, err = NewItemTbItem(buf) ; err != nil {
        return nil, err
    }
    return tables, nil
}



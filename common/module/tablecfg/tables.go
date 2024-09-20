package tablecfg

import (
	"origingame/table/tabledef"
	"sync/atomic"
)

type ProcessedTables struct {
}

type Tables struct {
	*tabledef.Tables
	*ProcessedTables
}

var tables Tables
var vTables atomic.Value

func SetTables(tables Tables) {
	vTables.Store(&tables)
}

func GetTables() *Tables {
	return vTables.Load().(*Tables)
}

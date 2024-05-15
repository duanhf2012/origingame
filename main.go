package main

import (
	"github.com/duanhf2012/origin/v2/node"
	"time"

	_ "origingame/gamecore/centerservice"
	_ "origingame/gamecore/dbservice"
	_ "origingame/gamecore/gameservice"
	_ "origingame/gamecore/gateservice"
	_ "origingame/gamemaster/authservice"
	_ "origingame/gamemaster/httpgateservice"
)

func main() {
	node.OpenProfilerReport(time.Second * 10)
	node.Start()
}

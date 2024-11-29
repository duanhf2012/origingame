package main

import (
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	_ "origingame/service/authservice"
	_ "origingame/service/botservice"
	_ "origingame/service/centerservice"
	_ "origingame/service/dbservice"
	_ "origingame/service/gameservice"
	_ "origingame/service/gateservice"
	_ "origingame/service/hotloadservice"
	_ "origingame/service/httpgateservice"
	"time"
)

func main() {
	// 使用文本日志格式,默认为json
	log.GetLogger().SetEncoder(log.GetTxtEncoder())

	node.OpenProfilerReport(time.Second * 10)
	node.Start()
}

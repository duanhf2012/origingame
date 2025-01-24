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
	// 使用文本日志格式,默认为json格式
	log.GetLogger().SetEncoder(log.GetTxtEncoder())

	// 默认使用rotatelogs日志分割，以下为Lumberjack分割
	//log.GetLogger().SetSyncers(log.GetLogger().NewLumberjackWriter)

	// 打开性能报告
	node.OpenProfilerReport(time.Second * 10)

	// 开启Node
	node.Start()
}

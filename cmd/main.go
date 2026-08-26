// OriginGame 是统一的游戏服务器程序入口。
// 实际运行的 Node 及其 Service 由配置和启动参数选择。
package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"origingame/service/dbservice"
	"origingame/service/gameservice"
	"origingame/service/gatewayservice"
	"origingame/service/loginservice"
	"origingame/service/robotservice"

	"github.com/duanhf2012/origin/v3/application"
)

// app 是最终可执行程序唯一的 Application 实例。
var app = application.New(application.Options{
	StartTimeout: 120 * time.Second,
	StopTimeout:  120 * time.Second,
})

// 此处只登记配置可以引用的 Service 类型；实际实例由所选 Node 的 services 配置创建。
func init() {
	app.Setup(
		&dbservice.DBService{},
		&gatewayservice.GatewayService{},
		&gameservice.GameService{},
		&loginservice.LoginService{},
		&robotservice.RobotService{},
	)
}

func main() {
	// Gin 的调试路由输出不属于 OriginGame 的结构化日志，最终入口统一关闭它。
	gin.SetMode(gin.ReleaseMode)
	app.Start()
}

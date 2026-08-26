# RobotService 蓝图工作区

使用 OriginBlueprint 打开本目录。`nodes/robot.json` 是 RobotService 节点契约，`blueprints/login_heartbeat.obp` 是首个真实客户端登录与心跳场景。

节点的 `name`、`port_id`、端口类型和顺序与 Go 实现共同构成兼容契约，不应直接重排。默认场景在登录后通过 `RobotStartHeartbeat` 启动 VirtualPlayer 后台心跳，然后进入独立的结构化 `While` 主业务循环；运行 Context 到期后由 RobotService 统一停止心跳并关闭连接。

结构化 `While` 的循环体执行到 `RobotHeartbeat` 成功出口末端后，由 VM 自动进入下一轮，不画回连线；直接回连 `While` 会形成非法 Exec 环。蓝图在节点属性中内嵌了 `WhileNode` 端口快照，确保缺少该内建可视化定义的编辑器版本也能完整恢复节点和连接。

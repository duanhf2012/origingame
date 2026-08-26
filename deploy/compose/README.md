# OriginGame v3 本地基础设施

本目录通过一个 Docker Compose 启动最小开发环境：

- 单节点 etcd；
- 单节点 Core NATS；
- 单节点 Redis，开启 AOF 持久化并使用 `noeviction`；
- 单节点 MongoDB 副本集，支持事务测试；
- MongoDB 三张基础表、校验器、索引和一组基础区服数据。

该编排用于学习、开发和集成测试，不是生产部署模板。etcd、NATS、Redis 和 MongoDB 默认绑定宿主机 `192.168.8.3`。

当前 OriginGame 使用 Origin 内置服务发现和 TCP RPC，etcd 与 NATS 不参与当前服务发现或服务间 RPC；它们仅作为可选基础设施保留。

当前 Compose 固定使用 `mongo:8.0.4`，用于兼容项目开发机的 Linux `7.0.0` 内核。MongoDB 官方说明 Linux `6.19` 到 `7.0.13` 与当前 TCMalloc 存在兼容问题；正式环境应使用官方支持的内核（例如 `7.0.14+`），并将 MongoDB 更新到受维护的新版本。

## 1. 启动

在本目录执行：

```bash
docker compose up -d
docker compose ps
```

首次启动会自动创建 MongoDB 数据库和基础数据。等待 MongoDB、Redis 状态变为 `healthy` 后，即可启动 OriginGame 服务。

## 2. 连接信息

| 服务 | 宿主机地址 | 容器网络地址 |
| --- | --- | --- |
| etcd | `http://192.168.8.3:2379` | `http://etcd:2379` |
| NATS（可选） | `nats://192.168.8.3:4222` | `nats://nats:4222` |
| NATS 监控（可选） | `http://192.168.8.3:8222` | `http://nats:8222` |
| Redis | `redis://192.168.8.3:6379` | `redis://redis:6379` |
| MongoDB | `192.168.8.3:27017` | `mongodb:27017` |

宿主机连接 URI：

```text
mongodb://192.168.8.3:27017/?replicaSet=rs0&directConnection=true
```

其他 Compose 容器连接 URI：

```text
mongodb://mongodb:27017/?replicaSet=rs0
```

MongoDB 当前没有配置认证，只能用于本地开发；账号域使用 `origingame_account`，角色域使用 `origingame_role`。

Redis 当前没有配置认证，只能用于本地开发。其淘汰策略固定为 `noeviction`：内存不足时写入失败，不允许静默淘汰可能承载在线归属等关键状态的键。

## 3. MongoDB 初始化内容

初始化脚本位于 `mongodb/init/01-init-origingame.js`，只会在 MongoDB 数据卷为空时由官方镜像自动执行。

创建内容：

- `Account`：账号表；`_id` 为 ObjectID，`(PlatType, PlatId)` 建立唯一索引；
- `RealAreaInfo`：真实区服及 TCP/KCP/WebSocket Gateway 公网地址；
- `ShowAreaInfo`：显示区服及真实区服映射，`RealAreaId` 建立索引；
- `UserInfo`：以 `AccountID + ShowAreaID` 组成的 PlayerKey 为主键；
- 真实区服 `1`；
- 显示区服 `1 / 体验1服`；
- Gateway 开发地址 `9001/TCP`、`9002/KCP`、`9003/WebSocket`。

`Account` 只创建集合、校验器和索引，不预先插入账号。

## 4. 验证

查看可选 NATS 健康状态：

```text
http://192.168.8.3:8222/healthz
```

查看可选 etcd 健康状态：

```bash
docker exec origingame-etcd etcdctl --endpoints=http://127.0.0.1:2379 endpoint health
```

查看 Redis 健康状态：

```bash
docker exec origingame-redis redis-cli ping
```

查看 MongoDB 基础数据：

```bash
docker exec origingame-mongodb mongosh origingame_account \
  --eval "db.RealAreaInfo.find(); db.ShowAreaInfo.find(); db.Account.getIndexes()"
```

PowerShell 可以将上述多行命令写成一行执行。

## 5. 停止与重建

停止容器但保留数据：

```bash
docker compose down
```

删除容器和数据卷并重新执行初始化脚本：

```bash
docker compose down -v
docker compose up -d
```

`down -v` 会删除本地 etcd、Redis 和 MongoDB 开发数据，只应在确认不再需要这些数据时执行。

如需调整绑定地址，可使用 `ORIGINGAME_ETCD_BIND_ADDRESS`、`ORIGINGAME_NATS_BIND_ADDRESS`、`ORIGINGAME_REDIS_BIND_ADDRESS` 和 `ORIGINGAME_MONGODB_BIND_ADDRESS`。由于本编排没有为 etcd、NATS 和 Redis 配置认证，不能将其暴露到公网。

宿主机标准端口已被占用时，可以覆盖映射端口而不修改 Compose：

```bash
ORIGINGAME_ETCD_PORT=32379 \
ORIGINGAME_NATS_PORT=34222 \
ORIGINGAME_NATS_MONITOR_PORT=38222 \
ORIGINGAME_REDIS_PORT=36379 \
ORIGINGAME_MONGODB_PORT=37017 \
docker compose up -d
```

// OriginGame v3 本地开发数据库初始化。
// Docker 官方 MongoDB 镜像只会在空数据目录首次启动时执行本目录脚本。

const appDB = db.getSiblingDB("origingame");

function ensureCollection(name, validator) {
    const exists = appDB.getCollectionInfos({name: name}).length > 0;
    const options = {
        validator: validator,
        validationLevel: "strict",
        validationAction: "error"
    };

    if (!exists) {
        appDB.createCollection(name, options);
        return;
    }

    appDB.runCommand({
        collMod: name,
        validator: validator,
        validationLevel: "strict",
        validationAction: "error"
    });
}

ensureCollection("Account", {
    $jsonSchema: {
        bsonType: "object",
        required: ["_id", "PlatType", "PlatId", "Gm", "CreateTime", "UpdateTime"],
        properties: {
            _id: {
                bsonType: "objectId",
                description: "游戏内部稳定 AccountID"
            },
            PlatType: {
                bsonType: ["int", "long"],
                description: "登录平台类型"
            },
            PlatId: {
                bsonType: "string",
                minLength: 1,
                description: "玩家在登录平台的身份标识"
            },
            Gm: {
                bsonType: "bool",
                description: "是否为 GM 账号"
            },
            CreateTime: {
                bsonType: "date",
                description: "账号创建时间"
            },
            UpdateTime: {
                bsonType: "date",
                description: "账号最后更新时间"
            }
        }
    }
});

appDB.Account.createIndex(
    {PlatType: 1, PlatId: 1},
    {name: "uk_platform_identity", unique: true}
);

ensureCollection("RealAreaInfo", {
    $jsonSchema: {
        bsonType: "object",
        required: ["_id", "GateList"],
        properties: {
            _id: {
                bsonType: ["int", "long"],
                minimum: 1,
                description: "真实区服 ID"
            },
            GateList: {
                bsonType: "array",
                minItems: 1,
                items: {
                    bsonType: "object",
                    required: ["Protocol", "Address"],
                    properties: {
                        Protocol: {
                            enum: ["tcp", "kcp", "websocket"],
                            description: "客户端连接协议"
                        },
                        Address: {
                            bsonType: "string",
                            minLength: 1,
                            description: "客户端可访问的 Gateway 公网地址"
                        }
                    }
                }
            }
        }
    }
});

ensureCollection("ShowAreaInfo", {
    $jsonSchema: {
        bsonType: "object",
        required: [
            "_id",
            "AreaName",
            "RealAreaId",
            "ServerMark",
            "ServerStatus",
            "OpenTime"
        ],
        properties: {
            _id: {
                bsonType: ["int", "long"],
                minimum: 1,
                description: "客户端显示区服 ID"
            },
            AreaName: {
                bsonType: "string",
                minLength: 1,
                description: "客户端显示的区服名称"
            },
            RealAreaId: {
                bsonType: ["int", "long"],
                minimum: 1,
                description: "关联的真实区服 ID"
            },
            ServerMark: {
                bsonType: ["int", "long"],
                minimum: 0,
                maximum: 2,
                description: "0 普通、1 新服、2 推荐"
            },
            ServerStatus: {
                bsonType: ["int", "long"],
                minimum: 0,
                maximum: 1,
                description: "0 正常、1 维护"
            },
            OpenTime: {
                bsonType: ["int", "long"],
                minimum: 0,
                description: "开服时间 Unix 毫秒；0 表示本地开发不限制"
            }
        }
    }
});

appDB.ShowAreaInfo.createIndex(
    {RealAreaId: 1},
    {name: "idx_real_area_id"}
);

// 基础数据只在不存在时插入，重新手工执行脚本不会覆盖使用者已经修改的地址或区服状态。
appDB.RealAreaInfo.updateOne(
    {_id: NumberInt(1)},
    {
        $setOnInsert: {
            GateList: [
                {Protocol: "tcp", Address: "127.0.0.1:9001"},
                {Protocol: "kcp", Address: "127.0.0.1:9002"},
                {Protocol: "websocket", Address: "ws://127.0.0.1:9003/ws"}
            ]
        }
    },
    {upsert: true}
);

appDB.ShowAreaInfo.updateOne(
    {_id: NumberInt(1)},
    {
        $setOnInsert: {
            AreaName: "体验1服",
            RealAreaId: NumberInt(1),
            ServerMark: NumberInt(2),
            ServerStatus: NumberInt(0),
            OpenTime: NumberLong("0")
        }
    },
    {upsert: true}
);

print("OriginGame database initialized: Account, RealAreaInfo, ShowAreaInfo");

// OriginGame v3 本地开发数据库初始化。
// Docker 官方 MongoDB 镜像只会在空数据目录首次启动时执行本目录脚本。

const accountDB = db.getSiblingDB("origingame_account");
const roleDB = db.getSiblingDB("origingame_role");

function ensureCollection(database, name, validator) {
    const exists = database.getCollectionInfos({name: name}).length > 0;
    const options = {
        validator: validator,
        validationLevel: "strict",
        validationAction: "error"
    };

    if (!exists) {
        database.createCollection(name, options);
        return;
    }

    database.runCommand({
        collMod: name,
        validator: validator,
        validationLevel: "strict",
        validationAction: "error"
    });
}

ensureCollection(accountDB, "Account", {
    $jsonSchema: {
        bsonType: "object",
        required: ["_id", "PlatType", "PlatId", "CreateTime", "UpdateTime"],
        properties: {
            _id: {bsonType: "objectId", description: "游戏内部稳定 AccountID"},
            PlatType: {bsonType: ["int", "long"], description: "登录平台类型"},
            PlatId: {bsonType: "string", minLength: 1, description: "玩家在登录平台的身份标识"},
            Ip: {bsonType: "string", description: "最近登录 IP"},
            CreateTime: {bsonType: "date", description: "账号创建真实系统时间"},
            UpdateTime: {bsonType: "date", description: "账号最后更新真实系统时间"}
        }
    }
});

accountDB.Account.createIndex(
    {PlatType: 1, PlatId: 1},
    {name: "uk_platform_identity", unique: true}
);

ensureCollection(accountDB, "RealAreaInfo", {
    $jsonSchema: {
        bsonType: "object",
        required: ["_id", "GateList"],
        properties: {
            _id: {bsonType: ["int", "long"], minimum: 1, description: "真实区服 ID"},
            GateList: {
                bsonType: "array",
                minItems: 1,
                items: {
                    bsonType: "object",
                    required: ["Protocol", "Address"],
                    properties: {
                        Protocol: {enum: ["tcp", "kcp", "websocket"], description: "客户端协议"},
                        Address: {bsonType: "string", minLength: 1, description: "Gateway 公网地址"}
                    }
                }
            }
        }
    }
});

ensureCollection(accountDB, "ShowAreaInfo", {
    $jsonSchema: {
        bsonType: "object",
        required: ["_id", "AreaName", "RealAreaId", "ServerMark", "ServerStatus", "OpenTime"],
        properties: {
            _id: {bsonType: ["int", "long"], minimum: 1, description: "显示区服 ID"},
            AreaName: {bsonType: "string", minLength: 1, description: "客户端显示区服名称"},
            RealAreaId: {bsonType: ["int", "long"], minimum: 1, description: "真实区服 ID"},
            ServerMark: {bsonType: ["int", "long"], minimum: 0, maximum: 2, description: "区服标记"},
            ServerStatus: {bsonType: ["int", "long"], minimum: 0, maximum: 1, description: "区服状态"},
            OpenTime: {bsonType: ["int", "long"], minimum: 0, description: "开服业务时间 Unix 毫秒"}
        }
    }
});

accountDB.ShowAreaInfo.createIndex(
    {RealAreaId: 1},
    {name: "idx_real_area_id"}
);

ensureCollection(roleDB, "UserInfo", {
    $jsonSchema: {
        bsonType: "object",
        required: ["_id", "account_id", "show_area_id", "nickname", "level", "created_at", "last_login_at", "last_logout_at"],
        properties: {
            _id: {bsonType: "string", minLength: 1, description: "AccountID 与 ShowAreaID 组成的 PlayerKey"},
            account_id: {bsonType: "string", minLength: 1, description: "账号 ID"},
            show_area_id: {bsonType: ["int", "long"], minimum: 1, description: "显示区服 ID"},
            nickname: {bsonType: "string", description: "玩家昵称"},
            level: {bsonType: ["int", "long"], minimum: 1, description: "玩家等级"},
            created_at: {bsonType: "date", description: "角色创建真实系统时间"},
            last_login_at: {bsonType: "date", description: "最近登录真实系统时间"},
            last_logout_at: {bsonType: "date", description: "最近离线真实系统时间"}
        }
    }
});

roleDB.UserInfo.createIndex(
    {account_id: 1, show_area_id: 1},
    {name: "uk_account_show_area", unique: true}
);

// 基础区服数据只在不存在时插入，重复执行不会覆盖开发者已经修改的地址或状态。
accountDB.RealAreaInfo.updateOne(
    {_id: NumberInt(1)},
    {$setOnInsert: {GateList: [
        {Protocol: "tcp", Address: "127.0.0.1:9001"},
        {Protocol: "kcp", Address: "127.0.0.1:9002"},
        {Protocol: "websocket", Address: "ws://127.0.0.1:9003/ws"}
    ]}},
    {upsert: true}
);

accountDB.ShowAreaInfo.updateOne(
    {_id: NumberInt(1)},
    {$setOnInsert: {
        AreaName: "体验1服",
        RealAreaId: NumberInt(1),
        ServerMark: NumberInt(2),
        ServerStatus: NumberInt(0),
        OpenTime: NumberLong("0")
    }},
    {upsert: true}
);

print("OriginGame databases initialized: account domain and role domain");

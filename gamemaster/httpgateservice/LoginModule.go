package httpgateservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/sysmodule/ginmodule"
	"github.com/duanhf2012/origin/v2/util/timer"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"origingame/common/collect"
	"origingame/common/db"
	global "origingame/common/keyword"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/common/util"
	"strconv"
	"time"
)

const MaxLoginCD = 3 * time.Second //登陆时间cd3秒

type GateStatus int

const (
	OnLine  GateStatus = 1
	Offline GateStatus = 0
)

type GateInfoResp struct {
	Url string
}

type LoginModule struct {
	service.Module
	jsonAreaInfo string

	forbidGuestLogin bool //是否限制游客和账号登录 true限制 false不限制

	mapPlatIdLoginTime map[string]int64 //账号登录时间-用于cd限制 map[PlatId]最后登录时间 PlatId:玩家在平台的openid
	lastResetTime      int64            //每过1一定时间(暂定10分钟)重置一次mapPlatIdLoginTime，避免map一直增大
}

func (login *LoginModule) OnInit() error {
	login.mapPlatIdLoginTime = make(map[string]int64, 1024)
	login.lastResetTime = 0

	return nil
}

func (login *LoginModule) OnRelease() {
}

// 设置是否限制游客和账号登录
func (login *LoginModule) SetForbidGuestLogin(forbid bool) {
	login.forbidGuestLogin = forbid
}

type HttpRespone struct {
	ECode    int
	Token    string
	AreaGate string
	AreaLoad string //区服负载
	AreaHis  map[string]int64
}

func (login *LoginModule) loginCheck(c *ginmodule.SafeContext, loginInfo *rpc.LoginInfo) {
	log.SDebug("http request PlatType:", int32(loginInfo.PlatType), " PlatId:", loginInfo.PlatId, " AccessToken:",
		loginInfo.AccessToken, " ActiveCode:", loginInfo.ActiveCode, " Sign:", loginInfo.Sign, " ChannelCode:", loginInfo.ChanneCode,
		" ChannePlat:", loginInfo.ChannePlat)

	//1.验证平台类型和Id
	platId := loginInfo.PlatId
	if loginInfo.PlatType < 0 || loginInfo.PlatType >= rpc.LoginType_LoginType_Max {
		log.SWarning("plat type ", loginInfo.PlatType, " is error!")
		c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_PlatTypeError})
		return
	}

	if len(platId) == 0 {
		log.SWarning("plat type ", loginInfo.PlatType, " is error!")
		c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_PlatIdError})
		return
	}
	loginInfo.LoginCheckTime = timer.Now().UnixMilli()
	
	//2.向验证服检查登陆
	if loginInfo.PlatType == rpc.LoginType_Gust || loginInfo.PlatType == rpc.LoginType_Account {
		//判断是否限制了游客和账号登录
		if login.forbidGuestLogin == true {
			log.SDebug("guest and account login limit, cannot login!!")
			c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_PlatTypeError})
			return
		}
		loginInfo.PlatId += "_" + strconv.FormatInt(int64(loginInfo.PlatType), 10)
		//验证通过从数据库生成或获取账号信息
		login.loginToDB(c, loginInfo)
	} else {
		err := login.AsyncCall("AuthService.RPC_Check", &loginInfo, func(loginResult *rpc.LoginResult, err error) {
			if err != nil {
				log.SError("call authservice.RPC_Check fail:", err.Error(), ",platId:", platId)
				c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_InterNalError})
				return
			}

			if loginResult.Ret != 0 {
				log.SWarning("authservice.RPC_Check fail Ret:", loginResult.Ret, ",platId:", platId)
				c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_TokenError})
				return
			}

			loginInfo.PlatId += "_" + strconv.FormatInt(int64(loginInfo.PlatType), 10)
			//验证通过从数据库生成或获取账号信息
			login.loginToDB(c, loginInfo)
		})

		//3.服务内部错误
		if err != nil {
			c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_InterNalError})
			log.SError("AsyncCall authservice.RPC_Check fail:", err.Error(), ",platId:", platId)
		}
	}
}

// 对登录信息进行验证: 验签(游戏自己的登录验签,防止客户端无限制进行请求) 和 CD判断
func (login *LoginModule) checkLoginParam(loginInfo *rpc.LoginInfo) msg.ErrCode {
	//1.验证CD
	retCode := login.checkLoginCd(loginInfo)
	if retCode != msg.ErrCode_OK {
		log.SError("LoginModule.checkLoginParam checkLoginCd retCode:", int32(retCode))
		return retCode
	}

	return msg.ErrCode_OK
}

// 验证登录PlatId的CD
func (login *LoginModule) checkLoginCd(loginInfo *rpc.LoginInfo) msg.ErrCode {
	//1.参数检查
	if loginInfo == nil {
		log.SError("LoginModule.checkLoginCd loginInfo is nil")
		return msg.ErrCode_LoginParamError
	}

	//2.重置记录的map,避免一直增大
	nowTime := time.Now().UnixNano()
	if login.lastResetTime == 0 {
		login.lastResetTime = nowTime
	} else {
		if nowTime-login.lastResetTime > int64(global.RefreshLoginCDMap) {
			login.mapPlatIdLoginTime = make(map[string]int64, 1024)
			return msg.ErrCode_OK
		}
	}

	//3.检查CD
	lastLoginTime, ok := login.mapPlatIdLoginTime[loginInfo.PlatId]
	if ok == false {
		login.mapPlatIdLoginTime[loginInfo.PlatId] = nowTime
		return msg.ErrCode_OK
	}

	diffTime := nowTime - lastLoginTime
	if diffTime > int64(MaxLoginCD) {
		login.mapPlatIdLoginTime[loginInfo.PlatId] = nowTime
		return msg.ErrCode_OK
	}

	log.SError("LoginModule.checkLoginCd PlatId:", loginInfo.PlatId, " login cd is ", diffTime)
	return msg.ErrCode_LoginCDError
}

func (login *LoginModule) loginToDB(c *ginmodule.SafeContext, loginInfo *rpc.LoginInfo) {
	//2.生成数据库请求
	var req db.DBControllerReq
	req.CollectName = collect.AccountCollectName
	req.Type = db.OptType_FindOneAndUpdate
	platId := loginInfo.PlatId
	req.Condition, _ = bson.Marshal(bson.D{{"_id", platId}})
	req.Key = loginInfo.PlatId
	req.Upsert = true

	upsertAccount := bson.M{"PlatType": int32(loginInfo.PlatType), "_id": loginInfo.PlatId, "Ip": c.GetHeader("X-Real-IP")} //todo 后续添加设备号、渠道
	data, err := bson.Marshal(upsertAccount)
	if err != nil {
		c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_InterNalError})
		log.SError("LoginToDB fail:", err.Error(), ",platId:", platId)
		return
	}

	req.Data = append(req.Data, data)
	//3.平台登陆验证成功，去DB创建或者查询账号
	err = login.GetService().GetRpcHandler().AsyncCall(util.AccDBRequest, &req, func(res *db.DBControllerRet, err error) {
		//返回账号创建结果
		if err != nil || len(res.Res) == 0 {
			if err != nil {
				log.SError("Call DBService.RPC_DBRequest platId:", platId, ", fail:", err.Error())
			} else {
				log.SError("Call DBService.RPC_DBRequest platId:", platId, ", fail res is empty!")
			}

			c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_InterNalError})
			return
		}

		//解析数据
		var account collect.CAccount
		err = bson.Unmarshal(res.Res[0], &account)
		if err != nil {
			c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_InterNalError})
			log.SError("Unmarshal fail:", err.Error(), ",platId:", platId)
			return
		}

		//登陆成功，向客户端返回登陆列表
		//登陆成功,返回结果
		var resp HttpRespone
		resp.Token, err = util.EncryptToken(account.PlatId)
		if err != nil {
			c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_InterNalError})
			log.SError("EncryptToken platId:", account.PlatId, ",fail:", err.Error())
			return
		}

		resp.AreaGate = login.GetShowAreaGateInfo()
		resp.AreaHis = account.AreaHis
		c.JSONAndDone(http.StatusOK, &resp)
	})

	if err != nil {
		c.JSONAndDone(http.StatusOK, gin.H{"ECode": msg.ErrCode_InterNalError})
		log.SError("AsyncCall DBService.RPC_DBRequest fail:", err.Error(), ",platId:", platId)
	}
}

func (login *LoginModule) GetShowAreaGateInfo() string {
	return login.jsonAreaInfo
}

func (login *LoginModule) SetJsonAreaGate(jsonAreaGate string) {
	login.jsonAreaInfo = jsonAreaGate
}

func (login *LoginModule) Login(c *ginmodule.SafeContext) {
	//1.验证Body请求内容
	var loginInfo rpc.LoginInfo
	err := c.ShouldBindBodyWith(&loginInfo, binding.JSON)
	if err != nil || loginInfo.AccessToken == "" {
		c.JSONAndDone(http.StatusBadRequest, gin.H{"ECode": msg.ErrCode_InterNalError})
		return
	}

	//2.平台登陆验证
	login.loginCheck(c, &loginInfo)
}

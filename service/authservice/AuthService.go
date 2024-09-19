package authservice

import (
	"crypto/md5"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/timer"
	"origingame/common/proto/rpc"
	"strings"

	"time"
)

func init() {
	node.Setup(&AuthService{})
}

type AuthService struct {
	service.Service
}

func (auth *AuthService) OnInit() error {
	//1.解析配置文件
	var authCfg struct {
		MinGoroutineNum   int32
		MaxGoroutineNum   int32
		MaxTaskChannelNum int
	}
	err := auth.ParseServiceCfg(&authCfg)
	if err != nil {
		return err
	}

	//2.打开协程模式
	auth.OpenConcurrent(authCfg.MinGoroutineNum, authCfg.MaxGoroutineNum, authCfg.MaxTaskChannelNum)

	//性能监控
	auth.OpenProfiler()
	auth.GetProfiler().SetOverTime(time.Second * 2)
	auth.GetProfiler().SetMaxOverTime(time.Second * 10)

	return nil
}

type LoginRet struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Status string `json:"status"`
	Data   string `json:"data"`
	Token  string `json:"token"`
}

const appKey = "a893e8177a5a4a5eb23833fa7a0278a9"

func (auth *AuthService) Sign(username, token, gameId string) string {
	rawStr := username + token + gameId + appKey
	return strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(rawStr))))
}

func (auth *AuthService) RPCCheck(responder originRpc.Responder, loginInfo *rpc.LoginInfo) error {
	if timer.Now().UnixMilli()-loginInfo.LoginCheckTime > 5000 {
		log.SError("time too lang ", int(loginInfo.PlatType))
		responder(nil, originRpc.RpcError("time too lang"))
		return nil
	}

	switch loginInfo.PlatType {
	case rpc.LoginType_Gust:
		auth.AsyncDo(func() bool {
			auth.guestCheck(responder, loginInfo)
			return true
		}, nil)

		return nil
	case rpc.LoginType_Account:
		auth.AsyncDo(func() bool {
			auth.accountCheck(responder, loginInfo)
			return true
		}, nil)
		return nil
	case rpc.LoginType_TapTap:
		auth.AsyncDo(func() bool {
			auth.tapTapCheck(responder, loginInfo)
			return true
		}, nil)

		return nil
	}

	responder(nil, originRpc.RpcError("unknown platType"))
	log.SError("platType is error:", int(loginInfo.PlatType))
	return nil
}

// guestCheck 游客登陆Check,非线程安全
func (auth *AuthService) guestCheck(responder originRpc.Responder, loginInfo *rpc.LoginInfo) {
	var loginResult rpc.LoginResult
	var rpcErr originRpc.RpcError
	responder(&loginResult, rpcErr)
}

// accountCheck 游戏账号登陆Check,非线程安全
func (auth *AuthService) accountCheck(responder originRpc.Responder, loginInfo *rpc.LoginInfo) {
	var loginResult rpc.LoginResult
	var rpcErr originRpc.RpcError
	responder(&loginResult, rpcErr)
}

// accountCheck tapTap SDK登陆Check,非线程安全
func (auth *AuthService) tapTapCheck(responder originRpc.Responder, loginInfo *rpc.LoginInfo) {
	var loginResult rpc.LoginResult
	var rpcErr originRpc.RpcError
	responder(&loginResult, rpcErr)
}

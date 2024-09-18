package gm

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/service"
	"origingame/common/env"
	"origingame/common/proto/msg"
	"origingame/gamecore/gameservice/msgrouter"
	"origingame/gamecore/gameservice/player"
	"reflect"
	"strings"
)

type gmFun func(player *player.Player, param []string) string

type gmMethodInfo struct {
	method       reflect.Method
	inParamValue [2]reflect.Value
	inParam      [2]interface{}
	st           any
}

type GmModule struct {
	service.Module
	mapGm map[string]*gmMethodInfo
}

func (gm *GmModule) OnInit() error {
	gm.mapGm = make(map[string]*gmMethodInfo)
	gm.analysisCommand(&Common{})

	msgrouter.RegMsgHandler(msg.MsgType_GM, gm.Call)
	return nil
}

func (gm *GmModule) analysisCommand(st any) {
	typ := reflect.TypeOf(st)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		err := gm.suitableMethods(st, method)
		if err != nil {
			panic(err)
		}
	}
}

func (gm *GmModule) suitableMethods(st any, method reflect.Method) error {
	typ := method.Type
	if typ.NumOut() != 1 {
		return fmt.Errorf("%s The number of returned arguments must be 1", method.Name)
	}

	if typ.Out(0).String() != "string" {
		return fmt.Errorf("%s The return parameter must be of type error", method.Name)
	}

	if typ.NumIn() != 3 {
		return fmt.Errorf("%s Unsupported parameter format", method.Name)
	}

	var methodInfo gmMethodInfo
	for i := 1; i < typ.NumIn(); i++ {

		methodInfo.inParamValue[i-1] = reflect.New(typ.In(i).Elem())
		methodInfo.inParam[i-1] = reflect.New(typ.In(i).Elem()).Interface()
	}

	methodInfo.method = method
	methodInfo.st = st
	gm.mapGm[strings.ToLower(method.Name)] = &methodInfo

	return nil
}

func (gm *GmModule) Call(p *player.Player, req *msg.MsgGmReq) {
	if !p.IsGm() && !env.IsRelease {
		return
	}

	var gmRes msg.MsgGmRes
	defer p.SendMsg(msg.MsgType_GM, &gmRes)
	cmd := strings.ToLower(req.Command)
	methodInfo := gm.mapGm[cmd]
	if methodInfo == nil {
		gmRes.Ret = "invalid command"
		return
	}

	var paramList []reflect.Value
	paramList = append(paramList, reflect.ValueOf(methodInfo.st))
	paramList = append(paramList, reflect.ValueOf(p))
	paramList = append(paramList, reflect.ValueOf(req.Param))

	returnValues := methodInfo.method.Func.Call(paramList)
	errInter := returnValues[0].Interface()
	if errInter != nil {
		gmRes.Ret = errInter.(string)
	}
}

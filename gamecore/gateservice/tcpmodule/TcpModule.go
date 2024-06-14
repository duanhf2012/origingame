package tcpmodule

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/bytespool"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"runtime"
	"sync"
	"time"
)

type TcpModule struct {
	tcpServer network.TCPServer
	service.Module

	mapClientLocker sync.RWMutex
	mapClient       map[string]*Client
	process         processor.IRawProcessor
}

type TcpPackType int8

const (
	TPTConnected    TcpPackType = 0
	TPTDisConnected TcpPackType = 1
	TPTPack         TcpPackType = 2
	TPTUnknownPack  TcpPackType = 3
)

type TcpPack struct {
	Type                TcpPackType //0表示连接 1表示断开 2表示数据
	ClientId            string
	Data                interface{}
	RecyclerReaderBytes func(data []byte)
}

type Client struct {
	id        string
	tcpConn   *network.TCPConn
	tcpModule *TcpModule
}

type TcpCfg struct {
	ListenAddr          string        //监听地址
	MaxConnNum          int           //最大连接数
	PendingWriteNum     int           //写channel最大消息数量
	LittleEndian        bool          //是否小端序
	LenMsgLen           int           //消息头占用byte数量，只能是1byte,2byte,4byte。如果是4byte，意味着消息最大可以是math.MaxUint32(4GB)
	MinMsgLen           uint32        //最小消息长度
	MaxMsgLen           uint32        //最大消息长度,超过判定不合法,断开连接
	ReadDeadlineSecond  time.Duration //读超时
	WriteDeadlineSecond time.Duration //写超时
}

func (tm *TcpModule) OnInit() error {
	//1解析配置
	iConfig := tm.GetService().GetServiceCfg()
	if iConfig == nil {
		return fmt.Errorf("%s config is error", tm.GetService().GetName())
	}
	mapTcpCfg := iConfig.(map[string]interface{})
	iTcpCfg, ok := mapTcpCfg["TcpCfg"]
	if ok == false {
		return fmt.Errorf("%s.TcpCfg config is error", tm.GetService().GetName())
	}

	var tcpCfg TcpCfg
	byteTcpCfg, err := json.Marshal(iTcpCfg)
	if err != nil {
		return fmt.Errorf("%s.TcpCfg config is error:%s", tm.GetService().GetName(), err.Error())
	}
	err = json.Unmarshal(byteTcpCfg, &tcpCfg)
	if err != nil {
		return fmt.Errorf("%s.TcpCfg config is error:%s", tm.GetService().GetName(), err.Error())
	}

	//2.初始化网络模块
	tm.tcpServer.Addr = tcpCfg.ListenAddr
	tm.tcpServer.MaxConnNum = tcpCfg.MaxConnNum
	tm.tcpServer.PendingWriteNum = tcpCfg.PendingWriteNum
	tm.tcpServer.LittleEndian = tcpCfg.LittleEndian
	tm.tcpServer.LenMsgLen = tcpCfg.LenMsgLen
	tm.tcpServer.MinMsgLen = tcpCfg.MinMsgLen
	tm.tcpServer.MaxMsgLen = tcpCfg.MaxMsgLen
	tm.tcpServer.ReadDeadline = tcpCfg.ReadDeadlineSecond * time.Second
	tm.tcpServer.WriteDeadline = tcpCfg.WriteDeadlineSecond * time.Second
	tm.mapClient = make(map[string]*Client, tm.tcpServer.MaxConnNum)
	tm.tcpServer.NewAgent = tm.NewClient

	//3.设置解析处理器
	tm.process.SetByteOrder(tcpCfg.LittleEndian)

	//4.设置网络事件处理
	tm.GetEventProcessor().RegEventReceiverFunc(event.Sys_Event_Tcp, tm.GetEventHandler(), tm.tcpEventHandler)

	//5.启动网络模块
	return tm.tcpServer.Start()
}

func (tm *TcpModule) tcpEventHandler(ev event.IEvent) {
	pack := ev.(*event.Event).Data.(TcpPack)
	switch pack.Type {
	case TPTConnected:
		tm.process.ConnectedRoute(pack.ClientId)
	case TPTDisConnected:
		tm.process.DisConnectedRoute(pack.ClientId)
	case TPTUnknownPack:
		tm.process.UnknownMsgRoute(pack.ClientId, pack.Data, pack.RecyclerReaderBytes)
	case TPTPack:
		tm.process.MsgRoute(pack.ClientId, pack.Data, pack.RecyclerReaderBytes)
	}
}

func (tm *TcpModule) SetProcessor(process processor.IRawProcessor) {
	tm.process = process
}

func (tm *TcpModule) NewClient(conn *network.TCPConn) network.Agent {
	tm.mapClientLocker.Lock()
	defer tm.mapClientLocker.Unlock()

	clientId := primitive.NewObjectID().Hex()
	pClient := &Client{tcpConn: conn, id: clientId}
	pClient.tcpModule = tm
	tm.mapClient[clientId] = pClient

	return pClient
}

func (slf *Client) GetId() string {
	return slf.id
}

func (slf *Client) Run() {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			errString := fmt.Sprint(r)
			log.Dump(string(buf[:l]), log.String("error", errString))
		}
	}()

	slf.tcpModule.NotifyEvent(&event.Event{Type: event.Sys_Event_Tcp, Data: TcpPack{ClientId: slf.id, Type: TPTConnected}})
	for {
		if slf.tcpConn == nil {
			break
		}

		slf.tcpConn.SetReadDeadline(slf.tcpModule.tcpServer.ReadDeadline)
		bytes, err := slf.tcpConn.ReadMsg()
		if err != nil {
			log.Debug("read client failed", log.ErrorAttr("error", err), log.String("clientId", slf.id))
			break
		}
		data, err := slf.tcpModule.process.Unmarshal(slf.id, bytes)
		if err != nil {
			slf.tcpModule.NotifyEvent(&event.Event{Type: event.Sys_Event_Tcp, Data: TcpPack{ClientId: slf.id, Type: TPTUnknownPack, Data: bytes, RecyclerReaderBytes: slf.tcpConn.GetRecyclerReaderBytes()}})
			continue
		}
		slf.tcpModule.NotifyEvent(&event.Event{Type: event.Sys_Event_Tcp, Data: TcpPack{ClientId: slf.id, Type: TPTPack, Data: data, RecyclerReaderBytes: slf.tcpConn.GetRecyclerReaderBytes()}})
	}
}

func (slf *Client) OnClose() {
	slf.tcpModule.NotifyEvent(&event.Event{Type: event.Sys_Event_Tcp, Data: TcpPack{ClientId: slf.id, Type: TPTDisConnected}})
	slf.tcpModule.mapClientLocker.Lock()
	defer slf.tcpModule.mapClientLocker.Unlock()
	delete(slf.tcpModule.mapClient, slf.GetId())
}

func (tm *TcpModule) SendMsg(clientId string, msg interface{}) error {
	tm.mapClientLocker.Lock()
	client, ok := tm.mapClient[clientId]
	if ok == false {
		tm.mapClientLocker.Unlock()
		return fmt.Errorf("client %d is disconnect!", clientId)
	}

	tm.mapClientLocker.Unlock()
	bytes, err := tm.process.Marshal(clientId, msg)
	if err != nil {
		return err
	}
	return client.tcpConn.WriteMsg(bytes)
}

func (tm *TcpModule) Close(clientId string) {
	tm.mapClientLocker.Lock()
	defer tm.mapClientLocker.Unlock()

	client, ok := tm.mapClient[clientId]
	if ok == false {
		return
	}

	if client.tcpConn != nil {
		client.tcpConn.Close()
	}

	log.SWarning("close client:", clientId)
	return
}

func (tm *TcpModule) GetClientIp(clientid string) string {
	tm.mapClientLocker.Lock()
	defer tm.mapClientLocker.Unlock()
	pClient, ok := tm.mapClient[clientid]
	if ok == false {
		return ""
	}

	return pClient.tcpConn.GetRemoteIp()
}

func (tm *TcpModule) SendRawMsg(clientId string, msg []byte) error {
	tm.mapClientLocker.Lock()
	client, ok := tm.mapClient[clientId]
	if ok == false {
		tm.mapClientLocker.Unlock()
		return fmt.Errorf("client %d is disconnect!", clientId)
	}
	tm.mapClientLocker.Unlock()
	return client.tcpConn.WriteMsg(msg)
}

func (tm *TcpModule) GetConnNum() int {
	tm.mapClientLocker.Lock()
	connNum := len(tm.mapClient)
	tm.mapClientLocker.Unlock()
	return connNum
}

func (tm *TcpModule) SetNetMempool(mempool bytespool.IBytesMempool) {
	tm.tcpServer.SetNetMempool(mempool)
}

func (tm *TcpModule) GetNetMempool() bytespool.IBytesMempool {
	return tm.tcpServer.GetNetMempool()
}

func (tm *TcpModule) ReleaseNetMem(byteBuff []byte) {
	tm.tcpServer.GetNetMempool().ReleaseBytes(byteBuff)
}

func (tm *TcpModule) GetProcessor() processor.IRawProcessor {
	return tm.process
}

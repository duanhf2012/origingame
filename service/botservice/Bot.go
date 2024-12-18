package botservice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/sysmodule/httpclientmodule"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"sync"
	"time"
)

type botStatus = int

const gateAddr = "127.0.0.1:9001"  // kcp,tcp
const wsAddr = "ws://0.0.0.0:9444" // ws
type NetType int

const (
	Tcp NetType = iota
	WS
	Kcp
)

var netType = Kcp

const (
	httpLogging       = 0
	gateLogging       = 1
	receivingUserInfo = 2
	finish            = 3
)

type IConn interface {
	ReadMsg() ([]byte, error)
	WriteMsg(args ...[]byte) error
}

type Bot struct {
	id     int
	token  string
	status botStatus

	tcpClient *network.TCPClient
	wsClient  *network.WSClient
	kcpClient *network.KCPClient

	tcpConn *network.NetConn
	wsConn  *network.WSConn
	kcpConn *network.NetConn

	conn      IConn
	connMutex sync.Mutex

	chanMsg        chan *MsgEvent
	pbRawProcessor processor.PBRawProcessor
}

var mapRegisterMsg map[msg.MsgType]*MsgEvent

type MsgEvent struct {
	conn   *network.NetConn
	msg    proto.Message
	funcDo func(bot *Bot, msg proto.Message)
}

func regMsg(msgType msg.MsgType, msg proto.Message, funcDo func(bot *Bot, msg proto.Message)) {
	mapRegisterMsg[msgType] = &MsgEvent{msg: msg, funcDo: funcDo}
}

func NewMsgByMsgType(msgType msg.MsgType) *MsgEvent {
	msg, ok := mapRegisterMsg[msgType]
	if ok == false {
		return nil
	}

	var msgEvent MsgEvent
	msgEvent.msg = proto.Clone(msg.msg)
	msgEvent.funcDo = msg.funcDo

	return &msgEvent
}

func (bt *Bot) runBot() bool {
	for {
		cxt, cancel := context.WithTimeout(context.Background(), time.Second*5)
		select {
		case <-cxt.Done():
			cancel()
			bt.doTimeOut()
		case receiveMsg := <-bt.chanMsg:
			if receiveMsg.funcDo != nil {
				receiveMsg.funcDo(bt, receiveMsg.msg)
			}
		}
	}

	return true
}

func (bt *Bot) doTimeOut() {
	switch bt.status {
	case httpLogging:
		bt.httpLogin()
	case gateLogging:
		bt.gateLogin()
	case receivingUserInfo:
		bt.waitReceivingUserInfo()
	case finish:
		bt.work()
	}
}

func (bt *Bot) setTcpConn(conn *network.NetConn) {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	bt.conn = conn
	bt.tcpConn = conn
}

func (bt *Bot) setWsConn(conn *network.WSConn) {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	bt.conn = conn
	bt.wsConn = conn
}

func (bt *Bot) setKcpConn(conn *network.NetConn) {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	bt.conn = conn
	bt.kcpConn = conn
}

func (bt *Bot) getTcpConn() *network.NetConn {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	return bt.tcpConn
}

func (bt *Bot) getWsConn() *network.WSConn {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	return bt.wsConn
}

func (bt *Bot) getConn() IConn {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()
	return bt.conn
}

func (bt *Bot) Init(id int) {
	bt.id = id
	bt.chanMsg = make(chan *MsgEvent, 100)
}

func (bt *Bot) httpLogin() {
	loginInfo := &rpc.LoginInfo{}
	loginInfo.PlatType = rpc.LoginType_Gust
	loginInfo.PlatId = fmt.Sprintf("bot_%d", bt.id)
	loginInfo.GameId = "test"
	loginInfo.UserName = "test"

	var httpClient httpclientmodule.HttpClientModule
	httpClient.Init("", 1, time.Second*5, time.Second*5, time.Second*5, time.Second*5)
	byteLogin, _ := json.Marshal(&loginInfo)
	response := httpClient.Request("POST", "http://127.0.0.1:9000/api/login", byteLogin, nil)
	log.Debug("response", log.String("body", string(response.Body)))

	var httpRespone struct {
		ECode    int
		Token    string
		AreaGate string
		AreaHis  map[string]int64
	}

	err := json.Unmarshal(response.Body, &httpRespone)
	if err != nil {
		log.Error("json unmarshal error", log.ErrorField("err", err))
		return
	}

	if httpRespone.ECode == 0 {
		bt.setGateLogging(httpRespone.Token)
	}
	//response.Body
}

func (bt *Bot) setGateLogging(token string) {
	bt.status = gateLogging
	bt.token = token
}

func (bt *Bot) setLoginReceive() {
	bt.status = receivingUserInfo
}

func (bt *Bot) setLoginFinish() {
	bt.status = finish
}

func (bt *Bot) initConnect() bool {
	if bt.tcpClient != nil {
		bt.tcpClient.Close(false)
	}

	if bt.wsClient != nil {
		bt.wsClient.Close()
	}

	switch netType {
	case Tcp:
		bt.tcpClient = &network.TCPClient{}
		bt.tcpClient.Addr = gateAddr
		bt.tcpClient.ConnectInterval = time.Second * 5
		bt.tcpClient.ConnNum = 1
		bt.tcpClient.AutoReconnect = false
		bt.tcpClient.ReadDeadline = time.Second * 600
		bt.tcpClient.WriteDeadline = time.Second * 600
		bt.tcpClient.NewAgent = func(conn *network.NetConn) network.Agent {
			agent := BotAgent{}
			agent.bt = bt

			bt.setTcpConn(conn)
			return &agent
		}

		bt.tcpClient.Start()
	case WS:
		bt.wsClient = &network.WSClient{}
		bt.wsClient.MessageType = websocket.BinaryMessage
		bt.wsClient.MaxMsgLen = 65535
		bt.wsClient.PendingWriteNum = 100
		bt.wsClient.AutoReconnect = false
		bt.wsClient.ConnNum = 1
		bt.wsClient.ConnectInterval = time.Second * 5
		bt.wsClient.Addr = wsAddr
		bt.wsClient.NewAgent = func(conn *network.WSConn) network.Agent {
			agent := BotAgent{}
			agent.bt = bt

			bt.setWsConn(conn)
			return &agent
		}
		bt.wsClient.Start()
	case Kcp:
		bt.kcpClient = &network.KCPClient{}
		bt.kcpClient.Addr = gateAddr
		bt.kcpClient.ConnectInterval = time.Second * 5
		bt.kcpClient.ConnNum = 1
		bt.kcpClient.AutoReconnect = false
		bt.kcpClient.ReadDeadline = time.Second * 600
		bt.kcpClient.WriteDeadline = time.Second * 600
		bt.kcpClient.NewAgent = func(conn *network.NetConn) network.Agent {
			agent := BotAgent{}
			agent.bt = bt

			bt.setKcpConn(conn)
			return &agent
		}

		bt.kcpClient.Start()
	}

	return true
}

func (bt *Bot) gateLogin() {
	if bt.initConnect() == false {
		time.Sleep(5 * time.Second)
		return
	}

	//发送消息
	time.Sleep(2 * time.Second)
	if bt.getConn() == nil {
		return
	}

	var msgLoginReq msg.MsgLoginReq
	msgLoginReq.UserId = fmt.Sprintf("bot_%d", bt.id)
	msgLoginReq.ChannePlat = fmt.Sprintf("bot_%d", bt.id)
	msgLoginReq.Token = bt.token
	msgLoginReq.ShowAreaId = 1

	bt.SendMsg(msg.MsgType_LoginReq, &msgLoginReq)

	bt.status = receivingUserInfo
}

func (bt *Bot) waitReceivingUserInfo() {
	log.Debug("wait...")
}

func (bt *Bot) work() {
	var ping msg.MsgPing
	bt.SendMsg(msg.MsgType_Ping, &ping)
}

func (bt *Bot) SendMsg(msgType msg.MsgType, msg proto.Message) error {
	var pbPackInfo processor.PBRawPackInfo
	byteMsg, _ := proto.Marshal(msg)
	pbPackInfo.SetPackInfo(uint16(msgType), byteMsg)
	bytes, err := bt.pbRawProcessor.Marshal("", &pbPackInfo)
	if err != nil {
		return err
	}

	conn := bt.getConn()
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}

	return conn.WriteMsg(bytes)
}

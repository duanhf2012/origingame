package botservice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/sysmodule/httpclientmodule"
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"sync"
	"time"
)

type botStatus = int

const gateAddr = "127.0.0.1:9001"
const (
	httpLogging       = 0
	gateLogging       = 1
	receivingUserInfo = 2
	finish            = 3
)

type Bot struct {
	id     int
	token  string
	status botStatus

	tcpClient *network.TCPClient

	conn      *network.TCPConn
	connMutex sync.Mutex

	chanMsg        chan *MsgEvent
	pbRawProcessor processor.PBRawProcessor
}

var mapRegisterMsg map[msg.MsgType]*MsgEvent

type MsgEvent struct {
	conn   *network.TCPConn
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

func (bt *Bot) setConn(conn *network.TCPConn) {
	bt.connMutex.Lock()
	defer bt.connMutex.Unlock()

	bt.conn = conn
}

func (bt *Bot) getConn() *network.TCPConn {
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
	log.Debug("response:%s", string(response.Body))

	var httpRespone struct {
		ECode    int
		Token    string
		AreaGate string
		AreaHis  map[string]int64
	}

	err := json.Unmarshal(response.Body, &httpRespone)
	if err != nil {
		log.Error("json unmarshal error:%s", err.Error())
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

	bt.tcpClient = &network.TCPClient{}
	bt.tcpClient.Addr = gateAddr
	bt.tcpClient.ConnectInterval = time.Second * 5
	bt.tcpClient.ConnNum = 1
	bt.tcpClient.AutoReconnect = false
	bt.tcpClient.ReadDeadline = time.Second * 600
	bt.tcpClient.WriteDeadline = time.Second * 600
	bt.tcpClient.NewAgent = func(conn *network.TCPConn) network.Agent {
		agent := BotAgent{}
		agent.bt = bt

		bt.setConn(conn)
		return &agent
	}

	bt.tcpClient.Start()
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

package botservice

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/sysmodule/httpclientmodule"
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
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

	tcpClient      *network.TCPClient
	conn           *network.TCPConn
	chanMsg        chan *MsgEvent
	pbRawProcessor processor.PBRawProcessor
}

var mapRegisterMsg map[uint16]*MsgEvent

func init() {
	//RegMsg()
}

type MsgEvent struct {
	msg    proto.Message
	funcDo func(msg proto.Message)
}

func RegMsg(msgType uint16, msg proto.Message, funcDo func(msg proto.Message)) {
	mapRegisterMsg[msgType] = &MsgEvent{msg: msg, funcDo: funcDo}
}

func NewMsgByMsgType(msgType uint16) *MsgEvent {
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
		if bt.status == httpLogging {
			bt.httpLogin()
			time.Sleep(5 * time.Second)
			continue
		}
		
		switch bt.status {
		case httpLogging:
			bt.httpLogin()
		case gateLogging:
			bt.gateLogin()
		case receivingUserInfo:
			bt.waitReceivingUserInfo()
		case finish:
			bt.work()
			return true
		}
	}

	return true
}

func (bt *Bot) SetId(id int) {
	bt.id = id
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

func (bt *Bot) initConnect() bool {
	if bt.tcpClient != nil {
		bt.tcpClient.Close(false)
	}

	bt.tcpClient = &network.TCPClient{}
	bt.tcpClient.Addr = gateAddr
	bt.tcpClient.ConnectInterval = time.Second * 5
	bt.tcpClient.ConnNum = 1
	bt.tcpClient.AutoReconnect = false
	bt.tcpClient.ReadDeadline = time.Second * 15
	bt.tcpClient.WriteDeadline = time.Second * 15
	bt.tcpClient.NewAgent = func(conn *network.TCPConn) network.Agent {
		agent := BotAgent{}
		agent.bt = bt
		agent.bt.conn = conn
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
}

func (bt *Bot) waitReceivingUserInfo() {

}

func (bt *Bot) work() {
}

func (bt *Bot) SendMsg(msgType msg.MsgType, msg proto.Message) error {
	var pbPackInfo processor.PBRawPackInfo
	bytes, err := bt.pbRawProcessor.Marshal("", pbPackInfo)
	if err != nil {
		return err
	}

	return bt.conn.Write(bytes)
}

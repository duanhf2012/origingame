package gateservice

import (
	"bytes"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"origingame/common/performance"
	"time"
)

type AnalyzerType = int

const (
	MsgAnalyzer        AnalyzerType = 0 //消息统计
	ConnectNumAnalyzer AnalyzerType = 1 //链接数统计
	MaxAnalyzer
)

// AnalyzerId定义
const (
	ClientConnectAnalyzerId int = 0
)

// Analyzer默认日志等级
const (
	AnalyzerLogLevel int = 1
)

// AnalyzerColumn定义
const (
	//MsgAnalyzer
	MsgReceiveNumColumn  int = 0 //消息接收数量
	MsgSendNumColumn     int = 1 //消息发送数量
	MsgReceiveByteColumn int = 2 //消息接收字节数
	MsgSendByteColumn    int = 3 //消息发送字节数

	//ConnectNumAnalyzer
	ClientConnectNumColumn int = 0 //链接数量
)

func InitPerformanceAnalyzer(analyzer *performance.PerformanceAnalyzer, analyzerInterval time.Duration, analyzerLogLevel int) {
	analyzer.InitAnalyzer(MaxAnalyzer+1, analyzerInterval, analyzerLogLevel)
	analyzer.CreateAnalyzer(MsgAnalyzer, performance.AnalyzerLogLevel1, FunMsgAsync)
	analyzer.CreateAnalyzer(ConnectNumAnalyzer, performance.AnalyzerLogLevel4, FunConnectNumAnalyzer)
}

func FunMsgAsync(performanceAnalyzer *performance.PerformanceAnalyzer, deltaTime int64, mapAnalyzer map[int]*performance.Analyzer) {
	if mapAnalyzer == nil {
		return
	}

	var buffer bytes.Buffer
	buffer.WriteString("\n[as]Message gate service statistical analysis:")
	var receiveNum, sendNum, receiveByte, sendByte int64
	isFind := false
	for id, an := range mapAnalyzer {
		receiveNum, sendNum, receiveByte, sendByte = an.FetchColumn4()
		if receiveNum == 0 && sendNum == 0 {
			continue
		}
		if receiveNum != 0 && sendNum != 0 {
			log.SError("Msg ", id, " can not receive and send")
			continue
		}

		isFind = true
		if receiveNum != 0 {
			buffer.WriteString(fmt.Sprintf("\n[as]Every %d Milliseconds msgtype[%d], recive count[%d], recevie byte len[%d], avg byte len[%d]",
				deltaTime/1000, id, receiveNum, receiveByte, receiveByte/receiveNum))
		} else {
			buffer.WriteString(fmt.Sprintf("\n[as]Every %d Milliseconds msgtype[%d], send count[%d], send byte len[%d], avg byte len[%d]",
				deltaTime/1000, id, sendNum, sendByte, sendByte/sendNum))
		}
	}

	if isFind {
		performanceAnalyzer.WriteLog(buffer.String())
	}

	return
}

func FunConnectNumAnalyzer(performanceAnalyzer *performance.PerformanceAnalyzer, deltaTime int64, mapAnalyzer map[int]*performance.Analyzer) {
	if mapAnalyzer == nil {
		return
	}

	analyzerInfo, ok := mapAnalyzer[ClientConnectAnalyzerId]
	if ok == false || analyzerInfo == nil {
		return
	}
	clientCount := analyzerInfo.GetColumn1()

	var buffer bytes.Buffer
	buffer.WriteString("\n[as]Gate service client connect statistical analysis:")
	buffer.WriteString(fmt.Sprintf("\n[as]gate service client connect num[%d]", clientCount))

	performanceAnalyzer.WriteLog(buffer.String())
}

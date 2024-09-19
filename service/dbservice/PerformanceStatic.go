package dbservice

import (
	"bytes"
	"fmt"
	"origingame/common/performance"
	"time"
)

const minServiceProcessNum = 10

func InitPerformanceAnalyzer(analyzer *performance.PerformanceAnalyzer, analyzerInterval time.Duration, analyzerLogLevel int) {
	analyzer.InitAnalyzer(performance.MaxAnalyzer+1, analyzerInterval, analyzerLogLevel)

	//1级，对象统计等，且没那么重要

	//2级，数量比较大的内容,如消息相关

	//3级耗时和数量统计相关

	//4.次最高优先级

	//5.最高优先级，一般是单条日志，且非常重要，必打印
	analyzer.CreateAnalyzer(performance.ServiceStateAnalyzer, performance.AnalyzerLogLevel5, FunServiceStateAnalyzer)
}

func FunServiceStateAnalyzer(performanceAnalyzer *performance.PerformanceAnalyzer, deltaTime int64, mapAnalyzer map[int]*performance.Analyzer) {
	if mapAnalyzer == nil {
		return
	}

	eventChannelNum := performanceAnalyzer.GetService().GetServiceEventChannelNum()
	timerChannelNum := performanceAnalyzer.GetService().GetServiceTimerChannelNum()
	if eventChannelNum+timerChannelNum < minServiceProcessNum {
		return
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("\n[as]Service Name[%s] static channelNum[%d],timerChannelNum[%d]", performanceAnalyzer.GetService().GetName(), eventChannelNum, timerChannelNum))
	performanceAnalyzer.WriteLog(buffer.String())
}

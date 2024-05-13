package performance

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/node"
	"time"

	"github.com/duanhf2012/origin/v2/console"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/timer"
)

//const MaxAnalyzerType = 1024

type PerformanceAnalyzer struct {
	service.Module

	analyzer         []map[int]*Analyzer
	analyzerReport   []ReportFunc
	analyzerInterval time.Duration
	analyzerLogLevel int
	analyzerTime     int64
	logger           log.ILogger
	//bOpen            bool
}

type ColumnDataExType = int

const (
	EXCount    ColumnDataExType = 0
	ExMinCount ColumnDataExType = 1
	ExMaxCount ColumnDataExType = 2
	ExMax
)

type Analyzer struct {
	Name         string
	ColumnData   [8]int64
	ColumnDataEx [8][ExMax + 1]int64
	//reportFunc   ReportFunc

	startTimeColumnId int
	startTime         int64
}

func (pa *PerformanceAnalyzer) OnInit() error {
	if pa.checkGlobalAnalyzerOpen() == false {
		return nil
	}

	pa.analyzerTime = time.Now().UnixNano()
	pa.NewTicker(pa.analyzerInterval, pa.analyzerTicker)
	var err error
	logPath := console.GetParamStringVal("logpath")
	if logPath == "" {
		logPath = "./log"
	}

	analyzerFileName := fmt.Sprintf("Node_%d_%s_Analyzer_", node.GetNodeId(), pa.GetService().GetName())
	pa.logger, err = log.NewTextLogger(log.LevelInfo, logPath, analyzerFileName, true, log.LogChannelCap)
	if err != nil {
		return err
	}

	/*
		mapCfg := pa.GetService().GetServiceCfg().(map[string]interface{})
		isOpen, okOpen := mapCfg["IsOpenStatic"]
		staticTime, okStatic := mapCfg["StaticTime"]
		if okOpen == false || okStatic == false {

		}
	*/
	//统计添加
	//mapCfg := ps.GetServiceCfg().(map[string]interface{})
	//isOpen, okOpen := mapCfg["IsOpenStatic"]
	//staticTime, okStatic := mapCfg["StaticTime"]
	//if okOpen == false || okStatic == false {
	//	return fmt.Errorf("Canot find IsOpenStatic/StaticTime from config.")
	//}
	//if isOpen.(bool) {
	//ps.staticDealModule = NewStaticMsgDeal(int64(staticTime.(float64)))
	//	ps.AddModule(ps.staticDealModule)
	//}
	return nil
}

func (pa *PerformanceAnalyzer) WriteLog(log string) {
	//pa.logger.SInfo(log)
}

func (pa *PerformanceAnalyzer) checkGlobalAnalyzerOpen() bool {
	if pa.analyzerInterval <= 0 {
		return false
	}

	return pa.analyzerInterval > 0
}

func (pa *PerformanceAnalyzer) InitAnalyzer(maxAnalyzer int, analyzerInterval time.Duration, analyzerLogLevel int) {
	pa.analyzerInterval = analyzerInterval
	pa.analyzerLogLevel = analyzerLogLevel

	if pa.checkGlobalAnalyzerOpen() == false {
		return
	}
	if pa.analyzerLogLevel > MaxAnalyzerLogLevel {
		pa.analyzerLogLevel = MaxAnalyzerLogLevel
	}

	pa.analyzer = make([]map[int]*Analyzer, maxAnalyzer)
	pa.analyzerReport = make([]ReportFunc, maxAnalyzer)

}

func (pa *PerformanceAnalyzer) analyzerTicker(ticker *timer.Ticker) {
	nowTm := time.Now().UnixNano()
	deltaTm := nowTm - pa.analyzerTime
	pa.analyzerTime = nowTm
	for idx, an := range pa.analyzerReport {
		if an != nil {
			an(pa, deltaTm, pa.analyzer[idx])
		}
	}
}

// 返回值为true时，则Reset，清空记录
type ReportFunc func(performanceAnalyzer *PerformanceAnalyzer, deltaTime int64, mapAnalyzer map[int]*Analyzer)

func (pa *PerformanceAnalyzer) CreateAnalyzer(analyzerType int, analyzerLogLevel int, reportFunc ReportFunc) bool {
	if pa.checkGlobalAnalyzerOpen() == false || analyzerLogLevel < pa.analyzerLogLevel {
		return false
	}

	pa.analyzerReport[analyzerType] = reportFunc
	pa.analyzer[analyzerType] = map[int]*Analyzer{}
	return true
}

func (pa *PerformanceAnalyzer) GetAnalyzer(analyzerType int, analyzerId int) *Analyzer {
	if pa.checkGlobalAnalyzerOpen() == false {
		return nil
	}

	mapAnalyzer := pa.analyzer[analyzerType]
	if mapAnalyzer == nil {
		return nil
	}

	analyzer := mapAnalyzer[analyzerId]
	if analyzer == nil {
		analyzer = &Analyzer{}
		mapAnalyzer[analyzerId] = analyzer
	}

	return analyzer
}

func (an *Analyzer) StartStatisticalTime() {
	an.startTime = time.Now().UnixNano()
}

func (an *Analyzer) Reset() {
	an.startTimeColumnId = 0
	an.startTime = 0
}

func (an *Analyzer) EndStatisticalTime(column int) {
	an.ColumnData[column] += time.Now().UnixNano() - an.startTime
	an.ColumnDataEx[column][EXCount] += 1
}

// return 总耗时ns,count
func (an *Analyzer) GetStatisticalTime(column int) (int64, int64) {
	return an.ColumnData[column], an.ColumnDataEx[column][EXCount]
}

func (an *Analyzer) FetchStatisticalTime(column int) (int64, int64) {
	costTm := an.ColumnData[column]
	count := an.ColumnDataEx[column][EXCount]
	an.ColumnData[column] = 0
	an.ColumnDataEx[column][EXCount] = 0

	return costTm, count
}

func (an *Analyzer) GetStatisticalTimeEx(column int) (int64, int64, int64, int64) {
	return an.ColumnData[column], an.ColumnDataEx[column][EXCount], an.ColumnDataEx[column][ExMinCount], an.ColumnDataEx[column][ExMaxCount]
}

func (an *Analyzer) FetchStatisticalTimeEx(column int) (int64, int64, int64, int64) {
	costTm := an.ColumnData[column]
	count := an.ColumnDataEx[column][EXCount]
	minCost := an.ColumnDataEx[column][ExMinCount]
	maxCost := an.ColumnDataEx[column][ExMaxCount]

	an.ColumnData[column] = 0
	an.ColumnDataEx[column][EXCount] = 0
	an.ColumnDataEx[column][ExMinCount] = 0
	an.ColumnDataEx[column][ExMaxCount] = 0

	return costTm, count, minCost, maxCost
}

func (an *Analyzer) EndStatisticalTimeEx(column int) {
	costTime := time.Now().UnixNano() - an.startTime
	if costTime == 0 {
		return
	}
	an.ColumnData[column] += costTime
	an.ColumnDataEx[column][EXCount] += 1
	if an.ColumnDataEx[column][ExMinCount] > costTime || an.ColumnDataEx[column][ExMinCount] == 0 {
		an.ColumnDataEx[column][ExMinCount] = costTime
	}

	if an.ColumnDataEx[column][ExMaxCount] < costTime || an.ColumnDataEx[column][ExMaxCount] == 0 {
		an.ColumnDataEx[column][ExMaxCount] = costTime
	}
}

func (an *Analyzer) GetColumn1() int64 {
	return an.ColumnData[0]
}

func (an *Analyzer) GetColumn2() (int64, int64) {
	return an.ColumnData[0], an.ColumnData[1]
}

func (an *Analyzer) GetColumn3() (int64, int64, int64) {
	return an.ColumnData[0], an.ColumnData[1], an.ColumnData[2]
}

func (an *Analyzer) GetColumn4() (int64, int64, int64, int64) {
	return an.ColumnData[0], an.ColumnData[1], an.ColumnData[2], an.ColumnData[3]
}

func (an *Analyzer) FetchColumn1() int64 {
	val0 := an.ColumnData[0]
	an.ColumnData[0] = 0
	return val0
}

func (an *Analyzer) FetchColumn2() (int64, int64) {
	val0 := an.ColumnData[0]
	val1 := an.ColumnData[1]
	an.ColumnData[0] = 0
	an.ColumnData[1] = 0

	return val0, val1
}

func (an *Analyzer) FetchColumn3() (int64, int64, int64) {
	val0 := an.ColumnData[0]
	val1 := an.ColumnData[1]
	val2 := an.ColumnData[2]
	an.ColumnData[0] = 0
	an.ColumnData[1] = 0
	an.ColumnData[2] = 0

	return val0, val1, val2
}

func (an *Analyzer) FetchColumn4() (int64, int64, int64, int64) {
	val0 := an.ColumnData[0]
	val1 := an.ColumnData[1]
	val2 := an.ColumnData[2]
	val3 := an.ColumnData[3]
	an.ColumnData[0] = 0
	an.ColumnData[1] = 0
	an.ColumnData[2] = 0
	an.ColumnData[3] = 0

	return val0, val1, val2, val3
}

func (pa *PerformanceAnalyzer) ChangeDeltaNum(analyzerType int, analyzerId int, columnIndex int, num int64) {
	if pa.checkGlobalAnalyzerOpen() == false {
		return
	}

	analyzer := pa.GetAnalyzer(analyzerType, analyzerId)
	if analyzer == nil {
		return
	}

	analyzer.ColumnData[columnIndex] += num
}

func (pa *PerformanceAnalyzer) Inc(analyzerType int, analyzerId int, columnIndex int) {
	pa.ChangeDeltaNum(analyzerType, analyzerId, columnIndex, 1)
}

func (pa *PerformanceAnalyzer) Dec(analyzerType int, analyzerId int, columnIndex int) {
	pa.ChangeDeltaNum(analyzerType, analyzerId, columnIndex, -1)
}

func (pa *PerformanceAnalyzer) Set(analyzerType int, analyzerId int, columnIndex int, value int64) {
	if pa.checkGlobalAnalyzerOpen() == false {
		return
	}

	analyzer := pa.GetAnalyzer(analyzerType, analyzerId) //pa.analyzer[analyzerType][analyzerId]
	if analyzer == nil {
		return
	}

	analyzer.ColumnData[columnIndex] = value
}

func (pa *PerformanceAnalyzer) SetDataEx(analyzerType int, analyzerId int, columnIndex int, dataExIndex int, value int64) {
	if pa.checkGlobalAnalyzerOpen() == false {
		return
	}

	analyzer := pa.GetAnalyzer(analyzerType, analyzerId) //pa.analyzer[analyzerType][analyzerId]
	if analyzer == nil {
		return
	}

	analyzer.ColumnDataEx[columnIndex][dataExIndex] = value
}

package botservice

import (
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
)

const botNum = 1

type BotService struct {
	service.Service
}

func init() {
	node.Setup(&BotService{})
}

func (ts *BotService) OnInit() error {
	ts.OpenConcurrent(1, botNum, 10)
	return nil
}

func (ts *BotService) OnStart() {
	for i := 0; i < botNum; i++ {
		var b Bot
		b.Init(i)

		ts.AsyncDo(b.runBot, nil)
	}

}

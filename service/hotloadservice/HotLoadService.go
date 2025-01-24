package hotloadservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"path/filepath"
)

var hotLoadService HotLoadService

func init() {
	node.Setup(&hotLoadService)
}

type HotLoadService struct {
	service.Service
	tableCfgModule TableCfgModule
}

func (gs *HotLoadService) OnInit() error {
	gs.tableCfgModule.SetJsonPath(filepath.Join(node.GetConfigDir(), "../datas"))
	_, err := gs.AddModule(&gs.tableCfgModule)
	if err != nil {
		return err
	}

	return err
}

func (gs *HotLoadService) RpcReload(req *rpc.Empty, res *rpc.Empty) {
	log.Debug("start load table config...")
	err := gs.tableCfgModule.LoadCfg()
	if err != nil {
		log.Error("load table config failed", log.ErrorField("err", err))
		return
	}
	log.Debug("finish load table config...")
}

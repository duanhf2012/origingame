package wsmodule

import "github.com/duanhf2012/origin/v2/service"

type WSModule struct {
	service.Module
}

func (ws *WSModule) OnInit() error {
	return nil
}

package interfacedef

import "github.com/duanhf2012/origin/v2/rpc"

type IGSService interface {
	rpc.IRpcHandler
	DestroyPlayer(playerId string) bool
}

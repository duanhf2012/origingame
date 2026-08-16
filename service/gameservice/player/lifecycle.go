package player

import (
	"errors"
	"time"
)

func (player *Player) initialize() error {
	initialized := 0
	for _, proxy := range player.proxies {
		if err := proxy.OnInit(player); err != nil {
			for index := initialized - 1; index >= 0; index-- {
				player.proxies[index].OnRelease()
			}
			player.released = true
			return err
		}
		initialized++
	}
	return nil
}

// FinishLoad 在全部 BSON 已应用后执行两道正序 Proxy 加载屏障。
func (player *Player) FinishLoad(isNewPlayer bool) error {
	if player == nil || player.released {
		return errors.New("Player 已释放")
	}
	ctx := PlayerLoadContext{IsNewPlayer: isNewPlayer}
	for _, proxy := range player.proxies {
		if err := proxy.OnLoaded(ctx); err != nil {
			player.Release()
			return err
		}
	}
	for _, proxy := range player.proxies {
		if err := proxy.OnAllLoaded(ctx); err != nil {
			player.Release()
			return err
		}
	}
	return nil
}

// Online 绑定连接并按注册顺序通知全部 Proxy。
func (player *Player) Online(gatewayNodeID string, connectionID string, now time.Time) error {
	if player == nil || player.released || gatewayNodeID == "" || connectionID == "" {
		return errors.New("Player 上线参数无效")
	}
	player.dataInfo.GatewayNodeID = gatewayNodeID
	player.dataInfo.GatewayConnectionID = connectionID
	player.dataInfo.State = StateOnline
	player.dataInfo.LastHeartbeatAt = now.UTC()
	player.dataInfo.ResidentDeadline = time.Time{}
	ctx := PlayerOnlineContext{GatewayNodeID: gatewayNodeID, GatewayConnectionID: connectionID}
	for _, proxy := range player.proxies {
		proxy.OnOnline(ctx)
	}
	return nil
}

// Offline 按注册倒序通知全部 Proxy，并清除当前连接关系。
func (player *Player) Offline(now time.Time, residentDuration time.Duration) {
	if player == nil || player.released || player.dataInfo.State != StateOnline {
		return
	}
	connectionID := player.dataInfo.GatewayConnectionID
	ctx := PlayerOfflineContext{GatewayConnectionID: connectionID}
	for index := len(player.proxies) - 1; index >= 0; index-- {
		player.proxies[index].OnOffline(ctx)
	}
	player.dataInfo.GatewayNodeID = ""
	player.dataInfo.GatewayConnectionID = ""
	player.dataInfo.State = StateResident
	player.dataInfo.ResidentDeadline = now.UTC().Add(residentDuration)
}

// Release 按注册倒序释放一次；回调不得继续修改持久化数据。
func (player *Player) Release() {
	if player == nil || player.released {
		return
	}
	player.released = true
	player.dataInfo.State = StateReleased
	for index := len(player.proxies) - 1; index >= 0; index-- {
		player.proxies[index].OnRelease()
	}
}

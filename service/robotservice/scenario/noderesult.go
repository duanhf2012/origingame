package scenario

import (
	"errors"
	"time"

	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/sysmodule/blueprintmodule"
)

const (
	nodeErrorInvalidState int64 = -1
	nodeErrorIO           int64 = -2
	nodeErrorTimeout      int64 = -3
	nodeErrorProtocol     int64 = -4
	nodeErrorOverloaded   int64 = -5
)

func durationMilliseconds(started time.Time) blueprintmodule.PortInt {
	duration := time.Since(started).Milliseconds()
	if duration < 0 {
		duration = 0
	}
	return blueprintmodule.PortInt(duration)
}

func (module *Module) resumeTo(handle *blueprintmodule.YieldHandle, output int, values ...any) {
	if handle == nil {
		return
	}
	if err := handle.ResumeTo(output, values...); err != nil &&
		!errors.Is(err, blueprintmodule.ErrExecutionCanceled) &&
		!errors.Is(err, blueprintmodule.ErrGraphReleased) {
		module.Logger().Warn("恢复机器人蓝图失败", log.Err(err))
	}
}

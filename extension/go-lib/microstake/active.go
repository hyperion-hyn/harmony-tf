package microstake

import (
	"fmt"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"time"
)

func WaitActive() {
	logger.Log(fmt.Sprintf("sleep %d second for map3Node active", config.Configuration.Network.WaitMap3ActiveTime), true)
	time.Sleep(time.Duration(config.Configuration.Network.WaitMap3ActiveTime) * time.Second)
}

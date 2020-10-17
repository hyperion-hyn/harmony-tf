package microstake

import (
	"fmt"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/utils"
	"github.com/hyperion-hyn/hyperion-tf/logger"
)

func WaitActive() {
	//logger.Log(fmt.Sprintf("sleep %d second for map3Node active", config.Configuration.Network.WaitMap3ActiveTime), true)
	//time.Sleep(time.Duration(config.Configuration.Network.WaitMap3ActiveTime) * time.Second)
	logger.Log(fmt.Sprintf("sleep 1 epoch for map3Node active"), true)
	rpc, _ := config.Configuration.Network.API.RPCClient()
	err := utils.WaitForEpoch(rpc, 1)
	if err != nil {
		fmt.Printf("Wait for skip epoch error")
	}
}

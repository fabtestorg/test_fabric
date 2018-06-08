package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/op/go-logging"
	"github.com/spf13/viper"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/fabtestorg/test_fabric/eventserver/handle"
	le "github.com/fabtestorg/test_fabric/eventserver/listenevent"
	"github.com/fabtestorg/test_fabric/apiserver/sdk"
)

const CHANNELBUFFER = 1000

var (
	logOutput  = os.Stderr
	configPath = flag.String("configPath", "./", "config path")
	configName = flag.String("configName", "client_sdk", "config file name")
)

func main() {
	flag.Parse()
	err := sdk.InitSDK(*configPath, *configName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//runtime.GOMAXPROCS(viper.GetInt("peer.gomaxprocs"))

	//setup system-wide logging backend based on settings from core.yaml
	flogging.InitBackend(flogging.SetFormat(viper.GetString("logging.format")), logOutput)
	logging.SetLevel(logging.DEBUG, "client_sdk")

	logger := logging.MustGetLogger("main")
	chainID := viper.GetString("other.channelId")
	mqAddresses := viper.GetStringSlice("other.mq_address")
	//queueName := viper.GetString("other.queue_name")
	//sysetmQueueName := viper.GetString("other.system_queue_name")
	userAlias := viper.GetString("user.alias")

	if len(mqAddresses) == 0 {
		logger.Panic("The mq_address is empty!")
	}

	handle.SetUserAliasWithFile(userAlias)

	if !handle.GetBlockHeight(chainID) {
		logger.Panic("Get block height failed!")
	}

	listenToHandle := make(chan handle.BlockInfoAll, CHANNELBUFFER)

	//check and recover the message
	//go handle.CheckAndRecoverEvent(chainID, handle.FilterEvent, listenToHandle)

	//listen the block event and parse the message
	err = le.ListenEvent("", chainID, handle.FilterEvent, listenToHandle)
	if err != nil{
		logger.Panic(err)
	}
	fmt.Println("end")
}

package handle

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"strings"

	"github.com/op/go-logging"
	"github.com/peersafe/factoring/apiserver/define"
	"github.com/peersafe/factoring/apiserver/utils"
	//"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"github.com/peersafe/gohfc"
	"time"
)

const prefixUserAlias = "user_alias:"

var (
	logger    = logging.MustGetLogger("filter-event")
	userAlias string
)

func SetUserAliasWithFile(alias string) {
	if "" == alias {
		logger.Warning("alias from file is empty!")
	} else {
		userAlias = alias
		logger.Infof("alias: %s from file is set to userAlias", userAlias)
	}

	return
}

func SetUserAlias(configPath, configFile, alias string) bool {
	var command, echoCommand string

	if "" == alias {
		logger.Error("alias is nil, cat set it to the file!")
		return false
	}

	// The command is just for ubuntu and centos, other operating system may use other command to do it.
	// TODO: adapt to other operating systems.
	// echoCommand just to adapt to docker environment
	if strings.HasSuffix(configPath, "/") {
		command = fmt.Sprintf("$(sed 's/^    alias.*/    alias: %s/g' %s%s.yaml)", alias, configPath, configFile)
		echoCommand = fmt.Sprintf("echo %q > %s%s.yaml", command, configPath, configFile)
	} else {
		command = fmt.Sprintf("$(sed 's/^    alias.*/    alias: %s/g' %s/%s.yaml)", alias, configPath, configFile)
		echoCommand = fmt.Sprintf("echo %q > %s/%s.yaml", command, configPath, configFile)
	}

	logger.Info("the echoCommand is:", echoCommand)

	cmd := exec.Command("/bin/bash", "-c", echoCommand)
	err := cmd.Run()
	if nil != err {
		logger.Error("set alias:", alias, "to the file failed:", err.Error())
		return false
	}

	logger.Info("set alias:", alias, "to the file successful")
	return true
}

func GetUserAlias(mqAddrs []string, mqQueue, configPath, configFile string) {
	if 0 == len(mqAddrs) || "" == mqQueue || "" == configPath || "" == configFile {
		logger.Error("mqAddr or mqQueue or configPath or configFile is nil!")
		return
	}
	logger.Info("mqAddr is:", mqAddrs)
	logger.Info("mqQueue is:", mqQueue)
	logger.Info("configPaht is:", configPath)
	logger.Info("configFile is:", configFile)

	var tryNum = 0
	var addrsNum = len(mqAddrs)
	logger.Infof("The systems has %d addresses.", addrsNum)

	for {
		tryNum++
		logger.Critical("Get userAlias try times:", tryNum)
		mqAddr := mqAddrs[(tryNum-1)%addrsNum]

		conn, err := amqp.Dial(mqAddr)
		if err != nil {
			logger.Error("Failed to connect to RabbitMQ:", err.Error())
			continue
		}
		defer conn.Close()

		channel, err := conn.Channel()
		if err != nil {
			logger.Error("Failed to open a channel:", err.Error())
			continue
		}
		err = channel.Qos(1, 0, false)
		if err != nil {
			logger.Error("Failed to set the channle's qos:", err.Error())
			continue
		}
		queue, err := channel.QueueDeclare(
			mqQueue, // name
			true,    // durable
			false,   // delete when unused
			false,   // exclusive
			false,   // no-wait
			nil,     // arguments
		)
		if err != nil {
			logger.Error("Failed to declare a queue:", err.Error())
			continue
		}

		msgs, err := channel.Consume(
			queue.Name, // queue
			"",         // consumer
			true,       // auto-ack
			false,      // exclusive
			false,      // no-local
			false,      // no-wait
			nil,        // args
		)
		if err != nil {
			logger.Error("Failed to consume a queue:", err.Error())
			continue
		}

		for msg := range msgs {
			logger.Debugf("Received a message: %s", msg.Body)
			aliasFromRabbitmq := string(msg.Body)
			if strings.HasPrefix(aliasFromRabbitmq, prefixUserAlias) {
				aliasFromRabbitmq = aliasFromRabbitmq[len(prefixUserAlias):len(aliasFromRabbitmq)]
			} else {
				logger.Error("The message form rabbitmq doesn't meet the pre-specified format:", aliasFromRabbitmq)
				continue
			}
			userAlias = aliasFromRabbitmq
			logger.Info("Get the alias:", userAlias, "from rabbitmq")
			if setOk := SetUserAlias(configPath, configFile, userAlias); !setOk {
				logger.Error("Get the alias:", userAlias, "from rabbitmq but set it to the config file failed.")
				continue
			}
			logger.Info("Get the alias", userAlias, "from rabbitmq and set it to the config file.")
		}
	}

	logger.Error("----------GetUserAlias exit----------")
}

func FilterEvent(event *gohfc.EventBlockResponseTransactionEvent) (interface{}, bool) {
	responseData := define.InvokeResponse{}
	iReqs := &[]string{}
	responseData.Payload = iReqs
	payload := []define.Factor{}
	err := json.Unmarshal(event.Value, &responseData)
	if err != nil {
		logger.Error(err)
		return nil, false
	} else {
		message := define.Message{}
		err := json.Unmarshal([]byte((*iReqs)[0]), &message)
		if err != nil {
			logger.Error(err)
			return nil, false
		} else {
			//fmt.Printf("the during c -- %d\n",message.CreateTime)
			//fmt.Printf("the during g-- %d\n",getCurTime())
			fmt.Println(getCurTime() - message.CreateTime/1000)
			err = utils.FormatResponseMessage(userAlias, &payload, &[]define.Message{message})
			if err != nil {
				if !strings.Contains(err.Error(), "receiver") {
					//The node is the messages' receiver, but retrieve data failed
					logger.Error(err)
				}
				return nil, false
			}
		}
	}

	eventResponse := Event{
		Header: Header{
			ResponseStatus: responseData.ResStatus,
			ContentDef: ContentDef{
				ContentType: "application/json",
				TrackId:     responseData.TrackId,
				Language:    "zh-CN"},
			Ack: Ack{
				Level:    "notRequired",
				Callback: ""}},
		Contents: Contents{
			Schema: "/schema/factorList.json",
			Command: Command{
				Uri:    "",
				Action: "Call",
				Desc:   ""},
			Payload: payload}}

	//fmt.Println("eventResponse==============", eventResponse)
	b, _ := json.Marshal(eventResponse)

	return b, true
}

//获取当前时间
func getCurTime() uint64 {
	return uint64(time.Now().UTC().Unix())
}
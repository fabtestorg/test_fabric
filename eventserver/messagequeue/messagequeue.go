package messagequeue

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/streadway/amqp"
)

type mqInfo struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

var (
	mqs    []*mqInfo
	logger = logging.MustGetLogger("filter-event")
)


func InitMQ(queueName string, addresses ...string) error {
	for _, addr := range addresses {
		var mq mqInfo
		var err error
		mq.conn, err = amqp.Dial(addr)
		if err != nil {
			logger.Errorf("Failed to connect to RabbitMQ: %s", err.Error())
			continue
		}
		// defer conn.Close()
		mq.channel, err = mq.conn.Channel()
		if err != nil {
			logger.Errorf("Failed to open a channel: %s", err.Error())
			continue
		}
		// defer ch.Close()
		mq.queue, err = mq.channel.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			logger.Errorf("Failed to declare a queue: %s", err.Error())
			continue
		}
		mqs = append(mqs, &mq)
	}
	if len(mqs) == 0 {
		return fmt.Errorf("There's no suitable mq!")
	}
	return nil
}

func Close() {
	for _, mq := range mqs {
		mq.conn.Close()
		mq.channel.Close()
	}
}

func SendMessage(msg interface{}) error {
	var data []byte
	if ret, ok := msg.([]byte); ok {
		data = ret
	} else if ret, ok := msg.(string); ok {
		data = []byte(ret)
	} else {
		return fmt.Errorf("Unexpect msg type !")
	}
	if len(data) == 0 {
		return fmt.Errorf("The message is empty!")
	}

	for _, mq := range mqs {
		if err := mq.channel.Publish(
			"",            // exchange
			mq.queue.Name, // routing key
			false,         // mandatory
			false,         // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        data,
			}); err == nil {
			return nil
		} else {
			logger.Errorf("mq publish failed: %s", err.Error())
		}
	}
	return fmt.Errorf("Send message to mq failed: the mq send error!")
}

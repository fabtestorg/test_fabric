package orderer

import (
	"fmt"
	"sync"

	"github.com/hyperledger/fabric/protos/common"
)

type OrdererClient struct {
	bc BroadcastClient
	sync.Mutex
}

var (
	broadcastClients []*OrdererClient
)

func InitBroadcastClient() error {
	orderers, err := newClients()
	if err != nil {
		return err
	}

	for _, order := range orderers {
		bc, err := GetBroadcastClient(&order)
		if err != nil {
			fmt.Println("InitBroadcastClient err")
			return err
		}
		client := &OrdererClient{
			bc: bc,
		}
		broadcastClients = append(broadcastClients, client)
	}

	return nil
}

func GetOrdererClients() []*OrdererClient {
	return broadcastClients
}

func (oc *OrdererClient) BroadcastClientSend(env *common.Envelope) error {
	oc.Lock()
	defer oc.Unlock()
	err := oc.bc.Send(env)
	if err != nil {
		fmt.Println("[(oc *OrdererClient)] BroadcastClientSend  err: ", err)
	}

	return err
}

func (oc *OrdererClient) Close() {
	oc.bc.Close()
}

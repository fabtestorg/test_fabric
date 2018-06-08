package orderer

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/utils"
	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/config"
)

func newClients() ([]utils.ClientInfo, error) {
	var clients []utils.ClientInfo

	orderers, err := config.GetOrderersConfig()
	if err != nil {
		return clients, fmt.Errorf("Error get orderer info from config file: %v", err)
	}

	for _, orderer := range orderers {
		c, err := utils.NewClient(orderer.Address, orderer.TLS.Certificate, orderer.TLS.ServerHostOverride, config.IsTLSEnabled())
		if err != nil {
			fmt.Printf("Error new  orderer client: %v\n", err)
			continue
		}
		clients = append(clients, *c)
	}
	return clients, nil
}

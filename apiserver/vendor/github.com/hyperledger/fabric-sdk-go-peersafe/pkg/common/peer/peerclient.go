package peer

import (
	"fmt"

	// "context"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/utils"
	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/config"
	protos_peer "github.com/hyperledger/fabric/protos/peer"
)

type EndorserClient interface {
	ProcessProposal(ctx context.Context, in *protos_peer.SignedProposal, opts ...grpc.CallOption) (*protos_peer.ProposalResponse, error)
	// Close()
}

type endorserClient struct {
	protos_peer.EndorserClient
}

func newEndorserClient(conn *grpc.ClientConn) EndorserClient {
	return &endorserClient{EndorserClient: protos_peer.NewEndorserClient(conn)}
}

func newClients() ([]utils.ClientInfo, error) {
	var clients []utils.ClientInfo

	peers, err := config.GetPeersConfig()
	if err != nil {
		return clients, fmt.Errorf("Error get peer info from config file: %v", err)
	}

	for _, peer := range peers {
		c, err := utils.NewClient(peer.Address, peer.TLS.Certificate, peer.TLS.ServerHostOverride, config.IsTLSEnabled())
		if err != nil {
			fmt.Printf("Error new peer client: %v\n", err)
			continue
		}
		c.LocalMspId = peer.LocalMspId
		c.Primary = peer.Primary

		clients = append(clients, *c)
	}
	return clients, nil
}

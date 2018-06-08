package peer

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/utils"
	"github.com/hyperledger/fabric/protos/peer"
	"golang.org/x/net/context"

	"github.com/haiheipijuan/grpc-pool"
)

type PeerClient struct {
	pool    *grpcpool.Pool
	MspId   string
	Primary bool
}

var peerClients []*PeerClient

func InitPeerClient() error {
	peers, err := newClients()
	if err != nil {
		return err
	}

	for _, c := range peers {
		p, err := utils.NewPool(c)
		if err != nil {
			return err
		}

		pool := &PeerClient{
			pool:    p,
			MspId:   c.LocalMspId,
			Primary: c.Primary,
		}
		peerClients = append(peerClients, pool)
	}
	return nil
}

func GetPeerClients(mspID string, onlyPrimary bool) (clients []*PeerClient) {
	for _, client := range peerClients {
		if (onlyPrimary == false || (onlyPrimary && client.Primary)) && (mspID == "" || client.MspId == mspID) {
			clients = append(clients, client)
		}
	}
	return
}

func (pc *PeerClient) ProcessProposal(signedProp *peer.SignedProposal) (*peer.ProposalResponse, error) {
	c, err := pc.pool.Get(context.Background())
	if err != nil {
		fmt.Println("[(pc *PeerClient) ProcessProposal] 11111 pc.pool.Get err: ", err)
		return nil, err
	}

	proposalResp, err := newEndorserClient(c.ClientConn).ProcessProposal(context.Background(), signedProp)
	if err != nil {
		fmt.Println("[(pc *PeerClient) ProcessProposal] 22222 ProcessProposal err: ", err)
		c.Unhealhty()
	}

	c.Close()

	return proposalResp, err
}

func (pc *PeerClient) Close() {
	pc.pool.Close()
}

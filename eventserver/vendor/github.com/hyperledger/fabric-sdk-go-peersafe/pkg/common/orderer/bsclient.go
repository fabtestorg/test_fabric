package orderer

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/utils"
	cb "github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"strings"
)

type BroadcastClient interface {
	//Send data to orderer
	Send(env *cb.Envelope) error
	Close() error
}

type broadcastClient struct {
	conn   *grpc.ClientConn
	client ab.AtomicBroadcast_BroadcastClient
}

// GetBroadcastClient creates a simple instance of the BroadcastClient interface
func GetBroadcastClient(o *utils.ClientInfo) (BroadcastClient, error) {
	if len(strings.Split(o.Url, ":")) != 2 {
		return nil, fmt.Errorf("Ordering service endpoint %s is not valid or missing", o.Url)
	}

	conn, err := grpc.Dial(o.Url, o.GrpcDialOption...)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to %s due to %s", o.Url, err)
	}
	client, err := ab.NewAtomicBroadcastClient(conn).Broadcast(context.TODO())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Error connecting to %s due to %s", o.Url, err)
	}

	return &broadcastClient{conn: conn, client: client}, nil
}

func (s *broadcastClient) getAck() error {
	msg, err := s.client.Recv()
	if err != nil {
		return err
	}
	if msg.Status != cb.Status_SUCCESS {
		return fmt.Errorf("Got unexpected status: %v", msg.Status)
	}
	return nil
}

//Send data to orderer
func (s *broadcastClient) Send(env *cb.Envelope) error {
	if err := s.client.Send(env); err != nil {
		return fmt.Errorf("Could not send :%s)", err)
	}

	err := s.getAck()

	return err
}

func (s *broadcastClient) Close() error {
	s.client.CloseSend()
	return s.conn.Close()
}

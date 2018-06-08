package utils

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/comm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ClientInfo struct {
	Url            string
	GrpcDialOption []grpc.DialOption
	// for peer
	Primary    bool
	LocalMspId string
}

func NewClient(url string, caFile string, serverHostOverride string, tlsEnabled bool) (*ClientInfo, error) {
	var opts []grpc.DialOption
	if tlsEnabled {
		creds, err := credentials.NewClientTLSFromFile(caFile, serverHostOverride)
		if err != nil {
			return nil, fmt.Errorf("Error connecting to %s due to %s", url, err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithTimeout(time.Second*3))
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(comm.MaxRecvMsgSize()),
		grpc.MaxCallSendMsgSize(comm.MaxSendMsgSize())))
	return &ClientInfo{Url: url, GrpcDialOption: opts}, nil
}

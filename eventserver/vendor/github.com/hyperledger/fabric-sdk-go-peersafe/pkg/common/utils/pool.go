package utils

import (
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/haiheipijuan/grpc-pool"
)

func NewPool(o ClientInfo) (*grpcpool.Pool, error) {
	factory := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial(o.Url, o.GrpcDialOption...)
		return conn, err
	}

	p, err := grpcpool.New(factory, 10, 100, time.Second*15)
	if err != nil {
		return nil, fmt.Errorf("Error new pool due to %s", err)
	}

	return p, nil
}

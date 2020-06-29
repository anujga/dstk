package partrpc

import "google.golang.org/grpc"

type PartitionClient interface {
	RpcClient() interface{}
	Close() error
}

type partClient struct {
	grpcClient interface{}
	// required only for close
	conn *grpc.ClientConn
	// mostly used for debugging
	url string
}

func (pc *partClient) Close() error {
	return pc.conn.Close()
}

func (pc *partClient) RpcClient() interface{} {
	return pc.grpcClient
}

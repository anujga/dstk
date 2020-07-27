package io

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func GrpcServer() *grpc.Server {
	grpc_prometheus.EnableHandlingTimeHistogram()
	//https://github.com/grpc-ecosystem/go-grpc-middleware
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			//grpc_ctxtags.StreamServerInterceptor(),
			//grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			//grpc_ctxtags.UnaryServerInterceptor(),
			//grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		)),
	)

	reflection.Register(s)
	return s
}

func DefaultClientOpts() []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	}
	return opts
}

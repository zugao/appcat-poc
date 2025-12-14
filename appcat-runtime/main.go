package main

import (
	"flag"
	"fmt"
	"net"

	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	addr := flag.String("addr", ":9443", "gRPC listen address")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(fmt.Errorf("listen: %w", err))
	}

	// Create gRPC server
	s := grpc.NewServer()

	// Create and register manager
	mgr := NewManager(zap.New())
	fnv1.RegisterFunctionRunnerServiceServer(s, mgr)

	// Enable reflection for debugging
	reflection.Register(s)

	fmt.Printf("Starting gRPC server on %s\n", *addr)
	if err := s.Serve(lis); err != nil {
		panic(fmt.Errorf("serve: %w", err))
	}
}
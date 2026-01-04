package main

import (
	"flag"
	"fmt"
	"os"

	function "github.com/crossplane/function-sdk-go"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	addr := flag.String("addr", ":9443", "gRPC listen address")
	tlsDirFlag := flag.String("tls-dir", "", "Directory containing tls.crt, tls.key, ca.crt (defaults to TLS_SERVER_CERTS_DIR)")
	flag.Parse()

	tlsDir := *tlsDirFlag
	if tlsDir == "" {
		tlsDir = os.Getenv("TLS_SERVER_CERTS_DIR")
	}
	if tlsDir == "" {
		panic("TLS server cert directory not set; set --tls-dir or TLS_SERVER_CERTS_DIR")
	}

	// Health server allows Crossplane to probe readiness
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("function-appcat-poc", healthpb.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Create and register manager
	mgr := NewManager(zap.New())

	opts := []function.ServeOption{
		function.Listen("tcp", *addr),
		function.MTLSCertificates(tlsDir),
		function.WithHealthServer(healthSrv),
	}

	fmt.Printf("Starting gRPC server on %s (mTLS dir: %s)\n", *addr, tlsDir)
	if err := function.Serve(mgr, opts...); err != nil {
		panic(fmt.Errorf("serve: %w", err))
	}
}
